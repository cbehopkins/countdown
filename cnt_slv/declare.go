package cnt_slv

import (
	"fmt"
	"sort"
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
	tmp_array := make([]int,len(item) )

	// Get the numbers, then sort them, then print theM
	for i, v := range item {
		tmp_array[i] = v.Val
	}
	//fmt.Println("Before:",tmp_array)
	sort.Ints(tmp_array)
	//fmt.Println("After:",tmp_array)
	for _,v := range tmp_array {
		//ret_str = fmt.Sprintf("%s%s%d", ret_str, comma, v.Val)
		ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		comma = ","
	}
	return ret_str
}
func (bob *NumCol) AddNum(input_num int, found_values *NumMap) {
	var empty_list NumCol

	a := new_number(input_num, empty_list, "I", 0)
	found_values.Add(input_num,a)
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

	a := list[0].Val
	b := list[1].Val
	// The thing that slows us down isn't calculations, but channel communications of generating new numbers
	// allocating memory for new numbers and garbage collecting the pointless old ones
	// So it's worth spending some CPU working out the useless calculations
	// And working out exactly what dimension of structure we need to generate
	
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
	num_to_make := 1

	var mul_res int
	if found_values.UseMult {
		mul_res = a*b
		num_to_make++
	}

	var divd bool
	var div_res int
	var sub_res int
	if a_gt_b {
		if !no_sub {
			sub_res = a-b
			num_to_make++
		}
		if (b > 0) && (!b1) && ((a % b) == 0) {
			divd = true
			div_res = a/b
                	num_to_make++
		}
	} else {
		if !no_sub {
	                sub_res = b-a
			num_to_make++

		}
		if (a > 0) && (!a1) && ((b % a) == 0) {
			divd = true
			div_res = b/a
	                num_to_make++

		}
	}
	//fmt.Println("Calling")
	ret_list = found_values.aquire_numbers(num_to_make)
	//fmt.Println("This is what we got:")
	//for i,j := range ret_list {                                                                                                                                                                        
        //        fmt.Printf("Item %x Pointer %p\n", i,j)                                                                                                                                                   
        //}
	current_number_loc := 0
	ret_list[current_number_loc].configure(a+b, list, "+",  1) 
	current_number_loc++

	if !no_sub {
		ret_list[current_number_loc].configure(sub_res, list, "-",  1)
		current_number_loc++
	}
        if found_values.UseMult {
		ret_list[current_number_loc].configure(mul_res, list, "*",  2)
                current_number_loc++
	}
	if divd {
		if a_gt_b {
			ret_list[current_number_loc].configure(div_res, list, "/",  3)
		} else {
                        ret_list[current_number_loc].configure(div_res, list, "\\",  3)
		}
		//current_number_loc++
	}

	return ret_list
}
//func (nm *NumMap) aquire_numbers (num_to_make int) []*Number {
//        tmp_list := make([]Number,num_to_make,4)        // Always allow 4 for cache lines                                                            
//        ret_list := make([]*Number, num_to_make,4)                                                                                                                                                                            
//        for i,l := range tmp_list {                                                                                                                  
//                ret_list[i] = &l                                                                                                                                                                                             
//        }
//	return ret_list
//}
func (num *Number) configure (input_a int, input_b []*Number, operation string, difficult int) {
     	num.Val = input_a                                                                                                                        
                                                                                                                                                                                                                             
        num.list = input_b                                                                                                                       
        num.operation = operation                                                                                                                                                                                        
        if len(input_b) > 1 {                                                                                                                        
                num.difficulty = input_b[0].difficulty + input_b[1].difficulty + difficult                                                                                                                               
        } else {                                                                                                                                     
                num.difficulty = difficult
        }

}
func new_number(input_a int, input_b []*Number, operation string, difficult int) *Number {
	var new_num Number
	new_num.configure(input_a,input_b,operation,difficult)
	return &new_num
}
