package marshal

type Item struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	Desc     string `xml:"description"`
	Content  string `xml:"content:encoded"`
	GUID     string `xml:"guid"`
	PubDate  string `xml:"pubDate"`
	Comments string `xml:"comments"`
}
