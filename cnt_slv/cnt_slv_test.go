package cntSlv

import (
	"fmt"
	"github.com/pkg/profile"
	"math/rand"
	"runtime"

	"sync"
	"testing"
)

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
func initMany() []testset {
	retLst := make([]testset, 9)
	retLst[0] = *NewTestSet(833, 50, 3, 3, 1, 10, 7)
	retLst[1] = *NewTestSet(78, 8, 9, 10, 75, 25, 100)
	retLst[2] = *NewTestSet(540, 4, 5, 7, 2, 4, 8)
	retLst[3] = *NewTestSet(952, 25, 50, 75, 100, 3, 6)
	retLst[4] = *NewTestSet(559, 75, 10, 5, 6, 1, 3)
	retLst[5] = *NewTestSet(406, 25, 50, 10, 7, 5, 1)
	retLst[6] = *NewTestSet(269, 100, 10, 8, 9, 7, 7)
	retLst[7] = *NewTestSet(277, 75, 10, 6, 3, 5, 4)
	// (9-1)*50 = 400
	// (100 + 9*3) = 327
	// (400+327)= 727
	retLst[8] = *NewTestSet(727, 50, 100, 9, 1, 9, 3)
	return retLst
}

func TestOne(t *testing.T) {
	defer profile.Start().Stop()
	//defer profile.Start(profile.MemProfile).Stop()
	var target int
	//target = 78
	target = 531000

	var proofList SolLst
	var bob NumCol
	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true
	foundValues.PermuteMode = NetMap
	bob.AddNum(8, foundValues)
	bob.AddNum(9, foundValues)
	bob.AddNum(10, foundValues)
	bob.AddNum(75, foundValues)
	bob.AddNum(25, foundValues)
	bob.AddNum(100, foundValues)

	foundValues.SetTarget(target)

	proofList = append(proofList, bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	returnProofs := permuteN(bob, foundValues)
	cleanupPacker := 0
	for v := range returnProofs {
		//fmt.Println("Proof Received")
		if foundValues.SelfTest {
			// This unused code is handy if we want a proof list
			proofList = append(proofList, v...)
			cleanupPacker++
			if cleanupPacker > 10 {
				checkReturnList(proofList, foundValues)
				proofList.RemoveDuplicates()
				cleanupPacker = 0
			}
		}
	}
	if foundValues.Solved() {
		t.Log("Proof Found")

	} else {
		t.Log("Couldn't solve")
		//fmt.Println(proof_list)
		foundValues.PrintProofs()
		t.Fail()
	}
	proofList = SolLst{}
	foundValues = &NumMap{}
	bob = NumCol{}
}

func TestMany(t *testing.T) {
	defer profile.Start(profile.MemProfile).Stop()
	testSet := initMany()
	for _, item := range testSet {
		proofList := *new(SolLst)
		bob := *new(NumCol)
		foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end
		foundValues.SelfTest = true
		foundValues.UseMult = true
		foundValues.PermuteMode = rand.Intn(3) // Select a random mode

		for _, itm := range item.Selected {
			bob.AddNum(itm, foundValues)
		}
		foundValues.SetTarget(item.Target)
		proofList = append(proofList, bob) // Add on the work item that is the source

		fmt.Println("Starting permute")
		returnProofs := permuteN(bob, foundValues)

		cleanupPacker := 0
		for v := range returnProofs {
			if foundValues.SelfTest {
				// This unused code is handy if we want a proof list
				proofList = append(proofList, v...)
				cleanupPacker++
				if cleanupPacker > 1000 {
					proofList.RemoveDuplicates()
					cleanupPacker = 0
				}
			}
		}
		if foundValues.Solved() {
			t.Log("Proof Found")

		} else {
			t.Log("Couldn't solve")
			t.Log(proofList)
			t.Fail()
		}
	}
	//	if false {
	//		f, err := os.Create("memprofile.prof")
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		pprof.WriteHeapProfile(f)
	//		f.Close()
	//	}
}
func initFailMany() []testset {
	retLst := make([]testset, 0, 7)
	retLst = append(retLst, *NewTestSet(1000, 8, 9, 10))
	retLst = append(retLst, *NewTestSet(824, 3, 7, 6, 2, 1, 7))
	retLst = append(retLst, *NewTestSet(974, 1, 2, 2, 3, 3, 7))
	return retLst
}

func TestFail(t *testing.T) {
	defer profile.Start(profile.MemProfile).Stop()

	testSet := initFailMany()
	for _, item := range testSet {
		proofList := *new(SolLst)
		bob := *new(NumCol)
		foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end
		foundValues.SelfTest = false
		foundValues.UseMult = true

		for _, itm := range item.Selected {
			bob.AddNum(itm, foundValues)
		}
		foundValues.SetTarget(item.Target)
		proofList = append(proofList, bob) // Add on the work item that is the source

		fmt.Println("Starting permute")
		returnProofs := permuteN(bob, foundValues)

		cleanupPacker := 0
		for v := range returnProofs {
			if foundValues.SelfTest {
				// This unused code is handy if we want a proof list
				proofList = append(proofList, v...)
				cleanupPacker++
				if cleanupPacker > 10 {
					proofList.RemoveDuplicates()
					cleanupPacker = 0
				}
			}
		}

		if foundValues.Solved() {
			t.Log("We found an impossible proof")
			//t.Log(proof_list)
			t.Fail()
		} else {
			t.Log("Failed Correctly")
		}
	}
}
func TestIt(t *testing.T) {
	//var target int
	//target = 78

	var proofList SolLst
	var bob NumCol
	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true

	bob.AddNum(8, foundValues)
	bob.AddNum(9, foundValues)
	bob.AddNum(10, foundValues)
	bob.AddNum(75, foundValues)
	bob.AddNum(25, foundValues)
	//bob.AddNum(100, found_values)

	//found_values.SetTarget(target)

	proofList = append(proofList, bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	returnProofs := permuteN(bob, foundValues)
	cleanupPacker := 0
	for v := range returnProofs {
		if foundValues.SelfTest {
			proofList = append(proofList, v...)
			cleanupPacker++
			if cleanupPacker > 1000 {
				proofList.RemoveDuplicates()
				cleanupPacker = 0
			}
		}
	}
}
func TestReduction(t *testing.T) {
	var proofList0 SolLst
	var proofList1 SolLst
	var bob0 NumCol
	var bob1 NumCol
	foundValues0 := NewNumMap() //pass it the proof list so it can auto-check for validity at the end
	foundValues1 := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues0.SelfTest = true
	foundValues0.UseMult = true

	foundValues1.SelfTest = true
	foundValues1.UseMult = true

	bob0.AddNum(8, foundValues0)
	bob0.AddNum(9, foundValues0)
	bob0.AddNum(10, foundValues0)
	bob0.AddNum(75, foundValues0)
	//bob0.AddNum(25, found_values0)
	bob0.AddNum(100, foundValues0)
	bob1.AddNum(8, foundValues1)
	bob1.AddNum(9, foundValues1)
	bob1.AddNum(10, foundValues1)
	bob1.AddNum(75, foundValues1)
	//bob1.AddNum(25, found_values1)
	bob1.AddNum(100, foundValues1)

	proofList0 = append(proofList0, bob0) // Add on the work item that is the source
	proofList1 = append(proofList1, bob1) // Add on the work item that is the source

	fmt.Println("Starting permute")
	returnProofs0 := permuteN(bob0, foundValues0)
	mwg := new(sync.WaitGroup)
	mwg.Add(2)
	go func() {
		for v := range returnProofs0 {
			proofList0 = append(proofList0, v...)
		}
		mwg.Done()
	}()
	returnProofs1 := permuteN(bob1, foundValues1)
	go func() {
		for v := range returnProofs1 {
			proofList1 = append(proofList1, v...)
		}
		mwg.Done()
	}()
	mwg.Wait()
	fmt.Println("Everything should have finished by now, start pringting proofs")
	// So by this point found_values* and proof_list* should both have the same contents - if not in the same order
	if foundValues1.Compare(foundValues0) {
	} else {
		fmt.Println("The new FV were different")
		t.Fail()
	}

	var proofList2 SolLst
	proofList2 = append(proofList2, proofList0...)
	proofList0.RemoveDuplicates()
	fmt.Printf("Size Before %d; Size after %d\n", len(proofList2), len(proofList0))
	foundValues2 := NewNumMap()
	foundValues2.SelfTest = true
	foundValues2.UseMult = true

	foundValues2.AddSol(proofList0, false)
	foundValues2.LastNumMap()
	if foundValues2.Compare(foundValues0) {
	} else {
		fmt.Println("The new FV were different")
		t.Fail()
	}

	//found_values0.PrintProofs()
}

func BenchmarkWorknMulti(b *testing.B) {

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		foundValues := NewNumMap()
		foundValues.SelfTest = true
		foundValues.UseMult = true
		var bob NumCol
		nuMap := make(map[int]struct{})
		for j := 0; j < 6; j++ {
			run := true
			var k int
			for run {
				k = rand.Intn(100)
				if k > 0 {
					_, run = nuMap[k] // If it exists generate anther
					nuMap[k] = struct{}{}
				}
			}
			bob.AddNum(k, foundValues)
		}
		target := rand.Intn(1000)
		foundValues.SetTarget(target)
		runtime.GC()
		b.StartTimer()
		workN(bob, foundValues, true)
	}
}
func BenchmarkWorknSingle(b *testing.B) {

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		foundValues := NewNumMap()
		foundValues.SelfTest = true
		foundValues.UseMult = true
		var bob NumCol
		nuMap := make(map[int]struct{})
		for j := 0; j < 6; j++ {
			run := true
			var k int
			for run {
				k = rand.Intn(100)
				if k > 0 {
					_, run = nuMap[k] // If it exists generate anther
					nuMap[k] = struct{}{}
				}
			}
			bob.AddNum(k, foundValues)
		}
		target := rand.Intn(1000)
		foundValues.SetTarget(target)
		runtime.GC()
		b.StartTimer()
		workN(bob, foundValues, false)
	}
}
