package regionlist

import (
	"testing"
)

func TestRLInsertRegion(t *testing.T) {

	// === case 1

	t.Run("InternalOverlap", func(t *testing.T) {
		// AAAAAA
		//   II
		rl := NewRegionList(6, "A")
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   3,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 6,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "I"},
				{Start: 4, End: 5, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	t.Run("InternalOverlapSame", func(t *testing.T) {
		// AAAAAA
		//   AA
		rl := NewRegionList(6, "A")
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   3,
			Value: "A",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 6,
			regions: []*Region{
				{Start: 0, End: 5, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// === case 3

	t.Run("ExternalOverlap", func(t *testing.T) {
		// AABBCCAA
		//   IIII
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 5, Value: "C"},
				{Start: 6, End: 7, Value: "A"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 5, Value: "I"},
				{Start: 6, End: 7, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case 3, colasce left
	t.Run("ExternalOverlapSameLeft", func(t *testing.T) {
		// AABBCCDD
		//   AAAA
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 5, Value: "C"},
				{Start: 6, End: 7, Value: "D"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "A",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 5, Value: "A"},
				{Start: 6, End: 7, Value: "D"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case 3, coalesce right
	t.Run("ExternalOverlapSameRight", func(t *testing.T) {
		// AABBCCDD
		//   DDDD
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 5, Value: "C"},
				{Start: 6, End: 7, Value: "D"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "D",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 7, Value: "D"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case 2/4

	// case 2 & 4
	t.Run("Span", func(t *testing.T) {
		// AAAABBBB
		//   IIII
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 3, Value: "A"},
				{Start: 4, End: 7, Value: "B"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 5, Value: "I"},
				{Start: 6, End: 7, Value: "B"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case 4, same left
	t.Run("SpanSameLeft", func(t *testing.T) {
		// AAAABBBB
		//   AAAA
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 3, Value: "A"},
				{Start: 4, End: 7, Value: "B"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "A",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 5, Value: "A"},
				{Start: 6, End: 7, Value: "B"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case 4, same right
	t.Run("SpanSameRight", func(t *testing.T) {
		// AAAABBBB
		//   BBBB
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 3, Value: "A"},
				{Start: 4, End: 7, Value: "B"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "B",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 7, Value: "B"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	//case 4
	t.Run("SpanLeftExact", func(t *testing.T) {
		// AAAAAA
		// II
		rl := NewRegionList(6, "A")
		err := rl.InsertRegion(&Region{
			Start: 0,
			End:   1,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 6,
			regions: []*Region{
				{Start: 0, End: 1, Value: "I"},
				{Start: 2, End: 5, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// case2
	t.Run("SpanRightExact", func(t *testing.T) {
		// AAAAAA
		//     II
		rl := NewRegionList(6, "A")
		err := rl.InsertRegion(&Region{
			Start: 4,
			End:   5,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 6,
			regions: []*Region{
				{Start: 0, End: 3, Value: "A"},
				{Start: 4, End: 5, Value: "I"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// combo scenarios

	t.Run("ExternalOverlapSpanRight", func(t *testing.T) {
		// AABBAAAA
		//   IIII
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 7, Value: "A"},
			},
		}
		err := rl.InsertRegion(&Region{
			Start: 2,
			End:   5,
			Value: "I",
		})
		if err != nil {
			t.Error(err)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 5, Value: "I"},
				{Start: 6, End: 7, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})
}

func TestRLResizeAt(t *testing.T) {
	t.Run("Grow", func(t *testing.T) {
		rl := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 7, Value: "A"},
			},
		}
		res, err := rl.ResizeAt(3, 2)
		if err != nil {
			t.Error(err)
		}
		if res != 10 {
			t.Errorf("Expected res to be 10, got %d", res)
		}
		rlExpected := &RegionList{
			len: 10,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 5, Value: "B"},
				{Start: 6, End: 9, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	t.Run("Shrink", func(t *testing.T) {
		rl := &RegionList{
			len: 10,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 5, Value: "B"},
				{Start: 6, End: 9, Value: "A"},
			},
		}
		res, err := rl.ResizeAt(3, -2)
		if err != nil {
			t.Error(err)
		}
		if res != 8 {
			t.Errorf("Expected res to be 8, got %d", res)
		}
		rlExpected := &RegionList{
			len: 8,
			regions: []*Region{
				{Start: 0, End: 1, Value: "A"},
				{Start: 2, End: 3, Value: "B"},
				{Start: 4, End: 7, Value: "A"},
			},
		}
		compareRegionLists(t, rlExpected, rl)
	})

	// TODO add test cases for invalid inputs
}

func compareRegionLists(t *testing.T, expected, actual *RegionList) {
	if expected.len != actual.len {
		t.Errorf("regionLists are different len.\nexpected: %s\nactual: %s\n", expected, actual)
	}
	if !expected.Equals(actual) {
		t.Errorf("regions aren't equal.\nexpected: %s\nactual: %s\n", expected, actual)
	}
}
