package cnt_slv

import (
	"fmt"
	"log"
	"sort"

	"github.com/tonnerre/golang-pretty"
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
func (i *Number) TidyDoubles() {
	// Remove any double notation in a proof
	// we use   our own special notation to make things easier for ourselves
	// However it's better to remove it at the tidy stage
	// To make reducing the proof sizes easier
	// Here's what our operands say
	// a-b == b--a
	// a/b == b\\a

	if (i.list == nil) || (len(i.list) == 0) {
		return
	}

	for _, v := range i.list {
		v.TidyDoubles()
	}

	if i.operation == "--" {
		if len(i.list) != 2 {
			log.Fatal("can't process -- on a list that is anything but 2 long")
		}

		i.operation = "-"
	} else if i.operation == "\\" {
		if len(i.list) != 2 {
			log.Fatal("can't process \\ on a list that is anything but 2 long")
		}

		i.operation = "/"
	} else {
		// Must not be a double operator
		return
	}
	// CBH get this from the central allocator
	temp_array := make([]*Number, 2)
	temp_array[0] = i.list[1]
	temp_array[1] = i.list[0]
	i.list = temp_array
	return
}
func (i *Number) TidyOperators() {
	// This one is sexy
	// we often in our proofs get things like:
	// (((1+2)+3)+(4/2)) or
	// (((8-2)-1)-2)
	// Which could of course both be simplified
	// So what we will do is re-write the tree structure of our proofs
	// Things are easy with + as we can just descend the tree and if the next level down uses a + as well
	// Then we can just combine them

	// Think about the use case:
	// ((1+2)+(3+4))
	// We will first read (3+7)
	// Look at the 3 and see how we got it.
	// we will see (1+2) uses the same operator
	// so we can pull that into ours
	// The same applies to multiples

	// When it comes to subtract and divide we have an issue
	// ((a-b)-c)-d == (a-(b+c+d)) <- much tidier
	// so let's look at: (a-b)-c as a starting point
	// Actually represented as something like:
	// g-c and we look at g and find it is a-b
	// but we could say that:
	// * if we are a subtract and the (first) leaf is a subtract
	// * Create a new number that is the leaf's second number + our second number
	// * Set our first number to the leaf's First number
	// * Set our second number to the new number we just made
	// Likewise for: a-(b-c) -> (a+c)-b
	// * if we are a subtract and the (second) leaf is a subtract
	// * Create a new number that is the leaf's second number + our First Number
	// * Set our first number to the  new number we just made
	// * Set our second number to the leaf's First number
	// Now we could get clever for things like merging addition if come the the
	// numbers in out subtraction turned into an addition, or we could just run ourselves
	// on that new number whch will merge up any additions

	// But of course the first thing we want is for their house to be in order
	tmp_list := make([]*Number, 0, 4) // CBH get this from the centra allocator
	list_modified := false
	for _, v := range i.list {
		v.TidyOperators()
		// Let's just combine +s for now
		if (i.operation == "+") && (v.operation == "+") {
			i.difficulty = i.difficulty + v.difficulty
			tmp_list = append(tmp_list, v.list...)
			list_modified = true
		} else if (i.operation == "*") && (v.operation == "*") {
			tmp_list = append(tmp_list, v.list...)
			i.difficulty = i.difficulty + v.difficulty
			list_modified = true
		} else {
			tmp_list = append(tmp_list, v)
		}
	}
	if list_modified {
		i.list = tmp_list
	}

	if (i.operation == "-") && (len(i.list) == 2) {
		// Play it safe and check first, work out optimisation later
		if (i.list[0].operation == "-") && (i.list[1].operation == "-") {
			// Fill in this later optimisaton
			// basically turn (a-b)-(c-d) -> (a+d)-(b+c)
		} else if i.list[0].operation == "-" {
			// Transform (a-b)-c -> a-(b+c)
			// in this terminology
			// a = i.list[0].list[0]
			// b = i.list[0].list[1]
			// c = i.list[1]
			// create b+c
			my_list0 := make([]*Number, 2)
			my_list0[0] = i.list[1]
			my_list0[1] = i.list[0].list[1]

			b_plus_c := new_number((i.list[1].Val + i.list[0].list[1].Val), my_list0, "+", (i.list[1].difficulty + i.list[0].list[1].difficulty + 1))

			my_list1 := make([]*Number, 2)
			my_list1[0] = i.list[0].list[0]
			my_list1[1] = b_plus_c
			new_num := new_number(i.Val, my_list1, "-", (b_plus_c.difficulty + i.list[0].list[0].difficulty + 1))
			i = new_num
			//i.TidyOperators()
			i.ProveSol() //CBH we've made serious modification so test it
		} else if i.list[1].operation == "-" {
			// Transform a-(b-c) -> (a+c)-b
			// in this terminology
			// a = i.list[0]
			// b = i.list[1].list[0]
			// c = i.list[1].list[1]

			// create a+c
			my_list0 := make(NumCol, 2)
			my_list0[0] = i.list[0]
			my_list0[1] = i.list[1].list[1]
			a_plus_c := new_number((my_list0[0].Val + my_list0[1].Val), my_list0, "+", (my_list0[0].difficulty + my_list0[1].difficulty + 1))

			my_list1 := make(NumCol, 2)
			my_list1[0] = a_plus_c
			my_list1[1] = i.list[1].list[0]
			new_num := new_number(i.Val, my_list1, "-", (a_plus_c.difficulty + my_list1[1].difficulty + 1))

			i = new_num
			//i.TidyOperators()
			i.ProveSol()
		}
	}

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

func (i *Number) ProveSol() int {
	// This function should go through the list and prove the solution
	// Also do other sanity checking like the ,/- operators only have 2 items in the list
	// That anything with a valid operator has >1 item in the list
	running_total := 0
	first_run := true
	if (i.list == nil) || (len(i.list) == 0) {
		// This is a source value
		return i.Val
	} else if len(i.list) == 1 {
		pretty.Print(i)
		log.Fatal("Error invalid list length")
		return 0
	} else {
		for _, v := range i.list {
			if first_run {
				//pretty.Print(v)
				first_run = false
				running_total = v.ProveSol()
			} else {
				switch i.operation {
				case "+":
					running_total = running_total + v.ProveSol()
				case "-":
					running_total = running_total - v.ProveSol()
				case "--":
					running_total = v.ProveSol() - running_total
				case "*":
					running_total = running_total * v.ProveSol()
				case "/":
					running_total = running_total / v.ProveSol()
				case "\\":
					running_total = v.ProveSol() / running_total
				default:
					log.Fatal("Unknown operation type")
				}
			}
		}
		if running_total != i.Val {
			pretty.Println(i)

			fmt.Println("We calculated ", running_total, i.ProveIt())
			log.Fatal("Failed to self check solution")
		}
		return running_total
	}
}

// CBH rename this function to String()
func (i *Number) ProveIt() string {
	var proof string
	var val int
	val = i.Val
	//pretty.Print(i)
	if i.list == nil {
		proof = fmt.Sprintf("%d", val)
	} else {
		proof = ""
		op := ""
		for _, v := range i.list {

			switch i.operation {
			case "--":

				proof = v.ProveIt() + op + proof
				op = "-"
				//proof = proof + "--" + v.ProveIt()
			case "\\":
				proof = v.ProveIt() + op + proof
				op = "/"
				//proof = proof + "//" + v.ProveIt()
			default:
				proof = proof + op + v.ProveIt()
				//proof = v.ProveIt() + op + proof
				op = i.operation
			}

		}
		proof = "(" + proof + ")"

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
	tmp_array := make([]int, len(item))

	// Get the numbers, then sort them, then print theM
	for i, v := range item {
		tmp_array[i] = v.Val
	}
	//fmt.Println("Before:",tmp_array)
	sort.Ints(tmp_array)
	//fmt.Println("After:",tmp_array)
	for _, v := range tmp_array {
		//ret_str = fmt.Sprintf("%s%s%d", ret_str, comma, v.Val)
		ret_str = ret_str + comma + fmt.Sprintf("%d", v)
		comma = ","
	}
	return ret_str
}
func (bob *NumCol) AddNum(input_num int, found_values *NumMap) {
	var empty_list NumCol

	a := new_number(input_num, empty_list, "I", 0)
	found_values.Add(input_num, a)
	*bob = append(*bob, a)

}
func (bob NumCol) Len() int {
	var array_len int
	array_len = len(bob)
	return array_len
}
func (item *SolLst) CheckDuplicates() {
	return
	sol_map := make(map[string]NumCol)
	var del_queue []int
	for i := 0; i < len(*item); i++ {
		var v NumCol
		var t SolLst
		t = *item
		v = *t[i]
		string := v.GetNumCol()

		_, ok := sol_map[string]
		if !ok {
			//fmt.Println("Added ", v)
			sol_map[string] = v
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
func (found_values *NumMap) DoMaths(list []*Number) (num_to_make,
	add_res, mul_res, sub_res, div_res int,
	add_set, mul_set, sub_set, div_set, a_gt_b bool) {

	a := list[0].Val
	b := list[1].Val
	add_res = a + b
	add_set = true
	mul_set = found_values.UseMult
	// The thing that slows us down isn't calculations, but channel communications of generating new numbers
	// allocating memory for new numbers and garbage collecting the pointless old ones
	// So it's worth spending some CPU working out the useless calculations
	// And working out exactly what dimension of structure we need to generate

	a1 := (a == 1)
	b1 := (b == 1)
	a_gt_b = (a > b)

	num_to_make = 1

	if mul_set {
		mul_res = a * b
		num_to_make++
	}

	if a_gt_b {
		sub_res = a - b
		if (sub_res != a) && (sub_res != 0) {
			sub_set = true
			num_to_make++
		}
		if (b > 0) && (!b1) && ((a % b) == 0) {
			div_set = true
			div_res = a / b
			num_to_make++
		}
	} else {
		sub_res = b - a
		if (b-a != b) && (sub_res != 0) {
			sub_set = true
			num_to_make++
		}
		if (a > 0) && (!a1) && ((b % a) == 0) {
			div_set = true
			div_res = b / a
			num_to_make++
		}
	}
	return
}

func (found_values *NumMap) AddItems(list []*Number, ret_list []*Number, current_number_loc int,
	add_res, mul_res, sub_res, div_res int,
	add_set, mul_set, sub_set, div_set, a_gt_b bool) {
	if add_set {
		ret_list[current_number_loc].configure(add_res, list, "+", 1)
		current_number_loc++
	}

	if sub_set {
		if a_gt_b {
			ret_list[current_number_loc].configure(sub_res, list, "-", 1)
		} else {
			ret_list[current_number_loc].configure(sub_res, list, "--", 1)
		}
		current_number_loc++
	}
	if mul_set {
		ret_list[current_number_loc].configure(mul_res, list, "*", 2)
		current_number_loc++
	}
	if div_set {
		if a_gt_b {
			ret_list[current_number_loc].configure(div_res, list, "/", 3)
		} else {
			ret_list[current_number_loc].configure(div_res, list, "\\", 3)
		}
		current_number_loc++
	}
}
func make_2_to_1(list []*Number, found_values *NumMap) NumCol {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	if len(list) != 2 {
		pretty.Print(list)
		log.Fatal("Invalid make2 list length")
	}
	var ret_list NumCol
	num_to_make,
		add_res, mul_res, sub_res, div_res,
		add_set, mul_set, sub_set, div_set,
		a_gt_b := found_values.DoMaths(list)

	// Now Grab the memory
	ret_list = found_values.aquire_numbers(num_to_make)

	current_number_loc := 0
	found_values.AddItems(list, ret_list, current_number_loc,
		add_res, mul_res, sub_res, div_res,
		add_set, mul_set, sub_set, div_set,
		a_gt_b)

	return ret_list
}

func (num *Number) configure(input_a int, input_b []*Number, operation string, difficult int) {
	num.Val = input_a

	num.list = input_b
	num.operation = operation
	//if len(input_b) > 1 {
	//	num.difficulty = input_b[0].difficulty + input_b[1].difficulty + difficult
	//} else {
	num.difficulty = difficult
	//}
	for _,v:= range input_b {
		num.difficulty = num.difficulty + v.difficulty
	}

}
func new_number(input_a int, input_b []*Number, operation string, difficult int) *Number {
	var new_num Number
	new_num.configure(input_a, input_b, operation, difficult)
	return &new_num
}
