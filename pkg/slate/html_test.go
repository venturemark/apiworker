package slate

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ToHTML(t *testing.T) {
	testCases := []struct {
		title              bool
		jsonInput          string
		expectedHtmlOutput string
		styles             map[string]string
	}{
		{
			jsonInput:          `{"type": "title","children":[{"text":"title\ntitle"}]}`,
			title:              true,
			expectedHtmlOutput: "<h3>title\ntitle</h3>",
		},
		{
			jsonInput:          `[{"type":"paragraph","children":[{"text":"<script>alert(123);</script>"}]},{"type":"paragraph","children":[{"text":"part2"}]}]`,
			title:              false,
			expectedHtmlOutput: "<p>&lt;script&gt;alert(123);&lt;/script&gt;</p><p>part2</p>",
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var html string

			if tc.title {
				var node Node
				if err := json.Unmarshal([]byte(tc.jsonInput), &node); err != nil {
					t.Fatal(err)
				}

				html = node.ToHTML(tc.styles)
			} else {
				var nodes Nodes
				if err := json.Unmarshal([]byte(tc.jsonInput), &nodes); err != nil {
					t.Fatal(err)
				}

				html = nodes.ToHTML(tc.styles)
			}

			if !cmp.Equal(tc.expectedHtmlOutput, html) {
				t.Fatal(cmp.Diff(tc.expectedHtmlOutput, html))
			}
		})
	}
}
