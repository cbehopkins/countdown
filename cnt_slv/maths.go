package cntSlv

import (
	"fmt"
	"log"
)

// maths.go contains the functions that actually do the maths on a pair of numbers
// Trivial I know, but we put effort into doing this minimising load on the rest of the
// system

func (foundValues *NumMap) make2To1(list NumCol) NumCol {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	if list.Len() != 2 {
		log.Println(list)
		log.Fatal("Invalid make2 list length")
	}
	var retList NumCol

	// Now grab the memory
	retList = make(NumCol, 4)
	for i := range retList {
		retList[i] = new(Number) // OPtimise to not make all of them
	}

	currentNumberLoc := 0
	_ = doCalcOn2(retList, list, currentNumberLoc)

	return retList
}
func (foundValues *NumMap) make3To1(list NumCol) NumCol {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	if foundValues.SelfTest && list.Len() != 3 {
		log.Println(list)
		log.Fatal("Invalid make3 list length")
	}
	var retList NumCol

	// Now grab the memory
	retList = make([]*Number, 4*3+4*4*3)
	// for i := range retList {
	// 	retList[i] = new(Number) // OPtimise to not make all of them
	// }
	currentNumberLoc := 0
	retList, _ = foundValues.make3To1bones(list, retList, currentNumberLoc)

	return retList
}
func (foundValues *NumMap) make3To1bones(list NumCol, retList []*Number, currentNumberLoc int) (NumCol, int) {
	for _, v := range list {
		if v == nil {
			log.Println("Odd")
		}
	}

	currentNumberLoc = doCalcOn2(retList, list[0:2], currentNumberLoc)
	currentNumberLoc = doCalcOn2(retList, list[1:3], currentNumberLoc)
	currentNumberLoc = doCalcOn2(retList, NumCol{list[0], list[2]}, currentNumberLoc)
	for j := 0; j < 3; j++ {
		for i := 0; i < 4; i++ {
			firstPart := (j * 4) + i
			secondPart := (j + 2) % 3
			if foundValues.SelfTest {
				if secondPart > 2 {
					log.Println("WTF", secondPart, i, j)
				}
				if firstPart > 11 {
					log.Println("WTF", firstPart, i, j)
				}
			}
			if retList[firstPart] == nil {
				continue
			}
			if retList[secondPart] == nil {
				continue
			}
			currentNumberLoc = doCalcOn2(retList, NumCol{retList[firstPart], list[secondPart]}, currentNumberLoc)
		}
	}
	if foundValues.SelfTest {
		for i, v := range retList {
			if v == nil {
				continue
			}
			if v.Val == 0 {
				fmt.Println("Returning 0 at index", i)
			}
		}
	}
	return retList, currentNumberLoc
}
func doCalcOn2(retList []*Number, list NumCol, currentNumberLoc int) int {
	a := list[0].Val
	b := list[1].Val
	initCurrentNumberLoc := currentNumberLoc
	if a == 0 || b == 0 {
		return currentNumberLoc
	}
	aGtB := a > b
	aEqB := a == b
	amb0 := ((a % b) == 0)
	bma0 := ((b % a) == 0)
	configure := func(inputA int, inputB []*Number, operation string, difficult int) *Number {
		nm := new(Number)
		nm.configure(inputA, inputB, operation, difficult)
		return nm
	}

	retList[currentNumberLoc] = configure(a+b, list, "+", 1)
	currentNumberLoc++
	retList[currentNumberLoc] = configure(a*b, list, "*", 2)
	currentNumberLoc++
	if aEqB {
		retList[currentNumberLoc] = nil
	} else if aGtB {
		retList[currentNumberLoc] = configure(a-b, list, "-", 1)
	} else {
		retList[currentNumberLoc] = configure(b-a, list, "--", 1)
	}
	currentNumberLoc++

	if aGtB {
		if amb0 {
			retList[currentNumberLoc] = configure(a/b, list, "/", 3)
		} else {
			retList[currentNumberLoc] = nil
		}
	} else {
		if bma0 {
			retList[currentNumberLoc] = configure(b/a, list, "\\", 3)
		} else {
			retList[currentNumberLoc] = nil
		}
	}
	currentNumberLoc++

	for i := initCurrentNumberLoc; i < currentNumberLoc; i++ {
		if retList[i] == nil {
			continue
		}
		if retList[i].Val == 0 {
			fmt.Println("doCalcOn2 fail at:", i)
		}
	}
	return currentNumberLoc
}
