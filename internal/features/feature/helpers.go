package feature

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

func WeightedRandomIndex(weights []int) int {
	var sum, i, w int
	for _, w = range weights {
		sum += w
	}
	r := rng.RandFromRange(sum)
	sum = 0
	for i, w = range weights {
		sum += w
		if r < sum {
			break
		}
	}
	return i
}

func RandomPermutation(arr []int) []int {
	a := make([]int, len(arr))
	for i, v := range arr {
		a[i] = v
	}
	ret := []int{}
	for len(a) > 0 {
		i := rng.RandFromRange(len(a))
		ret = append(ret, a[i])
		a0 := a[:i]
		a1 := a[i+1:]
		a = append(a0, a1...)
	}
	return ret
}

func (state FeatureState) GetCascadePositions() []int {

	if len(state.CascadePositions) > 0 {
		return state.CascadePositions
	} else {
		cascadePositions := make([]int, len(state.SymbolGrid))
		for r := range cascadePositions {
			cascadePositions[r] = len(state.SymbolGrid[r])
		}
		return cascadePositions
	}
}

func (state FeatureState) GetCandidatePositions() []int {

	cascadePos := state.GetCascadePositions()
	p := 0
	pos := []int{}
	for i, r := range state.SourceGrid {
		for j := range r {
			if j < cascadePos[i] {
				pos = append(pos, p)
			}
			p++
		}
	}

	return pos
}
