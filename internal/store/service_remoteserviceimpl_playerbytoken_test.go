package store

import (
	"bytes"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRemoteServiceForPlayerByToken(url string) Service {
	return New(&config.Config{
		DevMode:         false,
		DashurConfig: config.StoreConfig{
			StoreRemoteUrl:  url + "/v1/gnrc/maverick",
			StoreAppId:      "store-app-id",
			StoreAppPass:    "P@ssw0rd^^",
		},
		DefaultPlatform: "html5",
		DefaultLanguage: "en",
		DemoTokenPrefix: "demo-token",
		DemoCurrency:    "USD",
	})
}

func TestRemoteServiceImpl_PlayerByToken_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := uuid.NewV4().String()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			authRs := restAuthenticateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        token,
				ResponseCode: "0",
				Message:      "",
				Id:           uuid.NewV4().String(),
				Username:     uuid.NewV4().String(),
				BetLimit:     "",
				FreeGames: restFreeGame{
					NrGames:     0,
					CampaignRef: "",
				},
				Balance:  100,
				Currency: "USD",
			}

			b := new(bytes.Buffer)
			jsonErr := json.NewEncoder(b).Encode(authRs)

			if jsonErr != nil {
				rw.WriteHeader(500) // if unable to write json to string, throw err 500
			} else {
				rw.Write(b.Bytes())
			}
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	player, gameState, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if player.Token != Token(token) {
		t.Errorf("Found error, token is not equal [%v]", player)
	}

	if player.Balance.Currency != "USD" {
		t.Errorf("Found error, currency is not equal [%v]", player.Balance.Currency)
	}

	if player.Balance.Amount != engine.NewFixedFromFloat(float32(100/100)) {
		t.Errorf("Found error, balance is not equal [%v] - [%v]", player.Balance.Amount, engine.NewFixedFromFloat(float32(100/100)))
	}

	if len(gameState.GameState) != 0 {
		t.Errorf("Found error, game state not correct [%v]", gameState)
	}
}

func TestRemoteServiceImpl_PlayerByToken_2(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			rw.WriteHeader(500)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_3(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_4(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_5(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			rw.WriteHeader(402)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			authRs := restAuthenticateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "1",
				Message:      "",
				Id:           uuid.NewV4().String(),
				Username:     uuid.NewV4().String(),
				BetLimit:     "",
				FreeGames: restFreeGame{
					NrGames:     0,
					CampaignRef: "",
				},
				Balance:  100,
				Currency: "USD",
			}

			b := new(bytes.Buffer)
			jsonErr := json.NewEncoder(b).Encode(authRs)

			if jsonErr != nil {
				rw.WriteHeader(500) // if unable to write json to string, throw err 500
			} else {
				rw.Write(b.Bytes())
			}
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			authRs := restAuthenticateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "2",
				Message:      "",
				Id:           uuid.NewV4().String(),
				Username:     uuid.NewV4().String(),
				BetLimit:     "",
				FreeGames: restFreeGame{
					NrGames:     0,
					CampaignRef: "",
				},
				Balance:  100,
				Currency: "USD",
			}

			b := new(bytes.Buffer)
			jsonErr := json.NewEncoder(b).Encode(authRs)

			if jsonErr != nil {
				rw.WriteHeader(500) // if unable to write json to string, throw err 500
			} else {
				rw.Write(b.Bytes())
			}
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			authRs := restAuthenticateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "3",
				Message:      "",
				Id:           uuid.NewV4().String(),
				Username:     uuid.NewV4().String(),
				BetLimit:     "",
				FreeGames: restFreeGame{
					NrGames:     0,
					CampaignRef: "",
				},
				Balance:  100,
				Currency: "USD",
			}

			b := new(bytes.Buffer)
			jsonErr := json.NewEncoder(b).Encode(authRs)

			if jsonErr != nil {
				rw.WriteHeader(500) // if unable to write json to string, throw err 500
			} else {
				rw.Write(b.Bytes())
			}
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_PlayerByToken_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/auth" == req.URL.String() {
			authRs := restAuthenticateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "4",
				Message:      "",
				Id:           uuid.NewV4().String(),
				Username:     uuid.NewV4().String(),
				BetLimit:     "",
				FreeGames: restFreeGame{
					NrGames:     0,
					CampaignRef: "",
				},
				Balance:  100,
				Currency: "USD",
			}

			b := new(bytes.Buffer)
			jsonErr := json.NewEncoder(b).Encode(authRs)

			if jsonErr != nil {
				rw.WriteHeader(500) // if unable to write json to string, throw err 500
			} else {
				rw.Write(b.Bytes())
			}
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForPlayerByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, _, err := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

//func TestRemoteServiceImpl_PlayerByToken(t *testing.T) {
//	logger.NewLogger(logger.Configuration{})
//	mode := ModeDemo
//	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdCI6ImMwYjFmNmE4LTFkZmEtNDlhZC1hMWVlLTIxNjY0NmRiYWUzNiIsImN0eCI6OCwidXNlcl9uYW1lIjoiRHJhY28uQ08uMS5UMy1tYnItMDAxOjgwNzIiLCJ1aWQiOjI1ODg2MjIsImFpZCI6MjU3NDcxN30.5bcomka14fdzBI3SwNBwE7Yk9CiRWAMPPz_5LAM6St8"
//	gameIdStr := "MVRK-TEST-GAME-1"
//	serv := New(&config.Config{
//		DevMode:         false,
//		StoreRemoteUrl:  "https://gnrc-api.dashur.io/v1/gnrc/maverick",
//		StoreAppId:      "VrCLq4UqYBf39rJt",
//		StoreAppPass:    "V6Kg9muzJxsz3nWFvJeyccU7",
//		DefaultPlatform: "html5",
//		DefaultLanguage: "en",
//		DemoTokenPrefix: "demo-token-ezVVKaEv2nY7kP2gMCX7vfCH",
//		DemoCurrency:    "USD",
//	})
//	player, _, _ := serv.PlayerByToken(Token(tokenStr), mode, gameIdStr)
//
//	logger.Infof("Value of player  => [%v]", player)
//
//	balance, _ := serv.BalanceByToken(Token(player.Token), mode)
//
//	logger.Infof("Value of balance  => [%v]", balance)
//
//	roundId := uuid.NewV4().String()
//
//	balance2, _ := serv.Transaction(Token(player.Token), mode, TransactionStore{
//		TransactionId: uuid.NewV4().String(),
//		Token:         player.Token,
//		Mode:          mode,
//		Category:      CategoryWager,
//		RoundStatus:   RoundStatusOpen,
//		PlayerId:      player.PlayerId,
//		GameId:        gameIdStr,
//		RoundId:       roundId,
//		Amount: engine.Money{
//			Currency: player.Balance.Currency,
//			Amount:   10,
//		},
//		TxTime:    time.Now(),
//		GameState: nil,
//	})
//
//	logger.Infof("Value of balance2  => [%v]", balance2)
//
//	tx1, _ := serv.TransactionByGameId(Token(player.Token), mode, gameIdStr)
//
//	logger.Infof("Value of tx1  => [%v]", tx1)
//
//	bal3, _ := serv.CloseRound(Token(player.Token), mode, gameIdStr, roundId)
//
//	logger.Infof("Value of bal3  => [%v]", bal3)
//
//	tx2, _ := serv.TransactionByGameId(Token(player.Token), mode, gameIdStr)
//
//	logger.Infof("Value of tx2  => [%v]", tx2)
//}
