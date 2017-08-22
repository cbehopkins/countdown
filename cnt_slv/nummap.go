package cnt_slv

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
	map_lock            sync.RWMutex // The lock on nmp
	nmp                 map[int]*Number
	TargetSet           bool
	Target              int
	input_channel       chan NumMapAtom
	input_channel_array chan []NumMapAtom
	done_channel        chan bool
	num_struct_queue    chan *Number

	NumPool_2 sync.Pool

	const_lk    sync.RWMutex
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
	p.input_channel = make(chan NumMapAtom, 1000)
	p.input_channel_array = make(chan []NumMapAtom, 100)
	p.done_channel = make(chan bool)
	p.TargetSet = false

	go p.AddProc()

	p.num_struct_queue = make(chan *Number, 1024)

	p.NumPool_2 = sync.Pool{
		New: func() interface{} {
			return p.NewPoolI(2)
		},
	}

	return p
}
func (nmp *NumMap) Solved() bool {
	nmp.const_lk.RLock()
	defer nmp.const_lk.RUnlock()
	return *nmp.solved
}
func (nmp *NumMap) Keys() []int {
	ret_list := make([]int, len(nmp.nmp))
	i := 0
	for key, _ := range nmp.nmp {
		ret_list[i] = key
		i++
	}
	return ret_list
}
func (nmp *NumMap) Numbers() []*Number {
	ret_list := make([]*Number, len(nmp.nmp))
	i := 0
	for _, val := range nmp.nmp {
		ret_list[i] = val
		i++
	}
	return ret_list
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
func (nm *NumMap) NewPoolI(num_to_make int) NumCol {
	pool_num := make([]Number, num_to_make)
	pool_pnt := make([]*Number, num_to_make)
	for i, _ := range pool_num {
		j := &pool_num[i]
		pool_pnt[i] = j
	}
	return pool_pnt
}

func (nm *NumMap) aquire_numbers(num_to_make int) NumCol {
	if num_to_make == 2 {
		return nm.NumPool_2.Get().(NumCol)
	} else {
		return nm.NewPoolI(num_to_make)
	}
}

func (item *NumMap) Add(a int, b *Number) {
	var atomic NumMapAtom
	atomic.a = b.Val
	atomic.b = b
	atomic.report = false
	item.input_channel <- atomic
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
	item.input_channel_array <- arr
}

func (item *NumMap) AddSol(a SolLst, report bool) {
	item.map_lock.Lock()
	item.const_lk.RLock()
	for _, b := range a {
		for _, c := range b {
			//fmt.Println("Ading Value:", c.Val)
			item.add_item(c.Val, c, false)
		}
	}
	item.const_lk.RUnlock()
	item.map_lock.Unlock()
}
func (item *NumMap) Merge(a *NumMap, report bool) {
	a.map_lock.Lock()
	tmp_col := make(NumCol, len(a.nmp))
	i := 0
	for _, v := range a.nmp {
		tmp_col[i] = v
		i++
	}
	a.map_lock.Unlock()
	tmp_sol := SolLst{tmp_col}
	item.AddSol(tmp_sol, report)
}

func (item *NumMap) add_item(value int, stct *Number, report bool) {
	// The lock on the map structure must be grabbed outside
	retr, ok := item.nmp[value]
	if !ok {
		item.nmp[value] = stct
		if item.TargetSet {
			if value == item.Target {

				proof_string := stct.String()
				fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", value, proof_string, stct.ProofLen(), stct.difficulty)
				// Seeking the shortest, means run every combination we can
				if !item.SeekShort {
					item.const_lk.RUnlock()
					item.const_lk.Lock()
					*item.solved = true
					item.const_lk.Unlock()
					item.const_lk.RLock()
				}
				fmt.Println("Set Solved sucessfully")
			}
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
		for fred := range item.input_channel_array {
			item.map_lock.Lock()
			item.const_lk.RLock()
			for _, bob := range fred {
				item.add_item(bob.a, bob.b, false)
			}
			item.const_lk.RUnlock()
			item.map_lock.Unlock()
		}
		waiter.Done()
	}()
	go func() {
		for bob := range item.input_channel {
			item.map_lock.Lock()
			item.const_lk.RLock()
			item.add_item(bob.a, bob.b, false)
			item.const_lk.RUnlock()
			item.map_lock.Unlock()
		}
		waiter.Done()
	}()
	waiter.Wait()
	close(item.done_channel)

}
func (item *NumMap) GetVals() []int {
	ret_list := make([]int, len(item.nmp))
	i := 0
	for _, v := range item.nmp {
		//fmt.Printf("v:%d,%d\n",i, v.Val);
		ret_list[i] = v.Val
		i++
	}
	return ret_list
}

func (item *NumMap) LastNumMap() {
	//fmt.Println("Closing input_channel")
	close(item.input_channel_array)
	close(item.input_channel)
	<-item.done_channel
}
func (item *NumMap) SetTarget(target int) {
	//fmt.Println("Setting target to ", target)
	item.const_lk.Lock()
	item.TargetSet = true
	item.Target = target
	item.const_lk.Unlock()
}
func (item *NumMap) PrintProofs() {
	min_num := 1000
	max_num := 0
	num_num := 0
	for _, v := range item.nmp {
		// w is *Number
		var Value int
		Value = v.Val
		num_num++
		if Value > max_num {
			max_num = Value
		}
		if Value < min_num {
			min_num = Value
		}
	}
	for i := min_num; i <= max_num; i++ {
		Value, ok := item.nmp[i]
		if ok && (i < 1000) {
			proof_string := Value.String()
			fmt.Printf("Value %d, = %s, difficulty = %d\n", Value.Val, proof_string, Value.difficulty)
		}
	}
	fmt.Printf("There are:\n%d Numbers\nMin:%4d Max:%4d\n", num_num, min_num, max_num)
}
func (item *NumMap) GetProof(target int) string {
	val, ok := item.nmp[target]
	if ok {
		return val.String()
	} else {
		return "No Proof Found"
	}
}
