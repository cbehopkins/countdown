package cntSlv

import "errors"

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
