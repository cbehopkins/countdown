package cnt_slv

import (
	"fmt"
	"sort"
)

// So NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
// A solution List is a number of things you do with a set of numbers
type NumCol []*Number
type SolLst []*NumCol

// This function is used to add the initial numbers
func (bob *NumCol) AddNum(input_num int, found_values *NumMap) {
	var empty_list NumCol

	a := NewNumber(input_num, empty_list, "I", 0)
	found_values.Add(input_num, a)
	*bob = append(*bob, a)

}

func (item *SolLst) RemoveDuplicates() {
	//return
	sol_map := make(map[string]NumCol)
	var del_queue []int
	for i := 0; i < len(*item); i++ {
		var v NumCol
		var t SolLst
		t = *item
		v = *t[i]
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
func (item NumCol) TidyNumCol() {
	for _, v := range item {
		v.ProveSol()
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
		ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		comma = ","
	}
	return ret_str
}
func (proof_list SolLst) String() string {
	var ret_val string
	for _, v := range proof_list {
		// v is *NumCol
		for _, w := range *v {
			// w is *Number
			var Value int
			Value = w.Val

			ret_val = ret_val + fmt.Sprintf("Value %3d, = ", Value) + w.String() + "\n"
		}
	}
	ret_val = ret_val + "Done printing proofs\n"
	return ret_val
}
func (bob NumCol) Len() int {
	var array_len int
	array_len = len(bob)
	return array_len
}
func (bob SolLst) Len() int {
	return len(bob)
}
