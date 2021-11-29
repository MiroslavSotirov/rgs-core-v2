package api

import (
	"fmt"
	"net/http"
	"time"

	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
)

type initParamsRoulette struct {
	initParamsV3
	//	Bets string `json:"bets"`
}

func (i *initParamsRoulette) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i initParamsRoulette) validate() rgse.RGSErr {
	return nil
}

func (i *initParamsRoulette) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

type playParamsRoulette struct {
	playParamsV3

	Bets []BetRoulette `json:"bets"`
}

type BetRoulette struct {
	Index   string       `json:"index"`
	Stake   engine.Fixed `json:"stake"`
	Symbols []int        `json:"symbols"`
}
type PrizeRoulette BetRoulette

/*
type PrizeRoulette struct {
	Index   string       `json:"index"`
	Amount  engine.Fixed `json:"amount"`
	Symbols []int32      `json:"symbols"`
}
*/

type GameInitResponseRoulette struct {
	GameInitResponseV3
	LastRound IGamePlayResponseV3 `json:"lastRound"`
	Reel      []int               `json:"reel"`
	Bets      []BetRoulette       `json:"bets"`
}

func (resp GameInitResponseRoulette) base() GameInitResponseV3 {
	return resp.GameInitResponseV3
}

func (resp GameInitResponseRoulette) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GamePlayResponseRoulette struct {
	GamePlayResponseV3

	Symbol   int             `json:"number"`
	Position int             `json:"position"`
	Prizes   []PrizeRoulette `json:"wins"`
}

func (resp GamePlayResponseRoulette) Base() GamePlayResponseV3 {
	return resp.GamePlayResponseV3
}

func (resp GamePlayResponseRoulette) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GameStateRoulette struct {
	GameStateV3

	Position int
	Symbol   int
	Prizes   []*PrizeRoulette
}

func (s GameStateRoulette) GetTtl() int64 {
	return 3600
}

func initRoulette(engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token) (response IGameInitResponseV3, rgserr rgse.RGSErr) {

	var data initParamsRoulette
	if rgserr = data.deserialize(body); rgserr != nil {
		return nil, rgse.Create(rgse.JsonError)
	}

	gameState := initRouletteGS(data)
	playerID := ""
	gameState.Id = playerID + data.Game + "GSinit"

	balance := store.BalanceStore{
		Token: token,
	}

	playResponse := fillRoulettePlayResponse(gameState, balance)

	response = GameInitResponseRoulette{
		GameInitResponseV3: GameInitResponseV3{
			Name:   gameState.Game,
			Wallet: wallet,
		},
		LastRound: playResponse,
		Reel:      rouletteReel,
		Bets:      rouletteBets,
	}
	return
}

func playRoulette(engineId string, wallet string, body []byte, txStore store.TransactionStore) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {
	fmt.Printf("playRoulette begin\n")

	var data playParamsRoulette
	rgserr = data.deserialize(body)

	var latestState GameStateRoulette
	/*
		if txStore == nil {
			initParams := initParamsRoulette{
				initParamsV3: initParamsV3{
					Game: data.Game,
				},
			}
			latestState = initRouletteGS(initParams)
		} else {
			err := json.Unmarshal(txStore.GameState, &latestState)
			if err != nil {
				return nil, rgse.Create(rgse.JsonError)
			}
		}
	*/

	return getRouletteResults(data, latestState, txStore)
}

func initRouletteGS(data initParamsRoulette) GameStateRoulette {
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Game: data.Game,
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []*PrizeRoulette{},
	}
	return gameState
}

func validateRouletteBet(bet BetRoulette) bool {
	for _, b := range rouletteBets {
		if b.Index == bet.Index {
			if len(b.Symbols) != len(bet.Symbols) {
				return false
			}
			for j, s := range b.Symbols {
				if bet.Symbols[j] != s {
					return false
				}
			}
			return true
		}
	}
	return false
}

func getRouletteResults(data playParamsRoulette, latestState GameStateRoulette, txStore store.TransactionStore) (response GamePlayResponseRoulette, err rgse.RGSErr) {

	position := rng.RandFromRange(len(rouletteReel))
	symbol := rouletteReel[position]

	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			PreviousGamestate: data.PreviousID,
			//			NextGamestate:     string(token),
		},
		Position: position,
		Symbol:   symbol,
	}

	var balance store.BalanceStore
	var freeGameRef string = "" // check apiV2 for how to determine if this is a free game, and set
	/*
		= store.BalanceStore{
			PlayerId: txStore.PlayerId,
			Token:    txStore.Token,
		}
	*/
	stateBytes := gameState.Serialize()
	token := txStore.Token
	for _, t := range gameState.Transactions {
		tx := store.TransactionStore{
			TransactionId:       t.Id,
			Token:               token,
			Category:            store.Category(t.Type),
			RoundStatus:         store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              data.Game,
			RoundId:             gameState.RoundId,
			Amount:              t.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           stateBytes,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:                 gameState.GetTtl(),
		}
		balance, err = TransactionByWallet(token, data.Wallet, tx)
		if err != nil {
			return
		}
		token = balance.Token
	}

	response = fillRoulettePlayResponse(gameState, balance)

	return
}

