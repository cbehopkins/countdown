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

	values := arrayIn.Values()
	itm := newFastPermInt(values)
	go itm.Worker(foundValues)
	return itm
}

// newFastPermInt good way to generate proofs
func newFastPermInt(values []int) *fastPermStruct {
	itm := new(fastPermStruct)
	p, err := permutation.NewPerm(values, nil)
	if err != nil {
		fmt.Println(err)
	}
	itm.p = p
	itm.ch = make(chan Proofs)
	itm.wg = new(sync.WaitGroup)
	itm.wg.Add(1)
	return itm
}

func (ps *fastPermStruct) Work(target int) {
	p := ps.p
	targetFound := false
	for result, err := p.Next(); (err == nil) && (!targetFound); result, err = p.Next() {
		bob, ok := result.([]int)
		if !ok {
			log.Fatalf("Error Type conversion problem\n")
		}
		inP := newProofLstMany(bob)
		// Get a data structure to put the result into
		proofs := getProofs()
		// Populate it
		proofs.wrkFast(*inP)
		if target > 0 {
			targetFound = proofs.Exists(target)
			if targetFound {
				ps.ch <- proofs
			}
		} else {
			ps.ch <- proofs
		}
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
func (ps fastPermStruct) GetProofs(target int) Proofs {
	itm := getProofs()
	go ps.Work(target)
	for newPr := range ps.ch {
		itm.merge(newPr)
	}
	return itm
}
