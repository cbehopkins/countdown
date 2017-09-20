package cntSlv

import "log"

type operator []byte

var plusOperator operator
var multOperator operator
var minusOperator operator
var divOperator operator

func initOperators() {
	plusOperator = newOperator("+")
	multOperator = newOperator("*")
	minusOperator = newOperator("-")
	divOperator = newOperator("/")
}

func newOperator(in string) operator {
	return operator(in)
}

func (op operator) String() string {
	return string(op)
}
func (op operator) Bytes() []byte {
	return []byte(op)
}
func determineOperator(in string) func(int, int) int {
	switch in {
	case "+":
		return func(a, b int) int {
			return a + b
		}
	case "*":
		return func(a, b int) int {
			return a * b
		}
	case "-":
		return func(a, b int) int {
			return a - b
		}
	case "/":
		return func(a, b int) int {
			return a / b
		}
	default:
		log.Fatal("Invalid Operator", in)
		return func(a, b int) int { return -1 }
	}
}
