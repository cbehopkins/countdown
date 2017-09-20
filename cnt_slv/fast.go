package cntSlv

import (
	"errors"
	"log"
	"runtime"
	"strconv"
	"sync"
)

const (
	maxInputs         = 6
	maxValue          = 1024
	selfCheck         = false
	useExtraFast      = false
	useProofListCache = true
	useProofPool      = true
	usePool           = true
	useBa             = true
)

const opLen = 1

var oB = []byte{'('}
var cB = []byte{')'}
var bLen = len(oB) * 2
var maxCap = (maxInputs * 4) + ((bLen + opLen) * (maxInputs - 1))

var proofsPool sync.Pool
var nulPool sync.Pool
var poolBa *sync.Pool
var poolProofListArray map[int]*sync.Pool
var poolProofsArray []sync.Pool

func init() {
	poolBa = newBaPool(maxCap)
	poolProofListArray = make(map[int]*sync.Pool)
	proofsPool = sync.Pool{
		New: func() interface{} {
			return NewProofs()
		},
	}
	nulPool = sync.Pool{}
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
func newProofsArray(cnt int) []Proofs {
	tmp := make([]Proofs, cnt)
	for i := 0; i < cnt; i++ {
		tmp[i] = getProofs()
	}
	return tmp
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
func newProofLstArray(capa int) []proofLst {
	tmp := make([]proofLst, 0, capa)
	return tmp
}
func getProofListArray(capa int) []proofLst {
	if useProofListCache {
		pl, ok := poolProofListArray[capa]
		if !ok || pl == nil {
			pl = &sync.Pool{
				New: func() interface{} {
					return newProofLstArray(capa)
				},
			}
			poolProofListArray[capa] = pl
		}
		tmp := pl.Get().([]proofLst)

		// reslice to be  0 length
		tmp = tmp[0:0]
		//    if SelfCheck {
		//    if (len(tmp) != 0) || (cap(tmp) != capa) {
		//      log.Fatal("Learn how to write code Chris")
		//    }}
		return tmp
	}
	pl := newProofLstArray(capa)
	return pl

}
func putProofListArray(pl []proofLst) {
	if useProofListCache {
		capa := cap(pl)
		pool, ok := poolProofListArray[capa]
		if !ok {
			log.Fatal("Error put back a length there was not a map entry for", capa)
		}
		//pl = pl[0:0]
		pool.Put(pl)
	}
}

type operator []byte

func newOperator(in string) operator {
	return operator(in)
}

func (op operator) String() string {
	return string(op)
}
func (op operator) Bytes() []byte {
	return []byte(op)
}

// Proof A proof is saying how we got a number from a set of other Numbers
type Proof struct {
	tmp []byte
}

func (pr *Proof) merge(inp Proof) {
	oldProofLen := pr.Len()
	newProofLen := inp.Len()
	if oldProofLen == 0 || (oldProofLen > newProofLen) {
		if (pr.tmp == nil) || (cap(pr.tmp) == 0) {
			pr.tmp = getBa(len(inp.tmp))
		}
		//    if SelfCheck && cap(pr.tmp) < newProofLen {
		//      log.Fatal("Array too small:", cap(pr.tmp), newProofLen, string(inp.tmp), MaxCap)
		//    }
		pr.tmp = pr.tmp[:newProofLen]
		_ = copy(pr.tmp, inp.tmp)
		//    if SelfCheck && (cnt != newProofLen) {
		//      log.Fatal("Copied Incorrect Number")
		//    }
	}
}

// NewProof Create a single new proof from an integer
func NewProof(input int) *Proof {
	tmpBufArr := []byte(strconv.Itoa(input))
	tmp := tmpBufArr
	return &Proof{tmp: tmp}
}
func (prs *Proofs) mrg3(v int, pr, input Proof, op operator) {
	tmp := *prs
	buffer := tmp[v].tmp
	capNeeded := bLen + opLen + len(pr.tmp) + len(input.tmp)
	tmp[v].tmp = pr.concatCore(input, op, buffer, capNeeded)
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
func (pr Proof) String() string {
	return string(pr.tmp)
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

// Proofs is an array from the answer to the proof of how to get it
// i.e. index the array at the number you want the proof for
// it is always a list of maxValue in length
type Proofs []Proof

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

// NewProofs A New proof structure to add found proofs to
func NewProofs() Proofs {
	itm := make(Proofs, 0, maxValue)
	return itm
}

// proofLst is a list of proofs
// It is always 4 items long
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

// Reset the Proofs
// returns the buffers back to the pool
func (prs *Proofs) Reset() {
	tmp := *prs
	for i := range tmp {
		tmp[i].Reset()
	}
}
func (pl proofLst) check() {
	for i := range pl.prs {
		if pl.prs[i].tmp[0] == byte(0) {
			log.Fatal("Zero in pl", pl)
		}
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
func (prs *Proofs) wrkFast(inPr proofLst) {
	leng := inPr.Len()
	switch leng {
	//  case 0:
	//    if SelfCheck {
	//      log.Fatal("Zero Length Lists?")
	//    }
	//    prs.Clear()
	//    return
	case 1:
		prs.InitLst(inPr)
		return
	case 2:
		prs.InitLst(inPr)
		inPr.check()
		prs.wrkExtraFastPair(inPr)
		return
	default:
		prs.InitLst(inPr)
		prs.wrkFastGen(inPr, usePool, false)
		return
	}
	//return
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
func (prs *Proofs) mrg2(v int, bob Proof) {
	//  if SelfCheck {
	//    if prs.extend(v) {
	//      log.Fatal("I don't think we should have to extend in this function")
	//    }
	//    if v < 0 && v > maxValue {
	//      log.Fatal("Extend should have sanitized this for us")
	//    }
	//    // In theory this might be needed
	//    // But since we are always run after having this checked for us
	//    // Then we can comment out for execution speed
	//    //  } else {
	//    //    prs.extend(v)
	//  }
	tmp := *prs
	tmp[v].merge(bob)
	//putBa(bob.tmp)
}

// This looks to see if this will be an interesting proof
// i.e. do we think this will be a proof we'll be intersted in
// It means we have to do more work per calculation
// But reduces the amount of garbage we generate
func (prs *Proofs) interestingProof(v int, pr0l, pr1l int) bool {
	if v > 0 && v < maxValue {
		if prs.extend(v) {
			return true
		}
		tmp := *prs
		currLen := tmp[v].Len()
		if currLen == 0 {
			return true
		}
		possLen := pr0l + pr1l + 2
		if possLen < currLen {
			return true
		}
	}
	return false
}
func (prs Proofs) set(v int, pr Proof) {
	prs[v].tmp = newBa(pr.Len())
	prs[v].tmp = prs[v].tmp[:pr.Len()]
	_ = copy(prs[v].tmp, pr.tmp)
	//  if SelfCheck && cnt > 3 {
	//    log.Fatal("Leg set", pr)
	//  }
}

// InitLst will Re-Initialise the list with just the values provided
func (prs *Proofs) InitLst(inP proofLst) {
	prfLst := inP.Proofs()
	prs.empty()

	for i, v := range inP.Values() {
		// We have cleared it so we know this will be the shortest Proof
		// Worth not reusing functions for that!
		prs.extend(v)
		prs.set(v, prfLst[i])
	}
}

// mergeLst will merge a proofLst into a proof set
func (prs *Proofs) mergeLst(inP proofLst) {
	prfLst := inP.Proofs()
	for i, v := range inP.Values() {
		// For each value we can generate from the 2 inputs
		prs.mrg2(v, prfLst[i])
	}
}

func (prs *Proofs) merge(inP Proofs) {
	for newVal, newProof := range inP {
		if newProof.Valid() {
			if selfCheck && (newProof.Len() == 0) {
				log.Fatal("Zero length proof being merged in")
			}
			prs.extend(newVal)
			tmp := *prs
			tmp[newVal].merge(newProof)
		}
	}
}
func wrkFastGenParWorker(b0, b1 proofLst, proofChan chan Proofs, wg *sync.WaitGroup) {
	tmpP := getProofs()
	nL0 := getProofs()
	nL1 := getProofs()
	// Clear not needed as extend now does this function
	// Retained here for debug
	//nL0.Clear()
	//nL1.Clear()
	tmpP.loneFastCross(nL0, nL1, b0, b1)
	putProofs(nL0)
	putProofs(nL1)
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

//const NumWorkers = 8
func init() {
	NumWorkers := runtime.NumCPU() * 2
	crossChan = make(chan crossStruct, NumWorkers)
	for i := 0; i < NumWorkers; i++ {
		go crossWorker(crossChan)
	}
}

// We include the pool variable so that benchmarking can continue to show
// that setting pool as false is the best option
// Yeah, I know. Weird!
func (prs *Proofs) wrkFastGen(inP proofLst, pool, par bool) {
	// Let's say we're given {2,3,4} as inP
	// we need to work 2 against all the numbers possible given 3 and 4
	// i.e. our return list would be:
	// {2,3}, {2,4}, {2,(3+4)},{2,(4-3)} etc
	inLeng := inP.Len()
	// Fast path
	if selfCheck && inLeng == 1 {
		log.Fatal("Length of 1??")
	} else if inLeng == 2 {
		prs.wrkExtraFastPair(inP)
		return
	}
	// It's worth noting that WrkFastSplit successively reslices
	// inP into a series of other slices
	//workingList = WrkFastSplit(inP)
	splitter := newSplitter(&inP)
	var numList0 Proofs
	var numList1 Proofs
	var proofChan chan Proofs
	var wg sync.WaitGroup

	if !par {
		numList0 = getProofs()
		numList1 = getProofs()
	} else {
		proofChan = make(chan Proofs)
		wg.Add(splitter.cnt())
		waitForIt := func() {
			wg.Wait()
			close(proofChan)
		}
		go waitForIt()
		go func() {
			for v := range proofChan {
				prs.merge(v)
				putProofs(v)
			}
		}()
	}

	for pl, err := splitter.next(); err == nil; pl, err = splitter.next() {
		if par {
			//go wrkFastGenParWorker(pl[0], pl[1], proofChan, &wg)
			//go arbCross(pl, proofChan, &wg)
			crossChan <- crossStruct{pl, proofChan, &wg}
		} else {
			prs.loneFastCross(numList0, numList1, pl[0], pl[1])
		}
	}
	if par {
		wg.Wait()
	} else {
		putProofs(numList0)
		putProofs(numList1)
	}
}

func (prs *Proofs) loneFastCross(numList0, numList1 Proofs, bob0, bob1 proofLst) {
	// wrkFast will transform {4,5,6} into all the intermediate numbers that can be generated
	numList0.wrkFast(bob0)
	numList1.wrkFast(bob1)

	for outer, oProof := range numList0 {
		if outer > 0 && (oProof.Len() > 0) {
			for inner, iProof := range numList1 {
				if inner > 0 && (iProof.Len() > 0) {
					inP := newProofLstPair(outer, inner, oProof, iProof)
					prs.wrkExtraFastPair(inP)
				}
			}
		}
	}
}

// This is a minimised memory implementation of WrkFastSplit
type splitter struct {
	inPp     *proofLst
	numAdded int
	i        int
}

func newSplitter(inP *proofLst) *splitter {
	itm := new(splitter)
	itm.numAdded = (inP.Len() - 1)
	itm.inPp = inP
	return itm
}

var errSpEnd = errors.New("End of splitter")

func (sp *splitter) next() (pl []proofLst, err error) {
	if sp.i < sp.numAdded {
		pl = sp.inPp.sliceAt(sp.i + 1)
	} else {
		err = errSpEnd
	}
	sp.i++
	return
}
func (sp splitter) cnt() int {
	return sp.numAdded
}
func wrkFastSplit(inP proofLst) []proofLst {
	// Take the proof list and split it into the sub possibilities
	// i.e. {2,3,4} becomes:
	// {{2},{3,4}}
	// {{2,3},{4}}
	// Alternatively, {4,5,6,7} becomes:
	// {{{4},{5,6,7}}
	// {{{4,5},{6,7}}
	// {{{4,5,6},{7}}
	numAdded := (inP.Len() - 1)
	retListP := getProofListArray(numAdded * 2)
	for i := 0; i < numAdded; i++ {
		tmpArray := inP.sliceAt(i + 1)
		retListP = append(retListP, tmpArray...)
	}
	return retListP
}

var plusOperator operator
var multOperator operator
var minusOperator operator
var divOperator operator

func init() {
	plusOperator = newOperator("+")
	multOperator = newOperator("*")
	minusOperator = newOperator("-")
	divOperator = newOperator("/")
}

// wrkExtraFastPair Is a hand optimised version of wrkFastPair
// That merges all the functions together
func (prs *Proofs) wrkExtraFastPair(inP proofLst) {
	// Basically, generate new values and see if we should merge them in
	// prs.mrg2(value, proof)
	var input0, input1 int
	var pr0, pr1 Proof
	pr0 = inP.prs[0]
	pr1 = inP.prs[1]

	input0 = inP.intL[0]
	input1 = inP.intL[1]

	pr0l := pr0.Len()
	pr1l := pr1.Len()

	swapValues := input1 > input0
	plusValue := input0 + input1
	multValue := input0 * input1

	if prs.interestingProof(plusValue, pr0l, pr1l) {
		prs.mrg3(plusValue, pr0, pr1, plusOperator)
	}
	minusValue := input0 - input1
	if swapValues {
		input0, input1 = input1, input0
	}

	// Make sure to only generate new proofs when needed
	if prs.interestingProof(multValue, pr0l, pr1l) {
		prs.mrg3(multValue, pr0, pr1, multOperator)
	}
	modDivide := (input0 % input1)
	if swapValues {
		minusValue = -minusValue
		pr1, pr0 = pr0, pr1
		pr0l, pr1l = pr1l, pr0l
	}
	canDivide := modDivide == 0
	divValue := input0 / input1
	if prs.interestingProof(minusValue, pr0l, pr1l) {
		prs.mrg3(minusValue, pr0, pr1, minusOperator)
	}

	if canDivide {
		if prs.interestingProof(divValue, pr0l, pr1l) {
			prs.mrg3(divValue, pr0, pr1, divOperator)
		}
	}
	return
}
func determineOperator(in string) func(int, int) int {
	switch in {
	case "+":
		return func(a, b int) int {
			return a + b
		}
	case "*":
		return func(a, b int) int {
			return a * b
		}
	case "-":
		return func(a, b int) int {
			return a - b
		}
	case "/":
		return func(a, b int) int {
			return a / b
		}
	default:
		log.Fatal("Invalid Operator", in)
		return func(a, b int) int { return -1 }
	}
}
func (nm Number) calculate() int {
	if nm.Val > 0 {
		return nm.Val
	}
	var operateFunc func(a, b int) int

	if len(nm.list) > 1 {
		operateFunc = determineOperator(nm.operation)
	}
	switch len(nm.list) {
	case 0:
		log.Fatal("len 0")
	case 1:
		return nm.list[0].calculate()
	case 2:
		return operateFunc(nm.list[0].calculate(), nm.list[1].calculate())
	default:
		runningVal := 0
		for _, v := range nm.list {
			runningVal = operateFunc(runningVal, v.calculate())
		}
		return runningVal
	}
	return 0
}

func parseRunes(inA <-chan rune) *Number {
	var tmpNumArr []*Number
	var op string
	var numString string
	gotNum := func() {
		if numString != "" {
			// parse the number we have collected
			val, err := strconv.Atoi(numString)
			if err != nil {
				log.Fatal("Number conversion error:", err)
			}
			tmp := Number{Val: val}
			tmpNumArr = append(tmpNumArr, &tmp)
			numString = ""
		}
	}
	retVal := func() *Number {
		num := Number{
			list:      tmpNumArr,
			operation: op,
		}
		num.Val = num.calculate()

		return &num
	}
	for ch := range inA {
		switch ch {
		case '(':
			// call ourselves

			tmpNumArr = append(tmpNumArr, parseRunes(inA))

		case ')':

			gotNum()
			// return with what we have

			return retVal()
		case '+', '-', '*', '/':
			// is our operator
			op = string(ch)
			// parse the number we have collected
			gotNum()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// is a number
			numString += string(ch)
		default:
			log.Fatal("Unexpected Char", ch)
		}
	}
	gotNum()
	if len(tmpNumArr) == 1 {
		return tmpNumArr[0]
	}
	return retVal()
}
func parseString(in string) *Number {
	inA := make(chan rune)
	go func() {
		for _, ch := range in {
			inA <- ch
		}
		close(inA)
	}()
	tmp := parseRunes(inA)
	return tmp
}
