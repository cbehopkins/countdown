package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/cbehopkins/countdown/cnt_slv"
	"github.com/pkg/profile"
)

func main() {
	// CPU profiling by default
	defer profile.Start(profile.MemProfile).Stop()

	var target int
	var proof_list cnt_slv.SolLst
	var bob cnt_slv.NumCol

	var sources []int
	var tgflg = flag.Int("target", 0, "Define the target number to reach")
	var dmflg = flag.Bool("dism", false, "Disable nultiplication")
	var stflg = flag.Bool("selft", false, "Check our own internals as we go")
	var srflg = flag.Bool("seeks", false, "Seek the shortest proof, as opposed to the quickest one to find")
	var ntflg = flag.Bool("net", false, "Attempt to use network mode")
	flag.Parse()

	found_values := cnt_slv.NewNumMap()
	// Global control flags default to test mode
	found_values.SelfTest = *stflg
	found_values.SeekShort = *srflg
	if *ntflg {
		found_values.PermuteMode = cnt_slv.NetMap
	}
	if *tgflg <= 0 {

		log.Fatal("No target specified")
	}
	if *dmflg == false {
		found_values.UseMult = true
	}

	//fmt.Println("Set Target to ", *tgflg)
	target = *tgflg
	//fmt.Println("Other args are ", flag.Args())
	for _, j := range flag.Args() {
		value, err := strconv.ParseInt(j, 10, 32)
		if err != nil {
			log.Fatalf("Invalid command line ited, %s", j)
		} else {
			var smv int
			smv = int(value)
			fmt.Println("Found an Number ", smv)
			bob.AddNum(smv, found_values)
			sources = append(sources, smv)
		}
	}
	if false {
		return_proofs := make(chan cnt_slv.SolLst, 16)
		if *tgflg > 0 {
			found_values.SetTarget(target)
		}

		proof_list = append(proof_list, bob) // Add on the work item that is the source
		go cnt_slv.RunPermute(bob, found_values, return_proofs)
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
	} else {
		found_values := cnt_slv.NewNumMap()
		//found_values.SelfTest = true
		found_values.UseMult = true
		found_values.PermuteMode = cnt_slv.LonMap
		found_values.SeekShort = *srflg
		return_proofs := found_values.CountHelper(target, sources)
		for _ = range return_proofs {
			//fmt.Println("Proof Received", v)
		}
		profString := found_values.GetProof(target)
		fmt.Println("It's:", profString)
	}
	// TBD on seeks option add in tidy printing of the final solution
	//found_values.PrintProofs()

}
