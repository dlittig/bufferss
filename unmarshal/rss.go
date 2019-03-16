package unmarshal

import "encoding/xml"

type Rss struct {
	XMLName          xml.Name `xml:"rss"`
	Version          string   `xml:"version,attr"`
	ContentNamespace string   `xml:"xmlns:content,attr"`
	Channel          Channel  `xml:"channel"`
}
