package featureProducts

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

// init function to import package and evaluate static variable feature factories
func Register() {
	logger.Infof("Register feature products")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
