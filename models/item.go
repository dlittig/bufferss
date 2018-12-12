package models

type Item struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	Desc           string `xml:"description"`
	ContentEncoded string `xml:"content:encoded"`
	GUID           string `xml:"guid"`
	PubDate        string `xml:"pubDate"`
	Comments       string `xml:"comments"`
}
