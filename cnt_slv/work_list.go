package cnt_slv

import (
	"log"

	"github.com/tonnerre/golang-pretty"
)

type WrkLst struct {
	lst []SolLst
}

func (wl WrkLst) procWork(found_values *NumMap, wf func(a, b *Number) bool) {
	run := true
	for _, work_unit := range wl.lst {
		// Now we've extracted one work item,
		// so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

		if found_values.SelfTest {
			// Sanity check for programming errors
			work_unit_length := work_unit.Len()
			if work_unit_length != 2 {
				pretty.Println(wl)
				log.Fatalf("Invalid work unit length, %d", work_unit_length)
			}
		}

		unit_a := work_unit[0]
		unit_b := work_unit[1]
		list_a := work_n(unit_a, found_values)
		list_b := work_n(unit_b, found_values)
		gimmie_a := NewGimmie(list_a)
		gimmie_b := NewGimmie(list_b)

		for a_num, err_a := gimmie_a.Next(); (err_a == nil) && run; a_num, err_a = gimmie_a.Next() {
			for b_num, err_b := gimmie_b.Next(); (err_b == nil) && run; b_num, err_b = gimmie_b.Next() {
				run = wf(a_num, b_num)
			}
			gimmie_b.Reset()
		}
		if !run {
			return
		}
	}
}

func NewWrkLst(array_a NumCol) WrkLst {
	var work_list []SolLst
	// Easier to explain by example:
	// {2,3,4} -> {{2},{3,4}}
	// {2,3,4,5} -> {{2}, {3,4,5}}
	//           -> {{2,3},{4,5}}
	// {2,3,4,5,6} -> {{2},{3,4,5,6}}
	//             -> {{2,3},{4,5,6}}
	//             -> {{2,3,4},{5,6}}

	// The consumer of this list of list (of list) will then feed each list length >1 into a the work+_n function
	// In order to get down to a {{a},{b}} which can then be worked
	// The important point is that even though the list we return may be indefinitly long
	// each work unit within it is then a smaller unit
	// so an input array of 3 numbers only generates work units that contain number lists of length 2 or less

	len_array_m1 := array_a.Len() - 1

	for i := 0; i < (len_array_m1); i++ {
		var ar_a, ar_b NumCol
		// for 3 items in arrar
		// {0},{1,2}, {0,1}{2}
		ar_a = make(NumCol, i+1)
		copy(ar_a, array_a[0:i+1])
		ar_b = make(NumCol, (array_a.Len() - (i + 1)))

		copy(ar_b, array_a[(i+1):(array_a.Len())])
		var work_item SolLst // {{2},{3,4}};
		// a work item always contains 2 elements to the array
		work_item = append(work_item, ar_a, ar_b)
		work_list = append(work_list, work_item)
	}
	return WrkLst{lst: work_list}
}
func (wl WrkLst) Len() int {
	return len(wl.lst)
}

func (wl WrkLst) Get(itm int) SolLst {
	return wl.lst[itm]
}
func (wl WrkLst) Last() SolLst {
	return wl.Get(wl.Len() - 1)
}
