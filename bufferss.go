package main

import (
	"bufferss/models"
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
)

var url string
var port int64
var help bool

func main() {
	parseFlags()

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func parseFlags() {
	flag.StringVar(&url, "url", "", "<url> referes to the remote resource")
	flag.Int64Var(&port, "port", 8080, "<port> sets the port the service should listen to")
	flag.BoolVar(&help, "", false, "Prints this help screen")
	flag.Parse()

	if help == true {
		flag.PrintDefaults()
	}
}

func initialize() error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	rss := models.Rss{}

	err = decoder.Decode(&rss)
	if err != nil {
		return err
	}

	setTimer()

	return nil
}

func setTimer() {
	ticker := time.NewTicker(time.Hour)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				fetch()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	defer close(quit)
}

func fetch() error {
	// Check if file is present

	// Create new file if does not exist -> bufferss.feed

	// Read old file, and new file and create an entire new feed file

	// Compare upload dates to skip comparison for certain time

	return nil
}
