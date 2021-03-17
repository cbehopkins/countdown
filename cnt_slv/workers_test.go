package cntSlv

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cbehopkins/combination"
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

type howwy struct {
	Val int
	How string
}

func newHowwy(val int, operator string, ha, hb string) howwy {
	return howwy{val, "(" + ha + operator + hb + ")"}
}
func newHowwyArray(input []int) []howwy {
	array := make([]howwy, len(input))
	for i, v := range input {
		array[i] = howwy{v, strconv.Itoa(v)}
	}
	return array
}
func do2(input []howwy) (result []howwy) {
	if input[0].Val == 0 || input[1].Val == 0 {
		return input
	}
	result = make([]howwy, 6)
	a := input[0].Val
	b := input[1].Val
	aHow := input[0].How
	bHow := input[1].How

	result[0] = newHowwy(a+b, "+", aHow, bHow)
	result[1] = newHowwy(a*b, "*", aHow, bHow)
	eq := a == b
	gt := a > b
	amb0 := (a % b) == 0
	bma0 := (b % a) == 0
	if eq {
		result[2] = newHowwy(0, "-", aHow, bHow)
		result[3] = newHowwy(1, "/", aHow, bHow)
	} else {
		result[3] = newHowwy(0, "/", aHow, bHow) // technically not needed, but...
		if gt {
			result[2] = newHowwy(a-b, "-", aHow, bHow)
			if amb0 {
				result[3] = newHowwy(a/b, "/", aHow, bHow)
			}

		} else {
			result[2] = newHowwy(b-a, "-", bHow, aHow)
			if bma0 {
				result[3] = newHowwy(b/a, "/", bHow, aHow)
			}
		}
	}
	result[4] = newHowwy(a, "", aHow, "")
	result[5] = newHowwy(b, "", bHow, "")
	return result
}
func select1(input []howwy) howwy {
	nzFound := false
	for _, v := range input {
		if v.Val > 0 {
			nzFound = true
		}
	}
	if !nzFound {
		log.Fatal("Zeroes only")
	}

	for true {
		v := input[rand.Intn(len(input))]
		if v.Val != 0 {
			return v
		}
	}
	return input[0]
}
func allN(input []howwy) howwy {
	if len(input) == 1 {
		return input[0]
	}
	if len(input) == 2 {
		return select1(do2(input))
	}
	array := make([]howwy, len(input)-1)
	array[0] = select1(do2(input[:2]))
	copy(array[1:], input[2:])
	return allN(array)
}
func singleHowwyTst(candidateArray []int, t *testing.T) {
	// t.Log("Candidate Array:", candidateArray)
	for i := 0; i < 2; i++ {
		target := allN(newHowwyArray(candidateArray))
		// t.Log("Target:", target)
		if !runArray(target.Val, candidateArray) {
			t.Log("Candidate Array:", candidateArray)
			t.Log("Target:", target)
			t.Fatal("Unable to prove")
		}
	}
}
func TestAllCombinations(t *testing.T) {
	// t.Skip()
	bigNumbersSet := []int{25, 50, 75, 100}
	smallNumbersSet := []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10}

	dstArrayInner := make([]int, 4)
	dstArrayOuter := make([]int, 3)
	dstArray := make([]int, 6)

	copyFuncOuter := func(i, j int) {
		dstArrayOuter[i] = bigNumbersSet[j]
	}
	copyFuncInner := func(i, j int) {
		dstArrayInner[i] = smallNumbersSet[j]
	}
	cnt := 0
	for outerLength := 2; outerLength < 4; outerLength++ {
		gcOuter := combination.NewGeneric(len(bigNumbersSet), outerLength, copyFuncOuter)
		for err := gcOuter.Next(); err == nil; err = gcOuter.Next() {
			gcInner := combination.NewGeneric(len(smallNumbersSet), 6-outerLength, copyFuncInner)
			for err := gcInner.Next(); err == nil; err = gcInner.Next() {
				copy(dstArray[:outerLength], dstArrayOuter)
				copy(dstArray[outerLength:], dstArrayInner)
				singleHowwyTst(dstArray, t)
				cnt++
				if cnt >= 128 {
					t.Log("Done 128")
					cnt = 0
				}
			}
		}
	}
}

func runArray(target int, candidateArray []int) bool {
	var nc NumCol

	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true
	for _, v := range candidateArray {
		nc.AddNum(v, foundValues)
	}

	foundValues.SetTarget(target)
	tmpChan := permuteN(nc, foundValues)
	for tmp := range tmpChan {
		if tmp.Exists(target) {
			return true
		}
	}
	if *foundValues.solved {
		val, ok := foundValues.nmp[target]
		if ok {
			_ = val.ProveSol() // This does its own error reporting
		}
		return ok
	}
	return false
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
