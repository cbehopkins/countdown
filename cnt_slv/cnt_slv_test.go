package cnt_slv

import (
	"fmt"
	"testing"
)

//func TestSelf(t *testing.T) {
//	test_self()
//}
func TestThis(t *testing.T) {
	fmt.Printf("Test")
}
func TestOne(t *testing.T) {
	var target int
	target = 78

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

	proof_list = append(proof_list, &bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob, found_values, return_proofs)
	cleanup_packer := 0
	for v := range return_proofs {
		if found_values.SelfTest {
			// This unused code is handy if we want a proof list
			proof_list = append(proof_list, v...)
			cleanup_packer++
			if cleanup_packer > 1000 {
				proof_list.CheckDuplicates()
				cleanup_packer = 0
			}
		}
	}
	if found_values.Solved {
		t.Log("Proof Found")

	} else {
		t.Log("Couldn't solve")
		print_proofs(proof_list)
		t.Fail()
	}
}

func TestFail(t *testing.T) {
	var target int

	var proof_list SolLst
	var bob NumCol
	found_values := NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end

	found_values.SelfTest = true
	found_values.UseMult = true
	bob.AddNum(8, found_values)
	bob.AddNum(9, found_values)
	bob.AddNum(10, found_values)

	return_proofs := make(chan SolLst, 16)

	found_values.SetTarget(target)
	target = 1000 // You can't make 1000 from these input numbers

	proof_list = append(proof_list, &bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob, found_values, return_proofs)
	cleanup_packer := 0
	for v := range return_proofs {

		if found_values.SelfTest {
			// This unused code is handy if we want a proof list
			proof_list = append(proof_list, v...)
			cleanup_packer++
			if cleanup_packer > 1000 {
				proof_list.CheckDuplicates()
				cleanup_packer = 0
			}
		}
	}
	if found_values.Solved {
		t.Log("We found an impossible proof")
		print_proofs(proof_list)
		t.Fail()
	} else {
		t.Log("Failed Correctly")
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
	bob.AddNum(100, found_values)

	return_proofs := make(chan SolLst, 16)

	//found_values.SetTarget(target)

	proof_list = append(proof_list, &bob) // Add on the work item that is the source

	fmt.Println("Starting permute")
	go PermuteN(bob, found_values, return_proofs)
	cleanup_packer := 0
	for v := range return_proofs {
		if found_values.SelfTest {
			// This unused code is handy if we want a proof list
			proof_list = append(proof_list, v...)
			cleanup_packer++
			if cleanup_packer > 1000 {
				proof_list.CheckDuplicates()
				cleanup_packer = 0
			}
		}
	}
}
