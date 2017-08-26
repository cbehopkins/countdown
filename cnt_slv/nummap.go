package cntSlv

import (
	"fmt"
	"sync"
)

// nummap.go is our top level map of all the numbers we have generated
// every time a new number is generated, this is told about it
// It's also used for our top level config as it gets sent everywhere
// There are also a bunch of helper functions surrpunding the map here
// to efficiently and concisely extract the needed data
type NumMapAtom struct {
	a      int
	b      *Number
	report bool
}

type NumMap struct {
	mapLock   sync.RWMutex // The lock on nmp
	nmp       map[int]*Number
	TargetSet bool
	Target    int
	// When there are numbers to add, we queue them on the channel so
	// we can process them in batches
	inputChannel      chan NumMapAtom
	inputChannelArray chan []NumMapAtom
	doneChannel       chan bool

	constLk     sync.RWMutex
	solved      *bool
	SeekShort   bool
	UseMult     bool
	SelfTest    bool
	PermuteMode int
}

func NewNumMap() *NumMap {
	p := new(NumMap)
	p.nmp = make(map[int]*Number)
	p.solved = new(bool)
	p.inputChannel = make(chan NumMapAtom, 1000)
	p.inputChannelArray = make(chan []NumMapAtom, 100)
	p.doneChannel = make(chan bool)
	p.TargetSet = false
	go p.AddProc()
	return p
}
func (nmp *NumMap) Solved() bool {
	nmp.constLk.RLock()
	defer nmp.constLk.RUnlock()
	return *nmp.solved
}
func (nmp *NumMap) Keys() []int {
	retList := make([]int, len(nmp.nmp))
	i := 0
	for key := range nmp.nmp {
		retList[i] = key
		i++
	}
	return retList
}
func (nmp *NumMap) Numbers() []*Number {
	retList := make([]*Number, len(nmp.nmp))
	i := 0
	for _, val := range nmp.nmp {
		retList[i] = val
		i++
	}
	return retList
}

func (ref *NumMap) Compare(can *NumMap) bool {
	pass := true
	// Compare two number maps return true if they contain the same numbers
	for _, key := range can.Keys() {
		_, ok := ref.nmp[key]
		if !ok {
			fmt.Printf("The value %d was in the candidate, but not the reference\n", key)
			//return false
			pass = false
		}
	}

	for _, key := range ref.Keys() {
		_, ok := can.nmp[key]
		if !ok {
			fmt.Printf("The value %d was in the reference, but not the candidate\n", key)
			//return false
			pass = false
		}
	}
	return pass
}
func (nm *NumMap) NewPoolI(numToMake int) NumCol {
	poolNum := make([]Number, numToMake)
	poolPnt := make([]*Number, numToMake)
	for i := range poolNum {
		j := &poolNum[i]
		poolPnt[i] = j
	}
	return poolPnt
}

func (nm *NumMap) acquireNumbers(numToMake int) NumCol {
	return nm.NewPoolI(numToMake)
}

func (item *NumMap) Add(a int, b *Number) {
	var atomic NumMapAtom
	atomic.a = b.Val
	atomic.b = b
	atomic.report = false
	item.inputChannel <- atomic
}
func (item *NumMap) AddMany(b ...*Number) {
	arr := make([]NumMapAtom, len(b))
	for i, c := range b {
		var atomic NumMapAtom
		atomic.a = c.Val
		atomic.b = c
		atomic.report = false
		arr[i] = atomic
	}
	item.inputChannelArray <- arr
}

func (item *NumMap) AddSol(a SolLst, report bool) {
	item.mapLock.Lock()
	item.constLk.RLock()
	for _, b := range a {
		for _, c := range b {
			//fmt.Println("Ading Value:", c.Val)
			item.addItem(c.Val, c, false)
		}
	}
	item.constLk.RUnlock()
	item.mapLock.Unlock()
}
func (item *NumMap) Merge(a *NumMap, report bool) {
	a.mapLock.Lock()
	tmpCol := make(NumCol, len(a.nmp))
	i := 0
	for _, v := range a.nmp {
		tmpCol[i] = v
		i++
	}
	a.mapLock.Unlock()
	tmpSol := SolLst{tmpCol}
	item.AddSol(tmpSol, report)
}

func (item *NumMap) addItem(value int, stct *Number, report bool) {
	// The lock on the map structure must be grabbed outside
	retr, ok := item.nmp[value]
	if !ok {
		//item.nmp[value] = stct
		if item.TargetSet {
			if value == item.Target {
				// Store the solution we found
				item.nmp[value] = stct

				//proof_string := stct.String()
				//fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", value, proof_string, stct.ProofLen(), stct.difficulty)
				// Seeking the shortest, means run every combination we can
				if !item.SeekShort {
					item.constLk.RUnlock()
					item.constLk.Lock()
					*item.solved = true
					item.constLk.Unlock()
					item.constLk.RLock()
				}
				fmt.Println("Set Solved sucessfully")
			}
		} else {
			// When there is no target, the we care about every solution
			item.nmp[value] = stct
		}
	} else if item.SeekShort && (retr.difficulty > stct.difficulty) {
		// In seek short mode, then update when it has a shorter proof
		item.nmp[value] = stct
	}
}
func (item *NumMap) AddProc() {
	waiter := new(sync.WaitGroup)
	waiter.Add(2)
	go func() {
		for fred := range item.inputChannelArray {
			item.mapLock.Lock()
			item.constLk.RLock()
			for _, bob := range fred {
				item.addItem(bob.a, bob.b, false)
			}
			item.constLk.RUnlock()
			item.mapLock.Unlock()
		}
		waiter.Done()
	}()
	go func() {
		for bob := range item.inputChannel {
			item.mapLock.Lock()
			item.constLk.RLock()
			item.addItem(bob.a, bob.b, false)
			item.constLk.RUnlock()
			item.mapLock.Unlock()
		}
		waiter.Done()
	}()
	waiter.Wait()
	close(item.doneChannel)

}
func (item *NumMap) GetVals() []int {
	retList := make([]int, len(item.nmp))
	i := 0
	for _, v := range item.nmp {
		//fmt.Printf("v:%d,%d\n",i, v.Val);
		retList[i] = v.Val
		i++
	}
	return retList
}

func (item *NumMap) LastNumMap() {
	//fmt.Println("Closing input_channel")
	close(item.inputChannelArray)
	close(item.inputChannel)
	<-item.doneChannel
}
func (item *NumMap) SetTarget(target int) {
	//fmt.Println("Setting target to ", target)
	item.constLk.Lock()
	item.TargetSet = true
	item.Target = target
	item.constLk.Unlock()
}
func (item *NumMap) PrintProofs() {
	minNum := 1000
	maxNum := 0
	numNum := 0
	for _, v := range item.nmp {
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
		Value, ok := item.nmp[i]
		if ok && (i < 1000) {
			proofString := Value.String()
			fmt.Printf("Value %d, = %s, difficulty = %d\n", Value.Val, proofString, Value.difficulty)
		}
	}
	fmt.Printf("There are:\n%d Numbers\nMin:%4d Max:%4d\n", numNum, minNum, maxNum)
}
func (item *NumMap) GetProof(target int) string {
	val, ok := item.nmp[target]
	if ok {
		return val.String()
	} else {
		return "No Proof Found"
	}
}
