package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/cbehopkins/countdown/cnt_slv"
)

func main() {
	//test_self ()

	var target int
	target = 78

	var proof_list cnt_slv.SolLst
	var bob cnt_slv.NumCol
	found_values := cnt_slv.NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end

	var tgflg = flag.Int("target", 0, "Define the target number to reach")
	var muflg = flag.Bool("mult", false, "Turn on multiplication")
	var dmflg = flag.Bool("dism", false, "Disable nultiplication")
	var stflg = flag.Bool("selft", false, "Check our own internals as we go")
	var srflg = flag.Bool("seeks", false, "Seek the shortest proof, as opposed to the quickest one to find")
	flag.Parse()

	// Global control flags default to test mode
	found_values.UseMult = *muflg
	found_values.SelfTest = *stflg
	found_values.SeekShort = *srflg

	if *tgflg > 0 {
		if *dmflg == false {
			found_values.UseMult = true
		}
		fmt.Println("Set Target to ", *tgflg)
		target = *tgflg
		fmt.Println("Other args are ", flag.Args())
		for _, j := range flag.Args() {
			value, err := strconv.ParseInt(j, 10, 32)
			if err != nil {
				log.Fatalf("Invalid command line ited, %s", j)
			} else {
				var smv int
				smv = int(value)
				fmt.Println("Found an Number ", smv)
				bob.AddNum(smv, found_values)
			}

		}
	} else {
		found_values.UseMult = true
		bob.AddNum(8, found_values)
		bob.AddNum(9, found_values)
		bob.AddNum(10, found_values)
		bob.AddNum(75, found_values)
		bob.AddNum(25, found_values)
		bob.AddNum(100, found_values)
	}
	return_proofs := make(chan cnt_slv.SolLst, 16)

	if *tgflg > 0 {
		found_values.SetTarget(target)
	}

	proof_list = append(proof_list, &bob) // Add on the work item that is the source

	go cnt_slv.Permute_n(bob, found_values, return_proofs)
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
