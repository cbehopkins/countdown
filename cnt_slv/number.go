package cntSlv

import (
	"fmt"
	"log"
	"strconv"
)

// number.go contains the basics of manipulating our number type
// A number says where it comes from and effectively says how
// one can make it

// Number is the struct for a single number
// That is possible built up from other numbers
type Number struct {
	// A number consists of
	Val        int       `json:"val"` // a value
	list       []*Number // a pointer the the list of numbers used to obtain this
	operation  string    // The operation used on those numbers to get here
	difficulty int
}

//NewNumber - Create a new number to work on
func NewNumber(inputA int, inputB []*Number, operation string, difficult int) *Number {
	var newNum Number
	newNum.configure(inputA, inputB, operation, difficult)
	return &newNum
}

func lessNumber(i, j interface{}) bool {
	tmp, ok := i.(*Number)
	if !ok {
		log.Fatal("Can't compare an empty number")
	}
	v1 := tmp.Val
	tmp, ok = j.(*Number)
	if !ok {
		log.Fatal("Can't compare an empty number")
	}
	v2 := tmp.Val
	return v1 < v2
}
func (nm *Number) configure(inputA int, inputB []*Number, operation string, difficult int) {
	nm.Val = inputA
	nm.list = inputB
	nm.operation = operation
	nm.difficulty = difficult
	for _, v := range inputB {
		nm.difficulty = nm.difficulty + v.difficulty
	}
}

//ProofLen returns the length of a proof
func (nm *Number) ProofLen() int {
	var cumlen int
	if nm.list == nil {
		cumlen = 1
	} else {
		for _, v := range nm.list {
			cumlen += v.ProofLen()
		}
	}
	return cumlen
}

// TidyDoubles - is useful for prittier proofs
// Remove any double notation in a proof
// we use   our own special notation to make things easier for ourselves
// However it's better to remove it at the tidy stage
// To make reducing the proof sizes easier
// Here's what our operands say
// a-b == b--a
// a/b == b\\a
func (nm *Number) TidyDoubles() {

	if (nm.list == nil) || (len(nm.list) == 0) {
		return
	}

	for _, v := range nm.list {
		v.TidyDoubles()
	}

	if nm.operation == "--" {
		if len(nm.list) != 2 {
			log.Fatal("can't process -- on a list that is anything but 2 long")
		}

		nm.operation = "-"
	} else if nm.operation == "\\" {
		if len(nm.list) != 2 {
			log.Fatal("can't process \\ on a list that is anything but 2 long")
		}

		nm.operation = "/"
	} else {
		// Must not be a double operator
		return
	}
	nm.list = NumCol{nm.list[1], nm.list[0]}
	return
}
func (nm *Number) tidyOperators() {
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
	tmpList := make([]*Number, 0, 4) // CBH get this from the centra allocator
	listModified := false
	for _, v := range nm.list {
		v.tidyOperators()
		// Let's just combine +s for now
		if (nm.operation == "+") && (v.operation == "+") {
			nm.difficulty = nm.difficulty + v.difficulty
			tmpList = append(tmpList, v.list...)
			listModified = true
		} else if (nm.operation == "*") && (v.operation == "*") {
			tmpList = append(tmpList, v.list...)
			nm.difficulty = nm.difficulty + v.difficulty
			listModified = true
		} else {
			tmpList = append(tmpList, v)
		}
	}
	if listModified {
		nm.list = tmpList
	}

	if (nm.operation == "-") && (len(nm.list) == 2) {
		// Play it safe and check first, work out optimisation later
		if (nm.list[0].operation == "-") && (nm.list[1].operation == "-") {
			// Fill in this later optimisaton
			// basically turn (a-b)-(c-d) -> (a+d)-(b+c)
		} else if nm.list[0].operation == "-" {
			// Transform (a-b)-c -> a-(b+c)
			// in this terminology
			// a = i.list[0].list[0]
			// b = i.list[0].list[1]
			// c = i.list[1]
			// create b+c
			myList0 := make([]*Number, 2)
			myList0[0] = nm.list[1]
			myList0[1] = nm.list[0].list[1]

			bPlusC := NewNumber((nm.list[1].Val + nm.list[0].list[1].Val), myList0, "+", (nm.list[1].difficulty + nm.list[0].list[1].difficulty + 1))

			myList1 := make([]*Number, 2)
			myList1[0] = nm.list[0].list[0]
			myList1[1] = bPlusC
			newNum := NewNumber(nm.Val, myList1, "-", (bPlusC.difficulty + nm.list[0].list[0].difficulty + 1))
			nm = newNum
			//i.TidyOperators()
			nm.ProveSol() //CBH we've made serious modification so test it
		} else if nm.list[1].operation == "-" {
			// Transform a-(b-c) -> (a+c)-b
			// in this terminology
			// a = i.list[0]
			// b = i.list[1].list[0]
			// c = i.list[1].list[1]

			// create a+c
			myList0 := make(NumCol, 2)
			myList0[0] = nm.list[0]
			myList0[1] = nm.list[1].list[1]
			aPlusC := NewNumber((myList0[0].Val + myList0[1].Val), myList0, "+", (myList0[0].difficulty + myList0[1].difficulty + 1))

			myList1 := make(NumCol, 2)
			myList1[0] = aPlusC
			myList1[1] = nm.list[1].list[0]
			newNum := NewNumber(nm.Val, myList1, "-", (aPlusC.difficulty + myList1[1].difficulty + 1))

			nm = newNum
			nm.ProveSol()
		}
	}

}

