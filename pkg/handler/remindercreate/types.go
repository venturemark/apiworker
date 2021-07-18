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
	IDNumeric    int64              `json:"-"`
	Title        string             `json:"title,omitempty"`
	Body         string             `json:"body,omitempty"`
	AuthorName   string             `json:"authorName,omitempty"`
	RelativeTime string             `json:"relativeTime,omitempty"`
	Path         string             `json:"path,omitempty"`
	Venture      templateVenture    `json:"venture"`
	Timelines    []templateTimeline `json:"timelines,omitempty"`
}
