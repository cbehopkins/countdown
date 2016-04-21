package cnt_slv

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	//"github.com/fighterlyt/permutation"
	"github.com/cbehopkins/permutation"
	"github.com/tonnerre/golang-pretty"
)


func WorkN(array_in NumCol, found_values *NumMap) SolLst {
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
	var work_list []SolLst
	work_list = expand_n(array_in)
	// so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }
	var work_unit SolLst
	var top_src_to_make int
	var top_numbers_to_make int
	for _, work_unit = range work_list {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

		if found_values.SelfTest {
			// Sanity check for programming errors
			work_unit_length := work_unit.Len()
			if work_unit_length != 2 {
				pretty.Println(work_list)
				log.Fatalf("Invalid work unit length, %d", work_unit_length)
			}
		}
		var unit_a, unit_b NumCol
		unit_a = work_unit[0]
		unit_b = work_unit[1]

		var list_a SolLst
		var list_b SolLst
		list_a = work_n(unit_a, found_values) // return a list of everything that can be done with this set
		list_b = work_n(unit_b, found_values)

		// Now we want two list of numbers to cross against each other
		gimmie_a := NewGimmie(list_a)
		gimmie_b := NewGimmie(list_b)
		// Now Cross work then
		current_item := 0
		cross_len := gimmie_a.Items() * gimmie_b.Items()
		num_items_to_make := cross_len * 2
		top_src_to_make += num_items_to_make

		num_numbers_to_make := 0
		for a_num, err_a := gimmie_a.Next(); err_a == nil; a_num, err_a = gimmie_a.Next() {
			for b_num, err_b := gimmie_b.Next(); err_b == nil; b_num, err_b = gimmie_b.Next() {
				tmp,
					_, _, _, _, _ := found_values.do_maths([]*Number{a_num, b_num})
				num_numbers_to_make += tmp
				current_item = current_item + 2
			}
			gimmie_b.Reset()
		}
		top_numbers_to_make += num_numbers_to_make

	}
	//current_item = 0
	// Malloc the memory once!
	current_number_loc := 0
	// This is the list of numbers that calculations are done from
	src_list := found_values.aquire_numbers(top_src_to_make)
	// This is the list of numbers that will be used in the proof
	// i.e. the list that calculations results end up in
	num_list := found_values.aquire_numbers(top_numbers_to_make)
	// And this allocates the list that will point to those (previously allocated) numbers
	// uses top_src_to_make/2 as for each 2 items there is one solution
	ret_list = make(SolLst, 0, ((top_src_to_make / 2) + work_unit.Len() + ret_list.Len()))
	// Add on the work unit because that contains sub combinations that may be of use
	ret_list = append(ret_list, work_unit...)
	current_src := 0
	for _, work_unit = range work_list {
		unit_a := work_unit[0]
		unit_b := work_unit[1]
		list_a := work_n(unit_a, found_values)
		list_b := work_n(unit_b, found_values)
		gimmie_a := NewGimmie(list_a)
		gimmie_b := NewGimmie(list_b)

		for a_num, err_a := gimmie_a.Next(); err_a == nil; a_num, err_a = gimmie_a.Next() {
			for b_num, err_b := gimmie_b.Next(); err_b == nil; b_num, err_b = gimmie_b.Next() {
				// Here we have unrolled the functionality of make_2_to_1
				// So that it can use a single array
				// This is all to put less work on the malloc and gc
				found_values.const_lk.RLock()
				if found_values.Solved {
					found_values.const_lk.RUnlock()
					return ret_list
				}
				found_values.const_lk.RUnlock()
				// We have to re-caclulate

				src_list[current_src] = a_num
				src_list[current_src+1] = b_num
				// Shorthand to make code more readable
				bob_list := src_list[current_src : current_src+2]

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
			}
		}

		if false {
			ret_list.TidySolLst()
		}
	}
	// Add the entire solution list found in the previous loop in one go
	found_values.AddSol(ret_list)
	// This now doubles the runtime for no significant improvement
	//ret_list.CheckDuplicates()
	return ret_list
}

