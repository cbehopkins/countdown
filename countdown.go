package main

import (
  "os"
  "fmt"
  "log"
  "reflect"
  "runtime"
  "strconv"
  "flag"
//  "math/rand"
  "github.com/fighterlyt/permutation"
  "github.com/tonnerre/golang-pretty"
)

var use_mult bool
var self_test bool
var seek_short bool
type Number struct {
  // A number consists of 
  Val int // a value
  list []*Number // a pointer the the list of numbers used to obtain this
  operation string // The operation used on those numbers to get here
  
}
func (i *Number) ProofLen () int {
  var cumlen int
  if (i.list == nil) {
    cumlen = 1
  } else {
    l0 := i.list[0].ProofLen()
    l1 := i.list[1].ProofLen()

    cumlen = l0 + l1
  }
  return cumlen
}
func (i *Number) ProveIt () string {
  var proof string
  var val int
  val = i.Val
  if (i.list == nil) {
    proof = fmt.Sprintf("%d",val);
  } else {
    p0 := i.list[0].ProveIt();
    p1 := i.list[1].ProveIt();
    operation := i.operation;
    switch operation {
      case  "--" : 
        proof = fmt.Sprintf("(%s-%s)", p1, p0)
      case "\\" :
        proof = fmt.Sprintf("(%s/%s)", p1, p0)
      default : 
        proof = fmt.Sprintf("(%s%s%s)", p0, operation, p1)
    }
  }
  return proof
}

// So NumCol is a collection of numbers
// This says that (for a given inout) you can have all of these numbers
// at the same time
// A solution List is a number of things you do with a set of numbers
type NumCol []*Number
type SolLst []*NumCol

func (item NumCol) TestNum (to_test int) bool {
  for _, v:= range item {
    value := v.Val
    if (value==to_test) {
      return true
    }
  }
  return false
}

func (item NumCol) GetNumCol () string {
  var ret_str string
  comma := ""
  ret_str = ""
  for _, v:= range item {
    ret_str =fmt.Sprintf("%s%s%d",ret_str,comma,v.Val);
    comma=","
  }
  return ret_str
}
func (bob *NumCol) add_num (input_num int, found_values *NumMap) {
  var empty_list NumCol;

  a := new_number(input_num, empty_list,"I",found_values);
  *bob = append (*bob, a);

}
func (bob NumCol) Len () int {
   var array_len int
   array_len = len(bob)
   return array_len
}
func (item SolLst) CheckDuplicates () {
  sol_map := make (map [string] NumCol)
  var del_queue [] int


  for i:=0; i<len(item); i++ {
    var v NumCol
    v = *item[i]
    string:=v.GetNumCol()
    
    _,ok := sol_map[string]
    if (!ok) { 
      sol_map[string] = v;
    } else {
      //fmt.Printf("%s already exists, Length %d\n:", string,len(tpp));
      //pretty.Println(t1)
      //fmt.Printf("It is now, %d", i);
      //pretty.Println(t0);
      del_queue = append(del_queue, i) 
    }
  }
  
  for i:=len(del_queue); i>0; i-- {
    //fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
    v:=del_queue[i-1]
    //fmt.Println("You've asked to delete",v);
    l1:=item
    item=append(l1[:v], l1[v+1:]...)
  }

}



type NumMapAtom struct {   
 a int
 b *Number
}

type NumMap struct {
  nmp map [int]*Number 
  TargetSet bool
  Target int
  input_channel chan NumMapAtom
  done_channel chan bool
}

func (item *NumMap) Add (a int, b *Number){

  //fmt.Printf("Debugging NumMap.Add, %d\n", a);
  //pretty.Println (item)
  var atomic NumMapAtom
  atomic.a = a
  atomic.b = b
  item.input_channel <- atomic; 
}


func (item *NumMap) Merge (a NumMap){

  for i, v:= range a.nmp{

    var atomic NumMapAtom
  
    atomic.a = i
    atomic.b = v
    item.input_channel <- atomic;
  }
}

