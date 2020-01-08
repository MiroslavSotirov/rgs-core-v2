package store

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRemoteServiceForGameStateByGameId(url string) Service {
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

func TestRemoteServiceImpl_GameStateByGameId_1(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	token := uuid.NewV4().String()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			gameState := engine.Gamestate{
				GameID:      "Abc:Def",
				NextActions: []string{"fs1", "fs2", "fs3"},
				Transactions: []engine.WalletTransaction{engine.WalletTransaction{
					Id: uuid.NewV4().String(),
				}},
			}
			rs := restGameStateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        token,
				ResponseCode: "0",
				Message:      "",
				GameState:    base64.StdEncoding.EncodeToString(SerializeGamestateToBytes(gameState)),
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
			logger.Infof(string(b.Bytes()))
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	gameState, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err != nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if len(gameState.GameState) <= 0 {
		t.Errorf("Found error, game state is empyt")
	}
}

func TestRemoteServiceImpl_GameStateByGameId_2(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			rw.WriteHeader(500)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_3(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_4(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			rw.WriteHeader(401)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_5(t *testing.T) {
	logger.NewLogger(logger.Configuration{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			rw.WriteHeader(402)
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_6(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			gameState := engine.Gamestate{
				GameID:      "Abc:Def",
				NextActions: []string{"fs1", "fs2", "fs3"},
				Transactions: []engine.WalletTransaction{engine.WalletTransaction{
					Id: uuid.NewV4().String(),
				}},
			}
			rs := restGameStateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "1",
				Message:      "",
				GameState:    base64.StdEncoding.EncodeToString(SerializeGamestateToBytes(gameState)),
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeNotEnoughBalance {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_7(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			gameState := engine.Gamestate{
				GameID:      "Abc:Def",
				NextActions: []string{"fs1", "fs2", "fs3"},
				Transactions: []engine.WalletTransaction{engine.WalletTransaction{
					Id: uuid.NewV4().String(),
				}},
			}
			rs := restGameStateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "2",
				Message:      "",
				GameState:    base64.StdEncoding.EncodeToString(SerializeGamestateToBytes(gameState)),
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_8(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			gameState := engine.Gamestate{
				GameID:      "Abc:Def",
				NextActions: []string{"fs1", "fs2", "fs3"},
				Transactions: []engine.WalletTransaction{engine.WalletTransaction{
					Id: uuid.NewV4().String(),
				}},
			}
			rs := restGameStateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "3",
				Message:      "",
				GameState:    base64.StdEncoding.EncodeToString(SerializeGamestateToBytes(gameState)),
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeTokenExpired {
		t.Errorf("Error code not match [%v]", err)
	}
}

func TestRemoteServiceImpl_GameStateByGameId_9(t *testing.T) {
	logger.NewLogger(logger.Configuration{})
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if "/v1/gnrc/maverick/gamestate" == req.URL.String() {
			gameState := engine.Gamestate{
				GameID:      "Abc:Def",
				NextActions: []string{"fs1", "fs2", "fs3"},
				Transactions: []engine.WalletTransaction{engine.WalletTransaction{
					Id: uuid.NewV4().String(),
				}},
			}
			rs := restGameStateResponse{
				Metadata: restMetadata{
					ReqId:          uuid.NewV4().String(),
					ProcessingTime: 0,
				},
				Token:        uuid.NewV4().String(),
				ResponseCode: "4",
				Message:      "",
				GameState:    base64.StdEncoding.EncodeToString(SerializeGamestateToBytes(gameState)),
			}
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(rs)
			rw.Write(b.Bytes())
		}
	}))
	defer server.Close()

	serv := testRemoteServiceForGameStateByGameId(server.URL)
	mode := ModeDemo
	tokenStr := "refresh-token"
	gameIdStr := "MVRK-TEST-GAME-1"

	_, err := serv.GameStateByGameId(Token(tokenStr), mode, gameIdStr)

	if err == nil {
		t.Errorf("Found error, it shouldn't produce error [%v]", err)
	}

	if err.Code != ErrorCodeGeneralError {
		t.Errorf("Error code not match [%v]", err)
	}
}
