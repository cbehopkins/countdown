package cnt_slv

import (
	"fmt"
)

var lone_map = true

type Number struct {
	// A number consists of
	Val        int       // a value
	list       []*Number // a pointer the the list of numbers used to obtain this
	operation  string    // The operation used on those numbers to get here
	difficulty int
}

func (i *Number) ProofLen() int {
	var cumlen int
	if i.list == nil {
		cumlen = 1
	} else {
		l0 := i.list[0].ProofLen()
		l1 := i.list[1].ProofLen()

		cumlen = l0 + l1
	}
	return cumlen
}
func (i *Number) ProveIt() string {
	var proof string
	var val int
	val = i.Val
	if i.list == nil {
		proof = fmt.Sprintf("%d", val)
	} else {
		p0 := i.list[0].ProveIt()
		p1 := i.list[1].ProveIt()
		operation := i.operation
		switch operation {
		case "--":
			proof = fmt.Sprintf("(%s-%s)", p1, p0)
		case "\\":
			proof = fmt.Sprintf("(%s/%s)", p1, p0)
		default:
			proof = fmt.Sprintf("(%s%s%s)", p0, operation, p1)
		}
	}
	return proof
}

// So NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
// A solution List is a number of things you do with a set of numbers
type NumCol []*Number
type SolLst []*NumCol

func (item NumCol) TestNum(to_test int) bool {
	for _, v := range item {
		value := v.Val
		if value == to_test {
			return true
		}
	}
	return false
}

func (item NumCol) GetNumCol() string {
	var ret_str string
	comma := ""
	ret_str = ""
	for _, v := range item {
		//ret_str = fmt.Sprintf("%s%s%d", ret_str, comma, v.Val)
		ret_str = ret_str + comma + fmt.Sprintf("%d", v.Val)
		comma = ","
	}
	return ret_str
}
func (bob *NumCol) AddNum(input_num int, found_values *NumMap) {
	var empty_list NumCol

	a := new_number(input_num, empty_list, "I", found_values, 0)
	*bob = append(*bob, a)

}
func (bob NumCol) Len() int {
	var array_len int
	array_len = len(bob)
	return array_len
}
func (item SolLst) CheckDuplicates() {
	sol_map := make(map[string]NumCol)
	var del_queue []int

	for i := 0; i < len(item); i++ {
		var v NumCol
		v = *item[i]
		string := v.GetNumCol()

		_, ok := sol_map[string]
		if !ok {
			//fmt.Println("Added ", v)
			sol_map[string] = v
		} else {
			//fmt.Printf("%s already exists, Length %d\n:", string,len(tpp));
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
		l1 := item
		item = append(l1[:v], l1[v+1:]...)
	}

}

func make_2_to_1(list []*Number, found_values *NumMap) []*Number {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	var ret_list []*Number
	var plus_num *Number
	//var mult_num *Number
	//var minu_num *Number

	a := list[0].Val
	b := list[1].Val
	// The thing that slows us down isn't calculations, but channel communications of generating new numbers
	// and garbage collecting the pointless ones
	// So it's worth spending some CPU working out the useless calculations
	
	a1 := (a==1)
	b1 := (b==1)
	a_gt_b := (a>b)
	// If a-b=a then no pount calculating a-b
	var no_sub bool
	if a_gt_b {
		no_sub = (a-b==a)
	} else {
		no_sub = (b-a==b)
	}
	

	ret_list = make([]*Number, 1, 4)
	plus_num = new_number(a+b, list, "+", found_values, 1)
	ret_list[0] = plus_num

	if found_values.UseMult {
		mult_num := new_number(a*b, list, "*", found_values, 2)
		ret_list = append(ret_list, mult_num)
	}

	if a_gt_b {
		if !no_sub {
			minu_num := new_number(a-b, list, "-", found_values, 1)
			ret_list = append(ret_list,minu_num)
		}
		if (b > 0) && (!b1) && ((a % b) == 0) {
			tmp_div := new_number((a / b), list, "/", found_values, 3)
			ret_list = append(ret_list, tmp_div)
		}
	} else {
		if !no_sub {
			minu_num := new_number(b-a, list, "--", found_values, 1)
			ret_list = append(ret_list,minu_num)
		}
		if (a > 0) && (!a1) && ((b % a) == 0) {
			tmp_div := new_number((b / a), list, "\\", found_values, 3)
			ret_list = append(ret_list, tmp_div)
		}
	}
	//found_values.AddMany(ret_list...)
	//fmt.Printf("Values are: %d,%d\n",plus_num.Val,minu_num.Val)
	return ret_list
}

func new_number(input_a int, input_b []*Number, operation string, found_values *NumMap, difficult int) *Number {

	var new_num Number
	//new_num = <-found_values.num_struct_queue
	new_num.Val = input_a
	//found_values.AddMany(&new_num)

	new_num.list = input_b
	new_num.operation = operation
	if len(input_b) > 1 {
		new_num.difficulty = input_b[0].difficulty + input_b[1].difficulty + difficult
	} else {
		new_num.difficulty = difficult
	}
	//fmt.Printf("There are %d elements in the input_a list\n", len(input_a.list))
	return &new_num
}