func (item *NumMap) AddProc (proof_list *SolLst){

  //pretty.Println (item)
  for bob := range item.input_channel {
    //fmt.Printf("Debugging NumMap.AddProc, %d\n", bob.a);
    retr, ok := item.nmp[bob.a]
    if (!ok) {
      //fmt.Printf("Adding value %d\n", bob.a)
      item.nmp[bob.a] = bob.b;
      if (item.TargetSet) {
        if (bob.a == item.Target) {
          proof_string := bob.b.ProveIt()
          fmt.Printf("Value %d, = %s, Proof Len is %d\n",bob.b.Val, proof_string, bob.b.ProofLen())
          if (!seek_short) {os.Exit(0)}
        }
      }
    } else if (seek_short) {
      if (retr.ProofLen()>bob.b.ProofLen()) {
        // In seek short mode, then update when it has a shorter proof
        item.nmp[bob.a] = bob.b
      }
      if (item.TargetSet && (bob.a == item.Target)) {fmt.Printf("Value %d, = %s, Proof Len is %d\n",bob.b.Val, bob.b.ProveIt(), bob.b.ProofLen())}
    }
  }
  if (self_test) {check_return_list(*proof_list,item)}
  item.CheckDuplicates(proof_list)
  item.done_channel <- true
}
func (item *NumMap) GetVals () []int{
  ret_list := make ([]int, len(item.nmp))
  //fmt.Printf("\nThere are %d in list\n",len(item.nmp))
  i:=0
  for _, v:= range item.nmp {
    //fmt.Printf("v:%d,%d\n",i, v.Val);
    ret_list[i] = v.Val;
    i++
  }
  return ret_list
}
func  NewNumMap (proof_list *SolLst) *NumMap {
  p:= new(NumMap)
  fred := make (map [int]*Number);
  p.nmp = fred
  bob := make (chan NumMapAtom,1000)
  p.input_channel = bob
  dc := make (chan bool)
  p.done_channel = dc
  p.TargetSet = false

  go p.AddProc(proof_list);
  return p
}
func (item *NumMap) CheckDuplicates (proof_list *SolLst){
  set_list_map := make (map [string] NumCol);
  //fmt.Printf("Checking for duplicates in Proof\n");
  var tpp SolLst
  tpp = *proof_list
  var del_queue [] int

  for i:=0; i<len(*proof_list); i++ {
    v:=tpp[i]
    var t0 NumCol
    t0 = *v
    string:=t0.GetNumCol() 
    //fmt.Printf("Formatted into %s\n", string);
    _,ok := set_list_map[string]
    if (!ok) { 
      set_list_map[string] = t0;
    } else {
      //fmt.Printf("%s already exists, Length %d\n:", string,len(tpp));
      //pretty.Println(t1)
      //fmt.Printf("It is now, %d", i);
      //pretty.Println(t0);
      del_queue = append(del_queue, i) 
    }
  }

  for i:=len(del_queue); i>0; i-- {
    //fmt.Printf("DQ#%d, Len=%d\n",i, len(del_queue))
    v:=del_queue[i-1]
    //fmt.Println("You've asked to delete",v);
    l1:=*proof_list
    *proof_list=append(l1[:v], l1[v+1:]...)
  }


}
func (item *NumMap) LastNumMap () {
  fmt.Println("Closing Channel")
  close(item.input_channel)
  <- item.done_channel
}
func (item *NumMap) SetTarget (target int) {
    fmt.Println("Setting target to ",target)
    item.TargetSet = true
    item.Target = target
    fmt.Println("Target is now ", item.Target)
}
func (item *NumMap) PrintProofs () {
  min_num :=1000
  max_num := 0
  num_num := 0
  for _, v:= range item.nmp {
      // w is *Number
      var Value int;
      Value = v.Val;
      num_num++
      if (Value>max_num) {
        max_num = Value
      }
      if (Value<min_num) {
        min_num = Value
      }

//      proof_string := v.ProveIt()
//      fmt.Printf("Value %d, = %s\n",Value, proof_string);
      //pretty.Println(w);
  }
  for i:=min_num; i<=max_num;i++ {
    Value, ok := item.nmp[i]
    if (ok) {
       proof_string := Value.ProveIt()
       fmt.Printf("Value %d, = %s\n",Value.Val, proof_string);
    }
  }
  fmt.Printf("There are:\n%d Numbers\nMin:%4d Max:%4d\n",num_num,min_num,max_num)
}
func make_2_to_1 (a, b *Number, found_values *NumMap) []*Number{
  // This is (conceptually) returning a list of numbers 
  // That can be generated from 2 input numbers 
  // organised in such a way that we know how we created them
  var ret_list []*Number
  var plus_num *Number
  var mult_num *Number
  var minu_num *Number

  var list []*Number 

  list = append(list, a,b)
  ret_list = make([]*Number, 0, 4)

  plus_num = new_number(a.Val + b.Val,list, "+",found_values)
  if (use_mult) {mult_num = new_number(a.Val * b.Val,list, "*",found_values)}

  // REVISIT - generating both of these would give us more intersting numbers quicker
  if (a.Val > b.Val) {
    minu_num = new_number(a.Val - b.Val,list, "-",found_values)
    if ((b.Val>0) && ((a.Val%b.Val)==0)) {
      tmp_div := new_number((a.Val/b.Val),list, "/",found_values)
      ret_list = append(ret_list,tmp_div)
    }
  } else {
    minu_num = new_number(b.Val - a.Val,list, "--",found_values)
    if ((a.Val>0) && ((b.Val%a.Val)==0)) {
      tmp_div := new_number((b.Val/a.Val),list, "\\",found_values)
      ret_list = append(ret_list,tmp_div)
    }
  }

  ret_list = append(ret_list, plus_num)
  if (use_mult) {ret_list = append(ret_list, mult_num)}
  ret_list = append(ret_list, minu_num)
  return ret_list 
}

