package cntSlv

import (
	"log"
)

// workers contains the worker functions
// That is the functions that work on the lists to turn them into
// number pairs to process
// This means taking sets of numbers and permuting them to come up with
// all possible combinations
// and then taking these combinations and devolving them into smaller sets
// that can be worked on in turn
const (
	LonMap = iota
	ParMap
	NetMap
)

func WorkN(arrayIn NumCol, foundValues *NumMap) SolLst {
	for _, j := range arrayIn {
		if j.Val == 0 {
			log.Fatal("WorkN fed a 0 number")
		}
	}
	return workN(arrayIn, foundValues, false)
}

func workN(arrayIn NumCol, foundValues *NumMap, multipass bool) SolLst {
	var retList SolLst
	lenArrayIn := arrayIn.Len()
	if foundValues.Solved() {
		return retList
	}
	if lenArrayIn == 1 {
		//ret_list = append(ret_list, &array_in)
		return SolLst{arrayIn}
	} else if lenArrayIn == 2 {
		var tmpList NumCol
		tmpList = foundValues.make2To1(arrayIn[0:2])
		foundValues.AddMany(tmpList...)
		retList = append(retList, tmpList, arrayIn)
		return retList
	}

	// work_n takes
	// let's use work 3 as a first example {2,3,4} and should generate everything that can be done with these 3 numbers
	// Note: for these explanantions I'll assume we just add and subtract numbers
	// We do not return the supplied list with the return
	// we also do no permute the input numbers as we know that permute function will do this for us
	// So in this example we would look to do several steps first we feed to make_3
	// This will treat the input as {2,3),{4} it works the first list to get:
	// {5,1} (from 2+3 and 3-2) and therefore returns {{5,4}, {1,4}}
	// we then take each value in this list and work that to get {{9},{3}}
	// the final list we want to return is {{5,4}, {1,4}, {9},{3}}
	// the reason to not return {2,3,4} is so that in the grand scheme of things we can recurse these lists
	var workList WrkLst
	workList = NewWrkLst(arrayIn)
	// so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }

	if multipass {
		crossLen := 0
		numNumbersToMake := 0

		determineSizeFunc := func(aNum, bNum *Number) bool {
			if aNum.Val <= 0 || bNum.Val <= 0 {
				log.Fatalf("Gimmie gave %d, %d", aNum.Val, bNum.Val)
			}
			tmp,
				_, _, _, _, _ := foundValues.doMaths([]*Number{aNum, bNum})
			numNumbersToMake += tmp
			crossLen++
			return true
		}

		workList.procWork(foundValues, determineSizeFunc)

		topSrcToMake := crossLen * 2
		topNumbersToMake := numNumbersToMake
		//current_item = 0
		var workUnit SolLst
		// Last Item on work list contains sources
		workUnit = workList.Last()
		// Malloc the memory once!
		currentNumberLoc := 0
		// This is the list of numbers that calculations are done from
		srcList := foundValues.acquireNumbers(topSrcToMake)
		// This is the list of numbers that will be used in the proof
		// i.e. the list that calculations results end up in
		numList := foundValues.acquireNumbers(topNumbersToMake)
		// And this allocates the list that will point to those (previously allocated) numbers

		retList = make(SolLst, 0, (crossLen + len(workUnit)))
		// Add on the work unit because that contains sub combinations that may be of use
		retList = append(retList, workUnit...)
		currentSrc := 0
		workerFunc := func(aNum, bNum *Number) bool {
			// Here we have unrolled the functionality of make_2_to_1
			// So that it can use a single array
			// This is all to put less work on the malloc and gc

			if foundValues.Solved() {
				return false
			}

			// We have to recalculate

			srcList[currentSrc] = aNum
			srcList[currentSrc+1] = bNum
			// Shorthand to make code more readable
			bobList := srcList[currentSrc : currentSrc+2]
			if aNum.Val == 0 || bNum.Val == 0 {
				log.Fatalf("Gimmie gave %d, %d", aNum.Val, bNum.Val)
			}

			numToMake,
				addSet, mulSet, subSet, divSet,
				aGtB := foundValues.doMaths(bobList)

			// Shorthand
			tmpList := numList[currentNumberLoc:(currentNumberLoc + numToMake)]

			// Populate the part of the return list for this run
			// This is the arra AddItems will write into
			// num_list gets filled with numbers, tmp_list is an alias to the same data here
			foundValues.AddItems(bobList, numList, currentNumberLoc,
				addSet, mulSet, subSet, divSet,
				aGtB)
			currentNumberLoc += numToMake
			currentSrc += 2
			if foundValues.SelfTest {
				for _, v := range tmpList {
					v.ProveSol()
				}
			}
			retList = append(retList, tmpList)
			return true
		}
		workList.procWork(foundValues, workerFunc)

	} else {

		workerFunc := func(aNum, bNum *Number) bool {
			bobList := make([]*Number, 2)
			bobList[0] = aNum
			bobList[1] = bNum

			var tmpList NumCol
			tmpList = foundValues.make2To1(bobList)
			foundValues.AddMany(tmpList...)
			retList = append(retList, tmpList, arrayIn)
			return true
		}
		workList.procWork(foundValues, workerFunc)
	}
	// Add the entire solution list found in the previous loop in one go
	foundValues.AddSol(retList, false)
	return retList
}
