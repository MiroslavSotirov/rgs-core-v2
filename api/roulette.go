package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type initParamsRoulette struct {
	initParamsV3
	//	Bets map[string]BetRoulette `json:"bets"`
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

	Bets map[string]BetRoulette `json:"bets"`
}

func (i *playParamsRoulette) decode(request *http.Request) rgse.RGSErr {
	return decodeParams(i, request)
}

func (i playParamsRoulette) validate() rgse.RGSErr {
	return nil
}

func (i *playParamsRoulette) deserialize(b []byte) rgse.RGSErr {
	return deserializeParams(i, b)
}

type BetRoulette struct {
	//	Index   string       `json:"index"`
	Amount  engine.Fixed `json:"amount"`
	Symbols []int        `json:"symbols"`
}
type PrizeRoulette struct {
	Index  string       `json:"index"`
	Amount engine.Fixed `json:"amount"`
}

/*
type PrizeRoulette struct {
	Index   string       `json:"index"`
	Amount  engine.Fixed `json:"amount"`
	Symbols []int32      `json:"symbols"`
}
*/

type GameInitResponseRoulette struct {
	GameInitResponseV3
	LastRound IGamePlayResponseV3    `json:"lastRound"`
	Reel      []int                  `json:"reel"`
	Bets      map[string]BetRoulette `json:"bets"`
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
	Prizes   []PrizeRoulette
}

func (s GameStateRoulette) GetTtl() int64 {
	return 3600
}

func initRoulette(player store.PlayerStore, engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token, state []byte) (
	response IGameInitResponseV3, rgserr rgse.RGSErr) {

	var data initParamsRoulette
	if rgserr = data.deserialize(body); rgserr != nil {
		return nil, rgse.Create(rgse.JsonError)
	}

	var gameState GameStateRoulette
	if len(state) == 0 {
		gameState = initRouletteGS(data)
		gameState.Id = string(token) + data.Game + "GSinit"
	} else {
		err := json.Unmarshal(state, &gameState)
		if err != nil {
			rgserr = rgse.Create(rgse.StoreInitError)
		}
	}

	balance := store.BalanceStore{
		Balance:   player.Balance,
		Token:     token,
		FreeGames: player.FreeGames,
	}

	zero := engine.NewFixedFromInt(0)
	playResponse := fillRoulettePlayResponse(gameState, balance, zero, zero)

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
	var data playParamsRoulette
	rgserr = data.deserialize(body)

	fmt.Printf("playRoulette data= %#v\n", data)

	valid, stake := validateRouletteBets(data.Bets)
	if !valid || len(data.Bets) == 0 {
		fmt.Printf("not valid: valid=%v len bets=%d\n", valid, len(data.Bets))
		return nil, rgse.Create(rgse.InvalidStakeError)
	}

	var prevState GameStateRoulette

	if len(txStore.GameState) == 0 {
		logger.Debugf("no previous gamestate in playRoulette")
		initParams := initParamsRoulette{
			initParamsV3: initParamsV3{
				Game: data.Game,
			},
		}
		prevState = initRouletteGS(initParams)
	} else {
		err := json.Unmarshal(txStore.GameState, &prevState)
		if err != nil {
			return nil, rgse.Create(rgse.JsonError)
		}
	}
	logger.Debugf("prevState= %#v", prevState)

	return getRouletteResults(data, stake, prevState, txStore)
}

func initRouletteGS(data initParamsRoulette) GameStateRoulette {
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Game: data.Game,
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []PrizeRoulette{},
	}
	return gameState
}

func validateRouletteBets(bets map[string]BetRoulette) (bool, engine.Fixed) {
	sum := engine.NewFixedFromInt(0)
	for k, v := range bets {
		logger.Debugf("validating bet %#v", v)
		if !validateRouletteBet(k, v) {
			return false, engine.NewFixedFromInt(0)
		}
		sum += v.Amount
		logger.Debugf("valid roulette bet %s amount %.4f", k, v.Amount.ValueAsFloat())
	}
	return true, sum
}

