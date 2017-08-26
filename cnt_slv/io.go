package cntSlv

import (
	"encoding/json"
	"fmt"
	//    		"os"
	//	"github.com/tonnerre/golang-pretty"
)

//io.go is responsible for all the io stuff related to pushing the structures
// over the network

type JsonStruct struct {
	List []JsonNum `json:"l,omitempty"`
}

func NewJsonStruct(items int) (itm *JsonStruct) {
	itm = new(JsonStruct)
	if items > 0 {
		itm.List = make([]JsonNum, 0, items)
	}
	return itm
}
func (it *JsonStruct) Add(itm JsonNum) {
	it.List = append(it.List, itm)
}

type JsonNum struct {
	// The aim here is to keep the produced JSON as small as possible
	Val  int        `json:"v,attr"`
	Op   string     `json:"o,attr"`
	List JsonStruct `json:"s,omitempty"`
}

func AddNum(input Number) (result JsonNum) {
	result.Val = input.Val
	result.Op = input.operation
	result.List = *NewJsonStruct(len(input.list))
	for _, v := range input.list {
		result.List.Add(AddNum(*v))
	}
	return result
}
func (item *NumMap) AddJsonNum(input JsonNum) (newNumber Number) {
	newNumber.Val = input.Val
	newNumber.operation = input.Op
	if len(input.List.List) > 0 {
		newNumber.list = make([]*Number, len(input.List.List))
		for i, p := range input.List.List {
			tmpNum := item.AddJsonNum(p)
			newNumber.list[i] = &tmpNum
		}
	}
	item.Add(input.Val, &newNumber)

	return newNumber
}
func (item *NumMap) MarshalJson() (output []byte, err error) {
	thingList := NewJsonStruct(len(item.nmp))

	for _, v := range item.nmp {
		tmp := AddNum(*v)
		thingList.Add(tmp)
	}

	//output, err = json.MarshalIndent(thing_list, "", "    ")
	output, err = json.Marshal(thingList)
	//if err != nil {
	//	fmt.Printf("error: %v\n", err)
	//}

	//s := string(output)
	//fmt.Println(s)
	return
}
func (item *NumMap) UnMarshalJson(input []byte) (err error) {
	v := NewJsonStruct(0)
	err = json.Unmarshal(input, v)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return
	}
	for _, j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddJsonNum(j)
	}
	// At the end populate difficulty and prove the solutions for our sanity
	item.LastNumMap()
	for _, j := range item.Numbers() {
		j.ProveSol()
		j.SetDifficulty()
	}
	return
}

func (item *NumMap) FastUnMarshalJson(input []byte) (err error) {
	v := NewJsonStruct(0)
	err = json.Unmarshal(input, v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	for _, j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddJsonNum(j)
	}
	return
}

func ImportJson(message string) {
	fv := NewNumMap()
	err := fv.UnMarshalJson([]byte(message))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fv.PrintProofs()
}

func (item *NumMap) MergeJson(message string) {
	fv := NewNumMap()
	err := fv.UnMarshalJson([]byte(message))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	//fv.PrintProofs()
	item.Merge(fv, true)
}
