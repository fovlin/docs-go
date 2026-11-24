package markdown

import (
	"strings"
	"testing"
)

func TestSplitTableRowEmptyEdgeCells(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"leading empty", "|| a | b |", []string{"", "a", "b"}},
		{"trailing empty", "| a | b ||", []string{"a", "b", ""}},
		{"both empty", "|| a | b ||", []string{"", "a", "b", ""}},
		{"spaced edges", "| | a | b ||", []string{"", "a", "b", ""}},
		{"escaped trailing pipe", `| a \|`, []string{"a |"}},
		{"escaped trailing pipe with close", `| a \| |`, []string{"a |"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitTableRow(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("cells = %q, want %q", got, tc.want)
			}
			for index := range tc.want {
				if got[index] != tc.want[index] {
					t.Fatalf("cells = %q, want %q", got, tc.want)
				}
			}
		})
	}
}

func TestTableEmptyEdgeCells(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "leading empty header and body",
			input: "|| a | b |\n| --- | --- | --- |\n|| 1 | 2 |\n",
			want:  "<table><thead><tr><th></th><th>a</th><th>b</th></tr></thead><tbody><tr><td></td><td>1</td><td>2</td></tr></tbody></table>",
		},
		{
			name:  "trailing empty header and body",
			input: "| a | b ||\n| --- | --- | --- |\n| 1 | 2 ||\n",
			want:  "<table><thead><tr><th>a</th><th>b</th><th></th></tr></thead><tbody><tr><td>1</td><td>2</td><td></td></tr></tbody></table>",
		},
		{
			name:  "both edges empty",
			input: "|| a | b ||\n| --- | --- | --- | --- |\n|| 1 | 2 ||\n",
			want:  "<table><thead><tr><th></th><th>a</th><th>b</th><th></th></tr></thead><tbody><tr><td></td><td>1</td><td>2</td><td></td></tr></tbody></table>",
		},
		{
			name:  "spaced edge cells",
			input: "| | a | b |\n| --- | --- | --- |\n| 1 | 2 | 3 |\n",
			want:  "<table><thead><tr><th></th><th>a</th><th>b</th></tr></thead><tbody><tr><td>1</td><td>2</td><td>3</td></tr></tbody></table>",
		},
		{
			name:  "escaped trailing pipe",
			input: "| a \\|\n| --- |\n| 1 |\n",
			want:  "<table><thead><tr><th>a |</th></tr></thead><tbody><tr><td>1</td></tr></tbody></table>",
		},
		{
			name:  "regular table",
			input: "| a | b |\n| --- | --- |\n| 1 | 2 |\n",
			want:  "<table><thead><tr><th>a</th><th>b</th></tr></thead><tbody><tr><td>1</td><td>2</td></tr></tbody></table>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.ReplaceAll(string(Parse([]byte(tc.input)).Render(nil)), "\n", "")
			if got != tc.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}
