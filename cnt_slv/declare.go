package cntSlv

import (
	//	"fmt"
	"sort"
	"strconv"
)

// declare.go is responsible for declaring the interesting structures
// That is the NumCol and SolLst
// the types that contain our Sets of candidates and lists of possible solutions
// So NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
// A solution List is a number of things you do with a set of numbers
type NumCol []*Number
type SolLst []NumCol

func (foundValues *NumMap) CountHelper(target int, sources []int) chan SolLst {

	// Create a list of the input sources
	srcNumbers := foundValues.NewNumCol(sources)
	foundValues.SetTarget(target)

	return permuteN(srcNumbers, foundValues)
}

func (foundValues *NumMap) NewNumCol(input []int) NumCol {
	var list NumCol
	var emptyList NumCol

	for _, v := range input {
		a := NewNumber(v, emptyList, "I", 0)
		foundValues.Add(v, a)
		list = append(list, a)
	}
	return list
}

// This function is used to add the initial numbers
func (bob *NumCol) AddNum(inputNum int, foundValues *NumMap) {
	var emptyList NumCol

	a := NewNumber(inputNum, emptyList, "I", 0)
	foundValues.Add(inputNum, a)
	*bob = append(*bob, a)

}

func (item *SolLst) RemoveDuplicates() {
	// The purpose of this is to go through the supplied list
	// and modify the list to only include unique sets
	// any sets that produce the same string are considered identical
	// that is the collection contains the same values
	if false {
		solMap := make(map[string]NumCol)
		var delQueue []int
		for i := 0; i < len(*item); i++ {
			var v NumCol
			var t SolLst
			t = *item
			v = t[i]
			str := v.String()

			_, ok := solMap[str]
			if !ok {
				//fmt.Println("Added ", v)
				solMap[str] = v
			} else {
				//fmt.Printf("%s already exists\n", string)
				//pretty.Println(t1)
				//fmt.Printf("It is now, %d", i);
				//pretty.Println(t0);
				delQueue = append(delQueue, i)
			}
		}

		for i := len(delQueue); i > 0; i-- {
			//fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
			v := delQueue[i-1]
			//fmt.Println("You've asked to delete",v);
			l1 := *item
			*item = append(l1[:v], l1[v+1:]...)
		}
		//fmt.Printf("In Check, OrigLen %d, New Len %d\n",orig_len,len(*item))
	}
}
func (item NumCol) TidyNumCol() {
	for _, v := range item {
		//v.ProveSol()
		v.TidyDoubles()
		v.TidyOperators()
		v.ProveSol() // Just in case the Tidy functions have got things wrong
	}
}
func (item SolLst) TidySolLst() {
	for _, v := range item {
		v.TidyNumCol()
	}
}

func (item NumCol) String() string {
	var retStr string
	comma := ""
	retStr = ""
	tmpArray := make([]int, len(item))

	// Get the numbers, then sort them, then print them
	for i, v := range item {
		tmpArray[i] = v.Val
	}
	sort.Ints(tmpArray)
	for _, v := range tmpArray {
		//ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		retStr = retStr + comma + strconv.Itoa(v)
		comma = ","
	}
	return retStr
}
func (proofList SolLst) String() string {
	var retVal string
	if len(proofList) > 0 {
		for _, v := range proofList {
			// v is *NumCol
			for _, w := range v {
				// w is *Number
				var Value int
				Value = w.Val

				//ret_val = ret_val + fmt.Sprintf("Value %3d, = ", Value) + w.String() + "\n"
				retVal = retVal + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
		retVal = retVal + "Done printing proofs\n"
	} else {
		retVal = "No proofs found"
	}
	return retVal
}
func (proofList SolLst) StringNum(val int) string {
	var retVal string
	for _, v := range proofList {
		for _, w := range v {
			// w is *Number
			var Value int
			Value = w.Val
			if Value == val {
				retVal = retVal + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
	}
	return retVal
}
func (proofList SolLst) Exists(val int) bool {

	for _, v := range proofList {
		for _, w := range v {
			// w is *Number
			var Value int
			Value = w.Val
			if Value == val {
				return true
			}
		}
	}
	return false
}
func (bob NumCol) Len() int {
	var arrayLen int
	arrayLen = len(bob)
	return arrayLen
}
func (i0 NumCol) Equal(i1 NumCol) bool {
	if len(i0) != len(i1) {
		return false
	}
	for i := range i0 {
		i0Val := i0[i].Val
		i1Val := i1[i].Val
		if i0Val != i1Val {
			return false
		}
	}
	return true
}
func (bob SolLst) Len() int {
	return len(bob)
}
