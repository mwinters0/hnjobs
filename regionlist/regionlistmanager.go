package regionlist

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"math"
)

// RegionListManager allows you to track multiple parallel values over time, represented by RegionList s.
// We use this to track all formatting attributes (fg, bg, underline, etc) across the range of a document.
type RegionListManager struct {
	len              int
	regionLists      map[int]*RegionList
	emitManagerStart bool
	emitListStart    bool
	emitRegionStart  bool
	emitRegionEnd    bool
	emitListEnd      bool
	emitManagerEnd   bool
}

func NewRegionListManager(len int) *RegionListManager {
	return &RegionListManager{
		len:         len,
		regionLists: map[int]*RegionList{},
	}
}

func (rlm *RegionListManager) Len() int {
	return rlm.len
}

func (rlm *RegionListManager) Keys() []int {
	var k []int
	for key := range rlm.regionLists {
		k = append(k, key)
	}
	return k
}

func (rlm *RegionListManager) AddRegionList(key int, rl *RegionList) error {
	if rl.len != rlm.len {
		return errors.New(fmt.Sprintf(
			"new regionlist has len %d which does not match the manager len %d", rl.len, rlm.len,
		))
	}
	rlm.regionLists[key] = rl
	return nil
}

func (rlm *RegionListManager) CreateRegionList(key int, defaultValue string) error {
	rlm.regionLists[key] = NewRegionList(rlm.len, defaultValue)
	return nil
}

func (rlm *RegionListManager) InsertRegion(key int, region *Region) error {
	return rlm.regionLists[key].InsertRegion(region)
}

type MergedEvent struct {
	Offset int
	Values map[int]string
}

func (rlm *RegionListManager) MergedEvents() iter.Seq[MergedEvent] {
	rlLoc := make(map[int]int) // currently-considered region in each RL
	eligibleKeys := make(map[int]struct{})
	// init
	for key := range rlm.regionLists {
		// every RL starts at the 0th region
		rlLoc[key] = 0
		// every key is eligible
		eligibleKeys[key] = struct{}{}
	}
	return func(yield func(MergedEvent) bool) {
		// We only need to emit starts because that's when values change, but we can have multiple
		// regions with new starts ("winners") at the same offset.
		curValues := make(map[int]string)
		for {
			if len(eligibleKeys) == 0 {
				// done! we've emitted all RegionStarts
				break
			}
			var winningOffset = math.MaxInt
			var winningKeys []int
			for key := range eligibleKeys {
				thisOffset := rlm.regionLists[key].regions[rlLoc[key]].Start
				if thisOffset <= winningOffset {
					// a winner
					if thisOffset < winningOffset {
						//with better offset than previously seen - eject previous winners
						winningKeys = []int{}
					}
					winningOffset = thisOffset
					winningKeys = append(winningKeys, key)
				}
			}
			for _, winningKey := range winningKeys {
				curValues[winningKey] = rlm.regionLists[winningKey].regions[rlLoc[winningKey]].Value
				rlLoc[winningKey]++ // advance
				if rlLoc[winningKey] == len(rlm.regionLists[winningKey].regions) {
					// we just emitted the last region for this RL
					delete(eligibleKeys, winningKey)
				}
			}
			if !yield(MergedEvent{
				Offset: winningOffset,
				Values: maps.Clone(curValues),
			}) {
				return
			}
		}
	}
}

func (rlm *RegionListManager) ResizeAt(offset int, amount int) (newLen int, e error) {
	if amount < 0 && int(math.Abs(float64(amount))) > rlm.len {
		return rlm.len, errors.New(fmt.Sprintf("amount to shrink (%d) is greater than rlm len (%d)", amount, rlm.len))
	}
	if amount == 0 {
		return rlm.len, nil
	}
	if offset < 0 {
		return rlm.len, errors.New(fmt.Sprintf("index %d is less than 0", offset))
	}
	if offset > rlm.len {
		return rlm.len, errors.New(fmt.Sprintf("index %d is greater than rlm len %d", offset, rlm.len))
	}

	newLen = rlm.len + amount
	for key := range rlm.regionLists {
		nl, err := rlm.regionLists[key].ResizeAt(offset, amount)
		if err != nil {
			return rlm.len, err
		}
		if nl != newLen {
			return rlm.len, errors.New(fmt.Sprintf(
				"rl %d return newLen %d, doesn't match rlm newLen %d",
				key, nl, newLen,
			))
		}
	}
	rlm.len = newLen
	return rlm.len, nil
}
