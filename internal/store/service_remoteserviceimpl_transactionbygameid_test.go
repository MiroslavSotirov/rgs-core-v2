package store

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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
	token := rng.Uuid()
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
				Round:          "",
				Description:    "",
				InternalStatus: 0,
				Ttl:            3600,
				TtlStamp:       time.Now().Unix() + 3600,
				restTransactionDesc: restTransactionDesc{
					Amount:      100,
					BonusAmount: 0,
					JpAmount:    0,
					Category:    "",
					CampaignRef: "",
					CloseRound:  false,
					GameState:   "",
					TxRef:       "",
				},
			}
			rs := restQueryResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
					VendorInfo: restVendorResponse{
						LastAttemptedTx: lastTx,
					},
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "0",
					Message:      "",
				},
				ReqId: "",
				//CampaignRef:  "",
				FreeGames: restFreeGame{
					CampaignRef: "",
					NrGames:     0,
				},
				BetLimit: "",
				LastTx:   lastTx,
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "1",
					Message:      "",
				},
				Token:    rng.Uuid(),
				Balance:  100,
				Currency: "USD",
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "2",
					Message:      "",
				},
				Token:    rng.Uuid(),
				Balance:  100,
				Currency: "USD",
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.BadRequest {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "3",
					Message:      "",
				},
				Token:    rng.Uuid(),
				Balance:  100,
				Currency: "USD",
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_TransactionByGameId_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/query" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "4",
					Message:      "",
				},
				Token:    rng.Uuid(),
				Balance:  100,
				Currency: "USD",
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

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}
