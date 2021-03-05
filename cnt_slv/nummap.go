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
type NumMapAtom struct {
	a int // Document these fields when you understand them
	b *Number
}

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
	doneChannel       chan bool

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
	p.doneChannel = make(chan bool)
	p.TargetSet = false
	go p.addWorker()
	return p
}

// Solved returns true if the Target has been found
func (nmp *NumMap) Solved() bool {
	nmp.constLk.RLock()
	defer nmp.constLk.RUnlock()
	return *nmp.solved
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

// Compare two number maps return true if they contain the same numbers
func (nmp *NumMap) Compare(can *NumMap) bool {
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
func (nmp *NumMap) Add(a int, b *Number) {
	if nmp.SelfTest && (a == 0 || b.Val == 0) {
		fmt.Println("We should not add 0")
	}
	var atomic NumMapAtom
	atomic.a = b.Val
	atomic.b = b
	nmp.inputChannel <- atomic
}

// addMany allows adding several number at once
// It only takes a single lock to do a number of items
func (nmp *NumMap) addMany(b ...*Number) {
	arr := make([]NumMapAtom, len(b))
	for i, c := range b {
		if c == nil {
			continue
		}
		if c.Val == 0 {
			fmt.Println("We should not add many 0")
		}
		var atomic NumMapAtom
		atomic.a = c.Val
		atomic.b = c
		arr[i] = atomic
	}
	// for i, v := range arr {
	// 	if v.a == 0 || v.b.Val == 0 {
	// 		fmt.Println("Bugger:", i)
	// 	}
	// }
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
			if c.Val == 0 {
				fmt.Println("WTF?")
			}
			//fmt.Println("Ading Value:", c.Val)
			if c == nil {
				continue
			}
			nmp.addItem(c.Val, c, false)
		}
	}
	nmp.constLk.RUnlock()
	nmp.mapLock.Unlock()
}

// Merge allows you to merge two number maps together
// This is useful for parallel workers
func (nmp *NumMap) Merge(a *NumMap, report bool) {
	a.mapLock.Lock()
	tmpCol := make(NumCol, len(a.nmp))
	i := 0
	for _, v := range a.nmp {
		tmpCol[i] = v
		i++
	}
	a.mapLock.Unlock()
	tmpSol := SolLst{tmpCol}
	nmp.addSol(tmpSol, report)
}

func (nmp *NumMap) addItem(value int, stct *Number, report bool) {
	// The lock on the map structure must be grabbed outside
	if value == 0 {
		fmt.Println("We should not add 0")
	}
	retr, ok := nmp.nmp[value]
	if !ok {
		//item.nmp[value] = stct
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
				if number.a == 0 {
					continue
				}
				nmp.addItem(number.a, number.b, false)
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
			nmp.addItem(number.a, number.b, false)
			nmp.constLk.RUnlock()
			nmp.mapLock.Unlock()
		}
		waiter.Done()
	}()
	waiter.Wait()
	close(nmp.doneChannel)
}

// LastNumMap says we have done adding numbers
// Should be internal use only - FIXME
func (nmp *NumMap) LastNumMap() {
	close(nmp.inputChannelArray)
	close(nmp.inputChannel)
	<-nmp.doneChannel // FIXME move to closing the channel instead
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
		_ = val.ProveSol() // This does its own error reporting
		return val.String()
	}
	return "No Proof Found"
}
