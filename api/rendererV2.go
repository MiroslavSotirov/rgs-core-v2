package api

import (
	"fmt"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strings"
)

// GameInitResponse reponse
type GameInitResponseV2 struct {
	Name        string                        `json:"name"`
	Version     string                        `json:"version"`
	//Balance     engine.Money                  `json:"balance"`
	Wallet      string                        `json:"wallet"`
	StakeValues []engine.Fixed                `json:"stakeValues"`
	DefaultBet  engine.Fixed                  `json:"defaultBet"`
	LastRound   map[string]GameplayResponseV2 `json:"lastRound"`
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
	SessionID store.Token          `json:"host/verified-token"`
	StateID   string              `json:"stateID"`
	RoundID   string              `json:"roundID"`
	ReelsetID int                 `json:"reelset"`
	Stake     engine.Fixed        `json:"totalStake"`
	Win       engine.Fixed
	RoundWin    engine.Fixed `json:"cumulativeWin,omitempty"` // used for freespins/bonus rounds
	SpinWin     engine.Fixed   `json:"spinWin"` // total win only on this spin
	FSRemaining *int               `json:"freeSpinsRemaining,omitempty"`
	Balance     BalanceResponseV2 `json:"balance"`
	View        [][]int           `json:"view"` // includes row above and below
	Prizes      []engine.Prize    `json:"wins"` // []WinResponseV2
	NextAction  string            `json:"nextAction"`
	Closed      bool               `json:"closed"`
	RoundMultiplier int            `json:"roundMultiplier"`
	Gamification *engine.GamestatePB_Gamification `json:"gamification"`
	CascadePositions []int         `json:"cascadePositions,omitempty"`
}

type BalanceResponseV2 struct {
	Amount    engine.Money `json:"amount"`
	FreeGames int          `json:"freeGames"`
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
	Type       string  `json:"type"`
	BetMult    int     `json:"betMultiplier"`
}

func fillGamestateResponseV2(gamestate engine.Gamestate, balance store.BalanceStore) GameplayResponseV2 {

	var win engine.Fixed
	roundWin := gamestate.CumulativeWin
	var stake engine.Fixed

	for _, tx := range gamestate.Transactions {
		switch tx.Type {
		case "WAGER":
			stake += tx.Amount.Amount
		case "PAYOUT":
			win += tx.Amount.Amount
		}
	}
	var fsRemaining *int
	fsr := 0
	for i:=0;i<len(gamestate.NextActions); i++ {
		if strings.Contains(gamestate.NextActions[i], "freespin") {
			fsr ++
		}
	}
	fsRemaining = &fsr

	// artificially set cumulative win not to include spin win during cascade action unless it is the final round
	if gamestate.NextActions[0] == "cascade" {
		roundWin = roundWin.Sub(gamestate.SpinWin)
	}
	var cascadePositions []int
	if gamestate.Action == "cascade" {
		// this is a hack for now, needed for recovery. potentially in the future we add a proper cascade positions field to the gamestate message
		cascadePositions = gamestate.SelectedWinLines
	}
	logger.Infof("cum: %v, spin: %v, round: %v, win: %v", gamestate.CumulativeWin, gamestate.SpinWin, roundWin, win)

	return GameplayResponseV2{
		SessionID:   balance.Token,
		StateID:     gamestate.Id,
		RoundID:     gamestate.RoundID,
		ReelsetID:   gamestate.DefID,
		Stake:       stake,
		Win:         win,
		RoundWin:    roundWin,
		SpinWin:     gamestate.SpinWin,
		NextAction:  gamestate.NextActions[0],
		FSRemaining: fsRemaining,
		Balance:     BalanceResponseV2{
			Amount:    balance.Balance,
			FreeGames: balance.FreeGames.NoOfFreeSpins,
		},
		View:        gamestate.SymbolGrid,
		Prizes:      gamestate.Prizes,
		RoundMultiplier: gamestate.Multiplier,
		Closed:      gamestate.Closed,
		Gamification: gamestate.Gamification,
		CascadePositions: cascadePositions,
	}
}

func fillGameInitPreviousGameplay(previousGamestate engine.Gamestate, balance store.BalanceStore) GameInitResponseV2 {
	var resp GameInitResponseV2
	logger.Debugf("previousGamestate: %v; balance: %v; gameId: %v; auth: %v", previousGamestate, balance)

	lastRound := make(map[string]GameplayResponseV2, 2)
	lastRound[previousGamestate.Action] = fillGamestateResponseV2(previousGamestate, balance)
	// if last round was not base round, get triggering round ( for now no api support for this, so show default round)
	if ! strings.Contains(previousGamestate.Action, "base") {
		baseround := store.CreateInitGS(store.PlayerStore{PlayerId: balance.PlayerId, Balance:balance.Balance}, previousGamestate.Game)
		lastRound["base"] = fillGamestateResponseV2(baseround, balance)
	}
	resp.LastRound = lastRound
	resp.Name = previousGamestate.Game
	resp.Version = "2.0" // this is hardcoded for now
	return resp
}

func (initResp *GameInitResponseV2) FillEngineInfo(config engine.EngineConfig) {
	// todo: this doesn't handle when there are multiple reel sets for a single def (i.e. multiple defs with same ID)
	reelResp := make(map[string]ReelResponse, len(config.EngineDefs))
	for i, def := range config.EngineDefs {
		var reels ReelResponse

		reels.ID = make([][]int, len(def.ViewSize))
		reels.Count = make([][]int, len(def.ViewSize))
		reels.MaxVisible = make([][]int, len(def.ViewSize))
		reels.MaxStack = make([][]int, len(def.ViewSize))
		reels.BetMult = def.StakeDivisor

		for reel, reelContent := range def.Reels {

			//logger.Debugf("processing reel %v: %v", reel, reelContent)
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
		}
		//reels.Type = def.ID
		label := def.ID
		if _, ok := reelResp[label]; ok {
			label += fmt.Sprintf("%v", i)
		}
		reelResp[label] = reels
	}
	initResp.ReelSets = reelResp
	return
}
