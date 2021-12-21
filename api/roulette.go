package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
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
type PayoutRoulette struct {
	Multiplier int   `json:"multiplier"`
	Symbols    []int `json:"symbols"`
}

func MakePayoutRoulette(p engine.RoulettePayout) PayoutRoulette {
	return PayoutRoulette{
		Multiplier: p.Multiplier,
		Symbols:    p.Symbols,
	}
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
	LastRound IGamePlayResponseV3       `json:"lastRound"`
	Reel      []int                     `json:"reel"`
	MinBet    engine.Fixed              `json:"minBet"`
	MaxBet    engine.Fixed              `json:"maxBet"`
	Bets      map[string]PayoutRoulette `json:"bets"`
}

func (resp *GameInitResponseRoulette) Base() *GameInitResponseV3 {
	return &resp.GameInitResponseV3
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

	engineDef := engineConf.EngineDefs[0]

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

	stakeValues, defaultBet, minBet, maxBet, prmerr := parameterSelector.GetGameplayParameters(engine.Money{Currency: data.Ccy}, player.BetLimitSettingCode, data.Game)
	if prmerr != nil {
		rgserr = prmerr
		return
	}

	bets := make(map[string]PayoutRoulette, len(engineDef.RoulettePayouts))
	for k, v := range engineDef.RoulettePayouts {
		bets[k] = MakePayoutRoulette(v)
	}
	response = &GameInitResponseRoulette{
		GameInitResponseV3: GameInitResponseV3{
			Name:        gameState.Game,
			Wallet:      wallet,
			StakeValues: stakeValues,
			DefaultBet:  defaultBet,
		},
		LastRound: playResponse,
		Reel:      engineDef.Reels[0],
		Bets:      bets,
		MinBet:    minBet,
		MaxBet:    maxBet,
	}
	return
}

func playRoulette(engineId string, wallet string, body []byte, txStore store.TransactionStore) (response IGamePlayResponseV3, rgserr rgse.RGSErr) {
	var data playParamsRoulette
	rgserr = data.deserialize(body)

	fmt.Printf("playRoulette data= %#v\n", data)

	engineConf := engine.BuildEngineDefs(engineId)

	stakeValues, _, minBet, maxBet, prmerr := parameterSelector.GetGameplayParameters(engine.Money{0, txStore.Amount.Currency}, txStore.BetLimitSettingCode, data.Game)
	if prmerr != nil {
		return nil, prmerr
	}

	valid, stake := validateRouletteBets(data.Bets, engineConf.EngineDefs[0].RoulettePayouts, stakeValues)
	if minBet > 0 && stake < minBet {
		logger.Debugf("Total roulette bet %s is lower than limit %s", stake.ValueAsString(), minBet.ValueAsString())
		valid = false
	}
	if maxBet > 0 && stake > maxBet {
		logger.Debugf("Total roulette bet %s is higher than limit %s", stake.ValueAsString(), maxBet.ValueAsString())
		valid = false
	}
	if !valid || len(data.Bets) == 0 {
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

	var engineDef engine.EngineDef = engineConf.EngineDefs[0]

	return getRouletteResults(data, engineDef, stake, prevState, txStore)
}

func initRouletteGS(data initParamsRoulette) GameStateRoulette {
	id := uuid.NewV4().String()
	nextid := uuid.NewV4().String()
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Id:            id,
			NextGamestate: nextid,
			Game:          data.Game,
		},
		Position: 0,
		Symbol:   0,
		Prizes:   []PrizeRoulette{},
	}
	return gameState
}

func validateRouletteBets(bets map[string]BetRoulette, validBets map[string]engine.RoulettePayout, validStakes []engine.Fixed) (bool, engine.Fixed) {
	sum := engine.NewFixedFromInt(0)
	for k, v := range bets {
		logger.Debugf("validating bet %#v", v)
		if !validateRouletteBet(k, v, validBets) || !validateRouletteStake(v, validStakes) {
			return false, engine.NewFixedFromInt(0)
		}
		sum += v.Amount
		logger.Debugf("valid roulette bet %s amount %.4f", k, v.Amount.ValueAsFloat())
	}
	return true, sum
}

func validateRouletteBet(index string, bet BetRoulette, payouts map[string]engine.RoulettePayout) bool {
	v, ok := payouts[index]
	if ok {
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
	/*
		for k, v := range validBets { // rouletteBets {
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
	*/
	logger.Debugf("roulette bet with index %s is unknown", index)
	return false
}

func validateRouletteStake(bet BetRoulette, stakes []engine.Fixed) bool {
	logger.Debugf("validating stake in bet %#v", bet)
	amount := bet.Amount
	lastAmount := engine.Fixed(0)
	numstakes := len(stakes)
	for amount > 0 {
		if amount == lastAmount {
			logger.Warnf("bet %#v has a remainder of %s after validating against stakes %v", bet, amount.ValueAsString(), stakes)
			return false
		}
		lastAmount = amount
		for numstakes > 0 {
			numstakes--
			stake := stakes[numstakes]
			num := 0
			for amount >= stake {
				amount -= stake
				num++
			}
		}
	}
	return true
}

func processRouletteBets(symbol int, bets map[string]BetRoulette, payouts map[string]engine.RoulettePayout) (engine.Fixed, []PrizeRoulette) {
	sum, prizes := engine.NewFixedFromInt(0), []PrizeRoulette{}
	for k, v := range bets {
		for _, s := range v.Symbols {
			if s == symbol {
				win := v.Amount.Mul(engine.NewFixedFromInt(payouts[k].Multiplier))
				sum += win
				prizes = append(prizes, PrizeRoulette{Index: k, Amount: win})
				break
			}
		}
	}
	return sum, prizes
}

func getRouletteResults(
	data playParamsRoulette,
	engineDef engine.EngineDef,
	bet engine.Fixed,
	prevState GameStateRoulette,
	txStore store.TransactionStore) (response GamePlayResponseRoulette, err rgse.RGSErr) {

	gameState := rouletteRound(data, engineDef, prevState)

	bets := []engine.WalletTransaction{
		{
			Id:     prevState.NextGamestate,
			Amount: engine.Money{Amount: bet, Currency: "USD"},
			Type:   "WAGER",
		}}
	gameState.GameStateV3.Transactions = append(bets, gameState.GameStateV3.Transactions...)

	win, prizes := processRouletteBets(gameState.Symbol, data.Bets, engineDef.RoulettePayouts)
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
	logger.Debugf("processing state: %#v", gameState)
	stateBytes := gameState.Serialize()
	token := txStore.Token
	for _, t := range gameState.Transactions {
		logger.Debugf("performing transaction %#v", t)
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

func rouletteRound(data playParamsRoulette, engineDef engine.EngineDef, prevState GameStateRoulette) GameStateRoulette {
	reel := engineDef.Reels[0]
	position := rng.RandFromRange(len(reel))
	symbol := reel[position]

	id := prevState.NextGamestate
	roundId := id
	nextid := uuid.NewV4().String()
	gameState := GameStateRoulette{
		GameStateV3: GameStateV3{
			Id:                prevState.NextGamestate,
			Game:              data.Game,
			RoundId:           roundId,
			PreviousGamestate: data.PreviousID,
			NextGamestate:     nextid,
		},
		Position: position,
		Symbol:   symbol,
	}
	return gameState
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
			RoundId: gameState.RoundId,
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
