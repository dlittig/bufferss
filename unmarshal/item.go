package unmarshal

type Item struct {
	Title    string `xml:"title,omitempty"`
	Link     string `xml:"link,omitempty"`
	Desc     string `xml:"description,omitempty"`
	Content  Content
	GUID     string `xml:"guid,omitempty"`
	PubDate  string `xml:"pubDate,omitempty"`
	Comments string `xml:"comments,omitempty"`
}
