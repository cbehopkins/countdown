package cnt_slv

import (     
		"encoding/xml"
     		"fmt"
//    		"os"
	"github.com/tonnerre/golang-pretty"
)
type XmlStruct struct {
        XMLName   xml.Name `xml:"s"`
  	List [] XmlNum `xml:"l"`
}
func NewXmlStruct (items int) (itm *XmlStruct) {
	itm = new(XmlStruct)
	itm.List = make([] XmlNum, 0, items)
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
	List		XmlStruct
}

func (i *Number) MarshalXml () string {
  tmp := XmlNum{Val:i.Val, Proof:i.String()}

  output, err := xml.MarshalIndent(tmp, "", "    ")
  if err != nil {
    fmt.Printf("error: %v\n", err)
  }
  s := string(output)
  //fmt.Println(i)
  //fmt.Println(s)
  return s
}
func (it *XmlStruct) AddNum (input Number) (result XmlNum) {
	result.Val = input.Val
	result.Op = input.operation
	result.List = NewXmlStruct(len(input.list))
	for _, v := range input.list {
		result.List.Add(AddNum(v))
	}
	return result
}
func (item *NumMap) MarshalXml () (result string) {
	thing_list := NewXmlStruct(len(item.nmp))

        for _, v := range item.nmp {
		tmp := AddNum(v)
		thing_list.Add(tmp)
	}

	output, err := xml.MarshalIndent(thing_list, "", "    ")
  	if err != nil {
		fmt.Printf("error: %v\n", err)
  	}

  	s := string(output)
  	//fmt.Println(s)
	return s
}

func (item *NumMap) UnMarshalXml (input string) {
	v := NewXmlStruct(0)
	err := xml.Unmarshal([]byte(input), v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Printf("We've been given:\n%s\nand we turn this into:\n", input)
	pretty.Println(v)
	for _,j := range v.List {
		fmt.Printf("Value of %d, Proof of %s\n", j.Val, j.Proof)
	}
}
