package cnt_slv

import (
	"fmt"
	"log"
	"runtime"
	//"github.com/fighterlyt/permutation"
	"github.com/cbehopkins/permutation"
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
	return work_n(array_in, found_values)
}

func work_n(array_in NumCol, found_values *NumMap) SolLst {
	var ret_list SolLst
	len_array_in := array_in.Len()
	found_values.const_lk.RLock()
	if found_values.Solved {
		found_values.const_lk.RLock()
		return ret_list
	}
	found_values.const_lk.RUnlock()
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
		found_values.const_lk.RLock()
		if found_values.Solved {
			found_values.const_lk.RUnlock()
			return false
		}
		found_values.const_lk.RUnlock()
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
	// Add the entire solution list found in the previous loop in one go
	found_values.AddSol(ret_list, false)
	return ret_list
}
func permuteN(array_in NumCol, found_values *NumMap) (proof_list chan SolLst) {
	return_proofs := make(chan SolLst, 16)
	go PermuteN(array_in, found_values, return_proofs)
	return return_proofs
}
func PermuteN(array_in NumCol, found_values *NumMap, proof_list chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to have enough CPUs for this to be an issue
	// However the framework seems to be there
	// TBD make this a comannd line variable
	permute_mode := found_values.PermuteMode
	required_tokens := 16

	//fmt.Println("Start Permute")
	less := func(i, j interface{}) bool {
		tmp, ok := i.(*Number)
		if !ok {
			log.Fatal("Can't compare an empty number")
		}
		v1 := tmp.Val
		tmp, ok = j.(*Number)
		if !ok {
			log.Fatal("Can't compare an empty number")
		}
		v2 := tmp.Val
		return v1 < v2
	}
	p, err := permutation.NewPerm(array_in, less)
	if err != nil {
		fmt.Println(err)
	}

	pstrct := new_perm_struct(p, permute_mode == NetMap)
	pstrct.NumWorkers(16)

	if permute_mode == NetMap {
		extra_tokens, all_fail := pstrct.setup_conns(found_values)
		required_tokens += extra_tokens
		if all_fail {
			permute_mode = LonMap
		}
	}
	for i := 0; i < required_tokens; i++ {
		//fmt.Println("Adding token");
		pstrct.channel_tokens <- true
	}

	caller := func() {
		for result, err := p.Next(); err == nil; result, err = p.Next() {
			// To control the number of workers we run at once we need to grab a token
			// remember to return it later
			<-pstrct.channel_tokens
			fmt.Printf("%3d permutation: left %3d, GoRs %3d\r", p.Index()-1, p.Left(), runtime.NumGoroutine())
			bob, ok := result.(NumCol)
			if !ok {
				log.Fatalf("Error Type conversion problem")
			}

			if permute_mode == ParMap {
				go pstrct.worker_par(bob, found_values)
			}
			if permute_mode == LonMap {
				go pstrct.worker_lone(bob, found_values)
			}
			if permute_mode == NetMap {
				go pstrct.worker_net_send(bob, found_values)
			}

		}
	}
	go caller()

	pstrct.Workers(permute_mode, found_values, proof_list)

	pstrct.Wait()
	found_values.LastNumMap()
}
