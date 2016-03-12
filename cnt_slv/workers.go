package cnt_slv

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sort"

	"github.com/cheggaaa/pb"
	"github.com/fighterlyt/permutation"
	"github.com/tonnerre/golang-pretty"
)

func gimmie_1(array_in SolLst, found_values *NumMap) NumCol {
	var ret_list NumCol
	// This function takes in a list of numbers and tries to return a list of numbers
	//  fmt.Println("gimmie_1 called with")
	//  pretty.Println(array_in)
	len_array_needed := 0
	for _, v := range array_in {
		//REVISIT - this is a lot of copying and allocation
		//ret_list = append(ret_list, *v...)
		len_array_needed = len_array_needed + v.Len()
	}

	ret_list = make(NumCol, 0, len_array_needed) // Length is zero capacity is as needed
	for _, v := range array_in {
		// Append should only increase the size of the array if needed
		ret_list = append(ret_list, *v...)
	}

	// Note - Running a reduction here extends the run time, not reduces it!
	//var reduced_ret_list []*Number;
	//  reduction_map := make (map [int] *Number)
	//  for i,v:= range ret_list {
	//    //fmt.Printf ("working with value %d\n", i);
	//    reduction_map[i] = v;
	//  }
	//  i:=0
	//  reduced_ret_list := make ([]*Number, len (reduction_map))
	//  for _, v:= range reduction_map {
	//    reduced_ret_list[i] = v
	//    i++
	//  }
	return ret_list
}

