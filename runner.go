package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type queueJob struct {
	data source
}

type queue struct {
	runningCount int
	list         []source
	informer     io.Writer

	baseUrl     string
	timerMs     int
	concurrency int
	mu          sync.Mutex
}

func getMemPercent() float64 {
	memory, _ := mem.VirtualMemory()
	memPerc := memory.UsedPercent
	return memPerc
}

func getCpuAvgPercent() float64 {
	percents, _ := cpu.Percent(time.Second, false)

	sum := float64(0)
	for _, v := range percents {
		sum += v
	}

	avg := sum / float64(len(percents))

	return avg
}

func newQueue(baseUrl string, timerMs int, concurrency int, informer io.Writer) *queue {
	// make sure the lest character is a "/" so it is easy to join
	parsedBaseUrl := baseUrl
	if parsedBaseUrl[len(parsedBaseUrl)-1:] != "/" {
		parsedBaseUrl += "/"
	}

	return &queue{
		runningCount: 0,
		list:         []source{},
		informer:     informer,

		baseUrl:     parsedBaseUrl,
		timerMs:     timerMs,
		concurrency: concurrency,
	}
}

func (q *queue) doRequest(job source) error {
	var req *http.Request
	var err error

	method := strings.ToUpper(job.RequestMethod)

	// set the request
	if job.RequestBody == nil {
		req, err = http.NewRequest(method, job.RequestUrl, nil)
	} else {
		// prepare the body
		body, err := json.Marshal(job.RequestBody)
		if err != nil {
			return err
		}

		req, err = http.NewRequest(method, job.RequestUrl, bytes.NewReader(body))
	}

	if err != nil {
		return err
	}

	// set the headers
	for k, v := range job.RequestHeaders {
		if v == nil || reflect.TypeOf(v).String() != "string" {
			continue
		}

		req.Header.Set(k, v.(string))
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (q *queue) jobHandler(job source) (time.Duration, float64, float64, error) {
	start := time.Now()

	startCpuPercent := float64(0)
	startMemPercent := float64(0)
	endCpuPercent := float64(0)
	endMemPercent := float64(0)

	isLocal := strings.Contains(job.RequestUrl, "localhost")

	// we need access to the actual cpu and memory on the server for this one
	if isLocal {
		// no need to error this
		startCpuPercent = getCpuAvgPercent()
		startMemPercent = getMemPercent()
	}

	// actually do the request at hand
	err := q.doRequest(job)
	if err != nil {
		return time.Since(
			start,
		), endCpuPercent - startCpuPercent, endMemPercent - startMemPercent, err
	}

	// we need access to the actual cpu and memory on the server for this one
	if isLocal {
		// no need to error this
		endCpuPercent = getCpuAvgPercent()
		endMemPercent = getMemPercent()
	}

	return time.Since(start), endCpuPercent - startCpuPercent, endMemPercent - startMemPercent, nil
}

func (q *queue) nextJob() {
	// we dont have any more space to keep running
	if q.concurrency >= q.runningCount || len(q.list) == 0 {
		return
	}

	// pop out the first one in queue
	next := q.list[0]
	q.list = q.list[1:]

	go func(job source) {
		elapsed, cpuUsed, memUsed, err := q.jobHandler(job)

		if q.informer != nil {
			// setup a message to inform
			msg := fmt.Sprintf(
				"request_method:%s;;request_url:%s",
				job.RequestMethod,
				job.RequestUrl,
			)
			if err != nil {
				msg = fmt.Sprintf("%s;;err:%s", msg, err.Error())
			} else {
				msg = fmt.Sprintf("%s;;elapsed_time:%d;;cpu_usage:%.2f;;mem_usage:%.2f", elapsed, cpuUsed, memUsed)
			}

			_, _ = q.informer.Write([]byte(msg))
		}

		// make sure we remove the job from the running list
		q.mu.Lock()
		q.runningCount -= 1
		q.mu.Unlock()

		// before setting out the next request, wait
		if q.timerMs > 0 {
			time.Sleep(time.Duration(q.timerMs) * time.Millisecond)
		}

		q.nextJob()
	}(next)
}

func (q *queue) addToQueue(job source) {
	// construct protocol
	url := job.RequestUrl
	if strings.Contains(url, "https://") || strings.Contains(url, "http://") {
		job.RequestUrl = q.baseUrl + url
	}

	q.list = append(q.list, job)
	q.nextJob()
}

func (q *queue) getJobCount() int {
	return len(q.list) + q.runningCount
}

func isSourceFiltered(job source, ignorePatterns []string) bool {
	for _, p := range ignorePatterns {
		p = strings.ReplaceAll(p, " ", "")
		if len(p) == 0 {
			continue
		}

		method := "*"
		urlPattern := "*"

		// separate method and url pattern
		arr := strings.Split(p, ":")
		if len(arr) < 2 {
			urlPattern = arr[0]
		} else {
			method = strings.ToLower(arr[0])
			urlPattern = arr[1]
		}

		// check the method, if not the same, no point in going any further
		if method != "*" && strings.ToLower(job.RequestMethod) != method {
			continue
		}

		// if wildcard, already passed through the method, ignore
		if urlPattern == "*" {
			return true
		}

		// now check the url
		r, err := regexp.Compile(urlPattern)
		if err != nil {
			continue
		}

		i := r.FindIndex([]byte(job.RequestUrl))
		if len(i) > 0 {
			return true
		}
	}

	return false
}

func run(
	inputPath string,
	baseUrl string,
	concurrency int,
	timerMs int,
	ignorePatterns []string,
	informer io.Writer,
) error {
	if len(inputPath) == 0 {
		return errors.New("input path is required")
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	q := newQueue(baseUrl, timerMs, concurrency, informer)

	// run a scanner line by line on the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		v := strings.ToLower(scanner.Text())
		if len(v) == 0 || strings.Index(v, "#") == 0 {
			continue
		}

		sources, err := rawToSource(v)
		if err != nil {
			return err
		}

		for _, s := range sources {
			if isSourceFiltered(s, ignorePatterns) {
				continue
			}

			q.addToQueue(s)
		}

		// we might be loading a lot into memory, lets slow down a bit
		// so that we can remove some of the old sources
		if q.getJobCount() > 5000 {
			time.Sleep(time.Second * 2)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
