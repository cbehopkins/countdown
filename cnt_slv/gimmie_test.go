package cntSlv

import (
	"log"
	"math/rand"
	"testing"
)

func _generateTestSolList(fv *NumMap, numUse []int) (processList SolLst) {
	currentNumber := 0
	for i := 0; i < len(numUse); i++ {
		innerList := NumCol{}
		for j := 0; j < numUse[i]; j++ {
			if rand.Intn(10) == 1 {
				innerList = append(innerList, nil)
			} else {
				innerList.AddNum(currentNumber, fv)
			}
			currentNumber++
		}
		processList = append(processList, innerList)
	}
	return
}

func Test_gimmie_length_calc(t *testing.T) {
	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end
	foundValues.SelfTest = true
	foundValues.UseMult = true

	processList := _generateTestSolList(foundValues, []int{2, 4})
	gm := newGimmie(processList)
	calculatedLength := gm.items()
	t.Log("Calc as:", calculatedLength)
	for result, err := gm.next(); err == nil; result, err = gm.next() {
		calculatedLength--
		log.Println("got", result)
	}
	if calculatedLength != 0 {
		t.Log("Calculated length was incorrect", calculatedLength)
		t.Fail()
	}
}
func Test_gimmie_length_cross(t *testing.T) {

	foundValues := NewNumMap() //pass it the proof list so it can auto-check for validity at the end

	foundValues.SelfTest = true
	foundValues.UseMult = true

	processListA := _generateTestSolList(foundValues, []int{2, 4, 3})
	processListB := _generateTestSolList(foundValues, []int{1, 4, 9, 12, 23, 2, 1})
	gmA := newGimmie(processListA)
	gmB := newGimmie(processListB)

	calculatedLength := gmA.items() * gmB.items()

	t.Log("Calc as:", calculatedLength)
	for resultA, errA := gmA.next(); errA == nil; resultA, errA = gmA.next() {
		for resultB, errB := gmB.next(); errB == nil; resultB, errB = gmB.next() {
			calculatedLength--
			log.Println("got", resultA, resultB)
		}
		gmB.reset()
	}
	if calculatedLength != 0 {
		t.Log("Calculated length was incorrect", calculatedLength)
		t.Fail()
	}
}
