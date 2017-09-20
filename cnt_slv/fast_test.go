package cntSlv

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	//	"github.com/pkg/profile"
)

func ExampleProof() {
	proof1 := *NewProof(1)
	proof2 := *NewProof(2)
	proof3 := *NewProof(3)
	proof3a := proof1.concat(proof2, newOperator("+"))
	proof6 := proof2.concat(proof3, newOperator("*"))
	proof7 := proof6.concat(proof1, newOperator("+"))
	proof9 := proof3.concat(proof3a, newOperator("*"))

	fmt.Println("Received", proof9)
	fmt.Println("Received", proof7)
	//Output:
	// Received (3*(1+2))
	// Received ((2*3)+1)
}

func testExists(expect []int, proofs Proofs) {
	for _, v := range expect {
		if !proofs.Exists(v) {
			log.Fatalf("Error, expected %v to exist in proof, it doesn't\n", v)
		}
	}
	isIt := func(val int) bool {
		for _, v := range expect {
			if val == v {
				return true
			}
		}
		return false
	}
	for i, pr := range proofs {
		if i > 0 && (pr.Len() > 0) {
			if !isIt(i) {
				log.Fatalf("%v exists in proofs but not in reference", i)
			}
		}
	}
}
func TestWrkFast0(t *testing.T) {

	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(2)
	inP.Init(6)
	// What numbers can we prove from this?
	proofs := getProofs()
	defer putProofs(proofs)
	proofs.wrkFast(*inP)
	fmt.Println("Proofs:", proofs)
	// It should be possible to generate all these numbers fromt his input
	expectedResults := []int{2, 3, 4, 6, 8, 12}
	testExists(expectedResults, proofs)

}
func TestWrkFast1(t *testing.T) {

	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(100)
	inP.Init(25)
	inP.Init(8)
	inP.Init(1)
	inP.Init(9)
	inP.Init(10)
	// What numbers can we prove from this?
	proofs := getProofs()
	defer putProofs(proofs)
	proofs.wrkFast(*inP)
	fmt.Println("Proofs:", proofs)

	expectedResults := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
		41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
		51, 52, 53, 54, 55, 56, 57, 58, 59, 60,
		61, 62, 63, 64, 65, 66, 67, 68, 69, 70,
		71, 72, 73, 74, 75, 76, 77, 78, 79, 80,
		81, 82, 83, 84, 85, 86, 87, 88, 89, 90,
		91, 92, 93, 94, 95, 96, 97, 98, 99, 100,
		101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
		111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
		121, 122, 123, 124, 125, 126, 127, 128, 129, 130,
		131, 132, 133, 134, 135, 136, 137, 138, 139, 140,
		141, 142, 143, 144, 145, 146, 147, 148, 149, 150,
		151, 152, 153, 154, 155, 156, 157, 158, 160, 162,
		163, 164, 165, 166, 167, 169, 170, 171, 172, 173,
		174, 175, 176, 177, 178, 179, 180, 181, 182, 183,
		184, 185, 187, 188, 189, 190, 191, 192, 194, 195,
		196, 197, 198, 199, 200, 201, 202, 204, 205, 206,
		207, 208, 209, 210, 211, 212, 213, 214, 215, 216,
		217, 218, 219, 220, 222, 223, 224, 225, 226, 227,
		228, 230, 233, 234, 235, 236, 240, 242, 243, 244,
		245, 246, 247, 248, 250, 252, 253, 254, 255, 256,
		258, 260, 262, 263, 264, 265, 266, 269, 270, 271,
		272, 274, 275, 276, 277, 278, 279, 280, 281, 282,
		284, 285, 287, 288, 289, 290, 291, 292, 294, 295,
		296, 297, 298, 299, 300, 301, 302, 304, 305, 306,
		307, 308, 309, 310, 311, 314, 315, 316, 318, 319,
		320, 323, 324, 325, 326, 328, 330, 332, 334, 335,
		340, 342, 344, 350, 352, 354, 356, 360, 363, 364,
		365, 368, 370, 374, 375, 378, 380, 387, 388, 389,
		390, 391, 392, 394, 396, 397, 398, 400, 404, 406,
		407, 410, 415, 416, 420, 423, 425, 430, 432, 435,
		440, 442, 450, 456, 460, 463, 470, 475, 480, 490,
		494, 500, 505, 506, 508, 509, 510, 511, 515, 516,
		520, 524, 525, 526, 527, 530, 532, 534, 535, 536,
		540, 544, 546, 550, 555, 556, 560, 565, 570, 575,
		576, 580, 581, 582, 584, 585, 587, 589, 590, 591,
		592, 593, 594, 595, 598, 599, 600, 601, 602, 603,
		604, 605, 608, 609, 610, 611, 612, 613, 615, 618,
		619, 620, 622, 625, 626, 627, 628, 630, 634, 636,
		637, 640, 644, 645, 646, 650, 653, 654, 655, 656,
		660, 664, 665, 666, 670, 674, 675, 676, 680, 681,
		684, 685, 687, 689, 690, 691, 694, 695, 699, 700,
		701, 703, 705, 708, 709, 710, 711, 712, 715, 716,
		719, 720, 724, 725, 726, 727, 728, 730, 732, 735,
		736, 737, 738, 740, 745, 746, 747, 748, 750, 753,
		755, 756, 757, 760, 765, 766, 770, 775, 780, 781,
		782, 785, 787, 789, 790, 791, 792, 795, 796, 798,
		799, 800, 801, 802, 803, 804, 806, 808, 809, 810,
		811, 812, 814, 818, 819, 820, 825, 827, 828, 829,
		830, 835, 837, 838, 840, 845, 847, 850, 853, 854,
		856, 864, 865, 866, 870, 874, 875, 876, 880, 881,
		884, 885, 889, 890, 891, 894, 899, 900, 901, 909,
		910, 911, 913, 919, 920, 925, 926, 930, 935, 936,
		946, 953, 960, 962, 963, 965, 970, 971, 972, 973,
		980, 981, 982, 989, 990, 991, 992, 998, 999, 1000,
		1001, 1002, 1008, 1009, 1010, 1011, 1012, 1018, 1019, 1020}
	testExists(expectedResults, proofs)
}
func TestWrkFastGen0(t *testing.T) {
	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(2)
	inP.Init(6)
	// What numbers can we prove from this?
	proofs := NewProofs()
	if inP.Len() == 1 {
		log.Fatal("Yes Really!")
	} else {
		log.Println("Not really.")
	}
	proofs.wrkFastGen(*inP, false, false)
	fmt.Println("Proofs:", proofs)
	expectedResults := []int{4, 3, 8, 12}
	testExists(expectedResults, proofs)
}
func TestWrkFastGen1(t *testing.T) {
	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(2)
	inP.Init(6)
	inP.Init(4)
	// What numbers can we prove from this?
	proofs := NewProofs()
	fmt.Println("Before:", inP)

	proofs.wrkFastGen(*inP, false, false)
	fmt.Println("Proofs:", proofs)
	expectedResults := []int{1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 16, 20, 22, 24, 26, 32, 48}
	testExists(expectedResults, proofs)
}