// ProveSol will prove the soltion and return the value it calculated
func (nm *Number) ProveSol() int {
	// This function should go through the list and prove the solution
	// Also do other sanity checking like the ,/- operators only have 2 items in the list
	// That anything with a valid operator has >1 item in the list
	runningTotal := 0
	firstRun := true
	if (nm.list == nil) || (len(nm.list) == 0) {
		// This is a source value
		return nm.Val
	} else if len(nm.list) == 1 {
		fmt.Println(nm)
		log.Fatal("Error invalid list length")
		return 0
	} else {
		for _, v := range nm.list {
			if firstRun {
				firstRun = false
				runningTotal = v.ProveSol()
			} else {
				switch nm.operation {
				case "+":
					runningTotal = runningTotal + v.ProveSol()
				case "-":
					runningTotal = runningTotal - v.ProveSol()
				case "--":
					runningTotal = v.ProveSol() - runningTotal
				case "*":
					runningTotal = runningTotal * v.ProveSol()
				case "/":
					if (runningTotal % v.ProveSol()) != 0 {
						log.Println("Divide had a remainder", runningTotal, v.ProveSol())
						return 0
					}
					runningTotal = runningTotal / v.ProveSol()
				case "\\":
					if (v.ProveSol() % runningTotal) != 0 {
						log.Println("Divide had a remainder", v.ProveSol(), runningTotal)
						return 0
					}
					runningTotal = v.ProveSol() / runningTotal
				default:
					log.Fatal("Unknown operation type")
				}
			}
		}
		if runningTotal != nm.Val {
			fmt.Println(nm)

			fmt.Println("We calculated ", runningTotal, nm.String())
			log.Fatal("Failed to self check solution")
		}
		return runningTotal
	}
}

func (nm *Number) setDifficulty() int {
	if (nm.list == nil) || (len(nm.list) == 0) {
		nm.difficulty = 0
		return 0
	}
	switch nm.operation {
	case "+":
		nm.difficulty = 1
	case "-", "--":
		nm.difficulty = 1
	case "*":
		nm.difficulty = 2
	case "/", "\\":
		nm.difficulty = 3
	default:
		log.Fatal("Unknown operation type")
	}
	for _, v := range nm.list {
		nm.difficulty += v.setDifficulty()
	}
	return nm.difficulty
}

func (nm *Number) String() string {
	var proof string
	var val int
	val = nm.Val
	if nm.list == nil {
		//proof = fmt.Sprintf("%d", val)
		proof = strconv.Itoa(val)
	} else {
		proof = ""
		op := ""
		NumCol(nm.list).Tidy()

		for _, v := range nm.list {

			switch nm.operation {
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
				op = nm.operation
			}

		}
		proof = "(" + proof + ")"

	}
	return proof
}
func (n *Number) render() string {
	return "Value " + strconv.Itoa(n.Val) + ", = " + n.String()
}
