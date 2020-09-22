package api

import (
	"encoding/base64"
	"fmt"
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
	Wager      string
	Payout     string
	Currency   string
	BetPerLine string
	SymbolGrid [][]int
	ColSize    int
}

func playcheck(request *http.Request, w http.ResponseWriter) {
	// gets state for a certain gameplay
	w.Header().Set("Content-Type", "text/html")

	gameplayID := chi.URLParam(request, "gameplayID")

	gamestate := request.FormValue("state")
	var gsbytes []byte
	var err error
	var fm = template.FuncMap{
		"Iter": func(s int) []int {
			var Size []int
			for x := 0; x < (s); x++ {
				Size = append(Size, x)
			}
			return Size
		},
	}

	if gamestate == "" {
		gamestateStore, serr := store.ServLocal.GamestateById(gameplayID)
		if serr != nil {
			logger.Errorf("Error getting gamestate item : %v", serr)
			fmt.Fprint(w, "<center><h1>Bad Gamestate ID</h1></center>")
			return
		}
		gsbytes = gamestateStore.GameState

	} else {
		gsbytes, err = base64.StdEncoding.DecodeString(gamestate)
		if err != nil {
			logger.Errorf("Error decoding gs: %v", err)
			fmt.Fprint(w, "<center><h1>Gamestate Decoding Error</h1></center>")
			return
		}
	}
	gameplay := store.DeserializeGamestateFromBytes(gsbytes)
	logger.Infof("gameplay : %#v", gameplay)
	tpl := template.New("playcheck.html").Funcs(fm)
	t, err := tpl.ParseFiles("templates/api/playcheck/playcheck.html")
	if err != nil {
		logger.Errorf("Template parsing error: ", err)
		fmt.Fprint(w, "<center><h1>Template parsing error </h1></center>")
		return
	}

	//t := template.Must(template.New("api/playcheck").Funcs(tplFuncMap).ParseGlob("*.html"))
	var wager string
	var payout string
	for i := 0; i < len(gameplay.Transactions); i++ {
		switch gameplay.Transactions[i].Type {
		case "WAGER":
			wager = gameplay.Transactions[i].Amount.Amount.ValueAsString()
		case "PAYOUT":
			payout = gameplay.Transactions[i].Amount.Amount.ValueAsString()
		}
	}

	// transform symbolgrid
	symbolGrid := engine.TransposeGrid(gameplay.SymbolGrid)
	currency := gameplay.Transactions[0].Amount.Currency
	betPerLine := gameplay.BetPerLine.Amount.ValueAsString()
	var colSize int
	if len(symbolGrid) > 0 {
		colSize = len(symbolGrid[0])
	} else {
		colSize = 0
	}
	fields := PlaycheckFields{gameplay, gameplay.Game, wager, payout, currency, betPerLine, symbolGrid, colSize}
	err = t.Execute(w, fields)
	if err != nil {
		logger.Errorf("template executing error: ", err)
		fmt.Fprint(w, "<center><h1>Template Execution Error</h1></center>")
		return
	}
}
