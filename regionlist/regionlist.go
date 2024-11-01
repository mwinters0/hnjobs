package regionlist

import (
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
)

// Region represents a single span of a single attribute, e.g. foreground color.  Note that unlike slice, End is
// inclusive because this is more natural when working primarily with indices.
type Region struct {
	Start int
	End   int // not exactly necessary but simplifies management
	Value string
}

// RegionList is a simple Run-Length Encoding of the state of a state machine over time (or any other value across an
// integer range.)  It ensures that at every point in a range we will always have exactly one value.  In this project
// we use RegionList with a string value to represent the various "painted" regions of text, e.g. a range of text where
// the foreground color is blue.
//
// RegionList ensures that all added regions maintain contiguity of the range.  RegionList handles region overlaps as
// necessary by splitting regions / deleting / merging / etc.  Regions cannot be directly deleted, but they can be
// overwritten with the default value.  A RegionList can be expanded/contracted at a specific offset to represent
// document edits.
//
// The intended way to consume a RegionList is via a RegionListManager.
//
// (And yes, for this project we could have used an HTML-to-TTY library, but we still had to do a bunch of text
// processing to insert the correct markup anyway.  Plus for future projects I don't want to be tied to HTML as the
// only way to format parsed terminal text.)
type RegionList struct {
	len             int
	regions         []*Region
	emitListStart   bool
	emitRegionStart bool
	emitRegionEnd   bool
	emitListEnd     bool
}

func NewRegionList(len int, defaultValue string) *RegionList {
	return &RegionList{
		len:     len,
		regions: []*Region{{Start: 0, End: len - 1, Value: defaultValue}},
	}
}

func (rl *RegionList) Len() int {
	return rl.len
}

func (rl *RegionList) InsertRegion(newRegion *Region) error {
	// sanity check
	if newRegion.Start < 0 ||
		newRegion.End < newRegion.Start || //note: a one-character region is valid and will have the same start and end
		newRegion.Start > rl.len ||
		newRegion.End > rl.len {
		log.Fatal(errors.New(fmt.Sprintf("new region is invalid: %#v", newRegion)))
	}

	// Cases we have to handle, left to right:
	// 1: Internal Overlap
	//    cur: aaa
	//    new:  i
	//    action: split A
	// 2: Span End
	//    cur: aaa
	//    new:  iii
	//    action: resize A.End
	// 3: External Overlap
	//    cur: aaabbb
	//    new:    iii
	//    action: delete B
	// 4: Span Start
	//    cur: aaabbb
	//    new:   iii
	//    action: resize B.Start
	//
	// Special: we can get some variant of all of the above where the new region has the same value as one of
	// the existing regions, e.g.:
	//    cur: aaabbb
	//    new:   aa
	//    action: resize the current A.end

	// We avoid any "check the whole list" logic for performance since the list could be long, though that
	// approach would be much simpler.

	var deleteRegions []int //a range of regions marked for deletion.  [0] is start, [1] is end
	done := false           //need to call finish() from different places and know whether it's already been called
	doInsert := true
	coalescent := -1 //when expanding a region to avoid adjacency, need to keep track of which region to expand
	finish := func(i int) {
		if doInsert {
			rl.regions = slices.Insert(rl.regions, i, newRegion)
		}
		if len(deleteRegions) > 0 {
			rl.regions = slices.Concat(rl.regions[:deleteRegions[0]], rl.regions[deleteRegions[1]+1:])
		}
		done = true
	}
	for i, curRegion := range rl.regions {
		if curRegion.End < newRegion.Start {
			//we're not yet into the relevant range
			continue
		}
		if curRegion.Start > newRegion.End {
			//we're one iteration past the relevant range, so we are done.
			finish(i)
			break
		}

		// case 1: split curRegion
		// cur: aaaa
		// new:  ii
		if curRegion.Start < newRegion.Start && curRegion.End > newRegion.End {
			if curRegion.Value == newRegion.Value {
				// cur: aaaa
				// new:  aa
				// no reason to make any changes
				done = true
				break
			}
			// cur: aaaa
			// new:  ii
			spanEnd := rl.regions[i].End
			rl.regions[i].End = newRegion.Start - 1                //trim existing region
			rl.regions = slices.Insert(rl.regions, i+1, newRegion) //insert
			rl.regions = slices.Insert(rl.regions, i+2,
				&Region{Start: newRegion.End + 1, End: spanEnd, Value: rl.regions[i].Value},
			) //append remainder
			// case 1 is mutually exclusive with all others so we're done
			done = true
			break
		}

		// case 2: resize curRegion.End
		// cur: aaaa
		// new:  iiii
		if curRegion.Start < newRegion.Start && curRegion.End <= newRegion.End {
			if curRegion.Value == newRegion.Value {
				// cur: aaabbb
				// new:   aa
				// Simply resize curRegion.End to the right. We'll handle the B side in another iteration
				rl.regions[i].End = newRegion.End
				doInsert = false
				continue
			}
			// Resize curRegion.End to the left
			rl.regions[i].End = newRegion.Start - 1
			continue
		}

		// case 3: external overlap (possibly multiple) - delete curRegion
		// cur: aabbccaa
		// new:   iiii
		if curRegion.Start >= newRegion.Start && curRegion.End <= newRegion.End {
			//newRegion overlaps curRegion. Mark it for deletion.
			if len(deleteRegions) == 0 {
				deleteRegions = append(deleteRegions, i) //start
				deleteRegions = append(deleteRegions, i) //end
			} else {
				//this is not the first curRegion to be overlapped
				deleteRegions[1] = i //bump the end
			}
			if i > 0 && coalescent < 0 && rl.regions[i-1].Value == newRegion.Value {
				// Coalescence:
				// cur: aabbccdd
				// new:   aaaa
				// We want to expand the prior region rather than insert another.  Start a coalescence.
				coalescent = i - 1
			}
			if coalescent >= 0 {
				// Should we expand the coalescent or end it?
				if rl.regions[coalescent].Value == newRegion.Value {
					rl.regions[coalescent].End = curRegion.End // expand the coalescent
					newRegion.Start = curRegion.End + 1        // adjust the new region
					if newRegion.Start > newRegion.End {
						//we're done
						doInsert = false
						finish(i)
						break
					}
				} else {
					//stop the coalescence
					coalescent = -1
				}
			}
			// Check to see if the next region could be merged with this one.
			// cur:  aabbccdd
			// new:    dddd
			if i < len(rl.regions)-1 && // ending in the middle (there are regions after newRegion)
				newRegion.End <= rl.regions[i+1].Start && // last loop
				newRegion.Value == rl.regions[i+1].Value { // next region has same value
				rl.regions[i+1].Start = newRegion.Start
				doInsert = false
				finish(i)
				break
			}
			continue
		}

		// case 4: resize curRegion.Start
		// cur:   bbb
		// new: iii
		if newRegion.End < curRegion.End && newRegion.End >= curRegion.Start {
			if curRegion.Value == newRegion.Value {
				// cur:  bbb
				// new: bbb
				// Resize curRegion to the left
				rl.regions[i].Start = newRegion.Start
				doInsert = false
				finish(i)
				break
			}
			// Resize curRegion to the right
			rl.regions[i].Start = newRegion.End + 1
			finish(i)
			break
		}
		panic(errors.New(fmt.Sprintf("Unhandled region!\n  rl: %s\n  i: %d\n  newRegion: %v\n", rl, i, newRegion)))
	}
	if !done {
		// newRegion is the last one in the list, so the loop didn't have a chance to call finish()
		finish(len(rl.regions))
	}
	return nil
}

