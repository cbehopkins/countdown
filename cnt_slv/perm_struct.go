package cntSlv

import (
	"fmt"
	"log"
	"sync"

	"github.com/cbehopkins/permutation"
)

type permStruct struct {
	p            *permutation.Permutator
	fv           *NumMap
	coallateChan chan SolLst
}

func newPermStruct(arrayIn NumCol, foundValues *NumMap) *permStruct {

	p, err := permutation.NewPerm(arrayIn, lessNumber)
	if err != nil {
		fmt.Println(err)
	}

	itm := new(permStruct)
	itm.p = p
	itm.fv = foundValues
	itm.coallateChan = make(chan SolLst, 200)
	return itm
}

func (ps *permStruct) workerLone(it NumCol, fv *NumMap) {
	if !fv.Solved() {
		ps.coallateChan <- workN(it, fv)
	}
}

// Work the permutation struct
// That is get permulations and send them on the
// toWork Chan
func (ps *permStruct) generatePermutations() chan NumCol {
	permuteChan := make(chan NumCol)
	go func() {
		p := ps.p
		for result, err := p.Next(); err == nil; result, err = p.Next() {
			bob, ok := result.(NumCol)
			if !ok {
				log.Fatalf("Error Type conversion problem")
			}
			permuteChan <- bob
		}
		close(permuteChan)
	}()
	return permuteChan
}

func (ps *permStruct) Workers(proofList chan SolLst, numWorkers int) {
	permuteChan := ps.generatePermutations() // The thing that generates Permutations to work
	coallateWg := ps.Launch(numWorkers, permuteChan)
	var mwg sync.WaitGroup
	// one thing -  outputMerge - to wait for
	mwg.Add(1)

	// This will if needed merge together the results and then Done mwg
	go ps.outputMerge(proofList, &mwg)
	coallateWg.Wait()
	close(ps.coallateChan)
	// wait for all then Done on mwg
	mwg.Wait()
}
func (ps *permStruct) outputMerge(proofList chan SolLst, mwg *sync.WaitGroup) {
	for v := range ps.coallateChan {
		if proofList != nil {
			proofList <- v
		}
	}
	if proofList != nil {
		close(proofList)
	}
	mwg.Done()
}

// Launch a worker
// i.e. spawn the thing that will do the calc
func (ps *permStruct) Launch(cnt int, permuteChan chan NumCol) *sync.WaitGroup {
	var coallateWg sync.WaitGroup
	runner := func() {
		for set := range permuteChan {
			ps.workerLone(set, ps.fv)
		}
		coallateWg.Done()
	}

	coallateWg.Add(cnt)
	for i := 0; i < cnt; i++ {
		go runner()
	}
	return &coallateWg
}

// runPermute runs a permutation across a supplied set of numbers
func runPermute(arrayIn NumCol, foundValues *NumMap, proofList chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to has enough CPUs for this to be an issue
	// However the framework seems to be there

	pstrct := newPermStruct(arrayIn, foundValues)
	numWorkers := 8
	pstrct.Workers(proofList, numWorkers)
	foundValues.lastNumMap()
}
func permuteN(arrayIn NumCol, foundValues *NumMap) chan SolLst {
	returnProofs := make(chan SolLst, 16)
	go runPermute(arrayIn, foundValues, returnProofs)
	return returnProofs
}
