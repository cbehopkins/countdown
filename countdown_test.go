package main

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cbehopkins/combination"
	cntSlv "github.com/cbehopkins/countdown/cnt_slv"
)

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
func runArray(target int, candidateArray []int) bool {

	foundValues := cntSlv.NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true

	tmpChan := foundValues.CountHelper(target, candidateArray)
	for tmp := range tmpChan {
		if tmp.Exists(target) {
			return true
		}
	}
	return foundValues.Solved()
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
	var i uint64
	i = 0
	for outerLength := 2; outerLength < 3; outerLength++ {
		gcOuter := combination.NewGeneric(len(bigNumbersSet), outerLength, copyFuncOuter)
		for err := gcOuter.Next(); err == nil; err = gcOuter.Next() {
			gcInner := combination.NewGeneric(len(smallNumbersSet), 6-outerLength, copyFuncInner)
			for err := gcInner.Next(); err == nil; err = gcInner.NextSkipN(1024 * 256) {
				i++
				copy(dstArray[:outerLength], dstArrayOuter)
				copy(dstArray[outerLength:], dstArrayInner)
				singleHowwyTst(dstArray, t)
				cnt++
				if cnt >= 128 {
					t.Log("Done 128", i, dstArray)
					cnt = 0
				}
			}
			break
		}
	}
}
