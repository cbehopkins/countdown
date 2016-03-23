package cnt_slv

import (
	"fmt"
	"log"
)

func check_return_list(proof_list SolLst, found_values *NumMap) {
	value_check := make(map[int]int)
	found_values.const_lk.RLock()
	if found_values.TargetSet && !found_values.SeekShort && found_values.Solved {
		found_values.const_lk.RUnlock()
		// When we've aborted early because we found the proof
		// the proof list is incomplete
		return
	}
	found_values.const_lk.RUnlock()
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
			fmt.Printf("%d in Number map, but is not in the proof list, which has %d Items\n", v, proof_list.Len())
			fmt.Println(proof_list)
			log.Fatal("Done")
		}
	}
}
func (proof_list SolLst) find_proof(to_find int) {
	found_val := false
	for _, v := range proof_list {
		for _, w := range *v {
			Value := w.Val
			proof_string := w.String()
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
