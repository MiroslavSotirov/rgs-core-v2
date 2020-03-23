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

func testRemoteServiceForTransactionByGameId(url string) Service {
	return New(&config.Config{
		DevMode: false,
		DashurConfig: config.StoreConfig{
			StoreRemoteUrl: url + "/v1/gnrc/maverick",
			StoreAppId:     "store-app-id",
			StoreAppPass:   "P@ssw0rd^^",
		},
		DefaultPlatform: "html5",
		DefaultLanguage: "en",
		DemoTokenPrefix: "demo-token",
		DemoCurrency:    "USD",
	})
}

func TestRemoteServiceImpl_TransactionByGameId_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := uuid.NewV4().String()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			lastTx := restTransactionRequest{
				ReqId:          "",
				Token:          token,
				Game:           "",
				Platform:       "",
				Mode:           "",
				Session:        "",
				Currency:       "USD",
				Amount:         100,
				BonusAmount:    0,
				JpAmount:       0,
				Category:       "",
				CampaignRef:    "",
				CloseRound:     false,
				GameState:      "",
				Round:          "",
				TxRef:          "",
				Description:    "",
				InternalStatus: 0,
			}
			rs := restQueryResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
					VendorInfo: restVendorResponse{
						LastAttemptedTx: lastTx,
					},
				},
				ResponseCode: "0",
				Message:      "",
				ReqId:        "",
				//CampaignRef:  "",
				FreeGames: restFreeGame{
					CampaignRef: "",
					NrGames:     0,
				},
				BetLimit: "",
				LastTx:          lastTx,
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	txStore, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if txStore.Token != Token(token) {
		t.Errorf("Found error, token is not equal [%v]", txStore)
	}

	if txStore.Amount.Currency != "USD" {
		t.Errorf("Found error, currency is not equal [%v]", txStore.Amount.Currency)
	}

	if txStore.Amount.Amount != engine.NewFixedFromFloat(float32(100/100)) {
		t.Errorf("Found error, balance is not equal [%v] - [%v]", txStore.Amount.Amount, engine.NewFixedFromFloat(float32(100/100)))
	}
}

func TestRemoteServiceImpl_TransactionByGameId_2(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rw.WriteHeader(500)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_3(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_4(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_5(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rw.WriteHeader(402)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "1",
				Message:      "",
				Balance:      100,
				Currency:     "USD",
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "2",
				Message:      "",
				Balance:      100,
				Currency:     "USD",
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "3",
				Message:      "",
				Balance:      100,
				Currency:     "USD",
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "4",
				Message:      "",
				Balance:      100,
				Currency:     "USD",
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransactionByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.TransactionByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}
