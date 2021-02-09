package cntSlv

// workers contains the worker functions
// That is the functions that work on the lists to turn them into
// number pairs to process
// This means taking sets of numbers and permuting them to come up with
// all possible combinations
// and then taking these combinations and devolving them into smaller sets
// that can be worked on in turn
func workN(arrayIn NumCol, foundValues *NumMap) SolLst {
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
		// Generate every number by working the two
		tmpList = foundValues.make2To1(arrayIn[0:2])
		foundValues.addMany(tmpList...)
		// and add the origioanl list on
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
	// the final list we want to return is {{5,4}, {1,4}, {9},{3}, {2,3,4}}

	// It's important to note that is arrayIn has length of n
	// then workList will be a list of arrays of max length n-1
	workList := newWrkLst(arrayIn)
	// so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }
	workerFunc := func(aNum, bNum *Number) bool {
		bobList := make([]*Number, 2)
		bobList[0] = aNum
		bobList[1] = bNum

		var tmpList NumCol
		tmpList = foundValues.make2To1(bobList)
		//foundValues.addMany(tmpList...)
		retList = append(retList, tmpList)
		return true
	}
	workList.procWork(foundValues, workerFunc)
	// Make sure we include the list that started it
	retList = append(retList, arrayIn)

	// Add the entire solution list found in the previous loop in one go
	foundValues.addSol(retList, false)
	return retList
}
