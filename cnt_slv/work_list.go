package cntSlv

import (
	"log"
)

type wrkLst struct {
	lst []SolLst
}

func (wl wrkLst) procWork(foundValues *NumMap, wf func(a, b *Number) bool) {
	run := true
	for _, workUnit := range wl.lst {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2,3},{4,5,6}}

		if foundValues.SelfTest {
			// Sanity check for programming errors
			workUnitLength := workUnit.Len()
			if workUnitLength != 2 {
				log.Println(wl)
				log.Fatalf("Invalid work unit length, %d", workUnitLength)
			}
		}

		unitA := workUnit[0]
		unitB := workUnit[1]
		// Return a list of all the numbers that can be made with this set
		// i.e. {3,4} becomes {{3,4},{1},{7},{12}}
		listA := workN(unitA, foundValues)
		listB := workN(unitB, foundValues)
		// Give me all the numbers piossible from the solutions list
		// to cross with the others
		// i.e. {{3,4},{1},{7},{12}} becomes {3,4,1,7,12}
		// except that it is done without building a temporary array
		gimmieA := newGimmie(listA)
		gimmieB := newGimmie(listB)

		for aNum, errA := gimmieA.next(); (errA == nil) && run; aNum, errA = gimmieA.next() {
			for bNum, errB := gimmieB.next(); (errB == nil) && run; bNum, errB = gimmieB.next() {
				// FIXME pregenerate array that wf will write into
				run = wf(aNum, bNum)
			}
			gimmieB.reset()
		}
		if !run {
			return
		}
	}
}

func (wl wrkLst) procWorkSelf(foundValues *NumMap) SolLst {
	wf := func(aNum, bNum *Number, tmpList NumCol, currentNumberLoc int) int {
		if aNum == nil || aNum.Val == 0 {
			// Nothing new to add
			return currentNumberLoc
		}
		if bNum == nil || bNum.Val == 0 {
			// Nothing new to add
			return currentNumberLoc
		}

		return doCalcOn2(tmpList, NumCol{aNum, bNum}, currentNumberLoc)
	}
	var retList SolLst
	for _, workUnit := range wl.lst {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2,3},{4,5,6}}

		if foundValues.SelfTest {
			// Sanity check for programming errors
			workUnitLength := workUnit.Len()
			if workUnitLength != 2 {
				log.Println(wl)
				log.Fatalf("Invalid work unit length, %d", workUnitLength)
			}
		}

		unitA := workUnit[0]
		unitB := workUnit[1]
		// Return a list of all the numbers that can be made with this set
		// i.e. {3,4} becomes {{3,4},{1},{7},{12}}
		listA := workN(unitA, foundValues)
		listB := workN(unitB, foundValues)
		// Give me all the numbers piossible from the solutions list
		// to cross with the others
		// i.e. {{3,4},{1},{7},{12}} becomes {3,4,1,7,12}
		// except that it is done without building a temporary array
		gimmieA := newGimmie(listA)
		gimmieB := newGimmie(listB)

		wfCount := gimmieA.items() * gimmieB.items()
		// Now grab the memory
		tmpList := make(NumCol, 4*wfCount)
		tstLst := make([]Number, 4*wfCount)
		for i := range tmpList {
			tmpList[i] = &tstLst[i]
		}
		currentNumberLoc := 0
		for aNum, errA := gimmieA.next(); errA == nil; aNum, errA = gimmieA.next() {
			for bNum, errB := gimmieB.next(); errB == nil; bNum, errB = gimmieB.next() {
				currentNumberLoc = wf(aNum, bNum, tmpList, currentNumberLoc)
			}
			gimmieB.reset()
		}
		retList = append(retList, tmpList[0:currentNumberLoc])
	}
	return retList
}

// NewWrkLst returns a new work list from a Number Collection
// Easier to explain by example:
// {2,3,4} -> {{2},{3,4}}
//         -> {{2,3},{4}}
// {2,3,4,5} -> {{2}, {3,4,5}}
//           -> {{2,3},{4,5}}
//           -> {{2,3,4},{5}}
// {2,3,4,5,6} -> {{2},{3,4,5,6}}
//             -> {{2,3},{4,5,6}}
//             -> {{2,3,4},{5,6}}
// etc
// The consumer of this list of list (of list) will then feed each list length >1 into a the work+_n function
// In order to get down to a {{a},{b}} which can then be worked
// The important point is that even though the list we return may be indefinitly long
// each work unit within it is then a smaller unit
// so an input array of 3 numbers only generates work units that contain number lists of length 2 or less
func newWrkLst(arrayA NumCol) wrkLst {
	var workList []SolLst

	lenArrayM1 := arrayA.Len() - 1

	for i := 0; i < lenArrayM1; i++ {
		var arA, arB NumCol
		// for 3 items in arrar
		// {0},{1,2}, {0,1}{2}
		arA = make(NumCol, i+1)
		copy(arA, arrayA[0:i+1])
		arB = make(NumCol, (arrayA.Len() - (i + 1)))

		copy(arB, arrayA[(i+1):(arrayA.Len())])
		workItem := SolLst{arA, arB} // {{2},{3,4}};
		// a work item always contains 2 elements to the array
		//workItem = append(workItem, arA, arB)
		workList = append(workList, workItem)
	}
	return wrkLst{lst: workList}
}

// Len returns the length of the worklist
func (wl wrkLst) Len() int {
	return len(wl.lst)
}

// Get a specific item off the worklist
func (wl wrkLst) Get(itm int) SolLst {
	return wl.lst[itm]
}

// Last retrieves the last item
// This contains the sources that started this list
func (wl wrkLst) Last() SolLst {
	return wl.Get(wl.Len() - 1)
}
