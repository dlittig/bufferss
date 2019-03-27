package main

import (
	"bufferss/marshal"
	"bufferss/unmarshal"
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var url string
var port int64
var duration int
var help bool
var exPath string

/*
 * Main function
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
		content, _ := ioutil.ReadFile(exPath + "/bufferss.feed")

		w.Write([]byte(xml.Header[:len(xml.Header)-1] + string(content)))
	})

	log.Print("Successfully started.")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

/*
 * Parse passed flags
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
 * Initializes the feed to serve
 */
func initialize() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)

	// Do the first tick
	tick()
	setTimer()
}

/*
 * Sets the timer to start a function every tick
 */
func setTimer() {
	ticker := time.NewTicker(time.Hour)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				tick()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	defer close(quit)
}

/*
 * Opens and creates file, if not existing
 */
func openFile(name string) (*os.File, error) {
	if _, err := os.Stat(exPath + "/" + name); os.IsNotExist(err) {
		err := ioutil.WriteFile(exPath+"/"+name, []byte(""), 0644)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(exPath+"/"+name, os.O_RDWR, 0644) // For read access.
	return file, err
}

/*
 * Procedure that is being triggered every tick of the timer
 */
func tick() error {
	// Get latest rss version
	unmarshalRss, err := fetchCurrentFeed()

	// Write new feed
	file, err := openFile("bufferss.feed")
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if file is new
	info, _ := file.Stat()
	if info.Size() > 4 {
		addItemsFromFile(unmarshalRss, file)
	}

	//Now import the rss into the marshalling rss type
	marshalFeed(unmarshalRss, file)

	//Rewrite html entities
	rewrite()

	return nil
}

/*
 * Fetches current feed from the web and decodes it to a struct
 */
func fetchCurrentFeed() (unmarshal.Rss, error) {
	// Get latest rss version
	resp, err := http.Get(url)

	if err != nil {
		return unmarshal.Rss{}, err
	}

	defer resp.Body.Close()

	log.Println("Reading from web feed.")

	// Decode rss feed to models
	decoder := xml.NewDecoder(resp.Body)
	unmarshalRss := unmarshal.Rss{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
	}

	err = decoder.Decode(&unmarshalRss)
	if err != nil {
		return unmarshal.Rss{}, err
	}

	resp.Body.Close()

	return unmarshalRss, nil
}

/*
 * Read current feed file and append it to the new web feed.
 */
func addItemsFromFile(feed unmarshal.Rss, file *os.File) {
	log.Println("Updating feed file with new feeds.")

	// Initialize vars needed for processing feed
	fileDecoder := xml.NewDecoder(file)
	rssFile := unmarshal.Rss{}
	thresh := time.Now().AddDate(0, 0, -duration)

	// Read existing feed
	fileDecoder.Decode(&rssFile)

	newItems := make([]unmarshal.Item, 0)

	// Check all existing items
	for _, item := range rssFile.Channel.Items {
		// Prepare date
		creationDate, _ := time.Parse(time.RFC1123Z, item.PubDate)

		// Get the date of the last fetched item
		lastItem := feed.Channel.Items[len(feed.Channel.Items)-1]
		lastPub, _ := time.Parse(time.RFC1123Z, lastItem.PubDate)

		// Add if older than last item and still younger than threshold
		if creationDate.Before(lastPub) && creationDate.After(thresh) {
			// Copy content from item to item_export
			newItems = append(newItems, item)
		}
	}

	// Append new items to feed
	feed.Channel.Items = append(feed.Channel.Items, newItems...)
	rssFile = unmarshal.Rss{}
	newItems = nil
}

/*
 * Marshal the feed to file
 */
func marshalFeed(feed unmarshal.Rss, file *os.File) error {
	marshalRss := marshal.Rss{
		Version:          "2.0",
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
	}
	marshalRss.ImportFeed(feed)

	// Reset the file before writing
	file.Truncate(0)
	file.Seek(0, 0)
	file.Sync()

	log.Println("Writing new feed.")

	// Encode the rss feed
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	error := encoder.Encode(marshalRss)
	if error != nil {
		return error
	}

	file.Close()
	file.Sync()

	feed = unmarshal.Rss{}
	marshalRss = marshal.Rss{}

	return nil
}

/*
 * Dirty way to fix the issue that xml package does not offer a way to save unescaped content
 */
func rewrite() error {
	// Open feed for rewrite
	file, err := openFile("bufferss.feed")
	if err != nil {
		return err
	}
	defer file.Close()

	// Read by line and save in array
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := html.UnescapeString(scanner.Text())
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Join to huge string

	// Set pointer to first element in file
	file.Truncate(0)
	file.Seek(0, 0)
	file.Sync()

	log.Println("Decoding html entities.")
	for _, line := range lines {
		if _, err := file.Write([]byte(line + "\n")); err != nil {
			log.Fatal(err)
		}
	}

	file.Close()
	file.Sync()

	return nil
}
