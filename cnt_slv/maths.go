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
	INPUT_LEN := 3
	if foundValues.SelfTest {
		if list.Len() != INPUT_LEN {
			log.Println(list)
			log.Fatal("Invalid make3 list length")
		}
		for _, v := range list {
			if v == nil {
				log.Println("Odd")
			}
		}
	}
	var retList NumCol
	OPERAND_RESULT_CNT := 4
	// Now grab the memory
	retList = make([]*Number, OPERAND_RESULT_CNT*3+OPERAND_RESULT_CNT*4*3)
	currentNumberLoc := 0

	// Pairwise combination of 3 numbers, requires 3 calculations
	currentNumberLoc = doCalcOn2(retList, list[0:2], currentNumberLoc)                // secondPart = 2
	currentNumberLoc = doCalcOn2(retList, list[1:3], currentNumberLoc)                // secondPart = 0
	currentNumberLoc = doCalcOn2(retList, NumCol{list[0], list[2]}, currentNumberLoc) // secondPart = 1
	for j := 0; j < INPUT_LEN; j++ {
		// The outer loop selects a value from the input list
		secondPart := (j + (INPUT_LEN - 1)) % INPUT_LEN

		for i := 0; i < OPERAND_RESULT_CNT; i++ {
			firstPart := (j * OPERAND_RESULT_CNT) + i
			if foundValues.SelfTest {
				if firstPart > ((OPERAND_RESULT_CNT * INPUT_LEN) - 1) {
					log.Println("WTF", firstPart, i, j)
				}
			}
			if retList[firstPart] == nil {
				continue
			}
			if list[secondPart] == nil {
				continue
			}
			// Work one of the results of input X input against one of the inpuits
			// The celver bit is making sure we don't allow secondPart to have been involved
			// in the calculation of firstPart
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
	return retList
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

	retList[currentNumberLoc] = NewNumber(a+b, list, "+", 1)
	currentNumberLoc++
	retList[currentNumberLoc] = NewNumber(a*b, list, "*", 2)
	currentNumberLoc++
	if aEqB {
		retList[currentNumberLoc] = nil
	} else if aGtB {
		retList[currentNumberLoc] = NewNumber(a-b, list, "-", 1)
	} else {
		retList[currentNumberLoc] = NewNumber(b-a, list, "--", 1)
	}
	currentNumberLoc++
	if aEqB {
		retList[currentNumberLoc] = NewNumber(1, list, "/", 1)
	} else if aGtB {
		if amb0 {
			retList[currentNumberLoc] = NewNumber(a/b, list, "/", 3)
		} else {
			retList[currentNumberLoc] = nil
		}
	} else {
		if bma0 {
			retList[currentNumberLoc] = NewNumber(b/a, list, "\\", 3)
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