func validateRouletteBet(index string, bet BetRoulette) bool {
	for k, v := range rouletteBets {
		if k == index {
			if len(v.Symbols) != len(bet.Symbols) {
				logger.Debugf("roulette bet with index %s has the wrong number of symbols (expected %d got %d)",
					index, len(v.Symbols), len(bet.Symbols))
				return false
			}
			for j, s := range v.Symbols {
				if bet.Symbols[j] != s {
					logger.Debugf("roulette bet with index %s and symbols %v does not match symbol index %d in valid bet symbols %v",
						index, bet.Symbols, j, v.Symbols)
					return false
				}
			}
			return true
		}
	}
	logger.Debugf("roulette bet with index %s is unknown", index)
	return false
}

func processRouletteBets(symbol int, bets map[string]BetRoulette) (engine.Fixed, []PrizeRoulette) {
	sum, prizes := engine.NewFixedFromInt(0), []PrizeRoulette{}
	for k, v := range bets {
		for _, s := range v.Symbols {
			if s == symbol {
				sum += v.Amount
				prizes = append(prizes, PrizeRoulette{Index: k, Amount: v.Amount})
				break
			}
		}
	}
	return sum, prizes
}

func getRouletteResults(
	data playParamsRoulette,
	bet engine.Fixed,
	prevState GameStateRoulette,
	txStore store.TransactionStore) (response GamePlayResponseRoulette, err rgse.RGSErr) {

	position := rng.RandFromRange(len(rouletteReel))
	symbol := rouletteReel[position]

	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Id:                uuid.NewV4().String(), // prevState.NextGamestate,
			PreviousGamestate: data.PreviousID,
			//			NextGamestate:     string(token),
			Transactions: []engine.WalletTransaction{
				{
					Id:     prevState.NextGamestate,
					Amount: engine.Money{Amount: bet, Currency: "USD"},
					Type:   "WAGER",
				},
			},
		},
		Position: position,
		Symbol:   symbol,
	}

	win, prizes := processRouletteBets(symbol, data.Bets)
	gameState.Prizes = prizes
	if len(prizes) > 0 {
		gameState.Transactions = append(gameState.Transactions, engine.WalletTransaction{
			Id:     prevState.NextGamestate,
			Amount: engine.Money{Amount: win, Currency: "USD"},
			Type:   "PAYOUT",
		})
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
		//		gameState.Id = string(token)
	}

	response = fillRoulettePlayResponse(gameState, balance, bet, win)

	return
}

