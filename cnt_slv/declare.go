package cntSlv

import (
	"sort"
	"strconv"
)

// declare.go is responsible for declaring the interesting structures
// That is the NumCol and SolLst
// the types that contain our Sets of candidates and lists of possible solutions

// NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
type NumCol []*Number

// SolLst A solution List is a number of things you do with a set of numbers
type SolLst []NumCol

// Values is all the integher values in a number colleciton
func (nc NumCol) Values() []int {
	retInts := make([]int, len(nc))
	for i, v := range nc {
		retInts[i] = v.Val
	}
	return retInts
}

// CountHelper an exportable funciton to help externals work with us
func (nm *NumMap) CountHelper(target int, sources []int) chan SolLst {

	// Create a list of the input sources
	srcNumbers := nm.NewNumCol(sources)
	nm.SetTarget(target)

	return permuteN(srcNumbers, nm)
}

// NewNumCol Create a new number collection
func (nm *NumMap) NewNumCol(input []int) NumCol {
	var list NumCol
	var emptyList NumCol

	for _, v := range input {
		a := NewNumber(v, emptyList, "I", 0)
		nm.Add(v, a)
		list = append(list, a)
	}
	return list
}

// AddNum This function is used to add the initial numbers
func (nc *NumCol) AddNum(inputNum int, foundValues *NumMap) {
	var emptyList NumCol

	a := NewNumber(inputNum, emptyList, "I", 0)
	foundValues.Add(inputNum, a)
	*nc = append(*nc, a)

}

// RemoveDuplicates from the list
func (sl *SolLst) RemoveDuplicates() {
	// The purpose of this is to go through the supplied list
	// and modify the list to only include unique sets
	// any sets that produce the same string are considered identical
	// that is the collection contains the same values
	if false {
		solMap := make(map[string]NumCol)
		var delQueue []int
		for i := 0; i < len(*sl); i++ {
			var v NumCol
			var t SolLst
			t = *sl
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
			l1 := *sl
			*sl = append(l1[:v], l1[v+1:]...)
		}
		//fmt.Printf("In Check, OrigLen %d, New Len %d\n",orig_len,len(*item))
	}
}

// Tidy up the list
func (nc NumCol) Tidy() {
	for _, v := range nc {
		//v.ProveSol()
		v.TidyDoubles()
		v.tidyOperators()
		v.ProveSol() // Just in case the Tidy functions have got things wrong
	}
}

// Tidy up the list
func (sl SolLst) Tidy() {
	for _, v := range sl {
		v.Tidy()
	}
}

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
		//ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		retStr = retStr + comma + strconv.Itoa(v)
		comma = ","
	}
	return retStr
}
func (sl SolLst) String() string {
	var retVal string
	if len(sl) > 0 {
		for _, v := range sl {
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

// StringNum return the string for the supplied number
func (sl SolLst) StringNum(val int) string {
	var retVal string
	for _, v := range sl {
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

// Exists Does a value exist in the solution
func (sl SolLst) Exists(val int) bool {

	for _, v := range sl {
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

// Len of the count of number of solutions
func (sl SolLst) Len() int {
	return len(sl)
}
