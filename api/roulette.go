package api

import (
	"net/http"
	"time"

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

	Bets map[string]engine.BetRoulette `json:"bets"`
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

	Symbol   int                           `json:"number"`
	Position int                           `json:"position"`
	Bets     map[string]engine.BetRoulette `json:"bets"`
	Prizes   []engine.PrizeRoulette        `json:"wins"`
}

func (resp GamePlayResponseRoulette) Base() GamePlayResponseV3 {
	return resp.GamePlayResponseV3
}

func (resp GamePlayResponseRoulette) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func initRoulette(player store.PlayerStore, engineId string, wallet string, body []byte, engineConf engine.EngineConfig, token store.Token, state []byte) (
	response IGameInitResponseV3, rgserr rgse.RGSErr) {

	var data initParamsRoulette
	if rgserr = data.deserialize(body); rgserr != nil {
		return nil, rgse.Create(rgse.JsonError)
	}

	engineDef := engineConf.EngineDefs[0]

	var game store.GameRouletteV3
	var gameState engine.GameStateRoulette
	if len(state) == 0 {
		gameState = store.InitStateRoulette(data.Game, data.Ccy)
		gameState.Id = string(token) + data.Game + "GSinit"
	} else {
		gameState, rgserr = game.DeserializeStateRoulette(state)
		logger.Debugf("initRoulette state length:%s\ndeserialized:%#v", len(state), gameState)
	}

	balance := store.BalanceStore{
		Balance:   player.Balance,
		Token:     player.Token,
		FreeGames: player.FreeGames,
	}

	playResponse := fillRoulettePlayResponse(gameState, balance)

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

	logger.Debugf("playRoulette %#v\n", data)

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

	var game store.GameRouletteV3
	var prevState engine.GameStateRoulette

	if len(txStore.GameState) == 0 {
		logger.Debugf("no previous gamestate in playRoulette")
		initParams := initParamsRoulette{
			initParamsV3: initParamsV3{
				Game: data.Game,
			},
		}
		prevState = store.InitStateRoulette(initParams.Game, txStore.Amount.Currency)
	} else {
		prevState, rgserr = game.DeserializeStateRoulette(txStore.GameState)
		if rgserr != nil {
			return
		}
	}
	logger.Debugf("playRoulette prevState len:%d\ndeserialized:%#v", len(txStore.GameState), prevState)

	var engineDef engine.EngineDef = engineConf.EngineDefs[0]

	return getRouletteResults(data, engineDef, stake, prevState, txStore)
}

