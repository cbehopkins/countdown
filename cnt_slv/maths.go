package cntSlv

import (
	"fmt"
	"log"
)

// maths.go contains the functions that actually do the maths on a pair of numbers
// Trivial I know, but we put effort into doing this minimising load on the rest of the
// system

func (foundValues *NumMap) doMaths(list []*Number) (numToMake int,
	addSet, mulSet, subSet, divSet, aGtB bool) {
	// The thing that slows us down isn't calculations, but channel communications of generating new numbers
	// allocating memory for new numbers and garbage collecting the pointless old ones
	// So it's worth spending some CPU working out the useless calculations
	// And working out exactly what dimension of structure we need to generate

	a := list[0].Val
	b := list[1].Val
	a0 := a <= 0
	b0 := b <= 0
	aGtB = (a > b)

	a1 := (a == 1)
	b1 := (b == 1)

	if a0 || b0 {
		log.Fatal("We got 0 as an input to do_maths - who is feeding us rubbish??")
	}

	addSet = true
	mulSet = foundValues.UseMult
	numToMake = 1
	if mulSet {
		if (a * b) > 0 {
			numToMake = 2
		} else {
			mulSet = false
		}

	}
	if aGtB {
		subResAmb := a - b
		amb0 := ((a % b) == 0)
		if (subResAmb != a) && (subResAmb != 0) {
			subSet = true
			numToMake++
		}
		if !b1 && amb0 {
			divSet = true
			numToMake++
		}
	} else {
		subResBma := b - a
		bma0 := ((b % a) == 0)
		if (subResBma != b) && (subResBma != 0) {
			subSet = true
			numToMake++
		}
		if !a1 && bma0 {
			divSet = true
			numToMake++
		}
	}
	return
}

// AddItems is used to add several items at once
// Used by the calculation functions to efficiently add a bunch of stuff
func (foundValues *NumMap) AddItems(list []*Number, retList []*Number, currentNumberLoc int,
	addSet, mulSet, subSet, divSet, aGtB bool) {
	a := list[0].Val
	b := list[1].Val
	savedCurrentNumberLoc := currentNumberLoc
	if addSet {
		retList[currentNumberLoc].configure(a+b, list, "+", 1)
		currentNumberLoc++
	}

	if subSet {
		if aGtB {
			retList[currentNumberLoc].configure(a-b, list, "-", 1)
		} else {
			retList[currentNumberLoc].configure(b-a, list, "--", 1)
		}
		currentNumberLoc++
	}
	if mulSet {
		retList[currentNumberLoc].configure(a*b, list, "*", 2)
		currentNumberLoc++
	}
	if divSet {
		if aGtB {
			retList[currentNumberLoc].configure(a/b, list, "/", 3)
		} else {
			retList[currentNumberLoc].configure(b/a, list, "\\", 3)
		}
		currentNumberLoc++
	}
	for i := savedCurrentNumberLoc; i < currentNumberLoc; i++ {
		v := retList[i]
		if v.Val <= 0 {
			fmt.Println(v)
			fmt.Printf("value %d is %d, %d, %d\n", i, v.Val, a, b)
			fmt.Printf("add_set=%t, mul_set=%t, sub_set=%t, div_set=%t, a_gt_b=%t\n", addSet, mulSet, subSet, divSet, aGtB)
			for j := savedCurrentNumberLoc; j < currentNumberLoc; j++ {
				fmt.Printf("Val: %d\n", retList[j].Val)
			}
			log.Fatal("result <0")
		}
	}
}
func (foundValues *NumMap) make2To1(list NumCol) NumCol {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	if list.Len() != 2 {
		log.Println(list)
		log.Fatal("Invalid make2 list length")
	}
	var retList NumCol
	numToMake,
		addSet, mulSet, subSet, divSet,
		aGtB := foundValues.doMaths(list)

	// Now grab the memory
	//ret_list = found_values.aquire_numbers(num_to_make)
	retList = make([]*Number, numToMake)
	for i := range retList {
		retList[i] = new(Number)
	}

	currentNumberLoc := 0
	foundValues.AddItems(list, retList, currentNumberLoc,
		addSet, mulSet, subSet, divSet,
		aGtB)

	return retList
}
func (foundValues *NumMap) make2To1mod(list NumCol) NumCol {
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
func (foundValues *NumMap) make3To1mod(list NumCol) NumCol {
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
	for i := range retList {
		retList[i] = new(Number) // OPtimise to not make all of them
	}
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

	return retList, currentNumberLoc
}
func doCalcOn2(retList []*Number, list NumCol, currentNumberLoc int) int {
	a := list[0].Val
	b := list[1].Val
	if a == 0 || b == 0 {
		return currentNumberLoc
	}
	aGtB := a > b
	amb0 := ((a % b) == 0)
	bma0 := ((b % a) == 0)

	retList[currentNumberLoc].configure(a+b, list, "+", 1)
	currentNumberLoc++
	retList[currentNumberLoc].configure(a*b, list, "*", 2)
	currentNumberLoc++
	if aGtB {
		retList[currentNumberLoc].configure(a-b, list, "-", 1)
	} else {
		retList[currentNumberLoc].configure(b-a, list, "--", 1)
	}
	currentNumberLoc++

	if aGtB {
		if !amb0 {
			retList[currentNumberLoc].configure(a/b, list, "/", 3)
		} else {
			retList[currentNumberLoc] = nil
		}
	} else {
		if !bma0 {
			retList[currentNumberLoc].configure(b/a, list, "\\", 3)
		} else {
			retList[currentNumberLoc] = nil
		}
	}
	currentNumberLoc++

	return currentNumberLoc
}
