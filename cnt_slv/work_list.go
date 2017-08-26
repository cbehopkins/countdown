package cntSlv

import (
	"log"

	"github.com/tonnerre/golang-pretty"
)

type WrkLst struct {
	lst []SolLst
}

func (wl WrkLst) procWork(foundValues *NumMap, wf func(a, b *Number) bool) {
	run := true
	for _, workUnit := range wl.lst {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

		if foundValues.SelfTest {
			// Sanity check for programming errors
			workUnitLength := workUnit.Len()
			if workUnitLength != 2 {
				pretty.Println(wl)
				log.Fatalf("Invalid work unit length, %d", workUnitLength)
			}
		}

		unitA := workUnit[0]
		unitB := workUnit[1]
		listA := workN(unitA, foundValues, false)
		listB := workN(unitB, foundValues, false)
		gimmieA := NewGimmie(listA)
		gimmieB := NewGimmie(listB)

		for aNum, errA := gimmieA.Next(); (errA == nil) && run; aNum, errA = gimmieA.Next() {
			for bNum, errB := gimmieB.Next(); (errB == nil) && run; bNum, errB = gimmieB.Next() {
				run = wf(aNum, bNum)
			}
			gimmieB.Reset()
		}
		if !run {
			return
		}
	}
}

func NewWrkLst(arrayA NumCol) WrkLst {
	var workList []SolLst
	// Easier to explain by example:
	// {2,3,4} -> {{2},{3,4}}
	// {2,3,4,5} -> {{2}, {3,4,5}}
	//           -> {{2,3},{4,5}}
	// {2,3,4,5,6} -> {{2},{3,4,5,6}}
	//             -> {{2,3},{4,5,6}}
	//             -> {{2,3,4},{5,6}}

	// The consumer of this list of list (of list) will then feed each list length >1 into a the work+_n function
	// In order to get down to a {{a},{b}} which can then be worked
	// The important point is that even though the list we return may be indefinitly long
	// each work unit within it is then a smaller unit
	// so an input array of 3 numbers only generates work units that contain number lists of length 2 or less

	lenArrayM1 := arrayA.Len() - 1

	for i := 0; i < (lenArrayM1); i++ {
		var arA, arB NumCol
		// for 3 items in arrar
		// {0},{1,2}, {0,1}{2}
		arA = make(NumCol, i+1)
		copy(arA, arrayA[0:i+1])
		arB = make(NumCol, (arrayA.Len() - (i + 1)))

		copy(arB, arrayA[(i+1):(arrayA.Len())])
		var workItem SolLst // {{2},{3,4}};
		// a work item always contains 2 elements to the array
		workItem = append(workItem, arA, arB)
		workList = append(workList, workItem)
	}
	return WrkLst{lst: workList}
}
func (wl WrkLst) Len() int {
	return len(wl.lst)
}

func (wl WrkLst) Get(itm int) SolLst {
	return wl.lst[itm]
}
func (wl WrkLst) Last() SolLst {
	return wl.Get(wl.Len() - 1)
}
