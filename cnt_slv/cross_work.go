package cntSlv

import (
	"log"
	"runtime"
	"sync"
)

type crossStruct struct {
	b         []proofLst
	proofChan chan Proofs
	wg        *sync.WaitGroup
}

func crossWorker(ic chan crossStruct) {
	nLx := getProofsArray(2)
	for v := range ic {
		arbCrossCore(v.b, v.proofChan, v.wg, nLx)
		for i := range nLx {
			nLx[i].Reset()
		}
	}
	log.Fatal("Cross channel closed")
	putProofsArray(nLx)
}

var crossChan chan crossStruct

func initCrossWorkers() {
	NumWorkers := runtime.NumCPU() * 2
	crossChan = make(chan crossStruct, NumWorkers)
	for i := 0; i < NumWorkers; i++ {
		go crossWorker(crossChan)
	}
}
func arbCross(b []proofLst, proofChan chan Proofs, wg *sync.WaitGroup) {
	nLx := getProofsArray(len(b))
	arbCrossCore(b, proofChan, wg, nLx)
	putProofsArray(nLx)
}
func arbCrossCore(b []proofLst, proofChan chan Proofs, wg *sync.WaitGroup, nLx []Proofs) {
	tmpP := getProofs()
	for i := range b {
		nLx[i].wrkFast(b[i])
	}
	tmpP.recursiveCross(nLx, 0, Proof{})
	proofChan <- tmpP
	wg.Done()
}
func (prs *Proofs) recursiveCrossA(i, rv int, iP, rp Proof) {
	inP := newProofLstPair(i, rv, iP, rp)
	prs.wrkExtraFastPair(inP)
}

func (prs *Proofs) recursiveCross(pra []Proofs, refVal int, refProof Proof) {
	var toRun func(i, rv int, iP, rp Proof)

	// Do it like this so that we make the decision once
	if len(pra) == 1 {
		toRun = func(i int, rv int, iP, rp Proof) {
			//prs.recursiveCrossA(i, rv, iP, rp)
			inP := newProofLstPair(i, rv, iP, rp)
			prs.wrkExtraFastPair(inP)
		}
	} else {
		toRun = func(i int, rv int, iP, rp Proof) {
			prs.recursiveCross(pra[1:], i, iP)
		}
	}

	for iVal, iProof := range pra[0] {
		if iVal > 0 && (iProof.Len() > 0) {
			toRun(iVal, refVal, iProof, refProof)
		}
	}
}

// The only place this pool is used is here
//   so keep it local
var poolProofsArray []sync.Pool

func initPoolProofsArray() {
	poolProofsArray = make([]sync.Pool, maxInputs)
	for i := 1; i < maxInputs; i++ {
		j := i
		poolProofsArray[i] = sync.Pool{
			New: func() interface{} {
				return newProofsArray(j)
			},
		}
	}
}
func getProofsArray(cnt int) []Proofs {
	pl := poolProofsArray[cnt].Get().([]Proofs)
	if len(pl) != cnt {
		log.Fatal("Array pool returned wrong length", len(pl), cnt)
	}
	return pl
}

func putProofsArray(pla []Proofs) {
	//  for i := range pla {
	//    putProofs(pla[i])
	//  }
	cnt := len(pla)
	//  if SelfCheck {
	//    if cnt < 1 || cnt > MaxCap || cnt > len(poolProofsArray) {
	//      log.Fatal("Length problem", cnt, len(poolProofsArray))
	//    }
	//  }
	poolProofsArray[cnt].Put(pla)
}
func newProofsArray(cnt int) []Proofs {
	tmp := make([]Proofs, cnt)
	for i := 0; i < cnt; i++ {
		tmp[i] = getProofs()
	}
	return tmp
}
