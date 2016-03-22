package cnt_slv

import (
	"fmt"
	"log"
	"sync"

	//"github.com/tonnerre/golang-pretty"
)

type NumMapAtom struct {
	a      int
	b      *Number
	report bool
}

type NumMap struct {
	map_lock 	    sync.Mutex	// The lock on nmp
	nmp                 map[int]*Number
	TargetSet           bool
	Target              int
	input_channel       chan NumMapAtom
	input_channel_array chan []NumMapAtom
	done_channel        chan bool
	num_struct_queue    chan *Number

	NumPool_2 sync.Pool

	pool_lock sync.Mutex
	pool_num  []Number
	pool_pnt  []*Number
	pool_pos  int
	pool_cap  int
	pool_stat int

	Solved    bool
	SeekShort bool
	UseMult   bool
	SelfTest  bool
}

func NewNumMap(proof_list *SolLst) *NumMap {
	p := new(NumMap)
	p.nmp = make(map[int]*Number)
	p.input_channel = make(chan NumMapAtom, 1000)
	p.input_channel_array = make(chan []NumMapAtom, 100)
	p.done_channel = make(chan bool)
	p.TargetSet = false

	go p.AddProc(proof_list)

	p.num_struct_queue = make(chan *Number, 1024)
	//go p.generate_number_structs()

	p.NumPool_2 = sync.Pool{
		New: func() interface{} {
			return p.NewPoolI(2)
		},
	}

	p.pool_cap = 16
	p.pool_num = make([]Number, p.pool_cap)
	p.pool_pnt = make([]*Number, p.pool_cap)
	for i, _ := range p.pool_num {
		j := &p.pool_num[i]
		//fmt.Printf("Init %x,Pointer %p\n", i,j)
		p.pool_pnt[i] = j
	}

	return p
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
	if (num_to_make ==2 ) {
		return nm.NumPool_2.Get().(NumCol)
	} else {
	return nm.NewPoolI(num_to_make)
	}
}

