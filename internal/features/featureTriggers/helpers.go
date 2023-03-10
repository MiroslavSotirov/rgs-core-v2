package featureTriggers

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

// init function to import package and evaluate static variable feature factories
func Register() {
	logger.Infof("Register feature triggers")
}

func containsInt(array []int, value int) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func setGrid(dst [][]int, src [][]int) {
	for ir, r := range src {
		for is, s := range r {
			dst[ir][is] = s
		}
	}
}

func countSymbols(symbol int, grid [][]int) int {
	num := 0
	for _, r := range grid {
		for _, s := range r {
			if s == symbol {
				num++
			}
		}
	}
	return num
}
