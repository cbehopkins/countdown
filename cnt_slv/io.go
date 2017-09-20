package cntSlv

import (
	"encoding/json"
	"fmt"
	//    		"os"
	//	"github.com/tonnerre/golang-pretty"
)

//io.go is responsible for all the io stuff related to pushing the structures
// over the network

// JSONStruct holds a number of items
type JSONStruct struct {
	List []JSONNum `json:"l,omitempty"`
}

// NewJSONStruct creates a new structure to work on
func NewJSONStruct(items int) (itm *JSONStruct) {
	itm = new(JSONStruct)
	if items > 0 {
		itm.List = make([]JSONNum, 0, items)
	}
	return itm
}

// Add adds an item to the main structure
func (it *JSONStruct) Add(itm JSONNum) {
	it.List = append(it.List, itm)
}

// JSONNum is how we carry a number over the json format
type JSONNum struct {
	// The aim here is to keep the produced JSON as small as possible
	Val  int        `json:"v,attr"`
	Op   string     `json:"o,attr"`
	List JSONStruct `json:"s,omitempty"`
}

// AddNum - add a number to the struct
// returning a json struct
func AddNum(input Number) (result JSONNum) {
	result.Val = input.Val
	result.Op = input.operation
	result.List = *NewJSONStruct(len(input.list))
	for _, v := range input.list {
		result.List.Add(AddNum(*v))
	}
	return result
}

//AddJSONNum adds a number to the number map from the json struct
func (item *NumMap) AddJSONNum(input JSONNum) (newNumber Number) {
	newNumber.Val = input.Val
	newNumber.operation = input.Op
	if len(input.List.List) > 0 {
		newNumber.list = make([]*Number, len(input.List.List))
		for i, p := range input.List.List {
			tmpNum := item.AddJSONNum(p)
			newNumber.list[i] = &tmpNum
		}
	}
	item.Add(input.Val, &newNumber)

	return newNumber
}

// MarshalJSON takes the nummap and turns it into a json struct
func (item *NumMap) MarshalJSON() (output []byte, err error) {
	thingList := NewJSONStruct(len(item.nmp))

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

// UnMarshalJSON takes in a json struct and adds the
// numbers to the nummap
func (item *NumMap) UnMarshalJSON(input []byte) (err error) {
	v := NewJSONStruct(0)
	err = json.Unmarshal(input, v)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return
	}
	for _, j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddJSONNum(j)
	}
	// At the end populate difficulty and prove the solutions for our sanity
	item.LastNumMap()
	for _, j := range item.Numbers() {
		j.ProveSol()
		j.setDifficulty()
	}
	return
}

// FastUnMarshalJSON - quikcer lesss inteligable version of above
func (item *NumMap) FastUnMarshalJSON(input []byte) (err error) {
	v := NewJSONStruct(0)
	err = json.Unmarshal(input, v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	for _, j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddJSONNum(j)
	}
	return
}

// ImportJSON - just read in and print out
// in proof like format
func ImportJSON(message string) {
	fv := NewNumMap()
	err := fv.UnMarshalJSON([]byte(message))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fv.PrintProofs()
}

// MergeJSON Use merge to get in a new json message
func (item *NumMap) MergeJSON(message string) {
	fv := NewNumMap()
	err := fv.UnMarshalJSON([]byte(message))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	//fv.PrintProofs()
	item.Merge(fv, true)
}
