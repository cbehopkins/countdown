package cnt_slv

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/fighterlyt/permutation"
	"github.com/tonnerre/golang-pretty"
)

func gimmie_1(array_in SolLst, found_values *NumMap) NumCol {
	var ret_list NumCol
	// This function takes in a list of numbers and tries to return a list of numbers
	len_array_needed := 0
	for _, v := range array_in {
		// We're scanning through working out how big on an array we need to allocate
		len_array_needed = len_array_needed + v.Len()
	}

	ret_list = make(NumCol, 0, len_array_needed) // Length is zero capacity is as needed
	for _, v := range array_in {
		// Append should only increase the size of the array if needed
		// Because we have calculuated it previously this should not be needed
		// and this is reduced to a copying exercise
		ret_list = append(ret_list, *v...)
	}

	return ret_list
}

func work_n(array_in NumCol, found_values *NumMap) SolLst {
	var ret_list SolLst
	len_array_in := len(array_in)
	if found_values.Solved {
		return ret_list
	}
	if len_array_in == 1 {
		ret_list = append(ret_list, &array_in)
		return ret_list
	} else if len_array_in == 2 {
		//var a, b *Number
		//a = array_in[0]
		//b = array_in[1]
		var tmp_list NumCol

		//tmp_a := a.Val
		//tmp_b := b.Val
		//fmt.Println("a nd b", tmp_a,tmp_b);
		tmp_list = make_2_to_1(array_in[0:2], found_values)
		found_values.AddMany(tmp_list...)
		ret_list = append(ret_list, &tmp_list, &array_in)
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
	for _, work_unit = range work_list {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

		if found_values.SelfTest {
			// Sanity check for programming errors
			work_unit_length := len(work_unit)
			if work_unit_length != 2 {
				pretty.Println(work_list)
				log.Fatalf("Invalid work unit length, %d", work_unit_length)
			}
		}
		var unit_a, unit_b *NumCol
		unit_a = work_unit[0]
		unit_b = work_unit[1]

		var list_a SolLst
		var list_b SolLst
		list_a = work_n(*unit_a, found_values)
		list_b = work_n(*unit_b, found_values)

		// Now we want two list of numbers to cross against each other
		var list_of_1_a, list_of_1_b NumCol
		list_of_1_a = gimmie_1(list_a, found_values)
		list_of_1_b = gimmie_1(list_b, found_values)

		// Now Cross work then
		current_item := 0
		cross_len := len(list_of_1_a) * len(list_of_1_b)
		num_items_to_make := cross_len * 2
		list := make([]*Number, num_items_to_make)
		num_ret_to_make := 0
		// So scan through and work out how many items we are going to need
		for _, a_num := range list_of_1_a {
			for _, b_num := range list_of_1_b {
				list[current_item] = a_num
				list[current_item+1] = b_num
				tmp,
					_, _, _, _, _, _, _, _, _ := found_values.DoMaths(list[current_item:(current_item + 2)])
				num_ret_to_make += tmp
				current_item = current_item + 2
			}
		}
		current_item = 0
		// Malloc the memory once!
		current_number_loc := 0
		num_list := found_values.aquire_numbers(num_ret_to_make)
		ret_list = make(SolLst, 0, cross_len)

		for i := 0; i < cross_len; i++ {
			if found_values.Solved {
				return ret_list
			}
			// We have to re-caclulate
			num_to_make,
				add_res, mul_res, sub_res, div_res,
				add_set, mul_set, sub_set, div_set,
				a_gt_b := found_values.DoMaths(list[current_item:(current_item + 2)])

			// Populate the part of the return list for this run
			tmp_list := num_list[current_number_loc:(current_number_loc + num_to_make)]
			found_values.AddItems(list[current_item:(current_item+2)], num_list, &current_number_loc,
				add_res, mul_res, sub_res, div_res,
				add_set, mul_set, sub_set, div_set,
				a_gt_b)

			current_item = current_item + 2
			if found_values.SelfTest {
				for _, v := range tmp_list {
					v.ProveSol()
				}
			}
			ret_list = append(ret_list, &tmp_list)
		}

		// Add on the work unit because that contains sub combinations that may be of use
		ret_list = append(ret_list, work_unit...)
	}
	// Add the entire solution list found in the previous loop in one go
	found_values.AddSol(ret_list)
	// This adds about 10% to the run time, but reduces memory to 1/5th
	ret_list.CheckDuplicates()
	return ret_list
}

func PermuteN(array_in NumCol, found_values *NumMap, proof_list chan SolLst) {
	fmt.Println("Start Permute")
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
	channel_tokens = make(chan bool, 128)
	for i := 0; i < 64; i++ {
		//fmt.Println("Adding token");
		channel_tokens <- true
	}
	coallate_chan := make(chan SolLst, 200)
	coallate_done := make(chan bool, 8)

	var map_merge_chan chan NumMap
	map_merge_chan = make(chan NumMap)
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
			worker_par := func(it NumCol, fv *NumMap, curr_iten int) {
				// This is the parallel worker function
				// It creates a new number map, populates it by working the incoming number set
				// then merges the number map back into the main numbermap
				// This is useful if we have more processes than we know what to do with
				var arthur NumMap
				var prfl SolLst
				if found_values.Solved {
					coallate_done <- true
					channel_tokens <- true
					return
				}
				arthur = *NewNumMap(&prfl) //pass it the proof list so it can auto-check for validity at the en
				prfl = work_n(it, &arthur)
				coallate_chan <- prfl
				arthur.LastNumMap()
				channel_tokens <- true // Now we're done, add a token to allow another to start
				map_merge_chan <- arthur
				coallate_done <- true

			}
			worker_lone := func(it NumCol, fv *NumMap, curr_iten int) {
				if found_values.Solved {
					coallate_done <- true
					channel_tokens <- true
					return
				}
				coallate_chan <- work_n(it, fv)
				//fmt.Println("cdone send");
				coallate_done <- true
				//fmt.Println("cdone sent");
				//fmt.Println("Adding token");
				channel_tokens <- true // Now we're done, add a token to allow another to start

			}
			if !lone_map {
				go worker_par(bob, found_values, p.Index()-1)
			}
			if lone_map {
				worker_lone(bob, found_values, p.Index()-1)
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
		//fmt.Println("Closing  map_merge_chan")
		close(map_merge_chan)
		mwg.Done()
	}
	go done_control()

	output_merge := func() {
		for v := range coallate_chan {
			//v.CheckDuplicates()
			//fmt.Println("Received a proof")
			proof_list <- v
		}
		//fmt.Println("Closing proof list")
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

	len_array_m1 := len(array_a) - 1

	for i := 0; i < (len_array_m1); i++ {
		var ar_a, ar_b NumCol
		// for 3 items in arrar
		// {0},{1,2}, {0,1}{2}
		ar_a = make(NumCol, i+1)
		copy(ar_a, array_a[0:i+1])
		ar_b = make(NumCol, (len(array_a) - (i + 1)))

		copy(ar_b, array_a[(i+1):(len(array_a))])
		var work_item SolLst // {{2},{3,4}};
		// a work item always contains 2 elements to the array
		work_item = append(work_item, &ar_a, &ar_b)
		work_list = append(work_list, work_item)
	}
	return work_list
}

func check_return_list(proof_list SolLst, found_values *NumMap) {
	value_check := make(map[int]int)
	if found_values.TargetSet && !found_values.SeekShort && found_values.Solved {
		// When we've aborted early because we found the proof
		// the proof list is incomplete
		return
	}
	for _, v := range proof_list {
		// v is *NumLst
		for _, w := range *v {
			// w is *Number
			var Value int
			Value = w.Val
			value_check[Value] = 1
		}
	}

	tmp := found_values.GetVals()
	for _, v := range tmp {
		_, ok := value_check[v]
		// Every value in found_values should be in the list of values returned
		if !ok {
			fmt.Printf("%d in Number map, but is not in the proof list, which has %d Items\n", v, len(proof_list))
			print_proofs(proof_list)
		}
	}
}
func find_proof(proof_list SolLst, to_find int) {
	found_val := false
	for _, v := range proof_list {
		for _, w := range *v {
			Value := w.Val
			proof_string := w.ProveIt()
			if Value == to_find {
				found_val = true
				fmt.Printf("Found Value %d, = %s\n", Value, proof_string)
			}
		}
	}
	if !found_val {
		fmt.Println("Unable to find value :", to_find)
	}
}
func print_proofs(proof_list SolLst) {
	for _, v := range proof_list {
		// v is *NumCol
		for _, w := range *v {
			// w is *Number
			var Value int
			Value = w.Val
			proof_string := w.ProveIt()
			fmt.Printf("Value %3d, = %s\n", Value, proof_string)
		}
	}
	fmt.Println("Done printing proofs")
}
