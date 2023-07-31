package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

type source struct {
	RequestMethod  string            `json:"requestMethod"`
	RequestUrl     string            `json:"requestUrl"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	RequestBody    map[string]string `json:"requestBody"`
}

func sanitizeSources(raw []source) []source {
	newSources := []source{}

	for _, v := range raw {
		newSource := source{
			RequestMethod:  strings.ToUpper(v.RequestMethod),
			RequestUrl:     v.RequestUrl,
			RequestHeaders: v.RequestHeaders,
			RequestBody:    v.RequestBody,
		}

		if len(newSource.RequestMethod) == 0 {
			newSource.RequestMethod = "GET"
		}

		// no point in going further if we dont have a request url
		if len(newSource.RequestUrl) > 0 {
			newSources = append(newSources, newSource)
		}
	}

	return newSources
}

func removeSpaces(raw string) string {
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "\t", "")
	raw = strings.ReplaceAll(raw, "\n", "")

	return raw
}

func isRawSource(raw string) bool {
	return strings.Contains(strings.ToLower(raw), "requesturl") && strings.Contains(raw, ";") &&
		strings.Contains(raw, ":")
}

func isJsonSingleSource(raw string) bool {
	return strings.Index(raw, "{") == 0 && strings.LastIndex(raw, "}") == len(raw)-1
}

func isJsonSource(raw string) bool {
	return strings.Index(raw, "[") == 0 && strings.LastIndex(raw, "]") == len(raw)-1
}

func isRedisSource(raw string) bool {
	return strings.Index(raw, "redis://") == 0 || strings.Index(raw, "rediss://") == 0
}

// parseRawSource takes a raw string and converts it to
// the source array we use on the tool
func parseRawSource(raw string) ([]source, error) {
	data := []source{}

	rawArr := strings.Split(raw, "\n")
	for _, request := range rawArr {
		// it must be a comment
		if len(request) == 0 || strings.Index(request, "#") == 0 {
			continue
		}

		newSource := source{}

		// separate the properties and go one by one
		// DEV: we should probably do this differently but for now, i think it is ok
		//      to assume that the data coming in has a ";req" coming in
		//      we can't simply use ; because there could be a case (like user-agent)
		//      where the ; is required. we can embrace a different limiter
		// TODO: investigate on how to do this a little bit better
		properties := strings.Split(request, ";req")
		for _, property := range properties {
			propertyData := strings.SplitN(property, ":", 2)
			if len(propertyData) != 2 {
				continue
			}

			value := propertyData[1]

			// handle the raw per key, values are different, cache them on the source
			switch strings.ToLower(propertyData[0]) {
			case "requestmethod":
				newSource.RequestMethod = strings.ToUpper(value)
			case "requesturl":
				newSource.RequestUrl = value
			case "requestheaders":
				headers := make(map[string]string)
				if err := json.Unmarshal([]byte(value), &headers); err != nil {
					return data, err
				}

				newSource.RequestHeaders = headers
			case "requestbody":
				body := make(map[string]string)
				if err := json.Unmarshal([]byte(value), &body); err != nil {
					return data, err
				}

				newSource.RequestBody = body
			}
		}

		// no point in going further if we dont have a request url
		if len(newSource.RequestUrl) > 0 {
			data = append(data, newSource)
		}
	}

	return sanitizeSources(data), nil
}

// parseJsonSource takes a raw valid json string and converts it to
// the source array we use on the tool
func parseJsonSource(raw []byte) ([]source, error) {
	// normalize the single format
	if isJsonSingleSource(string(raw)) {
		raw = []byte("[" + string(raw) + "]")
	}

	// handle the array format
	var data []source
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, err
	}

	return sanitizeSources(data), nil
}

// parseRedisSource takes a redis url and fetches all keys with a pattern
// and then converts to a source array we use on the tool
func parseRedisSource(srcRaw string) ([]source, error) {
	var data []source

	arr := strings.Split(srcRaw, ";")
	pattern := "*"
	if len(arr) == 2 {
		pattern = arr[1]
	}

	opts, err := redis.ParseURL(arr[0])
	if err != nil {
		return data, err
	}

	rdb := redis.NewClient(opts)
	vals, err := redisScanAll(rdb, pattern, 0, 1000, make(map[string]bool))
	if err != nil {
		return data, err
	}

	raw := strings.Join(vals, "\n")
	os.WriteFile("foo", []byte(raw), 0644)
	return parseSource(raw)
}

// parseFileSource takes a source file and converts it to
// the source array we use on the tool
func parseFileSource(srcRaw string) ([]source, error) {
	var data []source
	if _, err := os.Stat(srcRaw); err != nil {
		return data, err
	}

	raw, err := os.ReadFile(srcRaw)
	if err != nil {
		return data, err
	}

	return parseSource(string(raw))
}

// parseSource takes a raw and tries to figure what kind of source is and
// converts to an array we use on the tool
func parseSource(srcRaw string) ([]source, error) {
	if len(srcRaw) == 0 {
		return []source{}, errors.New("source is required")
	}

	// remove spaces so we can easily check for indexes
	noSpacesRaw := removeSpaces(srcRaw)

	if isJsonSource(noSpacesRaw) || isJsonSingleSource(noSpacesRaw) {
		return parseJsonSource([]byte(srcRaw))
	}

	if isRedisSource(noSpacesRaw) {
		return parseRedisSource(srcRaw)
	}

	if isRawSource(noSpacesRaw) {
		return parseRawSource(srcRaw)
	}

	return parseFileSource(srcRaw)
}
