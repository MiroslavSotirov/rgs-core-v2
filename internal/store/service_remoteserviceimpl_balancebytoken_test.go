package store

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

func testRemoteServiceForBalanceByToken(url string) Service {
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

func TestRemoteServiceImpl_BalanceByToken_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := rng.Uuid()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				Token:        token,
				ResponseCode: "0",
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

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	balance, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if balance.Token != Token(token) {
		t.Errorf("Found error, token is not equal [%v]", balance)
	}

	if balance.Balance.Currency != "USD" {
		t.Errorf("Found error, currency is not equal [%v]", balance.Balance.Currency)
	}

	if balance.Balance.Amount != engine.NewFixedFromFloat(float32(100/100)) {
		t.Errorf("Found error, balance is not equal [%v] - [%v]", balance.Balance.Amount, engine.NewFixedFromFloat(float32(100/100)))
	}
}

func TestRemoteServiceImpl_BalanceByToken_2(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rw.WriteHeader(500)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_3(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_4(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_5(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rw.WriteHeader(402)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				Token:        rng.Uuid(),
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

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				Token:        rng.Uuid(),
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

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("No error, it should produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.BadRequest {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				Token:        rng.Uuid(),
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

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_BalanceByToken_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/balance" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				Token:        rng.Uuid(),
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

	serv := testRemoteServiceForBalanceByToken(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.BalanceByToken(Token(tokenStr), mode)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}
