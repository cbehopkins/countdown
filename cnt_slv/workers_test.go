package cnt_slv

import (
	"log"
	"testing"
)
func TestWeirdWork(t *testing.T) {
	var target int
	// (9-1)*50 = 400
	// (100 + 9*3) = 327
	// (400+327)= 727
	target = 727

	var proof400 SolLst
	var proof327 SolLst

	var mk400 NumCol
	var mk327 NumCol
	var combined NumCol

	found_values := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	found_values.SelfTest = true
	found_values.UseMult = true
	mk400.AddNum(50, found_values)
	mk400.AddNum(9, found_values)
	mk400.AddNum(1, found_values)
	mk327.AddNum(100, found_values)
	mk327.AddNum(9, found_values)
	mk327.AddNum(3, found_values)

	found_values.SetTarget(target)

	proof400 = append(proof400, mk400) // Add on the work item that is the source
	proof327 = append(proof327, mk327) // Add on the work item that is the source
	sol400 := work_n(mk400, found_values, false)
	sol327 := work_n(mk327, found_values, false)

	log.Println("Find 400", sol400.StringNum(400))
	log.Println("Find 327", sol327.StringNum(327))

	combined = append(mk400, mk327...)
	var work_list WrkLst
	work_list = NewWrkLst(combined)
	chkFunc := func() bool {
		for _, work_unit := range work_list.lst {
			var unit_a, unit_b NumCol
			unit_a = work_unit[0]
			unit_b = work_unit[1]
			if mk400.Equal(unit_a) {
				if mk327.Equal(unit_b) {
					tmp400 := work_n(unit_a, found_values, false)
					tmp327 := work_n(unit_b, found_values, false)
					if !tmp400.Exists(400) {
						return false
					}
					if !tmp327.Exists(327) {
						return false
					}
					return true
				}
			}
		}
		return false
	}
	log.Println("Its:", chkFunc())
	sol_combined := work_n(combined, found_values, false)
	log.Println("Find 727", sol_combined.StringNum(727))
  }
func tstWorker (fc func (NumCol, *NumMap) ) {
	var target int
	// (9-1)*50 = 400
	// (100 + 9*3) = 327
	// (400+327)= 727
	target = 727

	var mk400 NumCol

	found_values := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	found_values.SelfTest = true
	found_values.UseMult = true
	mk400.AddNum(50, found_values)
	mk400.AddNum(9, found_values)
	mk400.AddNum(1, found_values)
	mk400.AddNum(100, found_values)
	mk400.AddNum(9, found_values)
	mk400.AddNum(3, found_values)

	found_values.SetTarget(target)

  fc(mk400, found_values)
	//sol_combined := work_n(mk400, found_values, false)
  //return sol_combined
}

func TestWorkn(t *testing.T) {
  var tmp SolLst 
  fun := func (nc NumCol, fv *NumMap) {
    tmp = work_n(nc,fv,false)
  }
  tstWorker(fun)
	log.Println("Find 727", tmp.StringNum(727))
}
func TestPermute(t *testing.T) {
  var tmpChan chan SolLst 
  fun := func (nc NumCol, fv *NumMap) {
      tmpChan = permuteN(nc,fv)
  }
  tstWorker(fun)
  for tmp := range tmpChan {
	  log.Println("Find 727", tmp.StringNum(727))
  }
}

