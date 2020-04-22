package config

import (
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
)

func GetWalletFromOperatorAndMode(operator string, mode string) (string, rgse.RGSErr) {
	//todo make this method use a config parser, store info in yaml file
	//todo return interface
	if operator == "mav" {
		switch mode {
		case "real":
			return "dashur", nil
		case "demo":
			return "demo", nil
		}
	}
	return "", rgse.Create(rgse.BadOperatorConfig)
}
