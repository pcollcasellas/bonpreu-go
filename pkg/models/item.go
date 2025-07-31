package models

// ItemIds represents a product ID
type ItemIds struct {
	ProductID int `json:"product_id"`
}

// Sitemap represents the XML structure of the sitemap
type Sitemap struct {
	XMLName string `xml:"urlset"`
	URLs    []URL  `xml:"url"`
}

// URL represents a URL entry in the sitemap
type URL struct {
	Loc string `xml:"loc"`
}
