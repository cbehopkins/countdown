package cnt_slv

import (
		"encoding/json"
     		"fmt"
//    		"os"
//	"github.com/tonnerre/golang-pretty"
)
type JsonStruct struct {
  	List [] JsonNum `json:"l,omitempty"`
}
func NewJsonStruct (items int) (itm *JsonStruct) {
	itm = new(JsonStruct)
	if items>0 {
		itm.List = make([] JsonNum, 0, items)
	}
	return itm
}
func (it *JsonStruct) Add (itm JsonNum) () {
	it.List = append(it.List, itm)
}
type JsonNum struct {
	// The aim here is to keep the produced JSON as small as possible
	Val   		int	 `json:"v,attr"`
	Op		string `json:"o,attr"`
	List		JsonStruct `json:"s,omitempty"`
}

func AddNum (input Number) (result JsonNum) {
	result.Val = input.Val
	result.Op = input.operation
	result.List = *NewJsonStruct(len(input.list))
	for _, v := range input.list {
		result.List.Add(AddNum(*v))
	}
	return result
}
func (item *NumMap) AddJsonNum (input JsonNum) (new_number Number) {
	new_number.Val = input.Val
	new_number.operation = input.Op
	if (len(input.List.List)>0) {
		new_number.list = make([]*Number, len(input.List.List))
		for i,p := range input.List.List {
			tmp_num := item.AddJsonNum(p)
		        new_number.list[i] = &tmp_num
		}
	}
	item.Add(input.Val, &new_number)

	return new_number
}
func (item *NumMap) MarshalJson () (output []byte, err error) {
	thing_list := NewJsonStruct(len(item.nmp))

        for _, v := range item.nmp {
		tmp := AddNum(*v)
		thing_list.Add(tmp)
	}

	//output, err = json.MarshalIndent(thing_list, "", "    ")
        output, err = json.Marshal(thing_list)
  	//if err != nil {
	//	fmt.Printf("error: %v\n", err)
  	//}

  	//s := string(output)
  	//fmt.Println(s)
	return
}
func (item *NumMap) UnMarshalJson (input []byte ) (err error) {
	v := NewJsonStruct(0)
	err = json.Unmarshal(input, v)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return
	}
	for _,j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddJsonNum(j)
	}
	// At the end populate difficulty and prove the solutions for our sanity
	item.LastNumMap()
	for _,j := range item.Numbers() {
		j.ProveSol()
		j.SetDifficulty()
	}
	return
}

func (item *NumMap) FastUnMarshalJson (input []byte ) (err error) {
        v := NewJsonStruct(0)
        err = json.Unmarshal(input, v)
        if err != nil {
                fmt.Printf("error: %v", err)
                return
        }
        for _,j := range v.List {
                //fmt.Printf("Value of %d\n", j.Val)
                item.AddJsonNum(j)
        }
        return
}

func ImportJson (message string) () {
    var prl SolLst
    fv := NewNumMap(&prl)
    err := fv.UnMarshalJson([]byte(message))
    if err != nil {
          fmt.Printf("error: %v\n", err)
          return
    }
    fv.PrintProofs()
}

func (item *NumMap) MergeJson (message string) () {
    var prl SolLst
    fv := NewNumMap(&prl)
    err := fv.UnMarshalJson([]byte(message))
    if err != nil {
          fmt.Printf("error: %v\n", err)
          return
    }
    //fv.PrintProofs()
   item.Merge(fv, true)
}

