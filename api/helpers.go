package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

// GetURLScheme returns https if TLS is present else returns http
func GetURLScheme(r *http.Request) string {
	//return r.URL.Scheme+"//"
	//todo: find another way to handle this because TLS is missing sometimes from request object when it shouldn't be
	if config.GlobalConfig.Local == true {
		return "http://"
	}
	return "https://"
}

func processAuthorization(request *http.Request) (string, rgse.RGSErr) {

	tokenInfo := strings.Split(request.Header.Get("Authorization"), " ")
	switch tokenInfo[0] {
	default:
		return "", rgse.Create(rgse.InvalidCredentials)
	case "MAVERICK-Host-Token":
		logger.Debugf("Auth Token: %v; Auth Header: %v", tokenInfo[1], request.Header.Get("Authorization"))
	case "DUMMY-MAVERICK-Host-Token":
		logger.Debugf("Auth Token: %v; Auth Header: %v", tokenInfo[1], request.Header.Get("Authorization"))
	}
	if strings.Contains(tokenInfo[1], "token=\"") {
		return strings.Split(tokenInfo[1], "\"")[1], nil
	}

	return tokenInfo[1], nil
}

func parseMemID(token string) string {
	parts := strings.Split(token, "=")
	if len(parts) < 2 {
		return token
	}
	return strings.Trim(parts[1], "'")
}

func PlayerBalance(r *http.Request) (BalanceCheckResponse, rgse.RGSErr) {
	authToken, err := processAuthorization(r)
	if err != nil {
		return BalanceCheckResponse{}, err
	}

	logger.Debugf("AuthToken: [%v]", authToken)
	memID := parseMemID(authToken)
	logger.Debugf("MemID: %s", memID)

	wallet := chi.URLParam(r, "wallet")

	balance, err := store.PlayerBalance(memID, wallet)
	if err != nil {
		return BalanceCheckResponse{}, err
	}
	return BalanceCheckResponse{Token: balance.Token, BalanceResponseV2: BalanceResponseV2{Amount: balance.Balance, FreeGames: balance.FreeGames.NoOfFreeSpins, FreeSpinInfo: &FreespinResponse{CtRemaining: balance.FreeGames.NoOfFreeSpins, WagerAmt: balance.FreeGames.TotalWagerAmt}}}, nil
}

func AppendHistory(tx *store.TransactionStore, transaction engine.WalletTransaction) {
	if tx.Category == store.CategoryClose {
		tx.History = store.TransactionHistory{}
		tx.Category = store.Category("")
	}
	amount := transaction.Amount.Amount.ValueAsFloat64()
	switch store.Category(transaction.Type) {
	case store.CategoryWager:
		tx.History.NumWager++
		tx.History.SumWager += amount
	case store.CategoryPayout:
		tx.History.NumPayout++
		tx.History.SumPayout += amount
	}
	return
}

func DefaultTitleName(name string) string {
	parts := strings.Split(strings.ReplaceAll(name, "-", " "), " ")
	for i, p := range parts {
		parts[i] = strings.ToUpper(string(p[0])) + strings.TrimPrefix(p, string(p[0]))
	}
	return strings.Join(parts, " ")
}

func getCascadePositions(state engine.Gamestate) []int {
	if strings.Contains(state.Action, "cascade") || state.Action == "pushreels" {
		// this is a hack for now, needed for recovery. potentially in the future we add a proper cascade positions field to the gamestate message
		return state.SelectedWinLines
	}
	return nil
}

func countFreespinsRemaining(gamestate engine.Gamestate) int {
	fsr := 0
	for i := 0; i < len(gamestate.NextActions); i++ {
		if strings.Contains(gamestate.NextActions[i], "freespin") || strings.Contains(gamestate.NextActions[i], "shuffle") {
			fsr++
		}
	}
	return fsr
}

func adjustPrizes(gamestate engine.Gamestate) []engine.Prize {
	prizes := make([]engine.Prize, len(gamestate.Prizes))
	for p := 0; p < len(gamestate.Prizes); p++ {
		prizes[p] = gamestate.Prizes[p]
		prizes[p].Win = engine.NewFixedFromInt(gamestate.Prizes[p].Payout.Multiplier * gamestate.Prizes[p].Multiplier * gamestate.Multiplier).Mul(gamestate.BetPerLine.Amount)
	}
	return prizes
}