func ExampleWrkFastSplit() {
	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(2)
	inP.Init(4)
	inP.Init(6)
	inP.Init(8)
	resP := wrkFastSplit(*inP)
	fmt.Println("Res:", resP)
	putProofListArray(resP)
	//Output:
	// Res: [{2->2} {4->4,6->6,8->8} {2->2,4->4} {6->6,8->8} {2->2,4->4,6->6} {8->8}]
}
func BenchmarkBasic(b *testing.B) {
	//fer profile.Start().Stop()

	runInts := []int{2, 4, 6, 25, 100, 4}
	poolModes := []bool{true, false}
	parModes := []bool{true, false}
	for _, pool := range poolModes {
		for _, par := range parModes {
			for i := 2; i < 7; i++ {
				inP := newProofLst(0)
				// Initalise it with some numbers
				for j := 0; j < i; j++ {
					inP.Init(runInts[j])
				}

				runFunc := func(tb *testing.B) {
					for i := 0; i < tb.N; i++ {
						proofs := NewProofs()
						proofs.wrkFastGen(*inP, pool, par)
					}
				}

				runString := "Cnt:" + strconv.Itoa(i)
				if pool {
					runString += " Pooled"
				}
				if par {
					runString += " Par"
				}
				b.ResetTimer()
				b.Run(runString, runFunc)
			}
		}
	}
}

func TestDecompose0(t *testing.T) {
	inStrings :=
		[]string{
			"3", "(3)",
			"(2+3)",
			"(3-2)",
			"((3-2)*5)",
			"((8-2)/(1+2))",
		}
	for _, inString := range inStrings {
		derNum := parseString(inString)
		if derNum == nil {
			log.Fatal("Nil result for:", inString)
		}
		derString := derNum.String()
		if derString != inString {
			log.Fatal("Strings not equal:", derString, inString)
		} else {
			//log.Println("Strings are equal:", derString, inString)
		}
	}
}
func TestDecompose1(t *testing.T) {

	// Create an empty List
	inP := newProofLst(0)
	// Initalise it with some numbers
	inP.Init(100)
	inP.Init(25)
	inP.Init(8)
	inP.Init(1)
	inP.Init(9)
	inP.Init(10)
	// What numbers can we prove from this?
	proofs := getProofs()
	proofs.wrkFast(*inP)
	for i, pr := range proofs {
		derNum := pr.parseProof()
		if derNum != nil {
			dv := derNum.calculate()
			if dv != i {
				log.Fatal("Wrong Value", pr, i, dv)
			}
		}
	}
}
