package cntSlv

import "strconv"

// SolLst A solution List is a number of things you do with a set of numbers
type SolLst []NumCol

func (sl SolLst) String() string {
	var retVal string
	if len(sl) > 0 {
		for _, v := range sl {
			// v is *NumCol
			for _, w := range v {
				// w is *Number
				var Value int
				Value = w.Val

				//ret_val = ret_val + fmt.Sprintf("Value %3d, = ", Value) + w.String() + "\n"
				retVal = retVal + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
		retVal = retVal + "Done printing proofs\n"
	} else {
		retVal = "No proofs found"
	}
	return retVal
}

// Len of the count of number of solutions
func (sl SolLst) Len() int {
	return len(sl)
}

// Exists Does a value exist in the solution
func (sl SolLst) Exists(val int) bool {

	for _, v := range sl {
		for _, w := range v {
			// w is *Number
			if w == nil {
				continue
			}
			var Value int
			Value = w.Val
			if Value == val {
				return true
			}
		}
	}
	return false
}

// StringNum return the string for the supplied number
func (sl SolLst) StringNum(val int) string {
	var retVal string
	for _, v := range sl {
		for _, w := range v {
			// w is *Number
			if w == nil {
				continue
			}
			var Value int
			Value = w.Val
			if Value == val {
				retVal = retVal + "Value " + strconv.Itoa(Value) + ", = " + w.String() + "\n"
			}
		}
	}
	return retVal
}

// RemoveDuplicates from the list
func (sl *SolLst) RemoveDuplicates() {
	// The purpose of this is to go through the supplied list
	// and modify the list to only include unique sets
	// any sets that produce the same string are considered identical
	// that is the collection contains the same values
	if false {
		solMap := make(map[string]NumCol)
		var delQueue []int
		for i := 0; i < len(*sl); i++ {
			var v NumCol
			var t SolLst
			t = *sl
			v = t[i]
			str := v.String()

			_, ok := solMap[str]
			if !ok {
				//fmt.Println("Added ", v)
				solMap[str] = v
			} else {
				//fmt.Printf("%s already exists\n", string)
				//pretty.Println(t1)
				//fmt.Printf("It is now, %d", i);
				//pretty.Println(t0);
				delQueue = append(delQueue, i)
			}
		}

		for i := len(delQueue); i > 0; i-- {
			//fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
			v := delQueue[i-1]
			//fmt.Println("You've asked to delete",v);
			l1 := *sl
			*sl = append(l1[:v], l1[v+1:]...)
		}
		//fmt.Printf("In Check, OrigLen %d, New Len %d\n",orig_len,len(*item))
	}
}

// Tidy up the list
func (sl SolLst) Tidy() {
	for _, v := range sl {
		v.Tidy()
	}
}
