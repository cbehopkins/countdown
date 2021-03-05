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

// Tidy up the list
func (sl SolLst) Tidy() {
	for _, v := range sl {
		v.Tidy()
	}
}
