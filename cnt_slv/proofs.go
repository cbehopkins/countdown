package cntSlv

import (
	"strconv"
	"sync"
)

var proofsPool sync.Pool

func initProofsPool() {
	proofsPool = sync.Pool{
		New: func() interface{} {
			return NewProofs()
		},
	}
}

// getProofs might return a proof with items on it
// It shouldn't, but this is not safe to assume
// Anyone who calls this should clear it before use
func getProofs() Proofs {
	if useProofPool {
		pl := proofsPool.Get().(Proofs)
		//    if SelfCheck {
		//      if cap(pl) != maxValue {
		//        log.Fatal("Incorrect Capacity")
		//      }
		//      //pl.Clear()
		//    }
		return pl
	}
	pl := NewProofs()
	return pl

}
func putProofs(pl Proofs) {
	//pl.Clear()
	pl.Reset()
	if useProofPool {
		proofsPool.Put(pl)
	}
}

// Proofs is an array from the answer to the proof of how to get it
// i.e. index the array at the number you want the proof for
// it is always a list of maxValue in length
type Proofs []Proof

// NewProofs A New proof structure to add found proofs to
func NewProofs() Proofs {
	itm := make(Proofs, 0, maxValue)
	return itm
}

func (prs Proofs) String() string {
	var retString string
	for i, pr := range prs {
		if pr.Exists() {
			retString += strconv.Itoa(i) + ":" + pr.String() + "\n"
		}
	}
	return retString
}

// Exists Test a proof list to see if a value has been found
func (prs Proofs) Exists(val int) bool {
	if (val > 0) && (val < len(prs)) {
		if prs[val].Len() > 0 {
			return true
		}
	}
	return false
}

// Get proof for the number
func (prs Proofs) Get(val int) Proof {
	if (val > 0) && (val < len(prs)) {
		if prs[val].Len() > 0 {
			return prs[val]
		}
	}
	return Proof{}
}

// Reset the Proofs
// returns the buffers back to the pool
func (prs *Proofs) Reset() {
	tmp := *prs
	for i := range tmp {
		tmp[i].Reset()
	}
}

// Len is the number of valid proofs
func (prs Proofs) Len() int {
	cnt := 0
	for _, v := range prs {
		if v.Len() > 0 {
			cnt++
		}
	}
	return cnt
}
func (prs *Proofs) empty() {
	tmp := *prs
	*prs = tmp[0:0]
}

// Clear exists to clear the contents of a proof
// It does not return things tot he pool
func (prs *Proofs) Clear() {
	tmp := *prs
	for i := range tmp {
		tmp[i].Clear()
	}
	//  if SelfCheck {
	//    for i := range tmp {
	//      pr := tmp[i]
	//      if pr.tmp != nil {
	//        log.Fatal("Que")
	//      }
	//    }
	//  }
}
func (prs *Proofs) extend(v int) bool {
	tmp := *prs
	savedLen := len(tmp)

	if v >= savedLen {
		newEnd := v + 1
		*prs = tmp[:newEnd]

		// Because Clear ranges over Proofs
		// it doesn't clear items in the backing array
		// so if we extend the array, we could pick them up again
		// so we MUST ensure we are clear
		tmp2 := tmp[:newEnd]
		for i := savedLen; i < newEnd; i++ {
			tmp2[i].tmp = tmp2[i].tmp[0:0]
			//tmp2[i].Clear()
		}
		return true
	}
	return false
}
