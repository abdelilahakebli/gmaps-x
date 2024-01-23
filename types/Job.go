package types

type Job struct {
	Index     int    `json:"index" csv:"index"`
	Title     string `json:"title" csv:"title"`
	Link      string `json:"url" csv:"url"`
	Sponsored bool   `json:"sponsored" csv:"sponsored"`
	Place     *Place `json:"place" csv:"place"`
	Done      bool   `json:"done" csv:"done"`
}
