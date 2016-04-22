package cnt_slv

import (     
		"encoding/xml"
     		"fmt"
//    		"os"
//	"github.com/tonnerre/golang-pretty"
)
type XmlStruct struct {
        XMLName   xml.Name `xml:"s,omitempty"`
  	List [] XmlNum `xml:"l,omitempty"`
}
func NewXmlStruct (items int) (itm *XmlStruct) {
	itm = new(XmlStruct)
	if items>0 {
		itm.List = make([] XmlNum, 0, items)
	}
	return itm
}
func (it *XmlStruct) Add (itm XmlNum) () {
	it.List = append(it.List, itm)
}
type XmlNum struct {
	// The aim here is to keep the produced XML as small as possible
//	XMLName   	xml.Name `xml:"n"` 
	Val   		int	 `xml:"v,attr"`
	Op		string `xml:"o,attr"`
	List		XmlStruct `xml:"s,omitempty"`
}

func AddNum (input Number) (result XmlNum) {
	result.Val = input.Val
	result.Op = input.operation
	result.List = *NewXmlStruct(len(input.list))
	for _, v := range input.list {
		result.List.Add(AddNum(*v))
	}
	return result
}
func (item *NumMap) AddXmlNum (input XmlNum) (new_number Number) {
	new_number.Val = input.Val
	new_number.operation = input.Op
	if (len(input.List.List)>0) {
		new_number.list = make([]*Number, len(input.List.List))
		for i,p := range input.List.List {
			tmp_num := item.AddXmlNum(p)
		        new_number.list[i] = &tmp_num
		}
	}
	item.Add(input.Val, &new_number)

	return new_number
}
func (item *NumMap) MarshalXml () (output []byte, err error) {
	thing_list := NewXmlStruct(len(item.nmp))

        for _, v := range item.nmp {
		tmp := AddNum(*v)
		thing_list.Add(tmp)
	}

	//output, err = xml.MarshalIndent(thing_list, "", "    ")
        output, err = xml.Marshal(thing_list)
  	//if err != nil {
	//	fmt.Printf("error: %v\n", err)
  	//}

  	//s := string(output)
  	//fmt.Println(s)
	return 
}
func (item *NumMap) UnMarshalXml (input []byte ) (err error) {
	v := NewXmlStruct(0)
	err = xml.Unmarshal(input, v)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return
	}
	//fmt.Printf("We've been given:\n%s\nand we turn this into:\n", input)
	//pretty.Println(v)
	for _,j := range v.List {
		//fmt.Printf("Value of %d\n", j.Val)
		item.AddXmlNum(j)
	}
	// At the end populate difficulty and prove the solutions for our sanity
	item.LastNumMap()
	for _,j := range item.Numbers() {
		j.ProveSol()
		j.SetDifficulty()
	}
	return
}


func ImportXml (message string) () {
    var prl SolLst
    fv := NewNumMap(&prl)
    err := fv.UnMarshalXml([]byte(message))
    if err != nil {
          fmt.Printf("error: %v\n", err)
          return  
    }
    fv.PrintProofs()
}

func (item *NumMap) MergeXml (message string) () {
    var prl SolLst                                                                                                                                                                                                           
    fv := NewNumMap(&prl)
    err := fv.UnMarshalXml([]byte(message))                                                                                                                                                                                           
    if err != nil {
          fmt.Printf("error: %v\n", err)                                                                                                                                                                                             
          return
    }                                                                                                                                                                                                                                
    //fv.PrintProofs()
   item.Merge(fv, true)
}      

