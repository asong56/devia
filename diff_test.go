package core

import "testing"

func TestDiffTextIdenticalInput(t *testing.T) {
	lines, err := DiffText("a\nb\nc", "a\nb\nc")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		if l.Op != DiffEqual {
			t.Errorf("identical input should produce only DiffEqual lines, got %+v", l)
		}
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestDiffTextOneLineChanged(t *testing.T) {
	// Hand-traced against the LCS backtracking algorithm: "b" is
	// replaced by "x" in the middle, so the diff should keep the equal
	// lines around it and mark the change as a delete-then-add pair.
	lines, err := DiffText("a\nb\nc", "a\nx\nc")
	if err != nil {
		t.Fatal(err)
	}

	want := []DiffLine{
		{DiffEqual, "a"},
		{DiffDel, "b"},
		{DiffAdd, "x"},
		{DiffEqual, "c"},
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d diff lines, want %d: %+v", len(lines), len(want), lines)
	}
	for i, l := range lines {
		if l != want[i] {
			t.Errorf("line[%d] = %+v, want %+v", i, l, want[i])
		}
	}
}

func TestDiffTextAllAdded(t *testing.T) {
	lines, err := DiffText("", "x\ny")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		if l.Op != DiffAdd {
			t.Errorf("diffing from empty text should produce only DiffAdd lines, got %+v", l)
		}
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 added lines, got %d", len(lines))
	}
}

func TestDiffTextAllDeleted(t *testing.T) {
	lines, err := DiffText("x\ny", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		if l.Op != DiffDel {
			t.Errorf("diffing to empty text should produce only DiffDel lines, got %+v", l)
		}
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 deleted lines, got %d", len(lines))
	}
}

func TestDiffTextTooLarge(t *testing.T) {
	// n*m must exceed maxDiffCells (4,000,000) to trigger the guard.
	// 2100 lines on each side comfortably clears that (~4.4M), with
	// enough margin that the exact line count doesn't need to be exact.
	const numLines = 2100
	big := make([]byte, 0, numLines*2)
	for i := 0; i < numLines; i++ {
		big = append(big, 'x', '\n')
	}
	_, err := DiffText(string(big), string(big))
	if err == nil {
		t.Fatal("expected an error for inputs whose line-count product exceeds the limit")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("oversized diff should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestFormatDiff(t *testing.T) {
	lines := []DiffLine{
		{DiffEqual, "same"},
		{DiffDel, "removed"},
		{DiffAdd, "added"},
	}
	got := FormatDiff(lines)
	want := "  same\n- removed\n+ added\n"
	if got != want {
		t.Errorf("FormatDiff = %q, want %q", got, want)
	}
}