func validateRouletteBets(bets map[string]engine.BetRoulette, validBets map[string]engine.RoulettePayout, validStakes []engine.Fixed) (bool, engine.Fixed) {
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

func validateRouletteBet(index string, bet engine.BetRoulette, payouts map[string]engine.RoulettePayout) bool {
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
	logger.Debugf("roulette bet with index %s is unknown", index)
	return false
}

func validateRouletteStake(bet engine.BetRoulette, stakes []engine.Fixed) bool {
	logger.Debugf("validating stake in bet %#v", bet)
	if bet.Amount <= 0 {
		logger.Warnf("bet is zero or negative")
		return false
	}
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

func processRouletteBets(symbol int, bets map[string]engine.BetRoulette, payouts map[string]engine.RoulettePayout) (engine.Fixed, []engine.PrizeRoulette) {
	sum, prizes := engine.NewFixedFromInt(0), []engine.PrizeRoulette{}
	for k, v := range bets {
		for _, s := range v.Symbols {
			if s == symbol {
				win := v.Amount.Mul(engine.NewFixedFromInt(payouts[k].Multiplier))
				sum += win
				prizes = append(prizes, engine.PrizeRoulette{Index: k, Amount: win})
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
	prevState engine.GameStateRoulette,
	txStore store.TransactionStore) (response GamePlayResponseRoulette, err rgse.RGSErr) {

	var game store.GameRouletteV3
	gameState := rouletteRound(data, engineDef, prevState)

	logger.Debugf("getRouletteResults gameState.Id=%s prevState.NextGamestate.Id=%s gameState.RoundId=%s",
		gameState.Id, prevState.NextGamestate, gameState.RoundId)

	bets := []engine.WalletTransaction{
		{
			Id:     prevState.NextGamestate,
			Amount: engine.Money{Amount: bet, Currency: txStore.Amount.Currency},
			Type:   "WAGER",
		}}
	gameState.GameStateV3.Transactions = append(bets, gameState.GameStateV3.Transactions...)

	win, prizes := processRouletteBets(gameState.Symbol, data.Bets, engineDef.RoulettePayouts)
	gameState.Prizes = prizes
	if len(prizes) > 0 {
		gameState.GameStateV3.Transactions = append(gameState.GameStateV3.Transactions, engine.WalletTransaction{
			Id:     rng.Uuid(), // prevState.NextGamestate,
			Amount: engine.Money{Amount: win, Currency: txStore.Amount.Currency},
			Type:   "PAYOUT",
		})
	}
	gameState.Bet = bet
	gameState.Win = win

	var balance store.BalanceStore
	var freeGameRef string = "" // check apiV2 for how to determine if this is a free game, and set
	/*
		= store.BalanceStore{
			PlayerId: txStore.PlayerId,
			Token:    txStore.Token,
		}
	*/
	logger.Debugf("processing state: %#v", gameState)
	stateBytes := game.SerializeState(&gameState) // gameState.Serialize()
	token := txStore.Token
	for _, transaction := range gameState.Transactions {
		logger.Debugf("performing transaction %#v", transaction)
		AppendHistory(&txStore, transaction)
		tx := store.TransactionStore{
			TransactionId:       transaction.Id,
			Token:               token,
			Category:            store.Category(transaction.Type),
			RoundStatus:         store.RoundStatusOpen,
			PlayerId:            txStore.PlayerId,
			GameId:              data.Game,
			RoundId:             gameState.RoundId,
			Amount:              transaction.Amount,
			ParentTransactionId: "",
			TxTime:              time.Now(),
			GameState:           stateBytes,
			BetLimitSettingCode: txStore.BetLimitSettingCode,
			FreeGames:           store.FreeGamesStore{NoOfFreeSpins: 0, CampaignRef: freeGameRef},
			Ttl:                 gameState.GetTtl(),
			History:             txStore.History,
		}
		balance, err = TransactionByWallet(token, data.Wallet, tx)
		if err != nil {
			return
		}
		token = balance.Token
		//		gameState.Id = string(token)
	}

	response = fillRoulettePlayResponse(gameState, balance)

	return
}

func rouletteRound(data playParamsRoulette, engineDef engine.EngineDef, prevState engine.GameStateRoulette) engine.GameStateRoulette {
	reel := engineDef.Reels[0]
	position := rng.RandFromRange(len(reel))
	symbol := reel[position]

	id := prevState.NextGamestate
	roundId := id
	nextid := rng.Uuid()
	gameState := engine.GameStateRoulette{
		GameStateV3: engine.GameStateV3{
			Id:                prevState.NextGamestate,
			Game:              data.Game,
			Version:           "3",
			Currency:          prevState.Currency,
			RoundId:           roundId,
			PreviousGamestate: data.PreviousID,
			NextGamestate:     nextid,
		},
		Position: position,
		Symbol:   symbol,
		Bets:     data.Bets,
	}
	return gameState
}

func fillRoulettePlayResponse(gameState engine.GameStateRoulette, balance store.BalanceStore) GamePlayResponseRoulette {
	return GamePlayResponseRoulette{
		GamePlayResponseV3: GamePlayResponseV3{
			Token:   balance.Token,
			StateId: gameState.Id,
			RoundId: gameState.RoundId,
			Balance: BalanceResponseV3{
				Amount: balance.Balance,
			},
			Bet: gameState.Bet,
			Win: gameState.Win,
		},
		Symbol:   gameState.Symbol,
		Position: gameState.Position,
		Bets:     gameState.Bets,
		Prizes:   gameState.Prizes,
	}
}
