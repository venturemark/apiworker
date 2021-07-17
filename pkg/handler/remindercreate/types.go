package remindercreate

type templateVenture struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type templateTimeline struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type templateUpdate struct {
	ID           string             `json:"id,omitempty"`
	Title        string             `json:"title,omitempty"`
	Body         string             `json:"body,omitempty"`
	AuthorName   string             `json:"author_name,omitempty"`
	RelativeTime string             `json:"relative_time,omitempty"`
	Path         string             `json:"path,omitempty"`
	Venture      templateVenture    `json:"venture"`
	Timelines    []templateTimeline `json:"timelines,omitempty"`
}
