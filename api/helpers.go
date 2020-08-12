package api

import (
	"github.com/go-chi/chi"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strings"
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

func PlayerBalance(r *http.Request) (BalanceCheckResponse, rgse.RGSErr) {
	authToken, err := processAuthorization(r)
	if err != nil {
		return BalanceCheckResponse{}, err
	}

	logger.Debugf("AuthToken: [%v]", authToken)
	memID := strings.Split(authToken, "=")[1]
	memID = strings.Trim(memID, "'")
	logger.Debugf("MemID: %s", memID)

	wallet := chi.URLParam(r, "wallet")

	balance, err := store.PlayerBalance(memID, wallet)
	if err != nil {
		return BalanceCheckResponse{}, err
	}
	return BalanceCheckResponse{Token: balance.Token, BalanceResponseV2: BalanceResponseV2{Amount: balance.Balance, FreeGames: balance.FreeGames.NoOfFreeSpins, FreeSpinInfo: &FreespinResponse{CtRemaining:balance.FreeGames.NoOfFreeSpins, WagerAmt:balance.FreeGames.WagerAmt}}}, nil
}
