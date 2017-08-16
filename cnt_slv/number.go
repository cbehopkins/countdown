package cnt_slv

import (
	"fmt"
	"github.com/tonnerre/golang-pretty"
	"log"
	"strconv"
)

// number.go contains the basics of manipulating our number type
// A number says where it comes from and effectively says how
// one can make it
type Number struct {
	// A number consists of
	Val        int       `json:"val"` // a value
	list       []*Number // a pointer the the list of numbers used to obtain this
	operation  string    // The operation used on those numbers to get here
	difficulty int
}

func NewNumber(input_a int, input_b []*Number, operation string, difficult int) *Number {
	var new_num Number
	new_num.configure(input_a, input_b, operation, difficult)
	return &new_num
}

func (num *Number) configure(input_a int, input_b []*Number, operation string, difficult int) {
	num.Val = input_a

	num.list = input_b
	num.operation = operation
	num.difficulty = difficult
	for _, v := range input_b {
		num.difficulty = num.difficulty + v.difficulty
	}

}

func (i *Number) ProofLen() int {
	var cumlen int
	if i.list == nil {
		cumlen = 1
	} else {
		for _, v := range i.list {
			cumlen += v.ProofLen()
		}
	}
	return cumlen
}
func (i *Number) TidyDoubles() {
	// Remove any double notation in a proof
	// we use   our own special notation to make things easier for ourselves
	// However it's better to remove it at the tidy stage
	// To make reducing the proof sizes easier
	// Here's what our operands say
	// a-b == b--a
	// a/b == b\\a

	if (i.list == nil) || (len(i.list) == 0) {
		return
	}

	for _, v := range i.list {
		v.TidyDoubles()
	}

	if i.operation == "--" {
		if len(i.list) != 2 {
			log.Fatal("can't process -- on a list that is anything but 2 long")
		}

		i.operation = "-"
	} else if i.operation == "\\" {
		if len(i.list) != 2 {
			log.Fatal("can't process \\ on a list that is anything but 2 long")
		}

		i.operation = "/"
	} else {
		// Must not be a double operator
		return
	}
	i.list = NumCol{i.list[1], i.list[0]}
	return
}
func (i *Number) TidyOperators() {
	// This one is sexy
	// we often in our proofs get things like:
	// (((1+2)+3)+(4/2)) or
	// (((8-2)-1)-2)
	// Which could of course both be simplified
	// So what we will do is re-write the tree structure of our proofs
	// Things are easy with + as we can just descend the tree and if the next level down uses a + as well
	// Then we can just combine them

	// Think about the use case:
	// ((1+2)+(3+4))
	// We will first read (3+7)
	// Look at the 3 and see how we got it.
	// we will see (1+2) uses the same operator
	// so we can pull that into ours
	// The same applies to multiples

	// When it comes to subtract and divide we have an issue
	// ((a-b)-c)-d == (a-(b+c+d)) <- much tidier
	// so let's look at: (a-b)-c as a starting point
	// Actually represented as something like:
	// g-c and we look at g and find it is a-b
	// but we could say that:
	// * if we are a subtract and the (first) leaf is a subtract
	// * Create a new number that is the leaf's second number + our second number
	// * Set our first number to the leaf's First number
	// * Set our second number to the new number we just made
	// Likewise for: a-(b-c) -> (a+c)-b
	// * if we are a subtract and the (second) leaf is a subtract
	// * Create a new number that is the leaf's second number + our First Number
	// * Set our first number to the  new number we just made
	// * Set our second number to the leaf's First number
	// Now we could get clever for things like merging addition if come the the
	// numbers in out subtraction turned into an addition, or we could just run ourselves
	// on that new number whch will merge up any additions

	// But of course the first thing we want is for their house to be in order
	tmp_list := make([]*Number, 0, 4) // CBH get this from the centra allocator
	list_modified := false
	for _, v := range i.list {
		v.TidyOperators()
		// Let's just combine +s for now
		if (i.operation == "+") && (v.operation == "+") {
			i.difficulty = i.difficulty + v.difficulty
			tmp_list = append(tmp_list, v.list...)
			list_modified = true
		} else if (i.operation == "*") && (v.operation == "*") {
			tmp_list = append(tmp_list, v.list...)
			i.difficulty = i.difficulty + v.difficulty
			list_modified = true
		} else {
			tmp_list = append(tmp_list, v)
		}
	}
	if list_modified {
		i.list = tmp_list
	}

	if (i.operation == "-") && (len(i.list) == 2) {
		// Play it safe and check first, work out optimisation later
		if (i.list[0].operation == "-") && (i.list[1].operation == "-") {
			// Fill in this later optimisaton
			// basically turn (a-b)-(c-d) -> (a+d)-(b+c)
		} else if i.list[0].operation == "-" {
			// Transform (a-b)-c -> a-(b+c)
			// in this terminology
			// a = i.list[0].list[0]
			// b = i.list[0].list[1]
			// c = i.list[1]
			// create b+c
			my_list0 := make([]*Number, 2)
			my_list0[0] = i.list[1]
			my_list0[1] = i.list[0].list[1]

			b_plus_c := NewNumber((i.list[1].Val + i.list[0].list[1].Val), my_list0, "+", (i.list[1].difficulty + i.list[0].list[1].difficulty + 1))

			my_list1 := make([]*Number, 2)
			my_list1[0] = i.list[0].list[0]
			my_list1[1] = b_plus_c
			new_num := NewNumber(i.Val, my_list1, "-", (b_plus_c.difficulty + i.list[0].list[0].difficulty + 1))
			i = new_num
			//i.TidyOperators()
			i.ProveSol() //CBH we've made serious modification so test it
		} else if i.list[1].operation == "-" {
			// Transform a-(b-c) -> (a+c)-b
			// in this terminology
			// a = i.list[0]
			// b = i.list[1].list[0]
			// c = i.list[1].list[1]

			// create a+c
			my_list0 := make(NumCol, 2)
			my_list0[0] = i.list[0]
			my_list0[1] = i.list[1].list[1]
			a_plus_c := NewNumber((my_list0[0].Val + my_list0[1].Val), my_list0, "+", (my_list0[0].difficulty + my_list0[1].difficulty + 1))

			my_list1 := make(NumCol, 2)
			my_list1[0] = a_plus_c
			my_list1[1] = i.list[1].list[0]
			new_num := NewNumber(i.Val, my_list1, "-", (a_plus_c.difficulty + my_list1[1].difficulty + 1))

			i = new_num
			//i.TidyOperators()
			i.ProveSol()
		}
	}

}

