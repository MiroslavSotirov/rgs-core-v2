package api

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
)

// GameInitResponse reponse
type GameInitResponseV2 struct {
	Name        string                        `json:"name"`
	Version     string                        `json:"version"`
	Balance     engine.Money                  `json:"balance"`
	StakeValues []engine.Fixed                `json:"stakeValues"`
	DefaultBet  engine.Fixed                  `json:"defaultBet"`
	LastRound   map[string]GameplayResponseV2 `json:"lastRound"`
	Links       map[string]string             `json:"links"`    //name: url
	ReelSets    map[string]ReelResponse       `json:"reelSets"` // base, freeSpin, etc. as keys  might want to have this as ReelSetResponse
}

func (gi GameInitResponseV2) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (gi GameplayResponseV2) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GameLinkResponse ...
type GameLinkResponse struct {
	Results []LinkResponse `json:"results"`
}

//type GameplayResponseV2 struct {
//	Base BaseGameResponse `json:"base"`
//	FreeSpin FreeGameResponse `json:"freeSpin"`
//	Bonus PickGameResp `json:"bonusId"`
//}

type GameplayResponseV2 struct {
	SessionID   store.Token `json:"host/verified-token"`
	Stake       engine.Fixed
	Win         engine.Fixed
	CumWin      engine.Fixed `json:"cumulativeWin,omitempty"` // used for freespins/bonus rounds
	CurrentSpin int          `json:"currentSpin"`             // is this really needed ??
	FSRemaining int          `json:"freeSpinsRemaining,omitempty"`
	Balance     BalanceResponseV2
	View        [][]int        // includes row above and below
	Prizes      []engine.Prize `json:"wins"` // []WinResponseV2
}

type BalanceResponseV2 struct {
	Amount    engine.Money         `json:"amount"`
	FreeGames store.FreeGamesStore `json:"free_games"`
}

// todo: incorporate this into gameplay response
//type ChoiceResponse struct {
//	Position int `json:"position"`
//	Value string	`json:"value"`
//}
//type PickGameResp struct {
//	NumPicks int `json:"numPicks"`
//	PickedItems []ChoiceResponse `json:"pickedItems"`
//}

//type WinResponseV2 struct {
//	Type string
//	ID string // can include line id, prize name, etc
//	SymbolPositions []int `json:"symbols"`
//	Count int
//	SymbolId int
//	Amount float32
//	FreeSpins int `json:"freeSpins,omitempty"`
//}

type ReelResponse struct {
	ID         [][]int `json:"id"` // can omit if ordered properly
	MaxStack   [][]int `json:"maxStack"`
	MaxVisible [][]int `json:"maxVisible"`
	Count      [][]int `json:"count"`
}

func fillGamestateResponseV2(gamestate engine.Gamestate, balance store.BalanceStore) GameplayResponseV2 {

	var win engine.Fixed
	var cumWin engine.Fixed
	var stake engine.Fixed

	for _, tx := range gamestate.Transactions {
		switch tx.Type {
		case "WAGER":
			stake += tx.Amount.Amount
		case "PAYOUT":
			win += tx.Amount.Amount
			cumWin += tx.Amount.Amount
		}
	}

	resp := GameplayResponseV2{
		SessionID:   balance.Token,
		Stake:       stake,
		Win:         win,
		CumWin:      gamestate.CumulativeWin,
		CurrentSpin: gamestate.PlaySequence,         // zero-indexed
		FSRemaining: len(gamestate.NextActions) - 1, // for now, assume all future actions besides finish are fs (perhaps change this to bonusRdsRemaining in future)
		Balance:     BalanceResponseV2{
			Amount:    balance.Balance,
			FreeGames: store.FreeGamesStore{},
		},
		View:        gamestate.SymbolGrid,
		Prizes:      gamestate.Prizes,
	}
	return resp
}

func fillGameInitPreviousGameplay(previousGamestate engine.Gamestate, balance store.BalanceStore, gameId string) GameInitResponseV2 {
	var resp GameInitResponseV2
	logger.Debugf("previousGamestate: %v; balance: %v; gameId: %v; auth: %v", previousGamestate, balance, gameId)
	//if previousGamestate.Action != "" {
	// otherwise, assume this is first gp round

	lastRound := make(map[string]GameplayResponseV2, 1)
	lastRound[previousGamestate.Action] = fillGamestateResponseV2(previousGamestate, balance)
	resp.LastRound = lastRound
	//}
	resp.Name = gameId
	resp.Version = "2.0" // this is hardcoded for now
	resp.Balance = balance.Balance
	return resp
}

func (initResp *GameInitResponseV2) FillEngineInfo(config engine.EngineConfig) {
	// todo: this doesn't handle when there are multiple reel sets for a single def (i.e. multiple defs with same ID)
	reelResp := make(map[string]ReelResponse, len(config.EngineDefs))
	for _, def := range config.EngineDefs {
		var reels ReelResponse
		// todo do this smarter
		reels.ID = make([][]int, len(def.ViewSize))
		reels.Count = make([][]int, len(def.ViewSize))
		reels.MaxVisible = make([][]int, len(def.ViewSize))
		reels.MaxStack = make([][]int, len(def.ViewSize))
		//reels := make([]ReelResponse, len(def.Reels))
		for reel, reelContent := range def.Reels {
			//var reelResponse ReelResponse
			logger.Debugf("processing reel %v: %v", reel, reelContent)

			// todo: these maps are made too large right now
			symbolCounts := make(map[int]int, len(reelContent))
			symbolStacks := make(map[int]int, len(reelContent))
			symbolVisible := make(map[int]int, len(reelContent))
			var currStack int
			var currSymbol int
			//var maxSymbolId int
			wrappedReelContent := append(reelContent, reelContent[:def.ViewSize[reel]]...) // add first few symbols to end of reel

			for ii := 0; ii < len(reelContent); ii++ {
				symbol := reelContent[ii]
				visible := wrappedReelContent[ii : ii+def.ViewSize[reel]]
				var countVisible int
				for iii := 0; iii < len(visible); iii++ {
					if visible[iii] == symbol {
						countVisible++
					}
				}
				if countVisible > symbolVisible[symbol] {
					symbolVisible[symbol] = countVisible
				}

				symbolCounts[symbol]++
				if symbol == currSymbol {
					currStack++
				} else {
					currStack = 1
				}
				if currStack > symbolStacks[symbol] {
					symbolStacks[symbol] = currStack
				}
				currSymbol = symbol
				//if symbol > maxSymbolId {
				//	maxSymbolId = symbol
				//}
			}
			var ids []int
			var counts []int
			var maxStacks []int
			var maxVisibles []int
			for symbol := 0; symbol < len(symbolCounts); symbol++ {
				ids = append(ids, symbol)
				counts = append(counts, symbolCounts[symbol])
				maxStacks = append(maxStacks, symbolStacks[symbol])
				maxVisibles = append(maxVisibles, symbolVisible[symbol])
			}
			reels.ID[reel] = ids
			reels.Count[reel] = counts
			reels.MaxStack[reel] = maxStacks
			reels.MaxVisible[reel] = maxVisibles
			//logger.Debugf("made set of reel responses %v", reelResponses)

			//reels[reel] = reelResponse
		}

		reelResp[def.ID] = reels
	}
	initResp.ReelSets = reelResp
	return
}
