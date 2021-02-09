package cntSlv

import (
	"fmt"
	"log"
)

func checkReturnList(proofList SolLst, foundValues *NumMap) {
	valueCheck := make(map[int]int)
	foundValues.constLk.RLock()
	if foundValues.TargetSet && !foundValues.SeekShort && *foundValues.solved {
		foundValues.constLk.RUnlock()
		// When we've aborted early because we found the proof
		// the proof list is incomplete
		return
	}
	foundValues.constLk.RUnlock()
	for _, v := range proofList {
		// v is *NumLst
		for _, w := range v {
			// w is *Number
			var Value int
			Value = w.Val
			valueCheck[Value] = 1
		}
	}

	tmp := foundValues.GetVals()
	for _, v := range tmp {
		_, ok := valueCheck[v]
		// Every value in found_values should be in the list of values returned
		if !ok {
			fmt.Printf("%d in Number map, but is not in the proof list, which has %d Items\n", v, proofList.Len())
			fmt.Println(proofList)
			log.Fatal("Done")
		}
	}
}
func (proofList SolLst) findProof(toFind int) {
	foundVal := false
	for _, v := range proofList {
		for _, w := range v {
			Value := w.Val
			proofString := w.String()
			if Value == toFind {
				foundVal = true
				fmt.Printf("Found Value %d, = %s\n", Value, proofString)
			}
		}
	}
	if !foundVal {
		fmt.Println("Unable to find value :", toFind)
	}
}
