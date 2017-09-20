package cntSlv

import (
	"log"
	"strconv"
)

// proofLst is a list of proofs
// Think of it as a lightweight map
// there is a number and the proof for that number
// it means we can densly pack items in memory
// when we don't care about the order
// or having to find a particular solution
type proofLst struct {
	intL []int
	prs  []Proof
}

// newProofLst Create a new proof list with a single proof
func newProofLst(leng int) *proofLst {
	itm := new(proofLst)
	itm.prs = make([]Proof, leng)
	itm.intL = make([]int, leng)
	return itm
}

// newProofLstMany as NewProofLst
// but for an array of integers
func newProofLstMany(arr []int) *proofLst {
	itm := new(proofLst)
	leng := len(arr)
	itm.prs = make([]Proof, 0, leng)
	itm.intL = make([]int, 0, leng)
	for _, v := range arr {
		itm.Init(v)
	}
	return itm
}

func (pl proofLst) String() string {
	retStr := "{"
	nl := ""
	for i, v := range pl.intL {
		retStr += nl + strconv.Itoa(v) + "->" + pl.prs[i].String()
		nl = ","
	}
	retStr += "}"
	return retStr
}

// newProofLstPair Creare a new proof list with jsut 2 proofs
// Useful to make sure when we are crossing 2 we are producing the most
// efficient result
func newProofLstPair(val0, val1 int, pr0, pr1 Proof) proofLst {
	//  if pr0.tmp[0] == byte(0) {
	//    log.Fatalf("Zero in buffer pr0")
	//  }
	//  if pr1.tmp[0] == byte(0) {
	//    log.Fatalf("Zero in buffer pr1")
	//  }
	return proofLst{

		intL: []int{val0, val1},
		prs:  []Proof{pr0, pr1},
	}
}

// Proofs:Get the proofs in this list
func (pl proofLst) Proofs() []Proof {
	return pl.prs
}

// Values: What values are in the list
func (pl proofLst) Values() []int {
	return pl.intL
}

// Add a value and proof for it tot he list
func (pl *proofLst) Add(val int, pr Proof) {
	pl.intL = append(pl.intL, val)
	pl.prs = append(pl.prs, pr)
}

// Init initialises a proof list by addin a new number to it
func (pl *proofLst) Init(val int) {
	tmpP := *NewProof(val)
	pl.Add(val, tmpP)
}

// Len shows the length of a proof list
func (pl proofLst) Len() int {
	if selfCheck {
		intLen := len(pl.intL)
		prLen := len(pl.prs)
		if intLen != prLen {
			log.Fatal("Proof Lengths are not equal")
		}
		return intLen
	}
	return len(pl.intL)
}
func (pl proofLst) sliceAt(loc int) []proofLst {
	// Unable to use pool as the backing array is copied into other structures
	retListP := newProofLstArray(2)
	retListP = append(retListP,
		proofLst{
			intL: pl.intL[:loc],
			prs:  pl.prs[:loc],
		},
		proofLst{
			intL: pl.intL[loc:],
			prs:  pl.prs[loc:],
		})
	return retListP
}

func (pl proofLst) check() {
	for i := range pl.prs {
		if pl.prs[i].tmp[0] == byte(0) {
			log.Fatal("Zero in pl", pl)
		}
	}
}
