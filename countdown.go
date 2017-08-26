package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/cbehopkins/countdown/cnt_slv"
	//	"github.com/pkg/profile"
)

func main() {
	// CPU profiling by default
	//defer profile.Start(profile.MemProfile).Stop()

	var target int
	var proofList cntSlv.SolLst
	var bob cntSlv.NumCol

	var sources []int
	var tgflg = flag.Int("target", 0, "Define the target number to reach")
	var dmflg = flag.Bool("dism", false, "Disable nultiplication")
	var stflg = flag.Bool("selft", false, "Check our own internals as we go")
	var srflg = flag.Bool("seeks", false, "Seek the shortest proof, as opposed to the quickest one to find")
	var ntflg = flag.Bool("net", false, "Attempt to use network mode")
	flag.Parse()

	foundValues := cntSlv.NewNumMap()
	// Global control flags default to test mode
	foundValues.SelfTest = *stflg
	foundValues.SeekShort = *srflg
	if *ntflg {
		foundValues.PermuteMode = cntSlv.NetMap
	}
	if *tgflg <= 0 {

		log.Fatal("No target specified")
	}
	if *dmflg == false {
		foundValues.UseMult = true
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
			bob.AddNum(smv, foundValues)
			sources = append(sources, smv)
		}
	}
	if false {
		returnProofs := make(chan cntSlv.SolLst, 16)
		if *tgflg > 0 {
			foundValues.SetTarget(target)
		}

		proofList = append(proofList, bob) // Add on the work item that is the source
		go cntSlv.RunPermute(bob, foundValues, returnProofs)
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
	} else {
		foundValues := cntSlv.NewNumMap()
		//found_values.SelfTest = true
		foundValues.UseMult = true
		foundValues.PermuteMode = cntSlv.LonMap
		foundValues.SeekShort = *srflg
		returnProofs := foundValues.CountHelper(target, sources)
		for range returnProofs {
			fmt.Println("Proof Received")
		}
		fmt.Println("Finished looking for proofs")
		profString := foundValues.GetProof(target)
		fmt.Println("It's:", profString)
	}
	// TBD on seeks option add in tidy printing of the final solution
	if *tgflg == 0 {
		foundValues.PrintProofs()
	}

}
