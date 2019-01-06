package main

import (
	"bufferss/models"
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var url string
var port int64
var duration int
var help bool
var exPath string

func main() {
	parseFlags()

	log.Print("Starting bufferss ...")

	initialize()
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Print("Successfully started.")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func parseFlags() {
	flag.StringVar(&url, "url", "", "<url> referes to the remote resource")
	flag.Int64Var(&port, "port", 8080, "<port> sets the port the service should listen to")
	flag.IntVar(&duration, "duration", 2, "<duration> sets threshold the application should cache the entries")
	flag.BoolVar(&help, "", false, "Prints this help screen")
	flag.Parse()

	if help == true {
		flag.PrintDefaults()
	}
}

func initialize() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)

	fetch()
	setTimer()
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

func readFile(name string) (string, error) {
	filepath := []string{exPath, "/", name}
	content, err := ioutil.ReadFile(strings.Join(filepath, ""))
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("%s", content)
	return result, nil
}

func writeFile(name string, content string) error {
	filepath := []string{exPath, "/", name}
	err := ioutil.WriteFile(strings.Join(filepath, ""), []byte(content), 0644)

	return err
}

func openFile(name string) (*os.File, error) {
	filepath := []string{exPath, "/", name}

	if _, err := os.Stat(strings.Join(filepath, "")); os.IsNotExist(err) {
		writeFile(name, "")
	}

	file, err := os.Open(strings.Join(filepath, "")) // For read access.
	return file, err
}

func fetch() error {
	// Get latest rss version
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Decode rss feed to models
	decoder := xml.NewDecoder(resp.Body)
	rss := models.Rss{}

	err = decoder.Decode(&rss)
	if err != nil {
		return err
	}

	resp.Body.Close()

	// Write new feed
	file, err := openFile("bufferss.feed")
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if file is new
	info, _ := file.Stat()
	if info.Size() > 4 {
		// Initialize vars needed for processing feed
		fileDecoder := xml.NewDecoder(file)
		root := models.Rss{}
		thresh := time.Now().AddDate(0, 0, -duration)
		dateLayout := "2006-01-02T15:04:05.000Z"
		lastSync, err := time.Parse(dateLayout, getLastSync())

		// Read existing feed
		fileDecoder.Decode(&root)
		if err != nil {
			return err
		}

		newItems := make([]models.Item, 0)

		// Check all existing items
		for _, item := range root.Channel.Items {
			// Prepare date

			creationDate, err := time.Parse(dateLayout, item.PubDate)

			log.Print(err)

			// Add if older than last sync and still younger than threshold
			if creationDate.Before(lastSync) && creationDate.After(thresh) {
				newItems = append(newItems, item)
			}
		}

		// Append new items to feed
		rss.Channel.Items = append(rss.Channel.Items, newItems...)
		root = models.Rss{}
		newItems = nil
	}

	encoder := xml.NewEncoder(file)
	encoder.Encode(&rss)

	file.Close()
	rss = models.Rss{}
	setLastSync()

	return nil
}

func setLastSync() {
	writeFile(".sync", time.Now().String())
}

func getLastSync() string {
	timestamp, err := readFile(".sync")
	if err != nil {
		return time.Now().AddDate(0, 0, -1).String()
	} else {
		return timestamp
	}
}
