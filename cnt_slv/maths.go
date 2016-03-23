package cnt_slv

import (
	"log"

	"github.com/tonnerre/golang-pretty"
)

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
		log.Fatal("We got 0")
	}

	add_set = true
	mul_set = found_values.UseMult
	if mul_set {
		num_to_make = 2
	} else {
		num_to_make = 1 // add_set must be set to reach here
	}

	if a_gt_b {
		sub_res := a - b
		amb0 := ((a % b) == 0)
		if (sub_res != a) && (sub_res != 0) {
			sub_set = true
			num_to_make++
		}
		if !b1 && amb0 {
			div_set = true
			num_to_make++
		}
	} else {
		sub_res := b - a
		bma0 := ((b % a) == 0)
		if (b-a != b) && (sub_res != 0) {
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
	ret_list = found_values.aquire_numbers(num_to_make)

	current_number_loc := 0
	found_values.AddItems(list, ret_list, current_number_loc,
		add_set, mul_set, sub_set, div_set,
		a_gt_b)

	return ret_list
}
