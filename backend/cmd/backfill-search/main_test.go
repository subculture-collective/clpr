package main

import (
	"errors"
	"testing"
)

func TestRunPaged_ContinuesPastPagesWhereEveryRowFailed(t *testing.T) {
	pageDelay = 0
	// Page 2 fetches rows but indexes none; paging must continue to page 3.
	pages := [][2]int{{2, 2}, {2, 0}, {1, 1}, {0, 0}}
	var offsets []int
	call := 0
	indexed, failed, err := runPaged(2, func(offset int) (int, int, error) {
		offsets = append(offsets, offset)
		p := pages[call]
		call++
		return p[0], p[1], nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if indexed != 3 || failed != 2 {
		t.Fatalf("indexed=%d failed=%d, want 3 and 2", indexed, failed)
	}
	want := []int{0, 2, 4, 6}
	if len(offsets) != len(want) {
		t.Fatalf("offsets = %v, want %v", offsets, want)
	}
	for i := range want {
		if offsets[i] != want[i] {
			t.Fatalf("offsets = %v, want %v", offsets, want)
		}
	}
}

func TestRunPaged_PropagatesPageError(t *testing.T) {
	pageDelay = 0
	boom := errors.New("rows error")
	_, _, err := runPaged(10, func(int) (int, int, error) { return 0, 0, boom })
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