func new_number (input_a int, input_b []*Number, operation string, found_values *NumMap) *Number{

  var new_num Number
  new_num.Val = input_a
  found_values.Add(input_a,&new_num);
  
  new_num.list = input_b;
  new_num.operation = operation;
  //fmt.Printf("There are %d elements in the input_a list\n", len(input_a.list))
  return &new_num
}
func gimmie_1 (array_in SolLst, found_values *NumMap) NumCol {
  var ret_list NumCol;
  // This function takes in a list of numbers and tries to return a list of numbers
//  fmt.Println("gimmie_1 called with")
//  pretty.Println(array_in)
  len_array_needed := 0
  for _, v:= range array_in {
    //REVISIT - this is a lot of copying and allocation
    //ret_list = append(ret_list, *v...)
    len_array_needed = len_array_needed + v.Len()
  }

  ret_list = make (NumCol, 0,len_array_needed) // Length is zero capacity is as needed
  for _, v:= range array_in {
    // Append should only increase the size of the array if needed
    ret_list = append(ret_list, *v...)
  }


// Note - Running a reduction here extends the run time, not reduces it!
//var reduced_ret_list []*Number;
//  reduction_map := make (map [int] *Number)
//  for i,v:= range ret_list {
//    //fmt.Printf ("working with value %d\n", i);
//    reduction_map[i] = v;
//  }
//  i:=0
//  reduced_ret_list := make ([]*Number, len (reduction_map))
//  for _, v:= range reduction_map {
//    reduced_ret_list[i] = v
//    i++
//  }
  return ret_list
}

