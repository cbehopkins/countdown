package cntSlv

import (
	"fmt"
	"log"
	"sync"

	"github.com/cbehopkins/permutation"
)

type fastPermStruct struct {
	p               *permutation.Permutator
	numPermutations int
	ch              chan Proofs
	wg              *sync.WaitGroup
}

func newFastPermStruct(arrayIn NumCol, foundValues *NumMap) *fastPermStruct {
	itm := new(fastPermStruct)
	values := arrayIn.Values()
	p, err := permutation.NewPerm(values, nil)
	if err != nil {
		fmt.Println(err)
	}
	itm.p = p
	itm.ch = make(chan Proofs)
	itm.wg = new(sync.WaitGroup)
	itm.wg.Add(1)
	go itm.Worker(foundValues)
	return itm
}
func (ps *fastPermStruct) Work() {
	p := ps.p
	for result, err := p.Next(); err == nil; result, err = p.Next() {
		bob, ok := result.([]int)
		if !ok {
			log.Fatalf("Error Type conversion problem\n")
		}
		inP := newProofLstMany(bob)
		// Get a data structure to put the result into
		proofs := getProofs()
		// Populate it
		proofs.wrkFast(*inP)
		// Now convert the result into something we can use
		ps.ch <- proofs
	}
	close(ps.ch)
	ps.wg.Wait()
}

// Worker pulls completed proofs off the channel
// and stuffs them into the map
func (ps fastPermStruct) Worker(fv *NumMap) {
	for proofs := range ps.ch {
		proofs.addProofsNm(fv)
	}
	ps.wg.Done()
}
