package slate

import (
	"fmt"
	"html"
	"io"
	"strings"
)

func textNodeToHTML(node Node, builder io.StringWriter) {
	_, _ = builder.WriteString(html.EscapeString(node.Text))
}

func containerNodeToHTML(node Node, styles map[string]string, builder *strings.Builder) {
	tag := "div"
	style := ""
	if node.Type == "title" {
		tag = "h3"
		style = styles[node.Type]
	} else if node.Type == "paragraph" {
		tag = "p"
		style = styles[node.Type]
	} else if node.Type == "unordered-list" {
		tag = "ul"
		style = styles[node.Type]
	} else if node.Type == "list-item" {
		tag = "li"
		style = styles[node.Type]
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
			textNodeToHTML(childNode, builder)
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
