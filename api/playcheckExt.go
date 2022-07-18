package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type PlaycheckExtBaseReq struct {
	Id        string  `json:"id"`
	Start     string  `json:"start"`
	End       string  `json:"end"`
	BetAmount float64 `json:"betamount"`
	WinAmount float64 `json:"winamount"`
	Currency  string  `json:"currency"`
}

type PlaycheckExtRequest struct {
	PlaycheckExtBaseReq
	Rounds [][][]int `json:"rounds"`
}

type PlaycheckExtRouletteRequest struct {
	PlaycheckExtBaseReq
	Symbol int               `json:"symbol"`
	Bets   map[string]string `json:"bets"`
	Prizes map[string]string `json:"prizes"`
}

type PlaycheckExtResponse struct {
	Url string `json:"url"`
}

func (gb PlaycheckExtResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func playcheckExt(r *http.Request, w http.ResponseWriter, params PlayCheckExtParams) (PlaycheckExtResponse, error) {
	if len(params.Feeds) == 0 {
		return PlaycheckExtResponse{}, fmt.Errorf("empty feeds")
	}

	var txdata store.RestTransactiondata = params.Feeds[0]

	gsbytes, err := base64.StdEncoding.DecodeString(txdata.Metadata.Vendor.State)
	if err != nil {
		logger.Errorf("Error base64 decoding gamestate: %v", err)
		return PlaycheckExtResponse{}, nil
	}

	istate, rgserr := DeserializeV3Gamestate(gsbytes)
	if rgserr == nil {
		switch istate.Base().Game {
		case "dragon-roulette":
			return playcheckExtRoulette(r, w, params, istate)
		default:
			err = fmt.Errorf("Can not produce playcheckExt for unknown V3 game \"%s\"", istate.Base().Game)
		}
		return PlaycheckExtResponse{}, err
	}

	state := store.DeserializeGamestateFromBytes(gsbytes)
	logger.Debugf("gamestate: %#v", state)

	var tx store.FeedTransaction = store.FeedTransaction{
		Id:           txdata.Id,
		Category:     txdata.Category,
		ExternalRef:  txdata.ExternalRef,
		CurrencyUnit: txdata.CurrencyUnit,
		Amount:       txdata.Amount,
		Metadata: store.FeedRoundMetadata{
			RoundId:   txdata.Metadata.RoundId,
			ExtItemId: txdata.Metadata.ExtItemId,
			ItemId:    txdata.Metadata.ItemId,
			Vendor: store.FeedRoundVendordata{
				State: state,
			},
		},
	}

	bet := 0.0
	win := 0.0

	for _, wt := range tx.Metadata.Vendor.State.Transactions {
		amount := wt.Amount.Amount.ValueAsFloat64()
		switch wt.Type {
		case "WAGER":
			bet += amount
		case "PAYOUT":
			win += amount
		}
	}

	rounds := [][][]int{}
	if len(tx.Metadata.Vendor.State.FeatureView) > 0 {
		rounds = append(rounds, tx.Metadata.Vendor.State.FeatureView)
	} else {
		rounds = append(rounds, tx.Metadata.Vendor.State.SymbolGrid)
	}

	gameId := tx.Metadata.ExtItemId

	req := PlaycheckExtRequest{
		PlaycheckExtBaseReq: PlaycheckExtBaseReq{
			Id:        tx.Metadata.RoundId,
			Start:     tx.TxTime,
			End:       tx.TxTime,
			BetAmount: bet,
			WinAmount: win,
			Currency:  tx.CurrencyUnit,
		},
		Rounds: rounds,
	}

	js, err := json.Marshal(req)
	if err != nil {
		return PlaycheckExtResponse{}, err
	}
	data := base64.StdEncoding.EncodeToString(js)

	url := fmt.Sprintf(config.GlobalConfig.ExtPlaycheck+"?game=%s&d=%s", gameId, data)

	return PlaycheckExtResponse{
		Url: url,
	}, nil
}

func playcheckExtRoulette(r *http.Request, w http.ResponseWriter, params PlayCheckExtParams, istate engine.IGameStateV3) (PlaycheckExtResponse, error) {
	if len(params.Feeds) == 0 {
		return PlaycheckExtResponse{}, fmt.Errorf("empty feeds")
	}

	var txdata store.RestTransactiondata = params.Feeds[0]

	var state *engine.GameStateRoulette = istate.(*engine.GameStateRoulette)
	logger.Debugf("gamestate: %#v", state)

	var tx store.FeedTransaction = store.FeedTransaction{
		Id:           txdata.Id,
		Category:     txdata.Category,
		ExternalRef:  txdata.ExternalRef,
		CurrencyUnit: txdata.CurrencyUnit,
		Amount:       txdata.Amount,
		Metadata: store.FeedRoundMetadata{
			RoundId:   txdata.Metadata.RoundId,
			ExtItemId: txdata.Metadata.ExtItemId,
			ItemId:    txdata.Metadata.ItemId,
			Vendor: store.FeedRoundVendordata{
				StateV3: state,
			},
		},
	}

	bet := 0.0
	win := 0.0

	for _, wt := range tx.Metadata.Vendor.State.Transactions {
		amount := wt.Amount.Amount.ValueAsFloat64()
		switch wt.Type {
		case "WAGER":
			bet += amount
		case "PAYOUT":
			win += amount
		}
	}

	bets := make(map[string]string)
	for k, v := range state.Bets {
		bets[k] = v.Amount.ValueAsString()
	}
	prizes := make(map[string]string)
	for _, p := range state.Prizes {
		prizes[p.Index] = p.Amount.ValueAsString()
	}

	gameId := tx.Metadata.ExtItemId

	req := PlaycheckExtRouletteRequest{
		PlaycheckExtBaseReq: PlaycheckExtBaseReq{
			Id:        tx.Metadata.RoundId,
			Start:     tx.TxTime,
			End:       tx.TxTime,
			BetAmount: bet,
			WinAmount: win,
			Currency:  tx.CurrencyUnit,
		},
		Symbol: state.Symbol,
		Bets:   bets,
		Prizes: prizes,
	}

	js, err := json.Marshal(req)
	if err != nil {
		return PlaycheckExtResponse{}, err
	}
	data := base64.StdEncoding.EncodeToString(js)

	url := fmt.Sprintf(config.GlobalConfig.ExtPlaycheck+"?game=%s&d=%s", gameId, data)
	logger.Debugf("%s", url)

	return PlaycheckExtResponse{
		Url: url,
	}, nil
}
