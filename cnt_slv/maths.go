package cnt_slv

import (
	"fmt"
	"github.com/tonnerre/golang-pretty"
	"log"
)

// maths.go contains the functions that actually do the maths on a pair of numbers
// Trivial I know, but we put effort into doing this minimising load on the rest of the
// system

func (found_values *NumMap) do_maths(list []*Number) (num_to_make int,
	add_set, mul_set, sub_set, div_set, a_gt_b bool) {
	// The thing that slows us down isn't calculations, but channel communications of generating new numbers
	// allocating memory for new numbers and garbage collecting the pointless old ones
	// So it's worth spending some CPU working out the useless calculations
	// And working out exactly what dimension of structure we need to generate

	a := list[0].Val
	b := list[1].Val
	a0 := a <= 0
	b0 := b <= 0
	a_gt_b = (a > b)

	a1 := (a == 1)
	b1 := (b == 1)

	if a0 || b0 {
		log.Fatal("We got 0 as an input to do_maths - who is feeding us rubbish??")
	}

	add_set = true
	mul_set = found_values.UseMult
	num_to_make = 1
	if mul_set {
		if (a * b) > 0 {
			num_to_make = 2
		} else {
			mul_set = false
		}

	}
	if a_gt_b {
		sub_res_amb := a - b
		amb0 := ((a % b) == 0)
		if (sub_res_amb != a) && (sub_res_amb != 0) {
			sub_set = true
			num_to_make++
		}
		if !b1 && amb0 {
			div_set = true
			num_to_make++
		}
	} else {
		sub_res_bma := b - a
		bma0 := ((b % a) == 0)
		if (sub_res_bma != b) && (sub_res_bma != 0) {
			sub_set = true
			num_to_make++
		}
		if !a1 && bma0 {
			div_set = true
			num_to_make++
		}
	}
	return
}

func (found_values *NumMap) AddItems(list []*Number, ret_list []*Number, current_number_loc int,
	add_set, mul_set, sub_set, div_set, a_gt_b bool) {
	a := list[0].Val
	b := list[1].Val
	saved_current_number_loc := current_number_loc
	if add_set {
		ret_list[current_number_loc].configure(a+b, list, "+", 1)
		current_number_loc++
	}

	if sub_set {
		if a_gt_b {
			ret_list[current_number_loc].configure(a-b, list, "-", 1)
		} else {
			ret_list[current_number_loc].configure(b-a, list, "--", 1)
		}
		current_number_loc++
	}
	if mul_set {
		ret_list[current_number_loc].configure(a*b, list, "*", 2)
		current_number_loc++
	}
	if div_set {
		if a_gt_b {
			ret_list[current_number_loc].configure(a/b, list, "/", 3)
		} else {
			ret_list[current_number_loc].configure(b/a, list, "\\", 3)
		}
		current_number_loc++
	}
	for i := saved_current_number_loc; i < current_number_loc; i++ {
		v := ret_list[i]
		if v.Val <= 0 {
			pretty.Println(v)
			fmt.Printf("value %d is %d, %d, %d\n", i, v.Val, a, b)
			fmt.Printf("add_set=%t, mul_set=%t, sub_set=%t, div_set=%t, a_gt_b=%t\n", add_set, mul_set, sub_set, div_set, a_gt_b)
			for j := saved_current_number_loc; j < current_number_loc; j++ {
				fmt.Printf("Val: %d\n", ret_list[j].Val)
			}
			log.Fatal("result <0")
		}
	}
}
func (found_values *NumMap) make_2_to_1(list NumCol) NumCol {
	// This is (conceptually) returning a list of numbers
	// That can be generated from 2 input numbers
	// organised in such a way that we know how we created them
	if list.Len() != 2 {
		pretty.Print(list)
		log.Fatal("Invalid make2 list length")
	}
	var ret_list NumCol
	num_to_make,
		add_set, mul_set, sub_set, div_set,
		a_gt_b := found_values.do_maths(list)

	// Now Grab the memory
	//ret_list = found_values.aquire_numbers(num_to_make)
	ret_list = make([]*Number, num_to_make)
	for i, _ := range ret_list {
		ret_list[i] = new(Number)
	}

	current_number_loc := 0
	found_values.AddItems(list, ret_list, current_number_loc,
		add_set, mul_set, sub_set, div_set,
		a_gt_b)

	return ret_list
}