func (i *Number) ProveSol() int {
	// This function should go through the list and prove the solution
	// Also do other sanity checking like the ,/- operators only have 2 items in the list
	// That anything with a valid operator has >1 item in the list
	running_total := 0
	first_run := true
	if (i.list == nil) || (len(i.list) == 0) {
		// This is a source value
		return i.Val
	} else if len(i.list) == 1 {
		pretty.Print(i)
		log.Fatal("Error invalid list length")
		return 0
	} else {
		for _, v := range i.list {
			if first_run {
				//pretty.Print(v)
				first_run = false
				running_total = v.ProveSol()
			} else {
				switch i.operation {
				case "+":
					running_total = running_total + v.ProveSol()
				case "-":
					running_total = running_total - v.ProveSol()
				case "--":
					running_total = v.ProveSol() - running_total
				case "*":
					running_total = running_total * v.ProveSol()
				case "/":
					running_total = running_total / v.ProveSol()
				case "\\":
					running_total = v.ProveSol() / running_total
				default:
					log.Fatal("Unknown operation type")
				}
			}
		}
		if running_total != i.Val {
			pretty.Println(i)

			fmt.Println("We calculated ", running_total, i.String())
			log.Fatal("Failed to self check solution")
		}
		return running_total
	}
}
func (i *Number) SetDifficulty() int {
	if (i.list == nil) || (len(i.list) == 0) {
		i.difficulty = 0
		return 0
	}
	switch i.operation {
	case "+":
		i.difficulty = 1
	case "-", "--":
		i.difficulty = 1
	case "*":
		i.difficulty = 2
	case "/", "\\":
		i.difficulty = 3
	default:
		log.Fatal("Unknown operation type")
	}
	for _, v := range i.list {
		i.difficulty += v.SetDifficulty()
	}
	return i.difficulty
}

func (i *Number) String() string {
	var proof string
	var val int
	val = i.Val
	//pretty.Print(i)
	if i.list == nil {
		//proof = fmt.Sprintf("%d", val)
		proof = strconv.Itoa(val)
	} else {
		proof = ""
		op := ""
		for _, v := range i.list {

			switch i.operation {
			case "--":

				proof = v.String() + op + proof
				op = "-"
				//proof = proof + "--" + v.ProveIt()
			case "\\":
				proof = v.String() + op + proof
				op = "/"
				//proof = proof + "//" + v.ProveIt()
			default:
				proof = proof + op + v.String()
				//proof = v.ProveIt() + op + proof
				op = i.operation
			}

		}
		proof = "(" + proof + ")"

	}
	return proof
}