func (rl *RegionList) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("RegionList{\n  len: %d\n", rl.len))
	for i := range rl.regions {
		b.WriteString(fmt.Sprintf("  %#v,\n", *rl.regions[i]))
	}
	b.WriteString("}")
	return b.String()
}

func (rl *RegionList) Equals(other *RegionList) bool {
	// faster than reflect.DeepEquals, but maybe only needed for testing??
	if rl.len != other.len {
		return false
	}
	if len(rl.regions) != len(other.regions) {
		return false
	}
	for i := range rl.regions {
		if !(*rl.regions[i] == *other.regions[i]) {
			return false
		}
	}
	return true
}

func (rl *RegionList) ResizeAt(offset int, amount int) (newLen int, e error) {
	if amount < 0 && int(math.Abs(float64(amount))) > rl.len {
		return rl.len, errors.New(fmt.Sprintf("amount to shrink (%d) is greater than list len (%d)", amount, rl.len))
	}
	if amount == 0 {
		return rl.len, nil
	}
	if offset < 0 {
		return rl.len, errors.New(fmt.Sprintf("index %d is less than 0", offset))
	}
	if offset > rl.len {
		return rl.len, errors.New(fmt.Sprintf("index %d is greater than list len %d", offset, rl.len))
	}

	// find region at this offset
	startRegionIndex := -1
	for i, r := range rl.regions {
		if offset > r.End {
			continue
		}
		startRegionIndex = i
		break
	}
	if startRegionIndex < 0 {
		return rl.len, errors.New(fmt.Sprintf("couldn't find region at offset %d", offset))
	}
	//adjust this region
	if rl.regions[startRegionIndex].End+amount < rl.regions[startRegionIndex].Start {
		return rl.len, errors.New(fmt.Sprintf(
			"region at offset %d asked to shrink by %d but only has len %d",
			offset,
			amount,
			rl.regions[startRegionIndex].End-rl.regions[startRegionIndex].Start,
		))
	}
	rl.regions[startRegionIndex].End += amount
	//fixup remaining regions
	for i := startRegionIndex + 1; i < len(rl.regions); i++ {
		rl.regions[i].Start += amount
		rl.regions[i].End += amount
	}
	rl.len += amount
	return rl.len, nil
}