func fillRoulettePlayResponse(gameState GameStateRoulette, balance store.BalanceStore) GamePlayResponseRoulette {

	prizes := []PrizeRoulette{}
	for _, p := range gameState.Prizes {
		prizes = append(prizes, *p)
	}
	/*

		rouletteResponse := GamePlayResponseRoulette{
			GamePlayResponseV3: GamePlayResponseV3{
				SessionID: balance.Token,
				StateID:   gameState.Id,
			},
			Prizes: prizes,
		}
		return &playResponse
	*/

	return GamePlayResponseRoulette{
		GamePlayResponseV3: GamePlayResponseV3{
			SessionID: balance.Token,
			StateID:   gameState.Id,
		},
		Symbol:   gameState.Symbol,
		Position: gameState.Position,
		Prizes:   prizes,
	}
}

var rouletteReel = []int{0, 32, 16, 13, 21, 6, 19, 2, 27, 17, 36, 4, 25, 15, 34, 11, 28, 8, 23, 12, 5, 22, 18, 31, 00, 20, 14, 33, 7, 24, 16, 29, 9, 30, 10, 35, 1, 26}
var rouletteBets = []BetRoulette{
	{
		Index:   "1",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{1},
	},
	{
		Index:   "2",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{2},
	},
	{
		Index:   "3",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{3},
	},
	{
		Index:   "4",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{4},
	},
	{
		Index:   "5",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{5},
	},
	{
		Index:   "6",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{6},
	},
	{
		Index:   "7",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{7},
	},
	{
		Index:   "8",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{8},
	},
	{
		Index:   "9",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{9},
	},
	{
		Index:   "10",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{10},
	},
	{
		Index:   "11",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{11},
	},
	{
		Index:   "12",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{12},
	},
	{
		Index:   "13",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{13},
	},
	{
		Index:   "14",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{14},
	},
	{
		Index:   "15",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{15},
	},
	{
		Index:   "16",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{16},
	},
	{
		Index:   "17",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{17},
	},
	{
		Index:   "18",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{18},
	},
	{
		Index:   "19",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{19},
	},
	{
		Index:   "20",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{20},
	},
	{
		Index:   "21",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{21},
	},
	{
		Index:   "22",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{22},
	},
	{
		Index:   "23",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{23},
	},
	{
		Index:   "24",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{24},
	},
	{
		Index:   "25",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{25},
	},
	{
		Index:   "26",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{26},
	},
	{
		Index:   "27",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{27},
	},
	{
		Index:   "28",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{28},
	},
	{
		Index:   "29",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{29},
	},
	{
		Index:   "30",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{30},
	},
	{
		Index:   "31",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{31},
	},
	{
		Index:   "32",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{32},
	},
	{
		Index:   "33",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{33},
	},
	{
		Index:   "34",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{34},
	},
	{
		Index:   "35",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{35},
	},
	{
		Index:   "36",
		Stake:   engine.NewFixedFromInt(36),
		Symbols: []int{36},
	},
	{
		Index:   "dragonbet1",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{1, 5, 9, 12, 14, 16, 19, 23, 27, 30, 32, 34},
	},
	{
		Index:   "dragonbet2",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{3, 6, 8, 10, 13, 17, 20, 22, 25, 29, 33, 36},
	},
	{
		Index:   "dragonbet3",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 7, 12, 15, 17, 20, 24, 27, 28, 31, 35},
	},
	{
		Index:   "dragonbet4",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{3, 5, 7, 10, 14, 18, 21, 23, 25, 28, 32, 36},
	},
	{
		Index:   "dragonbet5",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 9, 11, 13, 18, 21, 22, 26, 30, 31, 35},
	},
	{
		Index:   "dragonbet6",
		Stake:   engine.NewFixedFromInt(3),
		Symbols: []int{1, 6, 8, 11, 15, 16, 19, 24, 26, 29, 33, 34},
	},
	{
		Index:   "red",
		Stake:   engine.NewFixedFromInt(2),
		Symbols: []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35},
	},
	{
		Index:   "black",
		Stake:   engine.NewFixedFromInt(2),
		Symbols: []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36},
	},
}
