package cntSlv

import (
	"errors"
)

// A gimmie is an efficient way to work through the large sets of numbers
// Rather than work out in advance everything we will need
// This supplies the next workload unit when requested.
// This is used to save memory
type Gimmie struct {
	solList SolLst
	inner   int
	outer   int
	sent    bool
}

func NewGimmie(arrayIn SolLst) *Gimmie {
	//type NumCol []*Number
	//type SolLst []*NumCol
	itm := new(Gimmie)
	itm.solList = arrayIn
	return itm
}
func (g *Gimmie) Items() (items int) {
	for _, v := range g.solList {
		items = items + v.Len()
	}
	return items
}
func (g *Gimmie) Reset() {
	g.sent = false
	g.outer = 0
	g.inner = 0
}

func (g *Gimmie) Next() (result *Number, err error) {
	for ; g.outer < g.solList.Len(); g.outer++ {
		inLstP := g.solList[g.outer]
		inLst := inLstP // It's okay these should be stack variables as they do not leave the scope
		for g.inner < inLst.Len() {
			result = inLst[g.inner]
			g.inner++
			return
		}
		g.inner = 0
	}
	err = errors.New("No More to give you")
	return
}
