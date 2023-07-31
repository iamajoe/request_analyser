package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

func help() {
	// TODO: should probably setup an help
	log.Fatal(errors.New("help! i need somebody"))
}

// stats checks the general statistics
func stats(srcRaw string) error {
	sourceArr, err := parseSource(srcRaw)
	if err != nil {
		return err
	}

	statObj := struct {
		count         int
		methodCount   map[string]int
		endpointCount map[string]int
	}{
		methodCount:   make(map[string]int),
		endpointCount: make(map[string]int),
		count:         0,
	}

	for _, src := range sourceArr {
		// count method
		method := strings.ToUpper(src.RequestMethod)
		oldCount := statObj.methodCount[method]
		statObj.methodCount[method] = oldCount + 1

		// count endpoint
		url := method + " " + strings.ToLower(src.RequestUrl)
		oldCount = statObj.endpointCount[url]
		statObj.endpointCount[url] = oldCount + 1

		statObj.count += 1
	}

	// now we can log the stat object
	tmpl := "\n\ncount: " + strconv.Itoa(statObj.count)

	if statObj.count > 0 {
		tmpl += "\n"
	}

	for k, v := range statObj.methodCount {
		tmpl += "\nmethod " + k + ": " + strconv.Itoa(v)
	}

	if statObj.count > 0 {
		tmpl += "\n"
	}

	for k, v := range statObj.endpointCount {
		tmpl += "\nendpoint " + k + ": " + strconv.Itoa(v)
	}

	log.Println(tmpl + "\n")

	return nil
}

func main() {
	statsFs := flag.NewFlagSet("stats", flag.ExitOnError)
	srcRaw := statsFs.String("s", "", "source of the records")

	switch os.Args[1] {
	case "stats":
		if err := statsFs.Parse(os.Args[2:]); err != nil {
			statsFs.PrintDefaults()
			log.Fatal(err)
		}

		if err := stats(*srcRaw); err != nil {
			log.Fatal(err)
		}
	default:
		help()
	}
}
