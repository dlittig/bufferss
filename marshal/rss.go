package marshal

import (
	unmarshal "bufferss/unmarshal"
	"encoding/xml"
)

type Rss struct {
	XMLName          xml.Name `xml:"rss"`
	Version          string   `xml:"version,attr"`
	ContentNamespace string   `xml:"xmlns:content,attr"`
	Channel          Channel  `xml:"channel"`
}

func (rss *Rss) ImportFeed(readRss unmarshal.Rss) {
	rss.Version = readRss.Version
	rss.ContentNamespace = readRss.ContentNamespace

	rss.Channel = Channel{
		readRss.Channel.Title,
		readRss.Channel.Link,
		readRss.Channel.Desc,
		make([]Item, 0),
	}

	for _, readItem := range readRss.Channel.Items {
		rss.Channel.Items = append(rss.Channel.Items, Item{
			Comments: readItem.Comments,
			Content:  readItem.Content.Data,
			Desc:     readItem.Desc,
			GUID:     readItem.GUID,
			Link:     readItem.Link,
			PubDate:  readItem.PubDate,
			Title:    readItem.Title,
		})
	}
}
