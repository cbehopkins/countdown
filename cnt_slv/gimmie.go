package cnt_slv

import (
	"errors"
)

// A gimmie is an efficient way to work through the large sets of numbers
// Rather than work out in advance everything we will need
// This supplies the next workload unit when requested.
// This is used to save memory
type Gimmie struct {
	sol_list SolLst
	inner    int
	outer    int
	sent     bool
}

func NewGimmie(array_in SolLst) *Gimmie {
	//type NumCol []*Number
	//type SolLst []*NumCol
	itm := new(Gimmie)
	itm.sol_list = array_in
	return itm
}
func (g *Gimmie) Items() (items int) {
	for _, v := range g.sol_list {
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
	for ; g.outer < g.sol_list.Len(); g.outer++ {
		in_lst_p := g.sol_list[g.outer]
		in_lst := in_lst_p // It's okay these should be stack variables as they do not leave the scope
		for g.inner < in_lst.Len() {
			result = in_lst[g.inner]
			g.inner++
			return
		}
		g.inner = 0
	}
	err = errors.New("No More to give you")
	return
}
