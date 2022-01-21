package api

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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

	var gameV3 store.GameV3
	var istate engine.IGameStateV3
	var rgserr rgse.RGSErr
	istate, rgserr = gameV3.DeserializeState(gsbytes)
	if rgserr == nil {
		var stateV3 *engine.GameStateV3 = istate.Base()
		logger.Debugf("creating V3 playcheck for state = %#v\n", stateV3)

		var igameV3 store.IGameV3
		igameV3, rgserr = store.CreateGameV3(stateV3.Game)
		if rgserr != nil {
			logger.Debugf("could not create V3 game for %s\n", stateV3.Game)
			fmt.Fprintf(w, "<center><h1>Internal Error</h1>Unknown game: %s</center>", stateV3.Game)
			return
		}
		//	gameV3.PlayCheck(istate, w)
		istate, rgserr = igameV3.DeserializeState(gsbytes)
		if rgserr != nil {
			logger.Infof("Could not deserialize state for game %s though it was possible to decode a GameStateV3 = %#v\n", stateV3.Game, stateV3)
			fmt.Fprintf(w, "<center><h1>Internal Error</h1>Decoded gameId: %s could not deserialize %#v</center>", stateV3, stateV3.Game)
			return
		}

		switch stateV3.Game {
		case "dragon-roulette":
			playcheckRoulette(istate, w)
		default:
			logger.Infof("Can not produce playcheck for unknown V3 game \"%s\"", stateV3.Game)
		}
		return
	} else {
		logger.Debugf("creating V1/V2 playcheck")
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

type PlaycheckRoulette struct {
	Gamestate engine.GameStateRoulette
	GameId    string
	Wager     string
	Payout    string
	Currency  string
	Bets      []PlaycheckBetRoulette
	Prizes    []PlaycheckPrizeRoulette
	Symbol    int
	ColSize   int
}

type PlaycheckBetRoulette struct {
	Index   string
	Amount  string
	Symbols []int
}

type PlaycheckPrizeRoulette struct {
	Index  string
	Amount string
}

func playcheckRoulette(istate engine.IGameStateV3, w http.ResponseWriter) {
	var state *engine.GameStateRoulette = istate.(*engine.GameStateRoulette)

	logger.Debugf("creating playcheck roulette for state %#v", state)

	var fm = template.FuncMap{
		"Iter": func(s int) []int {
			var Size []int
			for x := 0; x < (s); x++ {
				Size = append(Size, x)
			}
			return Size
		},
	}
	tpl := template.New("playcheckroulette.html").Funcs(fm)
	t, err := tpl.ParseFiles("templates/api/playcheck/playcheckroulette.html")
	if err != nil {
		logger.Errorf("Template parsing error: ", err)
		fmt.Fprint(w, "<center><h1>Template parsing error </h1></center>")
		return
	}

	bets := make([]PlaycheckBetRoulette, len(state.Bets))
	idx := 0
	for k, b := range state.Bets {
		bets[idx].Index = k
		bets[idx].Amount = b.Amount.ValueAsString() + " " + state.Currency
		bets[idx].Symbols = b.Symbols
		idx++
	}

	prizes := make([]PlaycheckPrizeRoulette, len(state.Prizes))
	for i, p := range state.Prizes {
		prizes[i].Index = p.Index
		prizes[i].Amount = p.Amount.ValueAsString() + " " + state.Currency
	}

	fields := PlaycheckRoulette{
		Gamestate: *state,
		GameId:    state.Game,
		Wager:     state.Bet.ValueAsString(),
		Payout:    state.Win.ValueAsString(),
		Currency:  state.Currency,
		Bets:      bets,
		Prizes:    prizes,
		Symbol:    state.Symbol,
		ColSize:   1,
	}
	err = t.Execute(w, fields)
	if err != nil {
		logger.Errorf("template executing error: ", err)
		fmt.Fprint(w, "<center><h1>Template Execution Error</h1></center>")
		return
	}
}
