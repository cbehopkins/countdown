package cntSlv

import (
	"log"
	"testing"
)

func TestWeirdWork(t *testing.T) {
	var target int
	// (9-1)*50 = 400
	// (100 + 9*3) = 327
	// (400+327)= 727
	target = 727

	var proof400 SolLst
	var proof327 SolLst

	var mk400 NumCol
	var mk327 NumCol
	var combined NumCol

	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true
	mk400.AddNum(50, foundValues)
	mk400.AddNum(9, foundValues)
	mk400.AddNum(1, foundValues)
	mk327.AddNum(100, foundValues)
	mk327.AddNum(9, foundValues)
	mk327.AddNum(3, foundValues)

	foundValues.SetTarget(target)

	proof400 = append(proof400, mk400) // Add on the work item that is the source
	proof327 = append(proof327, mk327) // Add on the work item that is the source
	sol400 := workN(mk400, foundValues)
	sol327 := workN(mk327, foundValues)

	log.Println("Find 400", sol400.StringNum(400))
	log.Println("Find 327", sol327.StringNum(327))

	combined = append(mk400, mk327...)
	var workList wrkLst
	workList = newWrkLst(combined)
	chkFunc := func() bool {
		for _, workUnit := range workList.lst {
			var unitA, unitB NumCol
			unitA = workUnit[0]
			unitB = workUnit[1]
			if mk400.Equal(unitA) {
				if mk327.Equal(unitB) {
					tmp400 := workN(unitA, foundValues)
					tmp327 := workN(unitB, foundValues)
					if !tmp400.Exists(400) {
						return false
					}
					if !tmp327.Exists(327) {
						return false
					}
					return true
				}
			}
		}
		return false
	}
	log.Println("Its:", chkFunc())
	solCombined := workN(combined, foundValues)
	log.Println("Find 727", solCombined.StringNum(727))
}

func tstWorker(target int, candidateArray []int, fc func(NumCol, *NumMap)) {
	// (9-1)*50 = 400
	// (100 + 9*3) = 327
	// (400+327)= 727

	var mk400 NumCol

	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true
	for _, v := range candidateArray {
		mk400.AddNum(v, foundValues)
	}

	foundValues.SetTarget(target)

	fc(mk400, foundValues)
}
func TestWorkn(t *testing.T) {
	var tmp SolLst
	fun := func(nc NumCol, fv *NumMap) {
		tmp = workN(nc, fv)
	}
	target := 727
	candidateArray := []int{50, 9, 1, 100, 9, 3}
	tstWorker(target, candidateArray, fun)
	if !tmp.Exists(target) {
		log.Fatal("Couldn't find", target)
	}

}
func TestPermute(t *testing.T) {
	target := 727
	candidateArray := []int{50, 9, 1, 100, 9, 3}

	var tmpChan chan SolLst
	fun := func(nc NumCol, fv *NumMap) {
		tmpChan = permuteN(nc, fv)
	}
	tstWorker(target, candidateArray, fun)
	var found bool
	var unfound bool
	for tmp := range tmpChan {
		if !tmp.Exists(target) {
			unfound = true
		} else {
			found = true
		}
	}
	if !found {
		log.Fatal("There should be at least on permutation where it is found")
	}
	if !unfound {
		log.Fatal("There should be at least on permutation where it is unfound")
	}
}
