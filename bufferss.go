package main

import (
	"bufferss/marshal"
	"bufferss/unmarshal"
	"encoding/xml"
	"flag"
	"fmt"
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

/*
 *
 */
func main() {
	ok := parseFlags()
	if !ok {
		log.Print("Missing parameter <url>. Aborting...")
		os.Exit(1)
	}

	log.Print("Starting bufferss ...")

	initialize()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filepath := []string{exPath, "/bufferss.feed"}
		content, _ := ioutil.ReadFile(strings.Join(filepath, ""))

		w.Write([]byte(xml.Header[:len(xml.Header)-1] + string(content)))
	})

	log.Print("Successfully started.")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

/*
 *
 */
func parseFlags() bool {
	flag.StringVar(&url, "url", "", "<url> refers to the remote resource")
	flag.Int64Var(&port, "port", 8080, "<port> sets the port the service should listen to")
	flag.IntVar(&duration, "duration", 2, "<duration> sets threshold the application should cache the entries")
	flag.BoolVar(&help, "", false, "Prints this help screen")
	flag.Parse()

	if help == true {
		flag.PrintDefaults()
	}

	if url == "" && help != true {
		return false
	}

	return true
}

/*
 *
 */
func initialize() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)

	fetch()
	setTimer()
}

/*
 *
 */
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

/*
 *
 */
func writeFile(name string, content string) error {
	filepath := []string{exPath, "/", name}
	err := ioutil.WriteFile(strings.Join(filepath, ""), []byte(content), 0644)
	log.Print(err)

	return err
}

/*
 *
 */
func openFile(name string) (*os.File, error) {
	filepath := []string{exPath, "/", name}

	if _, err := os.Stat(strings.Join(filepath, "")); os.IsNotExist(err) {
		writeFile(name, "")
	}

	file, err := os.OpenFile(strings.Join(filepath, ""), os.O_RDWR, 0644) // For read access.
	return file, err
}

/*
 *
 */
func fetch() error {
	// Get latest rss version
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Decode rss feed to models
	decoder := xml.NewDecoder(resp.Body)
	unmarshalRss := unmarshal.Rss{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
	}

	err = decoder.Decode(&unmarshalRss)
	if err != nil {
		return err
	}
	//log.Print(unmarshalRss)

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
		root := unmarshal.Rss{}
		thresh := time.Now().AddDate(0, 0, -duration)

		// Read existing feed
		fileDecoder.Decode(&root)
		if err != nil {
			return err
		}

		newItems := make([]unmarshal.Item, 0)

		// Check all existing items
		for _, item := range root.Channel.Items {
			// Prepare date
			creationDate, _ := time.Parse(time.RFC1123Z, item.PubDate)

			// Get the date of the last fetched item
			lastItem := unmarshalRss.Channel.Items[len(unmarshalRss.Channel.Items)-1]
			lastPub, _ := time.Parse(time.RFC1123Z, lastItem.PubDate)

			// Add if older than last item and still younger than threshold
			if creationDate.Before(lastPub) && creationDate.After(thresh) {
				// Copy content from item to item_export
				newItems = append(newItems, item)
			}
		}

		// Append new items to feed
		unmarshalRss.Channel.Items = append(unmarshalRss.Channel.Items, newItems...)
		root = unmarshal.Rss{}
		newItems = nil
	}

	//Now import the rss into the marshalling rss type
	marshalRss := marshal.Rss{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
	}
	marshalRss.ImportFeed(unmarshalRss)

	// Reset the file before writing
	file.Truncate(0)
	file.Seek(0, 0)
	file.Sync()

	// Encode the rss feed
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	error := encoder.Encode(marshalRss)
	if error != nil {
		return error
	}

	file.Close()
	file.Sync()
	unmarshalRss = unmarshal.Rss{}
	marshalRss = marshal.Rss{}

	return nil
}
