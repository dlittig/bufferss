package unmarshal

import "encoding/xml"

type Content struct {
	XMLName xml.Name `xml:"encoded"`
	Data    string   `xml:",innerxml"`
}
