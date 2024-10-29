package regionlist

import (
	"reflect"
	"testing"
)

func TestRLMRegions(t *testing.T) {
	t.Run("Insert", func(t *testing.T) {
		// 1:
		//   aaabbbccc
		// 2:
		//   ddeeeeeff

		rlm := NewRegionListManager(9)
		err := rlm.CreateRegionList(0, "a")
		if err != nil {
			t.Error(err)
		}
		_ = rlm.InsertRegion(0, &Region{
			Start: 3,
			End:   5,
			Value: "b",
		})
		_ = rlm.InsertRegion(0, &Region{
			Start: 6,
			End:   8,
			Value: "c",
		})
		expected0 := &RegionList{
			len: 9,
			regions: []*Region{
				{
					Start: 0,
					End:   2,
					Value: "a",
				},
				{
					Start: 3,
					End:   5,
					Value: "b",
				},
				{
					Start: 6,
					End:   8,
					Value: "c",
				},
			},
		}
		compareRegionLists(t, expected0, rlm.regionLists[0])

		err = rlm.CreateRegionList(1, "d")
		if err != nil {
			t.Error(err)
		}
		_ = rlm.InsertRegion(1, &Region{
			Start: 2,
			End:   6,
			Value: "e",
		})
		_ = rlm.InsertRegion(1, &Region{
			Start: 7,
			End:   8,
			Value: "f",
		})
		expected1 := &RegionList{
			len: 9,
			regions: []*Region{
				{
					Start: 0,
					End:   1,
					Value: "d",
				},
				{
					Start: 2,
					End:   6,
					Value: "e",
				},
				{
					Start: 7,
					End:   8,
					Value: "f",
				},
			},
		}
		compareRegionLists(t, expected1, rlm.regionLists[1])
	})

	// TODO test Add
}

func TestRLMMergedEvents(t *testing.T) {
	// 1:
	//   aaabbbccc
	// 2:
	//   ddeeeeeff
	rlm := NewRegionListManager(9)
	err := rlm.CreateRegionList(0, "a")
	if err != nil {
		t.Error(err)
	}
	_ = rlm.InsertRegion(0, &Region{
		Start: 3,
		End:   5,
		Value: "b",
	})
	_ = rlm.InsertRegion(0, &Region{
		Start: 6,
		End:   8,
		Value: "c",
	})
	err = rlm.CreateRegionList(1, "d")
	if err != nil {
		t.Error(err)
	}
	_ = rlm.InsertRegion(1, &Region{
		Start: 2,
		End:   6,
		Value: "e",
	})
	_ = rlm.InsertRegion(1, &Region{
		Start: 7,
		End:   8,
		Value: "f",
	})

	var emitted []MergedEvent
	for m := range rlm.MergedEvents() {
		emitted = append(emitted, m)
	}
	//for _, e := range emitted {
	//	fmt.Printf("%#v\n", e)
	//}

	expected := []MergedEvent{
		{Offset: 0, Values: map[int]string{0: "a", 1: "d"}},
		{Offset: 2, Values: map[int]string{0: "a", 1: "e"}},
		{Offset: 3, Values: map[int]string{0: "b", 1: "e"}},
		{Offset: 6, Values: map[int]string{0: "c", 1: "e"}},
		{Offset: 7, Values: map[int]string{0: "c", 1: "f"}},
	}

	if !reflect.DeepEqual(emitted, expected) {
		t.Errorf("expected:\n%#v\nactual:\n%#v", expected, emitted)
	}

}
