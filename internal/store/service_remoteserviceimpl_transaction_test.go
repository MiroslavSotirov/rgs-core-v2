package store

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

func testRemoteServiceForTransaction(url string) Service {
	return New(&config.Config{
		DevMode: false,
		DashurConfig: config.StoreConfig{
			StoreRemoteUrl:  url + "/v1/gnrc/maverick",
			StoreAppId:      "store-app-id",
			StoreAppPass:    "P@ssw0rd^^",
			StoreMaxRetries: 2,
			StoreTimeoutMs:  100,
		},
		DefaultPlatform: "html5",
		DefaultLanguage: "en",
		DemoTokenPrefix: "demo-token",
		DemoCurrency:    "USD",
	})
}

func testTransactinoStoreRemoteServiceForTransaction(token string) TransactionStore {
	return TransactionStore{
		TransactionId: "tx-2",
		Token:         Token(token),
		Mode:          ModeReal,
		Category:      CategoryPayout,
		RoundStatus:   RoundStatusClose,
		PlayerId:      rng.Uuid(),
		GameId:        "1",
		RoundId:       "1",
		Amount: engine.Money{
			Currency: "USD",
			Amount:   10,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           nil,
		Ttl:                 3600,
	}
}

func TestRemoteServiceImpl_Transaction_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := rng.Uuid()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: "0",
					Message:      "",
				},
				Token:    token,
				Balance:  100,
				Currency: "USD",
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	balance, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction(token))

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

func TestRemoteServiceImpl_Transaction_2(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rw.WriteHeader(500)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_3(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_4(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_5(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rw.WriteHeader(402)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
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

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.InsufficientFundError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
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

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.BadRequest {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
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

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.TokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
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

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func RemoteServiceImpl_Transaction_retry(failTries int, delayMs int64) rgse.RGSErr {
	logger.NewLogger(logger.Configuration{})
	try := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			code := func() string {
				try++
				if try > failTries {
					return "0"
				}
				return "4"
			}
			start := time.Now()
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          rng.Uuid(),
					ProcessingTime: 0,
				},
				restErrorResponse: restErrorResponse{
					ResponseCode: code(),
					Message:      "",
				},
				Token:    rng.Uuid(),
				Balance:  100,
				Currency: "USD",
			}
			for time.Now().Sub(start).Milliseconds() < delayMs {

			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForTransaction(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"

	_, err := serv.Transaction(Token(tokenStr), mode, testTransactinoStoreRemoteServiceForTransaction("token"))

	return err
}

/*
// disable retry until more testing
func TestRemoteServiceImpl_Transaction_10(t *testing.T) {
	err := RemoteServiceImpl_Transaction_retry(2, 0)
	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_11(t *testing.T) {
	err := RemoteServiceImpl_Transaction_retry(3, 0)
	if err == nil {
		t.Errorf("Error expected")
		return
	}
	if err.(*rgserror.RGSError).ErrCode != rgserror.GenericWalletError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_Transaction_12(t *testing.T) {
	err := RemoteServiceImpl_Transaction_retry(2, 80)
	if err == nil {
		t.Errorf("Error expected")
		return
	}
}

func TestRemoteServiceImpl_Transaction_13(t *testing.T) {
	err := RemoteServiceImpl_Transaction_retry(2, 40)
	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
		return
	}
}
*/
