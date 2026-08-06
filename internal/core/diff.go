package core

import "strings"

type DiffOp string

const (
	DiffEqual DiffOp = "equal"
	DiffAdd   DiffOp = "add"
	DiffDel   DiffOp = "del"
)

type DiffLine struct {
	Op   DiffOp `json:"op"`
	Text string `json:"text"`
}

// maxDiffCells caps the O(n*m) DP table (~4M ints ~= 32MB) so a diff
// between two huge inputs fails fast with a clear error instead of
// eating memory. Fine for anything up to a few thousand lines per side.
const maxDiffCells = 4_000_000

// DiffText computes a line-based diff between a and b via classic LCS
// backtracking.
func DiffText(a, b string) ([]DiffLine, error) {
	al := splitLines(a)
	bl := splitLines(b)
	n, m := len(al), len(bl)
	if n*m > maxDiffCells {
		return nil, NewInputError("input too large for diff (line count product exceeds limit)")
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if al[i] == bl[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var out []DiffLine
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case al[i] == bl[j]:
			out = append(out, DiffLine{DiffEqual, al[i]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			out = append(out, DiffLine{DiffDel, al[i]})
			i++
		default:
			out = append(out, DiffLine{DiffAdd, bl[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, DiffLine{DiffDel, al[i]})
	}
	for ; j < m; j++ {
		out = append(out, DiffLine{DiffAdd, bl[j]})
	}
	return out, nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

// FormatDiff renders diff lines as unified-style +/-/space prefixed text.
func FormatDiff(lines []DiffLine) string {
	var b strings.Builder
	for _, l := range lines {
		switch l.Op {
		case DiffAdd:
			b.WriteString("+ ")
		case DiffDel:
			b.WriteString("- ")
		default:
			b.WriteString("  ")
		}
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}
