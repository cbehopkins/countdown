package cntSlv

import (
	"fmt"
	"log"
	"sync"
)

// nummap.go is our top level map of all the numbers we have generated
// every time a new number is generated, this is told about it
// It's also used for our top level config as it gets sent everywhere
// There are also a bunch of helper functions surrounding the map here
// to efficiently and concisely extract the needed data

// NumMapAtom is the structure that holds the Number itself
type NumMapAtom *Number

// NumMap is the main map to a number and how we get there
// This is the main structure the solver adds numbers to
type NumMap struct {
	mapLock   sync.RWMutex    // The lock on nmp
	nmp       map[int]*Number // Only our internal worker function should operate on this
	TargetSet bool
	Target    int
	// When there are numbers to add, we queue them on the channel so
	// we can process them in batches
	inputChannel      chan NumMapAtom
	inputChannelArray chan []NumMapAtom

	// The constants are locked separately from the main stuff
	// This should be refactored to be a separate struct cause that's &^%$
	constLk     sync.RWMutex
	solved      *bool
	SeekShort   bool
	UseMult     bool
	SelfTest    bool
	PermuteMode int
}

// Duplicate returns a copy of the source
// FIXME so why not call it Copy then?
func (nmp *NumMap) Duplicate() *NumMap {
	itm := NewNumMap()

	// FIXME
	// Deep copy any numbers in the struct too
	itm.TargetSet = nmp.TargetSet
	itm.Target = nmp.Target

	itm.SeekShort = nmp.SeekShort
	itm.UseMult = nmp.UseMult
	itm.SelfTest = nmp.SelfTest
	return itm
}

// NewNumMap creates a new number map
// This maps from a desired number, to how we get it
func NewNumMap() *NumMap {
	p := new(NumMap)
	p.nmp = make(map[int]*Number)
	p.solved = new(bool)
	p.inputChannel = make(chan NumMapAtom, 1000)
	p.inputChannelArray = make(chan []NumMapAtom, 100)
	go p.addWorker()
	return p
}

// Solved returns true if the Target has been found
func (nmp *NumMap) Solved() bool {
	if nmp.SeekShort {
		nmp.constLk.RLock()
		defer nmp.constLk.RUnlock()
		return *nmp.solved
	}
	// FIXME This is messy - why not just always use the above?
	nmp.mapLock.RLock()
	defer nmp.mapLock.RUnlock()
	_, ok := nmp.nmp[nmp.Target]
	return ok
}

// Keys returns the integers that form the numbers
func (nmp *NumMap) Keys() []int {
	nmp.mapLock.RLock()
	defer nmp.mapLock.RUnlock()
	retList := make([]int, len(nmp.nmp))
	i := 0
	for key := range nmp.nmp {
		if key == 0 {
			fmt.Println("WTF")
		}
		retList[i] = key
		i++
	}
	return retList
}

// Numbers returns a list of the numbers found
func (nmp *NumMap) Numbers() []*Number {
	retList := make([]*Number, len(nmp.nmp))
	i := 0
	for _, val := range nmp.nmp {
		retList[i] = val
		i++
	}
	return retList
}

// compare two number maps return true if they contain the same numbers
func (nmp *NumMap) compare(can *NumMap) bool {
	for _, key := range can.Keys() {
		_, ok := nmp.nmp[key]
		if !ok {
			log.Println("Cannot find", key, "in reference")
			log.Println(can.GetProof(key))
			return false
		}
	}

	for _, key := range nmp.Keys() {
		_, ok := can.nmp[key]
		if !ok {
			log.Println("Cannot find ", key, "in candidate")
			return false
		}
	}
	return true
}

// Add a number we have found
// and how we found it
func (nmp *NumMap) Add(b *Number) {
	if nmp.SelfTest && (b.Val == 0) {
		fmt.Println("We should not add 0")
	}

	nmp.inputChannel <- NumMapAtom(b)
}

// addMany allows adding several number at once
// It only takes a single lock to do a number of items
func (nmp *NumMap) addMany(b ...*Number) {
	arr := make([]NumMapAtom, len(b))
	for i, c := range b {
		if c == nil {
			continue
		}
		if nmp.SelfTest && c.Val == 0 {
			fmt.Println("We should not add many 0")
		}
		arr[i] = c
	}
	if nmp.SelfTest {
		for i, v := range arr {
			if v != nil && v.Val == 0 {
				fmt.Println("Bugger b:", i)
			}
		}
	}
	nmp.inputChannelArray <- arr
}

