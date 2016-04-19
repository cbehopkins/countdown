package cnt_slv

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"os"
	"runtime/pprof"
	"log"
)

//func TestSelf(t *testing.T) {
//	test_self()
//}
func TestThis(t *testing.T) {
	fmt.Printf("Test")
}

type testset struct {
	Selected []int
	Target   int
}

func NewTestSet(target int, seld ...int) *testset {
	item := new(testset)
	item.Target = target
	item.Selected = make([]int, len(seld))
	for i, j := range seld {
		item.Selected[i] = j
	}
	return item
}
func init_many() []testset {
	ret_lst := make([]testset, 8)
	ret_lst[0] = *NewTestSet(833, 50, 3, 3, 1, 10, 7)
	ret_lst[1] = *NewTestSet(78, 8, 9, 10, 75, 25, 100)
	ret_lst[2] = *NewTestSet(540, 4, 5, 7, 2, 4, 8)
	ret_lst[3] = *NewTestSet(952, 25, 50, 75, 100, 3, 6)
	ret_lst[4] = *NewTestSet(559, 75, 10, 5, 6, 1, 3)
	ret_lst[5] = *NewTestSet(406, 25, 50, 10, 7, 5, 1)
	ret_lst[6] = *NewTestSet(269, 100, 10, 8, 9, 7, 7)
	ret_lst[7] = *NewTestSet(277, 75, 10, 6, 3, 5, 4)

	return ret_lst
}

func TestOne(t *testing.T) {
	var target int
	//target = 78
	target = 531000

	var proof_list SolLst
	var bob NumCol
	found_values := NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end

	found_values.SelfTest = true
	found_values.UseMult = true
	bob.AddNum(8, found_values)
	bob.AddNum(9, found_values)
	bob.AddNum(10, found_values)
	bob.AddNum(75, found_values)
	bob.AddNum(25, found_values)
	bob.AddNum(100, found_values)

	return_proofs := make(chan SolLst, 16)

	found_values.SetTarget(target)

	proof_list = append(proof_list, bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob, found_values, return_proofs)
	cleanup_packer := 0
	for v := range return_proofs {
		if found_values.SelfTest {
			// This unused code is handy if we want a proof list
			proof_list = append(proof_list, v...)
			cleanup_packer++
			if cleanup_packer > 1000 {
				check_return_list(proof_list, found_values)
				proof_list.RemoveDuplicates()
				cleanup_packer = 0
			}
		}
	}
	if found_values.Solved {
		t.Log("Proof Found")

	} else {
		t.Log("Couldn't solve")
		fmt.Println(proof_list)
		t.Fail()
	}
	found_values = &NumMap{}
	bob = NumCol{}
}
func TestMany(t *testing.T) {


	test_set := init_many()
	for _, item := range test_set {
		proof_list := *new(SolLst)
		bob := *new(NumCol)
		found_values := NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end
		found_values.SelfTest = true
		found_values.UseMult = true
		return_proofs := make(chan SolLst, 16)
		for _, itm := range item.Selected {
			bob.AddNum(itm, found_values)
		}
		found_values.SetTarget(item.Target)
		proof_list = append(proof_list, bob) // Add on the work item that is the source

		fmt.Println("Starting permute")
		go PermuteN(bob, found_values, return_proofs)

		cleanup_packer := 0
		for v := range return_proofs {
			if found_values.SelfTest {
				// This unused code is handy if we want a proof list
				proof_list = append(proof_list, v...)
				cleanup_packer++
				if cleanup_packer > 1000 {
					proof_list.RemoveDuplicates()
					cleanup_packer = 0
				}
			}
		}
		if found_values.Solved {
			t.Log("Proof Found")

		} else {
			t.Log("Couldn't solve")
			t.Log(proof_list)
			t.Fail()
		}
	}
	if false {
	f, err := os.Create("memprofile.prof")
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
	}
}
func init_fail_many() []testset {
	ret_lst := make([]testset, 3)
	ret_lst[0] = *NewTestSet(1000, 8, 9, 10)
	ret_lst[1] = *NewTestSet(824, 3, 7, 6, 2, 1, 7)
	ret_lst[2] = *NewTestSet(974, 1, 2, 2, 3, 3, 7)
	//ret_lst[3] = *NewTestSet(952, 25, 50, 75,100,  3,  6)
	//ret_lst[4] = *NewTestSet(559, 75, 10,  5,  6,  1,  3)
	//ret_lst[5] = *NewTestSet(406, 25, 50, 10,  7,  5,  1)
	//ret_lst[6] = *NewTestSet(269,100, 10,  8,  9,  7,  7)
	//ret_lst[7] = *NewTestSet(277, 75, 10,  6,  3,  5,  4)
	return ret_lst
}

