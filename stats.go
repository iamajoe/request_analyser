package main

import (
	"bufio"
	"errors"
	"os"
	"sort"
	"strings"
)

type reqStat struct {
	count  int
	method string
	url    string
}
type reqStatArr []reqStat

func (r reqStatArr) Len() int           { return len(r) }
func (r reqStatArr) Less(i, j int) bool { return r[i].count < r[j].count }
func (r reqStatArr) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

type statsData struct {
	count              int
	requests           []reqStat
	mostUsed           []reqStat
	requestMethodCount map[string]int
}

// stats fetches from the input the statistics
func stats(inputPath string, mostUsedRequestLimit int) (statsData, error) {
	data := statsData{
		count:              0,
		requests:           []reqStat{},
		mostUsed:           []reqStat{},
		requestMethodCount: make(map[string]int),
	}

	if len(inputPath) == 0 {
		return data, errors.New("output path is required")
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	reqStats := make(map[string]*reqStat)

	// for statistics, run a scanner line by line on the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		v := strings.ToLower(scanner.Text())
		if len(v) == 0 || strings.Index(v, "#") == 0 {
			continue
		}

		data.count += 1

		reqMethod := ""
		reqUrl := ""

		// count for request method
		for _, k := range strings.Split(v, ";") {
			arr := strings.Split(k, ":")
			if len(arr) != 2 {
				continue
			}

			// handle request method
			if strings.Index(k, "requestmethod") != -1 {
				reqMethod = arr[1]
				continue
			}

			// handle request url
			if strings.Index(k, "requesturl") != -1 {
				reqUrl = arr[1]
				continue
			}
		}

		// count method
		c, _ := data.requestMethodCount[reqMethod]
		data.requestMethodCount[reqMethod] = c + 1

		// count the request
		key := reqMethod + "_" + reqUrl
		req, ok := reqStats[key]
		if !ok {
			req = &reqStat{
				count:  0,
				method: reqMethod,
				url:    reqUrl,
			}
			reqStats[key] = req
		}
		req.count += 1
	}

	if err := scanner.Err(); err != nil {
		return data, err
	}

	// find the most used and count the rest
	for _, v := range reqStats {
		// cache the stats
		data.requests = append(data.requests, *v)

		// lets handle the most used now
		if len(data.mostUsed) < mostUsedRequestLimit {
			data.mostUsed = append(data.mostUsed, *v)
			continue
		}

		newMostUsed := append(data.mostUsed, *v)
		sort.Sort(reqStatArr(newMostUsed))
		// sort.Reverse(reqStatArr(newMostUsed))

		count := []int{}
		for _, k := range newMostUsed {
			count = append(count, k.count)
		}

		data.mostUsed = newMostUsed[1:]
	}

	return data, nil
}