// addSol adds a solution to the map
func (nmp *NumMap) addSol(a SolLst, report bool) {
	nmp.mapLock.Lock()
	nmp.constLk.RLock()
	for _, b := range a {
		for _, c := range b {
			if c == nil {
				// With the pre-allocated map, then we end up with some nil numbers
				continue
			}
			if nmp.SelfTest && c.Val == 0 {
				log.Fatal("logic error receiving 0s")
				continue
			}
			//fmt.Println("Ading Value:", c.Val)

			nmp.addItem(c, false)
		}
	}
	nmp.constLk.RUnlock()
	nmp.mapLock.Unlock()
}

func (nmp *NumMap) addItem(stct *Number, report bool) {
	// The lock on the map structure must be grabbed outside
	value := stct.Val
	if nmp.SelfTest && value == 0 {
		fmt.Println("We should not add 0")
	}
	retr, ok := nmp.nmp[value]
	if !ok {
		if nmp.TargetSet {
			if value == nmp.Target {
				// Store the solution we found
				nmp.nmp[value] = stct

				// Seeking the shortest, means run every combination we can
				if !nmp.SeekShort {
					nmp.constLk.RUnlock()
					nmp.constLk.Lock()
					*nmp.solved = true
					nmp.constLk.Unlock()
					nmp.constLk.RLock()
				}
			}
		} else {
			// When there is no target, the we care about every solution
			nmp.nmp[value] = stct
		}
	} else if nmp.SeekShort && (retr.difficulty > stct.difficulty) {
		// In seek short mode, then update when it has a shorter proof
		nmp.nmp[value] = stct
	}
}

// addWorker listens on the channels and populates the main map
func (nmp *NumMap) addWorker() {
	waiter := new(sync.WaitGroup)
	waiter.Add(2)
	go func() {
		for numberBlock := range nmp.inputChannelArray {
			nmp.mapLock.Lock()
			nmp.constLk.RLock()
			// Adding a number is an expensive task
			// so we grab a lock, and do several at once
			for _, number := range numberBlock {
				if number == nil || number.Val == 0 {
					continue
				}
				nmp.addItem(number, false)
			}
			nmp.constLk.RUnlock()
			nmp.mapLock.Unlock()
		}
		waiter.Done()
	}()
	go func() {
		for number := range nmp.inputChannel {
			nmp.mapLock.Lock()
			nmp.constLk.RLock()
			// We use this when we create a lone number
			nmp.addItem(number, false)
			nmp.constLk.RUnlock()
			nmp.mapLock.Unlock()
		}
		waiter.Done()
	}()
	waiter.Wait()
}

// lastNumMap says we have done adding numbers
// Should be internal use only - FIXME
func (nmp *NumMap) lastNumMap() {
	close(nmp.inputChannelArray)
	close(nmp.inputChannel)
}

// SetTarget for the search
// Failing to set this means all permutations will always run
func (nmp *NumMap) SetTarget(target int) {
	nmp.constLk.Lock()
	nmp.TargetSet = true
	nmp.Target = target
	nmp.constLk.Unlock()
}

// PrintProofs prints all of the proofs
// we have found across all runs on this map
func (nmp *NumMap) PrintProofs() {
	minNum := 1000
	maxNum := 0
	numNum := 0
	for _, v := range nmp.nmp {
		// w is *Number
		var Value int
		Value = v.Val
		numNum++
		if Value > maxNum {
			maxNum = Value
		}
		if Value < minNum {
			minNum = Value
		}
	}
	for i := minNum; i <= maxNum; i++ {
		Value, ok := nmp.nmp[i]
		if ok && (i < 1000) {
			proofString := Value.String()
			fmt.Printf("Value %d, = %s, difficulty = %d\n", Value.Val, proofString, Value.difficulty)
		}
	}
	fmt.Printf("There are:\n%d Numbers\nMin:%4d Max:%4d\n", numNum, minNum, maxNum)
}

// GetProof returns a specific proof for a target
func (nmp *NumMap) GetProof(target int) string {
	val, ok := nmp.nmp[target]
	if ok {
		if nmp.SelfTest {
			_ = val.ProveSol() // This does its own error reporting
		}
		return val.String()
	}
	return "No Proof Found"
}
