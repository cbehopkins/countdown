package cntSlv

func (nc NumCol) reducer() NumCol {
	loc := 0
	for _, v := range nc {
		if v == nil || v.Val == 0 {
			continue
		}

		loc++
	}
	if loc == len(nc) {
		return nc
	}

	newArray := make(NumCol, loc)
	loc = 0
	for _, v := range nc {
		if v == nil || v.Val == 0 {
			continue
		}
		newArray[loc] = v
		loc++
	}
	return newArray
}
func workN(arrayIn NumCol, foundValues *NumMap) SolLst {
	arrayIn = arrayIn.reducer()
	lenArrayIn := arrayIn.Len()
	if foundValues.Solved() {
		return SolLst{}
	}
	if lenArrayIn <= 1 {
		return SolLst{arrayIn}
	} else if lenArrayIn == 2 {
		var tmpList NumCol
		// Generate every number by working the two
		tmpList = foundValues.make2To1(arrayIn[0:2])
		foundValues.addMany(tmpList...)
		// and add the origioanl list on
		return append(SolLst{}, tmpList, arrayIn)
	} else if lenArrayIn == 3 {
		var tmpList NumCol
		// Generate every number by working three
		tmpList = foundValues.make3To1(arrayIn[0:3])

		foundValues.addMany(tmpList...)
		// and add the origioanl list on
		return append(SolLst{}, tmpList, arrayIn)
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

	retList := workList.procWorkSelf(foundValues)
	// Make sure we include the list that started it
	retList = append(retList, arrayIn)

	// Add the entire solution list found in the previous loop in one go
	foundValues.addSol(retList, false)
	return retList
}