func work_n(array_in NumCol, found_values *NumMap) SolLst {
	var ret_list SolLst
	len_array_in := len(array_in)
	//  fmt.Printf("Calling work_n with %d items\n",len_array_in);
	//  for _, v:= range array_in {
	//    value := v.Val;
	//    fmt.Printf("%d,",value);
	//  }
	//  fmt.Printf("\n");
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
	//  fmt.Println("expand_n returned:")
	//  pretty.Println(work_list)
	// so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }
	var work_unit SolLst
	for _, work_unit = range work_list {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

		// Sanity check for programming errors
		work_unit_length := len(work_unit)
		if work_unit_length != 2 {
			pretty.Println(work_list)
			log.Fatalf("Invalid work unit length, %d", work_unit_length)
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
		for _, a_num := range list_of_1_a {
			for _, b_num := range list_of_1_b {
				if found_values.Solved {
					return ret_list
				}
				var product_of_2 NumCol
				//tmp_a := a_num.Val
				//tmp_b := b_num.Val
				//fmt.Println("a nd b", tmp_a,tmp_b);
				var list []*Number
				list = append(list, a_num, b_num)
				product_of_2 = make_2_to_1(list, found_values)
				ret_list = append(ret_list, &product_of_2)
			}
		}
		// Add on the work unit because that contains sub combinations that may be of use
		ret_list = append(ret_list, work_unit...)
	}
	// This adds about 10% to the run time, but reduces memory to 1/5th
	ret_list.CheckDuplicates()
	return ret_list
}

func PermuteN(array_in NumCol, found_values *NumMap, proof_list chan SolLst) {
	fmt.Println("Start Permute")
	less := func(i, j interface{}) bool {
		//fmt.Println("Less func")
		v1 := reflect.ValueOf(i).Elem().FieldByName("Val").Addr().Interface().(*int)
		v2 := reflect.ValueOf(j).Elem().FieldByName("Val").Addr().Interface().(*int)
		return *v1 < *v2
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
	for i := 0; i < 1; i++ {
		//fmt.Println("Adding token");
		channel_tokens <- true
	}
	coallate_chan := make(chan SolLst, 200)
	coallate_done := make(chan bool, 8)

	var map_merge_chan chan NumMap
	map_merge_chan = make(chan NumMap)
	caller := func() {
		fmt.Println("Inside Caller")
		for result, err := p.Next(); err == nil; result, err = p.Next() {

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
				go worker_lone(bob, found_values, p.Index()-1)
			}

		}
	}
	hold_until_done := make(chan bool)
	fmt.Println("Starting caller")
	go caller()
	fmt.Println("Caller started")
	merge_report := false // Turn off reporting of new numbers for first run

	merge_func_worker := func() {
		for v := range map_merge_chan {
			found_values.Merge(v, merge_report)
			merge_report = true
		}
		hold_until_done <- true
	}
	if !lone_map {
		go merge_func_worker()
	}
	// This little go function waits for all the procs to have a done channel and then closes the channel
	done_control := func() {
		for i := 0; i < num_permutations; i++ {

			//fmt.Println("removing token");
			<-coallate_done
			if found_values.Solved {
				break
			}
			//fmt.Println("removed token");
		}
		close(coallate_chan)
		//fmt.Println("Closing  map_merge_chan")
		close(map_merge_chan)
		if lone_map {
			hold_until_done <- true
		} // when all process finished then we're done with the NumberMap
	}
	go done_control()

	output_merge := func() {
		for v := range coallate_chan {
			proof_list <- v
		}
		fmt.Println("Closing proof list")
		close(proof_list)
	}
	go output_merge()
	<-hold_until_done // don't exit permute until merge_func_worker has finished

	fmt.Println("Last Map all done")
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
			//pretty.Println(w);
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

func run_check(work_channel chan []int) {

	for data := range work_channel {
		var proof_list SolLst
		var bob NumCol
		found_values := *NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end
		comma := ""
		//fmt.Printf("Processing Permutation: ")
		for _, v := range data {
			fmt.Printf("%s%d", comma, v)
			comma = ","
			bob.AddNum(v, &found_values)
		}
		fmt.Printf("\n")
		return_proofs := make(chan SolLst, 16)

		proof_list = append(proof_list, &bob) // Add on the work item that is the source

		go PermuteN(bob, &found_values, return_proofs)
		cleanup_packer := 0
		for v := range return_proofs {
			if false {
				proof_list = append(proof_list, v...)
			}
			cleanup_packer++
			if cleanup_packer > 1000 {
				proof_list.CheckDuplicates()
			}
		}
	}
}

func test_self() {
	i := []int{100, 75, 50, 25, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	p, err := permutation.NewPerm(i, nil) //generate a Permutator
	if err != nil {
		fmt.Println(err)
		return
	}
	work_log := make(map[string]bool)

	work_channel := make(chan []int, 2)
	go run_check(work_channel)
	var last_pindex int
	last_pindex = -1
	running_difference := 1
	skip_count := 21
	var bar_divide int
	bar_divide = (p.Left() / 10000) // we want bar_divide to do about 10000 increments
	bar := pb.StartNew(p.Left() / bar_divide)
	//  bar.SetRefreshRate(0)
	top_count := 0
	bar_count := 0
	for i, err := p.Next(); err == nil; i, err = p.Next() {
		if bar_count < bar_divide {
			bar_count++
		} else {
			bar.Increment()
			bar_count = 0
		}
		var subs []int
		slice_copy := i.([]int)
		subs = slice_copy[:6]
		sort.Ints(subs)
		var stringy string
		stringy = fmt.Sprintf("%d,%d,%d,%d,%d,%d", subs[5], subs[4], subs[3], subs[2], subs[1], subs[0])
		_, ok := work_log[stringy]
		if !ok {
			// First of all record we have seen this
			work_log[stringy] = true

			// For expediting things make sure we check sparsely
			if skip_count < 20 {
				skip_count++
				continue
			}
			skip_count = 0
			if last_pindex > 0 {
				running_difference = p.Index() - 1 - last_pindex
			}
			idx := (p.Index() - 1) / running_difference
			lft := p.Left() / running_difference
			fmt.Printf("\n*** TOP %3d permutation: %3d left %3d, Running Diff  %8d\n", top_count, idx, lft, running_difference)
			top_count++
			work_channel <- subs
			last_pindex = p.Index() - 1
		}
	}
	close(work_channel)
}
