package cntSlv

import (
	"log"
	"strconv"
	"sync"
)

// Proof A proof is saying how we got a number from a set of other Numbers
type Proof struct {
	tmp []byte
}

// NewProof Create a single new proof from an integer
func NewProof(input int) *Proof {
	tmpBufArr := []byte(strconv.Itoa(input))
	tmp := tmpBufArr
	return &Proof{tmp: tmp}
}
func (pr Proof) String() string {
	return string(pr.tmp)
}
func (pr *Proof) set(prIn Proof) {
	if (pr.tmp == nil) || (cap(pr.tmp) == 0) {
		pr.tmp = newBa(prIn.Len())
	}

	pr.tmp = pr.tmp[:prIn.Len()]
	_ = copy(pr.tmp, prIn.tmp)
}
func (pr *Proof) merge(inp Proof) {
	oldProofLen := pr.Len()
	newProofLen := inp.Len()
	if oldProofLen == 0 || (oldProofLen > newProofLen) {
		pr.set(inp)
	}
}

// concat returns a new proof that is the result of
// two proofs and an operator
func (pr Proof) concat(input Proof, op operator) Proof {
	if selfCheck && pr.tmp == nil {
		log.Fatal("Concat to nil error")
	}
	if selfCheck && pr.tmp[0] == byte(0) {
		log.Fatalf("Zero in buffer")
	}
	var tmp []byte
	capNeeded := bLen + opLen + len(pr.tmp) + len(input.tmp)
	tmp = getBa(capNeeded)
	tmp = pr.concatCore(input, op, tmp, capNeeded)
	return Proof{tmp: tmp}
}

// The core of the concatonate functionality
// By this I mean that of taking two proofs and creating a new one
// that is the result of them being concatonated with an operator
func (pr Proof) concatCore(input Proof, op operator, byteStore []byte, capNeeded int) []byte {
	tmpOp := op.Bytes()
	//  if false {
	//    byteStore = append(byteStore, Ob...)
	//    byteStore = append(byteStore, pr.tmp...)
	//    byteStore = append(byteStore, tmpOp...)
	//    byteStore = append(byteStore, input.tmp...)
	//    byteStore = append(byteStore, Cb...)
	//    if SelfCheck {
	//      if len(byteStore) != capNeeded {
	//        log.Fatal("Made wrong capacity!", len(byteStore), capNeeded)
	//      }
	//    }

	//  } else {
	if cap(byteStore) == 0 {
		byteStore = getBa(capNeeded)
	}
	//    if SelfCheck {
	//      if capNeeded > cap(byteStore) {
	//        log.Fatal("Not Long enough", capNeeded, cap(byteStore))
	//      }
	//    }
	byteStore = byteStore[0:capNeeded]
	total := copy(byteStore[0:], oB)
	total += copy(byteStore[total:], pr.tmp)
	total += copy(byteStore[total:], tmpOp)
	total += copy(byteStore[total:], input.tmp)
	total += copy(byteStore[total:], cB)
	//    if SelfCheck && total != capNeeded {
	//      log.Fatal("Total wrong", total, tmp)
	//    }
	//}
	//  if SelfCheck {
	//    if string(pr.tmp) == "" || string(pr.tmp) == " " {
	//      log.Fatal("Found nil,", pr)
	//    }
	//    checkString := "(" + string(pr.tmp) + op.String() + string(input.tmp) + ")"
	//    if checkString != string(tmp) {
	//      log.Fatalf("%q!=%q\n", checkString, string(tmp))
	//    }
	//  }
	return byteStore
}

// Valid Is the selected proof valid
func (pr Proof) Valid() bool {
	if pr.tmp == nil {
		return true
	}
	return len(pr.tmp) != 0
}

// Len retuens the length of a proof
func (pr Proof) Len() int {
	return len(pr.tmp)
}

// Exists Test a proof list to see if a value has been found
func (pr Proof) Exists() bool {
	return pr.Len() > 0
}

// Clear the structure returning the buffer to the pool
func (pr *Proof) Clear() {
	if pr.tmp != nil {
		putBa(pr.tmp)
		//    for i := range pr.tmp {
		//      pr.tmp[i] = 0
		//    }
	}
	pr.tmp = nil
}

// Reset the proof keeping the buffers
func (pr *Proof) Reset() {
	if pr.Valid() {
		pr.tmp = pr.tmp[0:0]
	}
}

//////////////////////////////////
// Keep the memory pooling stuff at the end
// Very useful, but not interesting to main logic
var poolBa *sync.Pool
var nulPool sync.Pool

var bLen int
var maxCap int

func initProof() {
	bLen = len(oB) * 2
	maxCap = (maxInputs * 4) + ((bLen + opLen) * (maxInputs - 1))
	poolBa = newBaPool(maxCap)
	nulPool = sync.Pool{}
}
func getBa(capa int) []byte {
	if false {
		tmp := poolBa.Get().([]byte)
		//  if cap(tmp) != MaxCap {
		//    log.Fatal("Weird Buffer on pool", string(tmp), cap(tmp), len(tmp))
		//  }
		//  if len(tmp) != 0 {
		//    log.Fatal("Weird Buffer on pool", string(tmp), cap(tmp), len(tmp))
		//  }
		return tmp
	}
	return newBa(capa)

}

func putBa(ba []byte) {
	if useBa {
		if cap(ba) == maxCap {
			ba = ba[0:0]
			poolBa.Put(ba)
		} else {
			// Help the GC
			// Put useless things directly into the pool
			// where they will be removed
			nulPool.Put(ba)
		}
	}
}

func newBa(capa int) []byte {
	return make([]byte, 0, maxCap)
}
func newBaPool(capa int) *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return newBa(capa)
		},
	}
}
