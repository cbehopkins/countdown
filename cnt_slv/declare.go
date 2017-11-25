package cntSlv

// declare.go is where misc things are declared

// CountHelper an exportable funciton to help externals work with us
func (nm *NumMap) CountHelper(target int, sources []int) chan SolLst {

	// Create a list of the input sources
	srcNumbers := nm.NewNumCol(sources)
	nm.SetTarget(target)

	return permuteN(srcNumbers, nm)
}

// CountFastHelper is the fast version
// actually not really any faster it turns out! (but uses less memory)
func CountFastHelper(target int, sources []int, findShortest bool) string {
	ps := newFastPermInt(sources)
	tgt := target
	if findShortest {
		tgt = 0
	}
	rPs := ps.GetProofs(tgt)
	return rPs.Get(target).String()

}
