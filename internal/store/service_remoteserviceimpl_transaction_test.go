package store

import (
	"bytes"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testRemoteServiceForTransaction(url string) Service {
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

func testTransactinoStoreRemoteServiceForTransaction(token string) TransactionStore {
	return TransactionStore{
		TransactionId: "tx-2",
		Token:         Token(token),
		Mode:          ModeReal,
		Category:      CategoryPayout,
		RoundStatus:   RoundStatusClose,
		PlayerId:      uuid.NewV4().String(),
		GameId:        "1",
		RoundId:       "1",
		Amount: engine.Money{
			Currency: "USD",
			Amount:   10,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           nil,
	}
}

func TestRemoteServiceImpl_Transaction_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := uuid.NewV4().String()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/transaction" == req.URL.String() {
			rs := restBalanceResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
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