func (nm *NumMap) aquire_numbers_pool(num_to_make int) NumCol {
	log.Fatal("aquire_numbers_pool should not be used - FIXME and remove it")
	// This function seems like it would be a good idea to reduce load on the malloc
	// However it seems to make the garbage collector work harder
	// Which ends up costing us more
	//fmt.Println("Calling aquire_numbers:", num_to_make)
	nm.pool_lock.Lock()
	defer nm.pool_lock.Unlock()
	// TBD make this more efficient by using up the last elements in the previous structure
	num_in_queue := nm.pool_cap - nm.pool_pos

	if num_to_make > nm.pool_cap {
		log.Fatal("Requested us to make ", nm.pool_cap)
	} else if num_in_queue < num_to_make {
		nm.pool_num = make([]Number, nm.pool_cap)
		nm.pool_pnt = make([]*Number, nm.pool_cap)

		//fmt.Printf("Reallocared at %x\n", nm.pool_stat)
		for i, _ := range nm.pool_num {
			j := &nm.pool_num[i]
			nm.pool_pnt[i] = j
		}
		nm.pool_stat = 0
		nm.pool_pos = 0
	} else {
		//fmt.Printf("Saved allocation\n")
		nm.pool_stat++
	}
	new_end := nm.pool_pos + num_to_make

	//tmp_list := nm.pool_num[nm.pool_pos:new_end]
	//ret_list := nm.pool_pnt[nm.pool_pos:new_end]
	old_pool_pos := nm.pool_pos
	//fmt.Printf("Using position %x and %x\n", old_pool_pos, new_end)
	//for i,j := range nm.pool_pnt[old_pool_pos:new_end] {
	//	fmt.Printf("It's %x Pointer %p\n", i,j)
	//}

	nm.pool_pos = new_end
	//nm.pool_lock.Unlock()
	return nm.pool_pnt[old_pool_pos:new_end]
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

func (item *NumMap) AddSol(a SolLst) {
	arr_len := 0
	for _, b := range a {
		arr_len = arr_len + len(*b)
	}

	arr := make([]NumMapAtom, arr_len)
	i := 0
	//item.map_lock.Lock()
	for _, b := range a {
		for _, c := range *b {
			var atomic NumMapAtom
			atomic.a = c.Val
			atomic.b = c
			atomic.report = false
			//item.add_item(atomic)
			arr = append(arr,atomic)
			arr[i] = atomic
			i++
		}
	}
	//item.map_lock.Unlock()
	item.input_channel_array <- arr
}
func (item *NumMap) Merge(a *NumMap, report bool) {
	for i, v := range a.nmp {

		var atomic NumMapAtom

		atomic.a = i
		atomic.b = v
		atomic.report = report
		item.input_channel <- atomic
	}
}

func (item *NumMap) add_item (bob NumMapAtom) {
		retr, ok := item.nmp[bob.a]
		if !ok {
			item.nmp[bob.a] = bob.b
			if item.TargetSet {
				if bob.a == item.Target {
					proof_string := bob.b.ProveIt()
					fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", bob.b.Val, proof_string, bob.b.ProofLen(), bob.b.difficulty)
					if !item.SeekShort {
						item.Solved = true
					}
				}
			}
		} else if item.SeekShort && (retr.difficulty > bob.b.difficulty) {
			// In seek short mode, then update when it has a shorter proof
			item.nmp[bob.a] = bob.b
			//fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", bob.b.Val, bob.b.ProveIt(), bob.b.ProofLen(), bob.b.difficulty)
		}
	}
func (item *NumMap) AddProc(proof_list *SolLst) {
	waiter := new(sync.WaitGroup)
	waiter.Add(2)
	go func() {
		for fred := range item.input_channel_array {
			item.map_lock.Lock()
			for _, bob := range fred {
				item.add_item(bob)
			}
			item.map_lock.Unlock()
		}
		waiter.Done()
	}()
	go func() {
		for bob := range item.input_channel {
			item.map_lock.Lock()
			item.add_item(bob)
			item.map_lock.Unlock()
		}
		waiter.Done()
	}()
	waiter.Wait()
	if item.SelfTest {
		check_return_list(*proof_list, item)
		item.CheckDuplicates(proof_list)
	}
	item.done_channel <- true
}
func (item *NumMap) GetVals() []int {
	ret_list := make([]int, len(item.nmp))
	//fmt.Printf("\nThere are %d in list\n",len(item.nmp))
	i := 0
	for _, v := range item.nmp {
		//fmt.Printf("v:%d,%d\n",i, v.Val);
		ret_list[i] = v.Val
		i++
	}
	return ret_list
}

func (item *NumMap) CheckDuplicates(proof_list *SolLst) {
	// Each item in proof_list is a list of numbers
	// It's possible the same number list could be repeated
	// Delete these duplicates

	set_list_map := make(map[string]NumCol)
	//fmt.Printf("Checking for duplicates in Proof\n");
	var tpp SolLst
	tpp = *proof_list
	var del_queue []int

	for i := 0; i < len(tpp); i++ {
		v := tpp[i]
		var t0 NumCol
		t0 = *v
		string := t0.GetNumCol()
		//fmt.Printf("Formatted into %s\n", string);
		_, ok := set_list_map[string]
		if !ok {
			set_list_map[string] = t0
		} else {
			//fmt.Printf("%s already exists, Length %d\n:", string,len(tpp));
			//pretty.Println(t1)
			//fmt.Printf("It is now, %d", i);
			//pretty.Println(t0);
			del_queue = append(del_queue, i)
		}
	}

	for i := len(del_queue); i > 0; i-- {
		//fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
		v := del_queue[i-1]
		//fmt.Println("You've asked to delete",v);
		l1 := *proof_list
		*proof_list = append(l1[:v], l1[v+1:]...)
	}

}
func (item *NumMap) LastNumMap() {
	//fmt.Println("Closing input_channel")
	close(item.input_channel_array)
	close(item.input_channel)
	<-item.done_channel
}
func (item *NumMap) SetTarget(target int) {
	fmt.Println("Setting target to ", target)
	item.TargetSet = true
	item.Target = target
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

		//      proof_string := v.ProveIt()
		//      fmt.Printf("Value %d, = %s\n",Value, proof_string);
		//pretty.Println(w);
	}
	for i := min_num; i <= max_num; i++ {
		Value, ok := item.nmp[i]
		if ok {
			proof_string := Value.ProveIt()
			fmt.Printf("Value %d, = %s, difficulty = %d\n", Value.Val, proof_string, Value.difficulty)
		}
	}
	fmt.Printf("There are:\n%d Numbers\nMin:%4d Max:%4d\n", num_num, min_num, max_num)
}
