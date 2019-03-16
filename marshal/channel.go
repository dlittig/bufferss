package marshal

type Channel struct {
	Title string `xml:"title,omitempty"`
	Link  string `xml:"link,omitempty"`
	Desc  string `xml:"description,omitempty"`
	Items []Item `xml:"item,omitempty"`
}
