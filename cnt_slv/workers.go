package cnt_slv

import (
"log"
)

// workers contains the worker functions
// That is the functions that work on the lists to turn them into
// number pairs to process
// This means taking sets of numbers and permuting them to come up with
// all possible combinations
// and then taking these combinations and devolving them into smaller sets
// that can be worked on in turn
const (
	LonMap = iota
	ParMap
	NetMap
)

func WorkN(array_in NumCol, found_values *NumMap) SolLst {
	for _, j := range array_in {
		if j.Val == 0 {
			log.Fatal("WorkN fed a 0 number")
		}
	}
	return work_n(array_in, found_values, false)
}

func work_n(array_in NumCol, found_values *NumMap, multipass bool) SolLst {
	var ret_list SolLst
	len_array_in := array_in.Len()
	if found_values.Solved() {
		return ret_list
	}
	if len_array_in == 1 {
		//ret_list = append(ret_list, &array_in)
		return SolLst{array_in}
	} else if len_array_in == 2 {
		var tmp_list NumCol
		tmp_list = found_values.make_2_to_1(array_in[0:2])
		found_values.AddMany(tmp_list...)
		ret_list = append(ret_list, tmp_list, array_in)
		return ret_list
	}

	// work_n takes
	// let's use work 3 as a first example {2,3,4} and should generate everything that can be done with these 3 numbers
	// Note: for these explanantions I'll assume we just add and subtract numbers
	// We do not return the supplied list with the return
	// we also do no permute the input numbers as we know that permute function will do this for us
	// So in this example we would look to do several steps first we feed to make_3
	// This will treat the input as {2,3),{4} it works the first list to get:
	// {5,1} (from 2+3 and 3-2) and therefore returns {{5,4}, {1,4}}
	// we then take each value in this list and work that to get {{9},{3}}
	// the final list we want to return is {{5,4}, {1,4}, {9},{3}}
	// the reason to not return {2,3,4} is so that in the grand scheme of things we can recurse these lists
	var work_list WrkLst
	work_list = NewWrkLst(array_in)
	// so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }

	if multipass {
		cross_len := 0
		num_numbers_to_make := 0

		determineSizeFunc := func(a_num, b_num *Number) bool {
			if a_num.Val <= 0 || b_num.Val <= 0 {
				log.Fatalf("Gimmie gave %d, %d", a_num.Val, b_num.Val)
			}
			tmp,
				_, _, _, _, _ := found_values.do_maths([]*Number{a_num, b_num})
			num_numbers_to_make += tmp
			cross_len++
			return true
		}

		work_list.procWork(found_values, determineSizeFunc)

		top_src_to_make := cross_len * 2
		top_numbers_to_make := num_numbers_to_make
		//current_item = 0
		var work_unit SolLst
		// Last Item on work list contains sources
		work_unit = work_list.Last()
		// Malloc the memory once!
		current_number_loc := 0
		// This is the list of numbers that calculations are done from
		src_list := found_values.aquire_numbers(top_src_to_make)
		// This is the list of numbers that will be used in the proof
		// i.e. the list that calculations results end up in
		num_list := found_values.aquire_numbers(top_numbers_to_make)
		// And this allocates the list that will point to those (previously allocated) numbers

		ret_list = make(SolLst, 0, (cross_len + len(work_unit)))
		// Add on the work unit because that contains sub combinations that may be of use
		ret_list = append(ret_list, work_unit...)
		current_src := 0
		workerFunc := func(a_num, b_num *Number) bool {
			// Here we have unrolled the functionality of make_2_to_1
			// So that it can use a single array
			// This is all to put less work on the malloc and gc

			if found_values.Solved() {
				return false
			}

			// We have to re-caclulate

			src_list[current_src] = a_num
			src_list[current_src+1] = b_num
			// Shorthand to make code more readable
			bob_list := src_list[current_src : current_src+2]
			if a_num.Val == 0 || b_num.Val == 0 {
				log.Fatalf("Gimmie gave %d, %d", a_num.Val, b_num.Val)
			}

			num_to_make,
				add_set, mul_set, sub_set, div_set,
				a_gt_b := found_values.do_maths(bob_list)

			// Shorthand
			tmp_list := num_list[current_number_loc:(current_number_loc + num_to_make)]

			// Populate the part of the return list for this run
			// This is the arra AddItems will write into
			// num_list gets filled with numbers, tmp_list is an alias to the same data here
			found_values.AddItems(bob_list, num_list, current_number_loc,
				add_set, mul_set, sub_set, div_set,
				a_gt_b)
			current_number_loc += num_to_make
			current_src += 2
			if found_values.SelfTest {
				for _, v := range tmp_list {
					v.ProveSol()
				}
			}
			ret_list = append(ret_list, tmp_list)
			return true
		}
		work_list.procWork(found_values, workerFunc)

	} else {

		workerFunc := func(a_num, b_num *Number) bool {
			bob_list := make([]*Number, 2)
			bob_list[0] = a_num
			bob_list[1] = b_num

			var tmp_list NumCol
			tmp_list = found_values.make_2_to_1(bob_list)
			found_values.AddMany(tmp_list...)
			ret_list = append(ret_list, tmp_list, array_in)
			return true
		}
		work_list.procWork(found_values, workerFunc)
	}
	// Add the entire solution list found in the previous loop in one go
	found_values.AddSol(ret_list, false)
	return ret_list
}