func TestFail(t *testing.T) {
	test_set := init_fail_many()
	for _, item := range test_set {
		proof_list := *new(SolLst)
		bob := *new(NumCol)
		found_values := NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end
		found_values.SelfTest = true
		found_values.UseMult = true
		return_proofs := make(chan SolLst, 16)
		for _, itm := range item.Selected {
			bob.AddNum(itm, found_values)
		}
		found_values.SetTarget(item.Target)
		proof_list = append(proof_list, bob) // Add on the work item that is the source

		fmt.Println("Starting permute")
		go PermuteN(bob, found_values, return_proofs)

		cleanup_packer := 0
		for v := range return_proofs {
			if found_values.SelfTest {
				// This unused code is handy if we want a proof list
				proof_list = append(proof_list, v...)
				cleanup_packer++
				if cleanup_packer > 1000 {
					proof_list.RemoveDuplicates()
					cleanup_packer = 0
				}
			}
		}

		if found_values.Solved {
			t.Log("We found an impossible proof")
			t.Log(proof_list)
			t.Fail()
		} else {
			t.Log("Failed Correctly")
		}
	}
}
func TestIt(t *testing.T) {
	//var target int
	//target = 78

	var proof_list SolLst
	var bob NumCol
	found_values := NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end

	found_values.SelfTest = true
	found_values.UseMult = true

	bob.AddNum(8, found_values)
	bob.AddNum(9, found_values)
	bob.AddNum(10, found_values)
	bob.AddNum(75, found_values)
	bob.AddNum(25, found_values)
	//bob.AddNum(100, found_values)

	return_proofs := make(chan SolLst, 16)

	//found_values.SetTarget(target)

	proof_list = append(proof_list, bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob, found_values, return_proofs)
	cleanup_packer := 0
	for v := range return_proofs {
		if found_values.SelfTest {
			proof_list = append(proof_list, v...)
			cleanup_packer++
			if cleanup_packer > 1000 {
				proof_list.RemoveDuplicates()
				cleanup_packer = 0
			}
		}
	}
}
func TestReduction(t *testing.T) {
	var proof_list0 SolLst
	var proof_list1 SolLst
	var bob0 NumCol
	var bob1 NumCol
	found_values0 := NewNumMap(&proof_list0) //pass it the proof list so it can auto-check for validity at the end
	found_values1 := NewNumMap(&proof_list1) //pass it the proof list so it can auto-check for validity at the end

	found_values0.SelfTest = true
	found_values0.UseMult = true

	found_values1.SelfTest = true
	found_values1.UseMult = true

	bob0.AddNum(8, found_values0)
	bob0.AddNum(9, found_values0)
	bob0.AddNum(10, found_values0)
	bob0.AddNum(75, found_values0)
	//bob0.AddNum(25, found_values0)
	bob0.AddNum(100, found_values0)
	bob1.AddNum(8, found_values1)
	bob1.AddNum(9, found_values1)
	bob1.AddNum(10, found_values1)
	bob1.AddNum(75, found_values1)
	//bob1.AddNum(25, found_values1)
	bob1.AddNum(100, found_values1)

	return_proofs0 := make(chan SolLst, 16)
	return_proofs1 := make(chan SolLst, 16)
	proof_list0 = append(proof_list0, bob0) // Add on the work item that is the source
	proof_list1 = append(proof_list1, bob1) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob0, found_values0, return_proofs0)
	mwg := new(sync.WaitGroup)
	mwg.Add(2)
	go func() {
		for v := range return_proofs0 {
			proof_list0 = append(proof_list0, v...)
		}
		mwg.Done()
	}()
	go PermuteN(bob1, found_values1, return_proofs1)
	go func() {
		for v := range return_proofs1 {
			proof_list1 = append(proof_list1, v...)
		}
		mwg.Done()
	}()
	mwg.Wait()
	fmt.Println("Everything should have finished by now, start pringting proofs")
	// So by this point found_values* and proof_list* should both have the same contents - if not in the same order
	if found_values1.Compare(found_values0) {
	} else {
		fmt.Println("The new FV were different")
		t.Fail()
	}

	var proof_list2 SolLst
	proof_list2 = append(proof_list2, proof_list0...)
	proof_list0.RemoveDuplicates()
	fmt.Printf("Size Before %d; Size after %d\n", len(proof_list2), len(proof_list0))
	found_values2 := NewNumMap(&proof_list2)
	found_values2.SelfTest = true
	found_values2.UseMult = true

	found_values2.AddSol(proof_list0)
	found_values2.LastNumMap()
	if found_values2.Compare(found_values0) {
	} else {
		fmt.Println("The new FV were different")
		t.Fail()
	}

	//found_values0.PrintProofs()
}

func BenchmarkWorkn(b *testing.B) {

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var proof_list SolLst

		found_values := NewNumMap(&proof_list)
		found_values.SelfTest = true
		found_values.UseMult = true
		var bob NumCol
		nu_map := make(map[int]int)
		for j := 0; j < 6; j++ {
			run := true
			var k int
			for run {
				k = rand.Intn(100)
				_, run = nu_map[k] // If it exists generate anther
			}
			bob.AddNum(k, found_values)
		}

		//found_values.SetTarget(target)
		runtime.GC()
		fmt.Println("Starting work_n")
		b.StartTimer()
		work_n(bob, found_values)
	}
}
