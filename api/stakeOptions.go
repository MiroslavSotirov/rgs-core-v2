package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type GameStakeInfo struct {
	GameName  string    `json:"name"`
	GameID    string    `json:"id"`
	BetLevels []BetInfo `json:"betLevels"`
}

// Render Gameplay Resonse
func (gp GameStakeInfo) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type BetInfo struct {
	Currency string   `json:"currency"`
	Values   []string `json:"values"`
}

func stakeInfo(request *http.Request, w http.ResponseWriter) {

	//basic auth
	un, pw, ok := request.BasicAuth()
	if !ok || un != "dashur" || pw != "Q3F5tMQSnGTsWsG8" {
		_ = render.Render(w, request, ErrRender(rgserror.Create(rgserror.InvalidCredentials)))
		return
	}

	// gets allowed stakes in base configuration for a certain game
	w.Header().Set("Content-Type", "info/json")

	gameSlug := request.FormValue("game")
	companyId := request.FormValue("companyId")
	var err error
	var stakeInfo GameStakeInfo
	stakeInfo.GameID = gameSlug
	stakeInfo.GameName = strings.Title(strings.ReplaceAll(gameSlug, "-", " "))
	EC, err := engine.GetEngineDefFromGame(gameSlug)

	if err != nil {
		logger.Errorf("Error getting engine def for game %v: %v", gameSlug, err)
		return
	}

	for i, ccy := range engine.Ccy_name {
		ccyInfo := BetInfo{Currency: ccy, Values: []string{}}
		if i == 0 {
			continue
		}
		stakeValues, _, _, _, err := parameterSelector.GetGameplayParameters(engine.Money{0, ccy}, "", gameSlug, companyId)
		if err != nil {
			logger.Errorf("Error getting stake values with ccy %v: %v", ccy, err)
			continue
		}

		// multiply stake per line by num lines or bet multiplier
		for j := 0; j < len(stakeValues); j++ {
			ccyInfo.Values = append(ccyInfo.Values, stakeValues[j].Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor)).ValueAsString())
		}
		stakeInfo.BetLevels = append(stakeInfo.BetLevels, ccyInfo)
	}
	if err := render.Render(w, request, stakeInfo); err != nil {
		_ = render.Render(w, request, ErrRender(err))
	}
}
