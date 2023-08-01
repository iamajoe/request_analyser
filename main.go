package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"os"
	"strings"
)

func help() {
	// TODO: should probably setup an help
	log.Fatal(errors.New("help! i need somebody"))
}

// stats checks the general statistics
func stats(srcRaw string) error {
	filePath, err := sourceToFilePath(srcRaw)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	count := 0
	methodCounts := make(map[string]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		v := strings.ToLower(scanner.Text())
		if len(v) == 0 || strings.Index(v, "#") == 0 {
			continue
		}

		count += 1

		// count for request method
		for _, k := range strings.Split(v, ";") {
			if strings.Index(k, "requestmethod") == -1 {
				continue
			}

			arr := strings.Split(k, ":")
			if len(arr) != 2 {
				continue
			}

			c, ok := methodCounts[arr[1]]
			if !ok {
				methodCounts[arr[1]] = 0
			}

			methodCounts[arr[1]] = c + 1
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.Println("count:", count)
	for k, v := range methodCounts {
		log.Println("method", k, v)
	}

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
