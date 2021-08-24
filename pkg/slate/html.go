package slate

import (
	"fmt"
	"html"
	"strings"
)

func textNodeToHTML(node Node, styles map[string]string, builder *strings.Builder) {
	builder.WriteString(html.EscapeString(node.Text))
}

func containerNodeToHTML(node Node, styles map[string]string, builder *strings.Builder) {
	tag := "div"
	style := ""
	if node.Type == "title" {
		tag = "h3"
		style = styles["title"]
	} else if node.Type == "paragraph" {
		tag = "p"
		style = styles["paragraph"]
	} else if node.Type == "" {
		tag = ""
	}

	if tag != "" {
		if style != "" {
			builder.WriteString(fmt.Sprintf("<%s style=\"%s\">", tag, style))
		} else {
			builder.WriteString(fmt.Sprintf("<%s>", tag))
		}
	}

	for _, childNode := range node.Children {
		if childNode.Text != "" {
			textNodeToHTML(childNode, styles, builder)
		} else {
			containerNodeToHTML(childNode, styles, builder)
		}
	}

	if tag != "" {
		builder.WriteString(fmt.Sprintf("</%s>", tag))
	}
}

func (n Node) ToHTML(styles map[string]string) string {
	var builder strings.Builder
	containerNodeToHTML(n, styles, &builder)
	return builder.String()
}

func (n Nodes) ToHTML(styles map[string]string) string {
	var builder strings.Builder
	containerNodeToHTML(Node{
		Children: n,
	}, styles, &builder)
	return builder.String()
}
