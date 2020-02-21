package api

import (
	"encoding/base64"
	"github.com/go-chi/chi"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"html/template"
	"net/http"
)

type PlaycheckFields struct {
	Gamestate  engine.Gamestate
	GameID     string
	Wager      float32
	Payout     float32
	Currency   string
	BetPerLine float32
	SymbolGrid [][]int
	ColSize    int
}

func playcheck(request *http.Request, w http.ResponseWriter) {
	// gets state for a certain gameplay

	gameplayID := chi.URLParam(request, "gameplayID")

	gamestate := request.FormValue("state")
	var gsbytes []byte
	var err error

	if gamestate == "" {
		gamestateStore, serr := store.ServLocal.GamestateById(gameplayID)
		if serr != nil {
			logger.Errorf("Error getting gamestate item : %v", serr)
			return
		}
		gsbytes = gamestateStore.GameState

	} else {
		gsbytes, err = base64.StdEncoding.DecodeString(gamestate)
		if err != nil {
			logger.Errorf("Error decoding gs: %v", err)
			return
		}
	}
	gameplay := store.DeserializeGamestateFromBytes(gsbytes)
	t, err := template.ParseFiles("templates/api/playcheck/playcheck.html")
	if err != nil {
		logger.Errorf("template parsing error: ", err)
		return
	}

	gameID, _ := engine.GetGameIDAndReelset(gameplay.GameID)
	//t := template.Must(template.New("api/playcheck").Funcs(tplFuncMap).ParseGlob("*.html"))
	var wager float32
	var payout float32
	for i := 0; i < len(gameplay.Transactions); i++ {
		switch gameplay.Transactions[i].Type {
		case "WAGER":
			wager = gameplay.Transactions[i].Amount.Amount.ValueAsFloat()
		case "PAYOUT":
			payout = gameplay.Transactions[i].Amount.Amount.ValueAsFloat()
		}
	}

	// transform symbolgrid
	symbolGrid := engine.TransposeGrid(gameplay.SymbolGrid)
	currency := gameplay.Transactions[0].Amount.Currency
	betPerLine := gameplay.BetPerLine.Amount.ValueAsFloat()
	colSize := len(symbolGrid[0])
	fields := PlaycheckFields{gameplay, gameID, wager, payout, currency, betPerLine, symbolGrid, colSize}
	err = t.Execute(w, fields)
	if err != nil {
		logger.Errorf("template executing error: ", err)
		return
	}
}
