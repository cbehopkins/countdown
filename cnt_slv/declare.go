package cnt_slv

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

func (found_values *NumMap) CountHelper(target int, sources []int) chan SolLst {

	// Create a list of the input sources
	srcNumbers := found_values.NewNumCol(sources)
	found_values.SetTarget(target)

	return permuteN(srcNumbers, found_values)
}

func (found_values *NumMap) NewNumCol(input []int) NumCol {
	var list NumCol
	var empty_list NumCol

	for _, v := range input {
		a := NewNumber(v, empty_list, "I", 0)
		found_values.Add(v, a)
		list = append(list, a)
	}
	return list
}

// This function is used to add the initial numbers
func (bob *NumCol) AddNum(input_num int, found_values *NumMap) {
	var empty_list NumCol

	a := NewNumber(input_num, empty_list, "I", 0)
	found_values.Add(input_num, a)
	*bob = append(*bob, a)

}

func (item *SolLst) RemoveDuplicates() {
	// The purpose of this is to go through the supplied list
	// and modify the list to only include unique sets
	// any sets that produce the same string are considered identical
	// that is the collection contains the same values
  if false {
	sol_map := make(map[string]NumCol)
	var del_queue []int
	for i := 0; i < len(*item); i++ {
		var v NumCol
		var t SolLst
		t = *item
		v = t[i]
		str := v.String()

		_, ok := sol_map[str]
		if !ok {
			//fmt.Println("Added ", v)
			sol_map[str] = v
		} else {
			//fmt.Printf("%s already exists\n", string)
			//pretty.Println(t1)
			//fmt.Printf("It is now, %d", i);
			//pretty.Println(t0);
			del_queue = append(del_queue, i)
		}
	}

	for i := len(del_queue); i > 0; i-- {
		//fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
		v := del_queue[i-1]
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
	var ret_str string
	comma := ""
	ret_str = ""
	tmp_array := make([]int, len(item))

	// Get the numbers, then sort them, then print them
	for i, v := range item {
		tmp_array[i] = v.Val
	}
	sort.Ints(tmp_array)
	for _, v := range tmp_array {
		//ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		ret_str = ret_str + comma + strconv.Itoa(v)
		comma = ","
	}
	return ret_str
}
func (proof_list SolLst) String() string {
	var ret_val string
	if len(proof_list) > 0 {
		for _, v := range proof_list {
			// v is *NumCol
			for _, w := range v {
				// w is *Number
				var Value int
				Value = w.Val

				//ret_val = ret_val + fmt.Sprintf("Value %3d, = ", Value) + w.String() + "\n"
				ret_val = ret_val + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
		ret_val = ret_val + "Done printing proofs\n"
	} else {
		ret_val = "No proofs found"
	}
	return ret_val
}
func (proof_list SolLst) StringNum(val int) string {
	var ret_val string
	for _, v := range proof_list {
		for _, w := range v {
			// w is *Number
			var Value int
			Value = w.Val
			if Value == val {
				ret_val = ret_val + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
	}
	return ret_val
}
func (proof_list SolLst) Exists(val int) bool {

	for _, v := range proof_list {
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
	var array_len int
	array_len = len(bob)
	return array_len
}
func (i0 NumCol) Equal(i1 NumCol) bool {
	if len(i0) != len(i1) {
		return false
	}
	for i, _ := range i0 {
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