func work_n (array_in NumCol, found_values *NumMap) SolLst  {
  var ret_list SolLst ;
  len_array_in := len(array_in);
//  fmt.Printf("Calling work_n with %d items\n",len_array_in);
//  for _, v:= range array_in {
//    value := v.Val;
//    fmt.Printf("%d,",value);
//  }
//  fmt.Printf("\n");
 
  if (len_array_in==1) {
    ret_list = append(ret_list, &array_in)
    return ret_list
  } else if (len_array_in == 2) {
    var a, b *Number;
    a = array_in[0];
    b = array_in[1];
    var tmp_list NumCol
    tmp_list = make_2_to_1(a,b,found_values);
    ret_list = append(ret_list, &tmp_list, &array_in)
    return ret_list
  } 

 
  // work_n takes 
  // let's use work 3 as a first example {2,3,4} and should generate everything that can be done with these 3 numbers
  // Note: for these explanantions I'll assume we just add and subtract numbers
  // We do not return the supplied list with the return
  // we also do no permute the input numbers as we know that permute function will do this for us
  // So in this example we would look to do several steps first we feed to make_3
  // This will treat the input as {2,3),{4} it works the first list to get:
  // {5,1} (from 2+3 and 3-2) and therefore returns {{5,4}, {1,4}}
  // we then take each value in this list and work that to get {{9},{3}}
  // the final list we want to return is {{5,4}, {1,4}, {9},{3}}
  // the reason to not return {2,3,4} is so that in the grand scheme of things we can recurse these lists
  var work_list []SolLst;
  work_list = expand_n (array_in);
//  fmt.Println("expand_n returned:")
//  pretty.Println(work_list)
  // so by this stage we have something like {{{2},{3,4}}} or for a 4 variable: { {{2}, {3,4,5}}, {{2,3},{4,5}} }
  var work_unit SolLst;
  for _, work_unit = range work_list {
    // Now we've extracted one work item,
    // so conceptually  here we have {{2},{3,4,5,6}} or perhaps {{2},{3,4}} or {{2,3},{4,5}}

    // Sanity check for programming errors
    work_unit_length := len(work_unit);
    if (work_unit_length !=2) {
      pretty.Println(work_list);
      log.Fatalf("Invalid work unit length, %d", work_unit_length);
    }
    var unit_a,unit_b  *NumCol;
    unit_a  = work_unit[0];
    unit_b  = work_unit[1];

    var list_a SolLst;
    var list_b SolLst;
    list_a = work_n(*unit_a, found_values)
    list_b = work_n(*unit_b, found_values)



    // Now we want two list of numbers to cross against each other
    var list_of_1_a,list_of_1_b NumCol;
    list_of_1_a  = gimmie_1(list_a, found_values);
    list_of_1_b  = gimmie_1(list_b, found_values);



    // Now Cross work then
    for _, a_num := range list_of_1_a {
      for _, b_num := range list_of_1_b {
        var product_of_2 NumCol;
        product_of_2 = make_2_to_1(a_num,b_num, found_values);
        ret_list = append (ret_list, &product_of_2);
      }    
    }
    // Add on the work unit because that contains sub combinations that may be of use
    ret_list = append (ret_list, work_unit...)
  }
  // This adds about 10% to the run time, but reduces memory to 1/5th
  ret_list.CheckDuplicates()
  return ret_list;
}



func permute_n (array_in NumCol, found_values *NumMap, proof_list chan SolLst)  {
  less := func(i, j interface{}) bool {
    v1 := reflect.ValueOf(i).Elem().FieldByName("Val").Addr().Interface().(*int)
    v2 := reflect.ValueOf(j).Elem().FieldByName("Val").Addr().Interface().(*int)
    return *v1<*v2
  }
  fmt.Printf("*** SOP cot: %d\n", runtime.NumGoroutine()) 
  p,err:=permutation.NewPerm(array_in,less)
  if err!=nil{
    fmt.Println(err)
  }
  num_procs := p.Left()
  var comms_channels []chan SolLst
  comms_channels = make([]chan SolLst, num_procs)
  for i:= range comms_channels {
    comms_channels[i] = make (chan SolLst,2)
  }
  var channel_tokens chan bool
  channel_tokens = make (chan bool, 128)
  for i:= 0; i<1; i++ {
    channel_tokens<-true
  }
  coallate_chan := make(chan SolLst, 2)
  coallate_done := make(chan bool, 8)


  caller := func () {
   for result,err:=p.Next();err==nil;result,err=p.Next(){
    <- channel_tokens
    //var bob NumCol;
    fmt.Printf("%3d permutation: left %d, GoRs %d\n",p.Index()-1,p.Left(), runtime.NumGoroutine()) 
    bob, ok := result.(NumCol)
    if (!ok) {
      log.Fatalf("Error Type conversion problem")  
    }
    //pretty.Println(bob)
    worker := func (it NumCol, fv *NumMap,curr_iten int)  {
      coallate_chan <- work_n(it, fv)
      coallate_done <- true
      channel_tokens<-true // Now we're done, add a token to allow another to start

    }
    go worker(bob,found_values, p.Index()-1)
   }
  }
  go caller()

  // This little go function waits for all the procs to have a done channel and then closes the channel
  done_control := func () {
    for i:=0; i<num_procs; i++ {
      <- coallate_done
    }
    close (coallate_chan);
  }
  go done_control();

  output_merge := func () {
    for v:= range coallate_chan {
        proof_list <- v
    }
    close (proof_list)
  }
  go output_merge();
}



