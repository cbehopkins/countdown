package cntSlv

import (
	"errors"
)

// A gimmie is an efficient way to work through the large sets of numbers
// Rather than work out in advance everything we will need
// This supplies the next workload unit when requested.
// This is used to save memory
type gimmie struct {
	solList SolLst
	inner   int
	outer   int
	sent    bool
}

func newGimmie(arrayIn SolLst) *gimmie {
	//type NumCol []*Number
	//type SolLst []*NumCol
	itm := new(gimmie)
	itm.solList = arrayIn
	return itm
}
func (g *gimmie) items() int {
	items := 0
	for _, v := range g.solList {
		for _, w := range v {
			if w != nil {
				items++
			}
		}
	}
	return items
}
func (g *gimmie) reset() {
	g.sent = false
	g.outer = 0
	g.inner = 0
}

func (g *gimmie) next() (result *Number, err error) {
	for ; g.outer < g.solList.Len(); g.outer++ {
		inLst := g.solList[g.outer]
		for g.inner < inLst.Len() {
			result = inLst[g.inner]
			g.inner++
			if result != nil {
				return
			}
		}
		g.inner = 0
	}
	err = errors.New("No More to give you")
	return
}
