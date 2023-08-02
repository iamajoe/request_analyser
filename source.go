package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

func rawToSource(raw string) ([]source, error) {
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

	return data, nil
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

// convertRawSource takes a raw string and saves it to a file to be used
// the source array we use on the tool
func convertRawSource(raw string, filePath string) error {
	// time to append to file
	f, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return err
	}

	if _, err := f.WriteString(raw); err != nil {
		return err
	}

	return nil
}

// convertJsonSource takes a valid json string and converts it to a file
// that the software know how to use
func convertJsonSource(raw []byte, filePath string) error {
	// normalize the single format
	if isJsonSingleSource(string(raw)) {
		raw = []byte("[" + string(raw) + "]")
	}

	// handle the array format
	var data []source
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	// time to convert and send to a file
	newRaw := ""
	for _, req := range data {
		headers, err := json.Marshal(req.RequestHeaders)
		if err != nil {
			return err
		}

		body, err := json.Marshal(req.RequestBody)
		if err != nil {
			return err
		}

		newRaw += fmt.Sprintf(
			"requestUrl:%s;requestMethod:%s;requestHeaders:%s;requestBody:%s\n",
			req.RequestUrl,
			req.RequestMethod,
			headers,
			body,
		)
	}

	return convertRawSource(newRaw, filePath)
}

// parseRedisSource takes a redis url and fetches all keys with a pattern
// and then converts to a source array we use on the tool
func convertRedisSource(srcRaw string, filePath string) error {
	arr := strings.Split(srcRaw, ";")
	pattern := "*"
	if len(arr) == 2 {
		pattern = arr[1]
	}

	opts, err := redis.ParseURL(arr[0])
	if err != nil {
		return err
	}

	rdb := redis.NewClient(opts)
	return redisScanAll(rdb, pattern, func(raw string) error {
		return convertRawSource(raw, filePath)
	})
}

// sourceToFilePath takes a string, makes sure it is on the raw format we are
// expecting on the tool and saves to a file path, returns that file path
func sourceToFilePath(srcRaw string, outputPath string) error {
	if len(srcRaw) == 0 {
		return errors.New("source is required")
	}

	if len(outputPath) == 0 {
		return errors.New("output path is required")
	}

	// remove spaces so we can easily check for indexes
	noSpacesRaw := removeSpaces(srcRaw)

	if isJsonSource(noSpacesRaw) || isJsonSingleSource(noSpacesRaw) {
		err := convertJsonSource([]byte(srcRaw), outputPath)
		return err
	}

	if isRedisSource(noSpacesRaw) {
		err := convertRedisSource(srcRaw, outputPath)
		return err
	}

	if isRawSource(noSpacesRaw) {
		err := convertRawSource(srcRaw, outputPath)
		return err
	}

	return nil
}
