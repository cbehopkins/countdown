package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	cntslv "github.com/cbehopkins/countdown/cnt_slv"
)

func main() {
	// CPU profiling by default
	//defer profile.Start(profile.MemProfile).Stop()

	var target int
	var bob cntslv.NumCol

	var sources []int
	var tgflg = flag.Int("target", 0, "Define the target number to reach")
	var dmflg = flag.Bool("dism", false, "Disable nultiplication")
	var stflg = flag.Bool("selft", false, "Check our own internals as we go")
	var srflg = flag.Bool("seeks", false, "Seek the shortest proof, as opposed to the quickest one to find")
	flag.Parse()

	foundValues := cntslv.NewNumMap()
	// Global control flags default to test mode
	foundValues.SelfTest = *stflg
	foundValues.SeekShort = *srflg
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
	foundValues = cntslv.NewNumMap()
	//found_values.SelfTest = true
	// foundValues.UseMult = true
	// foundValues.SeekShort = *srflg
	returnProofs := foundValues.CountHelper(target, sources)
	for range returnProofs {
		//fmt.Println("Proof Received")
	}
	fmt.Println("Finished looking for proofs")
	profString := foundValues.GetProof(target)
	fmt.Println("It's:", profString)
	if *tgflg == 0 {
		foundValues.PrintProofs()
	}
}