func fillRoulettePlayResponse(gameState GameStateRoulette, balance store.BalanceStore, bet engine.Fixed, win engine.Fixed) GamePlayResponseRoulette {

	/*
		prizes := []PrizeRoulette{}
		for _, p := range gameState.Prizes {
			prizes = append(prizes, *p)
		}

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
			Token:   balance.Token,
			StateId: gameState.Id,
			RoundId: gameState.Id,
			Balance: BalanceResponseV3{
				Amount: balance.Balance,
			},
			Bet: bet,
			Win: win,
		},
		Symbol:   gameState.Symbol,
		Position: gameState.Position,
		Prizes:   gameState.Prizes,
	}
}

var rouletteReel = []int{0, 32, 16, 13, 21, 6, 19, 2, 27, 17, 36, 4, 25, 15, 34, 11, 28, 8, 23, 12, 5, 22, 18, 31, 37, 20, 14, 33, 7, 24, 16, 29, 9, 30, 10, 35, 1, 26}

/*
var rouletteBets = []BetRoulette{
	{
		Index:   "1",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{1},
	},
	{
		Index:   "2",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{2},
	},
	{
		Index:   "3",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{3},
	},
	{
		Index:   "4",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{4},
	},
	{
		Index:   "5",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{5},
	},
	{
		Index:   "6",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{6},
	},
	{
		Index:   "7",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{7},
	},
	{
		Index:   "8",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{8},
	},
	{
		Index:   "9",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{9},
	},
	{
		Index:   "10",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{10},
	},
	{
		Index:   "11",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{11},
	},
	{
		Index:   "12",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{12},
	},
	{
		Index:   "13",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{13},
	},
	{
		Index:   "14",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{14},
	},
	{
		Index:   "15",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{15},
	},
	{
		Index:   "16",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{16},
	},
	{
		Index:   "17",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{17},
	},
	{
		Index:   "18",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{18},
	},
	{
		Index:   "19",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{19},
	},
	{
		Index:   "20",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{20},
	},
	{
		Index:   "21",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{21},
	},
	{
		Index:   "22",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{22},
	},
	{
		Index:   "23",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{23},
	},
	{
		Index:   "24",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{24},
	},
	{
		Index:   "25",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{25},
	},
	{
		Index:   "26",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{26},
	},
	{
		Index:   "27",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{27},
	},
	{
		Index:   "28",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{28},
	},
	{
		Index:   "29",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{29},
	},
	{
		Index:   "30",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{30},
	},
	{
		Index:   "31",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{31},
	},
	{
		Index:   "32",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{32},
	},
	{
		Index:   "33",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{33},
	},
	{
		Index:   "34",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{34},
	},
	{
		Index:   "35",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{35},
	},
	{
		Index:   "36",
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{36},
	},
	{
		Index:   "dragonbet1",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{1, 5, 9, 12, 14, 16, 19, 23, 27, 30, 32, 34},
	},
	{
		Index:   "dragonbet2",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{3, 6, 8, 10, 13, 17, 20, 22, 25, 29, 33, 36},
	},
	{
		Index:   "dragonbet3",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 7, 12, 15, 17, 20, 24, 27, 28, 31, 35},
	},
	{
		Index:   "dragonbet4",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{3, 5, 7, 10, 14, 18, 21, 23, 25, 28, 32, 36},
	},
	{
		Index:   "dragonbet5",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 9, 11, 13, 18, 21, 22, 26, 30, 31, 35},
	},
	{
		Index:   "dragonbet6",
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{1, 6, 8, 11, 15, 16, 19, 24, 26, 29, 33, 34},
	},
	{
		Index:   "red",
		Amount:  engine.NewFixedFromInt(2),
		Symbols: []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35},
	},
	{
		Index:   "black",
		Amount:  engine.NewFixedFromInt(2),
		Symbols: []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36},
	},
}
*/
var rouletteBets = map[string]BetRoulette{
	"1": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{1},
	},
	"2": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{2},
	},
	"3": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{3},
	},
	"4": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{4},
	},
	"5": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{5},
	},
	"6": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{6},
	},
	"7": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{7},
	},
	"8": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{8},
	},
	"9": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{9},
	},
	"10": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{10},
	},
	"11": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{11},
	},
	"12": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{12},
	},
	"13": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{13},
	},
	"14": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{14},
	},
	"15": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{15},
	},
	"16": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{16},
	},
	"17": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{17},
	},
	"18": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{18},
	},
	"19": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{19},
	},
	"20": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{20},
	},
	"21": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{21},
	},
	"22": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{22},
	},
	"23": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{23},
	},
	"24": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{24},
	},
	"25": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{25},
	},
	"26": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{26},
	},
	"27": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{27},
	},
	"28": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{28},
	},
	"29": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{29},
	},
	"30": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{30},
	},
	"31": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{31},
	},
	"32": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{32},
	},
	"33": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{33},
	},
	"34": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{34},
	},
	"35": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{35},
	},
	"36": {
		Amount:  engine.NewFixedFromInt(36),
		Symbols: []int{36},
	},
	"dragonbet1": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{1, 5, 9, 12, 14, 16, 19, 23, 27, 30, 32, 34},
	},
	"dragonbet2": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{3, 6, 8, 10, 13, 17, 20, 22, 25, 29, 33, 36},
	},
	"dragonbet3": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 7, 12, 15, 17, 20, 24, 27, 28, 31, 35},
	},
	"dragonbet4": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{3, 5, 7, 10, 14, 18, 21, 23, 25, 28, 32, 36},
	},
	"dragonbet5": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{2, 4, 9, 11, 13, 18, 21, 22, 26, 30, 31, 35},
	},
	"dragonbet6": {
		Amount:  engine.NewFixedFromInt(3),
		Symbols: []int{1, 6, 8, 11, 15, 16, 19, 24, 26, 29, 33, 34},
	},
	"red": {
		Amount:  engine.NewFixedFromInt(2),
		Symbols: []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35},
	},
	"black": {
		Amount:  engine.NewFixedFromInt(2),
		Symbols: []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36},
	},
}
