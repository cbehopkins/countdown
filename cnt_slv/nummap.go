package cnt_slv

import (
	"fmt"
	"sync"
)

type NumMapAtom struct {
	a      int
	b      *Number
	report bool
}

type NumMap struct {
	nmp              map[int]*Number
	TargetSet        bool
	Target           int
        input_channel    chan NumMapAtom
	input_channel_array    chan []NumMapAtom
	done_channel     chan bool
	num_struct_queue chan *Number
	Solved           bool
	SeekShort        bool
	UseMult          bool
	SelfTest         bool
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

	return p
}
func (item *NumMap) Add(a int, b *Number) {
	var atomic NumMapAtom
	atomic.a = a
	atomic.b = b
	atomic.report = false
	item.input_channel <- atomic
}
func (item *NumMap) AddMany( b ...*Number) {
	arr := make([]NumMapAtom, len(b))
	for i,c := range b {
	        var atomic NumMapAtom
        	atomic.a = c.Val
	        atomic.b = c
        	atomic.report = false
		arr[i] = atomic
	}
	item.input_channel_array <- arr
} 

func (item *NumMap) AddSol( a SolLst) {
	arr_len := 0
        for _,b := range a {
		arr_len = arr_len + len(*b)
        }

        arr := make([]NumMapAtom,arr_len)
	i:=0
	for _,b := range a {
	        for _,c := range *b {
        	        var atomic NumMapAtom                                                                                                            
                	atomic.a = c.Val                                                                                                                 
	                atomic.b = c                                                                                                                     
        	        atomic.report = false                                                                                                            
                	//arr = append(arr,atomic)                                                                                                    
			arr[i] = atomic
			i++
	        }
	}
        item.input_channel_array <- arr                                                                                                          
} 
func (item *NumMap) Merge(a NumMap, report bool) {

	for i, v := range a.nmp {

		var atomic NumMapAtom

		atomic.a = i
		atomic.b = v
		atomic.report = report
		item.input_channel <- atomic
	}
}

func (item *NumMap) AddProc(proof_list *SolLst) {
	add_item := func (bob NumMapAtom) {
		retr, ok := item.nmp[bob.a]
		if !ok {

			item.nmp[bob.a] = bob.b
			//fmt.Printf("New Value %d\n", bob.b.Val);
			if item.TargetSet {
				//fmt.Printf("Target has been set to %d, we have:%d\n\n",item.Target,bob.a)
				if bob.a == item.Target {
					proof_string := bob.b.ProveIt()
					fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", bob.b.Val, proof_string, bob.b.ProofLen(), bob.b.difficulty)
					if !item.SeekShort {
						item.Solved = true
						//os.Exit(0)
						//item.CheckDuplicates(proof_list)
						//item.done_channel <- true
						//return
					}
				}
			}
		} else if item.SeekShort {
			//if (retr.ProofLen()>bob.b.ProofLen()) {
			if retr.difficulty > bob.b.difficulty {

				// In seek short mode, then update when it has a shorter proof
				item.nmp[bob.a] = bob.b
				//fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", bob.b.Val, bob.b.ProveIt(), bob.b.ProofLen(), bob.b.difficulty)
			}
			//if item.TargetSet && (bob.a == item.Target) {
			//	fmt.Printf("Value %d, = %s, Proof Len is %d, Difficulty is %d\n", bob.b.Val, bob.b.ProveIt(), bob.b.ProofLen(), bob.b.difficulty)
			//}
		}
	}
	waiter := new(sync.WaitGroup)
	waiter.Add(2)
	var local_lock sync.Mutex
	go func () {
		for fred := range item.input_channel_array{                                                                                                    
			local_lock.Lock()
			for _,bob := range fred {
                		add_item(bob)          
			}                                                                                                           
  			local_lock.Unlock() 
	        }
		waiter.Done()
	} ()
	go func () {
		for bob := range item.input_channel {
			local_lock.Lock()
			add_item(bob)
			local_lock.Unlock()
		}
		waiter.Done()
	} ()
	waiter.Wait()
	if item.SelfTest {
		check_return_list(*proof_list, item)
	}
	item.CheckDuplicates(proof_list)
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
	fmt.Println("Closing input_channel")
	close(item.input_channel)
	<-item.done_channel
}
func (item *NumMap) SetTarget(target int) {
	fmt.Println("Setting target to ", target)
	item.TargetSet = true
	item.Target = target
	fmt.Println("Target is now ", item.Target)
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
func (item *NumMap) generate_number_structs() {
	for {
		//var new_num_list [] Number
		//new_num_list = make([]Number, 16)
		//fmt.Println("new_num_list is::::")
		//pretty.Println(new_num_list)
		//for i,v:= range new_num_list {
		//  var ttmp *Number
		//  ttmp = &v
		//  fmt.Printf("Adding %d, %p\n",i, ttmp)
		//  item.num_struct_queue <- ttmp
		//}
		var tmp_var Number
		//var_array := make([]Number, 1024)
		//for _, itm := range var_array {
		item.num_struct_queue <- &tmp_var
		//}
	}
}
