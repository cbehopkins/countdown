package cntSlv

import (
	"sort"
	"strconv"
)

// NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
type NumCol []*Number

func (nc NumCol) String() string {
	var retStr string
	comma := ""
	retStr = ""
	tmpArray := make([]int, len(nc))

	// Get the numbers, then sort them, then print them
	for i, v := range nc {
		tmpArray[i] = v.Val
	}
	sort.Ints(tmpArray)
	for _, v := range tmpArray {
		retStr = retStr + comma + strconv.Itoa(v)
		comma = ","
	}
	return retStr
}

// NewNumCol Create a new number collection
func (nm *NumMap) NewNumCol(input []int) NumCol {
	var list NumCol
	var emptyList NumCol

	for _, v := range input {
		a := NewNumber(v, emptyList, "I", 0)
		nm.Add(a)
		list = append(list, a)
	}
	return list
}

// AddNum This function is used to add the initial numbers
func (nc *NumCol) AddNum(inputNum int, foundValues *NumMap) {
	var emptyList NumCol

	a := NewNumber(inputNum, emptyList, "I", 0)
	foundValues.Add(a)
	*nc = append(*nc, a)
}

// Len is the number of items in the number collection
func (nc NumCol) Len() int {
	var arrayLen int
	arrayLen = len(nc)
	return arrayLen
}

// Equal returns if both are equal
func (nc NumCol) Equal(i1 NumCol) bool {
	if len(nc) != len(i1) {
		return false
	}
	for i := range nc {
		i0Val := nc[i].Val
		i1Val := i1[i].Val
		if i0Val != i1Val {
			return false
		}
	}
	return true
}

// Tidy up the list
func (nc NumCol) Tidy() {
	for _, v := range nc {
		v.TidyDoubles()
		v.tidyOperators()
		v.ProveSol() // Just in case the Tidy functions have got things wrong
	}
}
