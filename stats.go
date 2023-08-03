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
		return data, errors.New("input path is required")
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
		sources, err := rawToSource(v)
		if err != nil {
			return data, err
		}

		if len(sources) == 0 {
			continue
		}

		// move per source on the raw line (in theory, only 1)
		for _, s := range sources {
			data.count += 1

			// count method
			c, _ := data.requestMethodCount[s.RequestMethod]
			data.requestMethodCount[s.RequestMethod] = c + 1

			// count the request
			key := s.RequestMethod + "_" + s.RequestUrl
			req, ok := reqStats[key]
			if !ok {
				req = &reqStat{
					count:  0,
					method: s.RequestMethod,
					url:    s.RequestUrl,
				}
				reqStats[key] = req
			}
			req.count += 1
		}
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
