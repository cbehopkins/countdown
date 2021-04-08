package cntSlv

// SolLst A solution List is a number of things you do with a set of numbers
type SolLst []NumCol

func (sl SolLst) String() string {
	var retVal string
	if len(sl) > 0 {
		for _, v := range sl {
			for _, w := range v {
				// FIXME strings.Builder?
				retVal = retVal + w.render() + "\n"
			}
		}
		retVal = retVal + "Done printing proofs\n"
	} else {
		retVal = "No proofs found"
	}
	return retVal
}

// exists Does a value exist in the solution
func (sl SolLst) exists(val int) bool {
	return sl.Get(val) != nil
}

// Exists Does a value exist in the solution
func (sl SolLst) Get(val int) *Number {
	for _, v := range sl {
		for _, w := range v {
			// w is *Number
			if w == nil {
				continue
			}
			if w.Val == val {
				return w
			}
		}
	}
	return nil
}

// StringNum return the string for the supplied number
func (sl SolLst) StringNum(val int) string {
	return sl.Get(val).render() + "\n"
}

// Tidy up the list
func (sl SolLst) Tidy() {
	for _, v := range sl {
		v.Tidy()
	}
}