func PermuteN(array_in NumCol, found_values *NumMap, proof_list chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to have enough CPUs for this to be an issue
	// However the framework seems to be there
	// TBD make this a comannd line variable
	var lone_map = true

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

	num_permutations := p.Left()
	fmt.Println("Num permutes:", num_permutations)
	var comms_channels []chan SolLst
	comms_channels = make([]chan SolLst, num_permutations)
	for i := range comms_channels {
		comms_channels[i] = make(chan SolLst, 200)
	}
	var channel_tokens chan bool
	channel_tokens = make(chan bool, 512)
	for i := 0; i < 16; i++ {
		//fmt.Println("Adding token");
		channel_tokens <- true
	}
	coallate_chan := make(chan SolLst, 200)
	coallate_done := make(chan bool, 8)

	var map_merge_chan chan *NumMap
	map_merge_chan = make(chan *NumMap)
	caller := func() {
		for result, err := p.Next(); err == nil; result, err = p.Next() {
			// To control the number of workers we run at once we need to grab a token
			// remember to return it later
			<-channel_tokens
			fmt.Printf("%3d permutation: left %3d, GoRs %3d\r", p.Index()-1, p.Left(), runtime.NumGoroutine())
			bob, ok := result.(NumCol)
			if !ok {
				log.Fatalf("Error Type conversion problem")
			}
			worker_par := func(it NumCol, fv *NumMap) {
				// This is the parallel worker function
				// It creates a new number map, populates it by working the incoming number set
				// then merges the number map back into the main numbermap
				// This is useful if we have more processes than we know what to do with
				var arthur NumMap
				var prfl SolLst
				fv.const_lk.RLock()
				if found_values.Solved {
					fv.const_lk.RUnlock()
					coallate_done <- true
					channel_tokens <- true
					return
				}
				fv.const_lk.RUnlock()
				arthur = *NewNumMap(&prfl) //pass it the proof list so it can auto-check for validity at the en
				prfl = work_n(it, &arthur)
				coallate_chan <- prfl
				arthur.LastNumMap()
				channel_tokens <- true // Now we're done, add a token to allow another to start
				map_merge_chan <- &arthur
				coallate_done <- true

			}
			worker_lone := func(it NumCol, fv *NumMap) {
				fv.const_lk.RLock()
				if fv.Solved {
					fv.const_lk.RUnlock()
					coallate_done <- true
					channel_tokens <- true
					return
				}
				fv.const_lk.RUnlock()
				coallate_chan <- work_n(it, fv)
				coallate_done <- true
				channel_tokens <- true // Now we're done, add a token to allow another to start

			}
			if !lone_map {
				go worker_par(bob, found_values)
			}
			if lone_map {
				go worker_lone(bob, found_values)
			}

		}
	}
	//fmt.Println("Starting caller")
	go caller()
	//fmt.Println("Caller started")
	merge_report := false // Turn off reporting of new numbers for first run
	mwg := new(sync.WaitGroup)
	mwg.Add(2)
	merge_func_worker := func() {
		for v := range map_merge_chan {
			found_values.Merge(v, merge_report)
			merge_report = true
		}
		mwg.Done()
	}
	if !lone_map {
		mwg.Add(1)
		go merge_func_worker()
	}
	// This little go function waits for all the procs to have a done channel and then closes the channel
	done_control := func() {
		for i := 0; i < num_permutations; i++ {
			<-coallate_done
		}
		//fmt.Println("All workers completed so closing coallate channel")
		close(coallate_chan)
		close(map_merge_chan)
		mwg.Done()
	}
	go done_control()

	output_merge := func() {
		for v := range coallate_chan {
			v.RemoveDuplicates()
			proof_list <- v
		}
		close(proof_list)
		mwg.Done()
	}
	go output_merge()
	mwg.Wait()

	found_values.LastNumMap()

}

func expand_n(array_a NumCol) []SolLst {
	var work_list []SolLst
	// Easier to explain by example:
	// {2,3,4} -> {{2},{3,4}}
	// {2,3,4,5} -> {{2}, {3,4,5}}
	//           -> {{2,3},{4,5}}
	// {2,3,4,5,6} -> {{2},{3,4,5,6}}
	//             -> {{2,3},{4,5,6}}
	//             -> {{2,3,4},{5,6}}

	// The consumer of this list of list (of list) will then feed each list length >1 into a the work+_n function
	// In order to get down to a {{a},{b}} which can then be worked
	// The important point is that even though the list we return may be indefinitly long
	// each work unit within it is then a smaller unit
	// so an input array of 3 numbers only generates work units that contain number lists of length 2 or less

	len_array_m1 := array_a.Len() - 1

	for i := 0; i < (len_array_m1); i++ {
		var ar_a, ar_b NumCol
		// for 3 items in arrar
		// {0},{1,2}, {0,1}{2}
		ar_a = make(NumCol, i+1)
		copy(ar_a, array_a[0:i+1])
		ar_b = make(NumCol, (array_a.Len() - (i + 1)))

		copy(ar_b, array_a[(i+1):(array_a.Len())])
		var work_item SolLst // {{2},{3,4}};
		// a work item always contains 2 elements to the array
		work_item = append(work_item, ar_a, ar_b)
		work_list = append(work_list, work_item)
	}
	return work_list
}
