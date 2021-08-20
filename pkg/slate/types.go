package slate

type Node struct {
	Children []Node `json:"children"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type Nodes []Node