func expand_n (array_a NumCol) []SolLst {
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
  


  len_array_m1 := len(array_a)-1;

  for i:=0; i<(len_array_m1); i++ {
    var ar_a,ar_b NumCol
    // for 3 items in arrar
    // {0},{1,2}, {0,1}{2}
    ar_a = make(NumCol, i+1)
    copy (ar_a, array_a[0:i+1]);
    ar_b = make(NumCol, ( len(array_a)-(i+1)));
    
    copy (ar_b, array_a[(i+1):(len(array_a))]);
    var work_item  SolLst; // {{2},{3,4}};
    // a work item always contains 2 elements to the array
    work_item = append(work_item, &ar_a,&ar_b);
    work_list = append(work_list, work_item);
  }
  return work_list
}

func check_return_list (proof_list SolLst, found_values *NumMap ) {
  value_check := make (map[int]int)

  for _, v:= range proof_list {
    // v is *NumLst
    for _,w:=range *v {
      // w is *Number
      var Value int;
      Value = w.Val;
      value_check[Value] = 1;
      //pretty.Println(w);
    }
  }

  tmp := found_values.GetVals()
  for _, v := range tmp{
    _, ok := value_check[v]
    // Every value in found_values should be in the list of values returned
    if (!ok) {
      fmt.Printf("%d in Number map, but is not in the proof list, which has %d Items\n",v,len(proof_list))
      //pretty.Println(found_values)
      //pretty.Println(proof_list)
      print_proofs(proof_list)
    }
  }
}
func find_proof(proof_list SolLst, to_find int) {
  found_val := false
  for _, v:= range proof_list {
    for _,w:=range *v {
      Value := w.Val
      proof_string := w.ProveIt()
      if (Value == to_find) {
        found_val = true
        fmt.Printf("Found Value %d, = %s\n",Value, proof_string);
      }
    }
  }
  if (!found_val) {fmt.Println("Unable to find value :",to_find);}
}
func print_proofs (proof_list SolLst) {
  for _, v:= range proof_list {
    // v is *NumCol
    for _,w:=range *v {
      // w is *Number
      var Value int;
      Value = w.Val
      proof_string := w.ProveIt()
      fmt.Printf("Value %3d, = %s\n",Value, proof_string);
    }
  }
  fmt.Println("Done printing proofs")
}

func main() {

  var target int
  target = 78

  var proof_list SolLst;
  var bob NumCol;
  found_values := NewNumMap(&proof_list);  //pass it the proof list so it can auto-check for validity at the end

  var tgflg = flag.Int("target", 0, "Define the target number to reach")
  var muflg = flag.Bool("mult", false, "Turn on multiplication")
  var dmflg = flag.Bool("dism", false, "Disable nultiplication")
  var stflg = flag.Bool("selft", false, "Check our own internals as we go")
  var srflg = flag.Bool("seeks", false, "Seek the shortest proof, as opposed to the quickest one to find")
  flag.Parse()

  // Global control flags default to test mode
  use_mult = *muflg
  self_test = *stflg
  seek_short = *srflg

  if (*tgflg>0) {
    if (*dmflg==false) {use_mult = true}
    fmt.Println("Set Target to ", *tgflg)
    target = *tgflg
    fmt.Println("Other args are ",flag.Args())
    for _,j:=range flag.Args() {
      value,err := strconv.ParseInt(j, 10,32)
      if (err!=nil) {
       log.Fatalf("Invalid command line ited, %s",j); 
      } else {
        var smv int
        smv = int(value)
        fmt.Println("Found an Number ",smv)
        bob.add_num(smv,found_values)
      }
      
    } 
  } else {
   use_mult = true
  bob.add_num(  8,found_values)
  bob.add_num(  9,found_values)
  bob.add_num( 10,found_values)
  bob.add_num( 75,found_values)
  bob.add_num( 25,found_values)
  bob.add_num(100,found_values)
  }
  return_proofs := make(chan SolLst, 16) 
  
  found_values.SetTarget(target)

  proof_list = append(proof_list, &bob);  // Add on the work item that is the source

  go permute_n(bob, found_values,return_proofs) 
  cleanup_packer:=0
  for v:= range return_proofs {
    if (self_test) {
      // This unused code is handy if we want a proof list
      proof_list = append(proof_list, v...)
      cleanup_packer++
      if (cleanup_packer>1000) {
        proof_list.CheckDuplicates()
        cleanup_packer = 0
      }
    }
  }
  // Close off the number map queue
  // also checks the proof_list passed to NewNumMap
  found_values.LastNumMap()

 
  found_values.PrintProofs()
}

