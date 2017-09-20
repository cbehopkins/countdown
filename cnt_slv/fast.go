package cntSlv

import (
	"log"
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

//func (prs *Proofs) mrg2(v int, bob Proof) {
//	tmp := *prs
//	tmp[v].merge(bob)
//	//putBa(bob.tmp)
//}
// 3rd attempt at a merge method
//
func (prs *Proofs) mrg3(v int, pr, input Proof, op operator) {
	tmp := *prs
	buffer := tmp[v].tmp
	capNeeded := bLen + opLen + len(pr.tmp) + len(input.tmp)
	tmp[v].tmp = pr.concatCore(input, op, buffer, capNeeded)
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
	prs[v].set(pr)
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
//func (prs *Proofs) mergeLst(inP proofLst) {
//	prfLst := inP.Proofs()
//	for i, v := range inP.Values() {
//		// For each value we can generate from the 2 inputs
//		prs.mrg2(v, prfLst[i])
//	}
//}

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
func (pr Proof) parseProof() *Number {
	inString := pr.String()
	if inString != "" {
		derNum := parseString(inString)
		if derNum == nil {
			log.Fatal("Nil result for:", inString)
		}
		return derNum
	}
	return nil
}

///////////////////////////////////////
// Putting memory pooling logic at end
// Useful, but distracting
var poolProofListArray map[int]*sync.Pool

func init() {

	initPoolProofsArray()
	initProofsPool()
	initCrossWorkers()
	initProof()
	poolProofListArray = make(map[int]*sync.Pool)
	initOperators()
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
