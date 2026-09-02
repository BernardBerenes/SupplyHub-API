package presenter

import "testing"

func TestMapToResponseListPaginate_LastPartialPage(t *testing.T) {
	items := []int{1, 2, 3} // 3 items on the last page, out of 13 total, limit 10

	mapped, metadata := MapToResponseListPaginate(items, 13, 2, 10, func(i int) int { return i })

	if len(mapped) != 3 {
		t.Fatalf("expected 3 mapped items, got %d", len(mapped))
	}
	if metadata.Size != 3 {
		t.Fatalf("expected Size 3 (actual returned count), got %d", metadata.Size)
	}
	if metadata.TotalPage != 2 {
		t.Fatalf("expected TotalPage 2 (ceil(13/10)), got %d", metadata.TotalPage)
	}
}
