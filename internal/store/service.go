package store

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	"github.com/travelaudience/go-promhttp"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	CategoryWager  Category = "WAGER"
	CategoryPayout Category = "PAYOUT"
	CategoryRefund Category = "REFUND"
	CategoryClose  Category = "ENDROUND"

	ApiTypeVersion     ApiType = "version"
	ApiTypeAuth        ApiType = "auth"
	ApiTypeBalance     ApiType = "balance"
	ApiTypeTransaction ApiType = "transaction"
	ApiTypeGameState   ApiType = "gamestate"
	ApiTypeQuery       ApiType = "query"

	ModeDemo Mode = "DEMO"
	ModeReal Mode = "REAL"

	RoundStatusOpen  RoundStatus = "OPEN"
	RoundStatusClose RoundStatus = "CLOSE"

	ErrorCodeGeneralError     ErrorCode = "ERR-001" // general error.
	ErrorCodeNotEnoughBalance ErrorCode = "ERR-002"
	ErrorCodeTokenExpired     ErrorCode = "ERR-003"
	ErrorCodeEntityNotFound   ErrorCode = "ERR-004"

	ResponseCodeOk                 ResponseCode = "0"
	ResponseCodeInsufficentBalance ResponseCode = "1"
	ResponseCodeDataError          ResponseCode = "2"
	ResponseCodeSessionExpired     ResponseCode = "3"
	ResponseCodeUnknownError                    = "4"
)

type (
	Token        string
	Category     string
	Mode         string
	RoundStatus  string
	ErrorCode    string
	ApiType      string
	ResponseCode string

	PlayerStore struct {
		PlayerId            string
		Token               Token
		Mode                Mode
		Username            string
		Balance             engine.Money
		BetLimitSettingCode string
		NoOfFreeSpins       int
	}

	GameStateStore struct {
		GameState []byte
	}

	BalanceStore struct {
		PlayerId string
		Token    Token
		Mode     Mode
		Balance  engine.Money
	}

	TransactionStore struct {
		TransactionId       string
		Token               Token
		Mode                Mode
		Category            Category
		RoundStatus         RoundStatus
		PlayerId            string
		GameId              string
		RoundId             string
		Amount              engine.Money
		ParentTransactionId string
		TxTime              time.Time
		GameState           []byte
	}

	Error struct {
		Code    ErrorCode
		Message string
	}

	LocalData struct {
		Token                   map[Token]string
		Player                  map[string]PlayerStore
		Transaction             map[string]TransactionStore
		TransactionByPlayerGame map[string]TransactionStore
		Lock                    sync.RWMutex
	}

	Service interface {
		// authenticate token, given the game id, it will also retrieve the latest gamestate from latest transaction.
		PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, *Error)

		// retrieve latest balance by token
		BalanceByToken(token Token, mode Mode) (BalanceStore, *Error)

		// create transaction.
		Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, *Error)

		// retrieve latest transcation by player and by game id
		TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error)

		// retrieve latest game state by game id
		GameStateByGameId(token Token, mode Mode, gameId string) (GameStateStore, *Error)

		// close round.
		CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error)

		//// gamestate by id
		//GamestateById(gamestateId string) (GameStateStore, *Error)
	}

	RemoteServiceImpl struct {
		serverUrl       string
		appId           string
		appCredential   string
		defaultPlatform string
		defaultLanguage string
		demoTokenPrefix string
		demoCurrency    string
	}

	// local service eq implemenation of service. so that unit test of services can be easily mocked.
	LocalService interface {
		PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, *Error)
		PlayerSave(token Token, mode Mode, player PlayerStore) (PlayerStore, *Error)
		BalanceByToken(token Token, mode Mode) (BalanceStore, *Error)
		Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, *Error)
		TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error)
		GameStateByGameId(token Token, mode Mode, gameId string) (GameStateStore, *Error)
		CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error)
		GamestateById(gamestateId string) (GameStateStore, *Error)
	}

	LocalServiceImpl struct{}

	restMetadata struct {
		ReqId          string `json:"req_id"`
		ProcessingTime int    `json:"processing_time"`
	}

	restAuthenticateRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
		Language string `json:"language"`
	}

	restFreeGame struct {
		CampaignRef string `json:"campaign_ref"`
		NrGames     int    `json:"nr_games"`
	}

	restPlayerMessage struct {
		Title    string `json:"title"`
		Link     string `json:"link"`
		Message  string `json:"message"`
		Location int    `json:"location"`
	}

	restAuthenticateResponse struct {
		Metadata      restMetadata      `json:"metadata"`
		Token         string            `json:"token"`
		ResponseCode  string            `json:"code"`
		Message       string            `json:"message"`
		Id            string            `json:"id"`
		Username      string            `json:"username"`
		BetLimit      string            `json:"bet_limit"`
		FreeGames     restFreeGame      `json:"free_games"`
		Balance       int64             `json:"balance"`
		Currency      string            `json:"currency"`
		LastGameState string            `json:"last_game_state"`
		PlayerMessage restPlayerMessage `json:"player_message"`
		Urls          map[string]string `json:"urls"`
	}

	restBalanceRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restBalanceResponse struct {
		Metadata      restMetadata      `json:"metadata"`
		Token         string            `json:"token"`
		ResponseCode  string            `json:"code"`
		Message       string            `json:"message"`
		PlayerId      string            `json:"player_id"`
		Balance       int64             `json:"balance"`
		Currency      string            `json:"currency"`
		PlayerMessage restPlayerMessage `json:"player_message"`
	}

	restGameStateRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restGameStateResponse struct {
		Metadata     restMetadata `json:"metadata"`
		Token        string       `json:"token"`
		ResponseCode string       `json:"code"`
		Message      string       `json:"message"`
		GameState    string       `json:"game_state"`
	}

	restVersionRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restVersionResponse struct {
		Metadata     restMetadata `json:"metadata"`
		Token        string       `json:"token"`
		ResponseCode string       `json:"code"`
		Message      string       `json:"message"`
		Version      string       `json:"version"`
	}

	restTransactionRequest struct {
		ReqId       string `json:"req_id"`
		Token       string `json:"token"`
		Game        string `json:"game"`
		Platform    string `json:"platform"`
		Mode        string `json:"mode"`
		Session     string `json:"session"`
		Currency    string `json:"currency"`
		Amount      int64  `json:"amount"`
		BonusAmount int64  `json:"bonus_amount"`
		JpAmount    int64  `json:"jp_amount"`
		Category    string `json:"category"`
		CampaignRef string `json:"campaign_ref"`
		CloseRound  bool   `json:"close_round"`
		GameState   string `json:"game_state"`
		Round       string `json:"round"`
		TxRef       string `json:"tx_ref"`
		Description string `json:"description"`
	}

	restTransactionResponse struct {
		Metadata      restMetadata      `json:"metadata"`
		Token         string            `json:"token"`
		ResponseCode  string            `json:"code"`
		Message       string            `json:"message"`
		PlayerId      string            `json:"player_id"`
		Balance       int64             `json:"balance"`
		Currency      string            `json:"currency"`
		TxId          string            `json:"tx_id"`
		PlayerMessage restPlayerMessage `json:"player_message"`
	}

	restQueryRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restQueryResponse struct {
		Metadata     restMetadata `json:"metadata"`
		Token        string       `json:"token"`
		ResponseCode string       `json:"code"`
		Message      string       `json:"message"`
		ReqId        string       `json:"req_id"`
		Game         string       `json:"game"`
		Platform     string       `json:"platform"`
		Mode         string       `json:"mode"`
		Session      string       `json:"session"`
		Currency     string       `json:"currency"`
		Amount       int64        `json:"amount"`
		BonusAmount  int64        `json:"bonus_amount"`
		JpAmount     int64        `json:"jp_amount"`
		Category     string       `json:"category"`
		CampaignRef  string       `json:"campaign_ref"`
		CloseRound   bool         `json:"close_round"`
		GameState    string       `json:"game_state"`
		Round        string       `json:"round"`
		TxRef        string       `json:"tx_ref"`
		Description  string       `json:"description"`
	}
)

var ld *LocalData
var remoteServiceImplHttpClient *promhttp.Client

func (i *LocalServiceImpl) PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, *Error) {
	logger.Debugf("LocalServiceImpl.PlayerByToken([%v], [%v])", token, mode)

	err := internalCheck()
	if err != nil {
		panic(err)
	}

	if ModeReal == mode {
		playerId, _ := i.getToken(token)
		player, _ := i.getPlayer(playerId)
		newToken := i.renewToken(token)
		key := player.PlayerId + "::" + gameId
		tx, txExists := i.getTransactionByPlayerGame(key)

		if txExists && tx.GameState != nil && len(tx.GameState) > 0 {
			return PlayerStore{
					PlayerId: player.PlayerId,
					Token:    newToken,
					Mode:     player.Mode,
					Username: player.Username,
					Balance: engine.Money{
						Currency: player.Balance.Currency,
						Amount:   player.Balance.Amount,
					},
					NoOfFreeSpins:       player.NoOfFreeSpins,
					BetLimitSettingCode: player.BetLimitSettingCode,
				},
				GameStateStore{GameState: tx.GameState},
				nil
		} else {
			return PlayerStore{
					PlayerId: player.PlayerId,
					Token:    newToken,
					Mode:     player.Mode,
					Username: player.Username,
					Balance: engine.Money{
						Currency: player.Balance.Currency,
						Amount:   player.Balance.Amount,
					},
					NoOfFreeSpins:       player.NoOfFreeSpins,
					BetLimitSettingCode: player.BetLimitSettingCode,
				},
				GameStateStore{},
				nil
		}
	} else if ModeDemo == mode {
		storePlayerId, playerIdExists := i.getToken(token)
		logger.Debugf("player id: %v", storePlayerId)
		if playerIdExists {
			player, _ := i.getPlayer(storePlayerId)
			newToken := i.renewToken(token)
			key := player.PlayerId + "::" + gameId
			tx, txExists := i.getTransactionByPlayerGame(key)

			if txExists && tx.GameState != nil && len(tx.GameState) > 0 {
				return PlayerStore{
						PlayerId: player.PlayerId,
						Token:    newToken,
						Mode:     player.Mode,
						Username: player.Username,
						Balance: engine.Money{
							Currency: player.Balance.Currency,
							Amount:   player.Balance.Amount,
						},
						NoOfFreeSpins:       player.NoOfFreeSpins,
						BetLimitSettingCode: player.BetLimitSettingCode,
					},
					GameStateStore{GameState: tx.GameState},
					nil
			} else {
				// this is likely an error, if player exists, there should be a previous gameplay unless init was called and never spun, which will throw an error
				logger.Warnf("DEMO WALLET PLAYER EXISTS BUT NO PREVIOUS TX")
				return PlayerStore{
						PlayerId: player.PlayerId,
						Token:    newToken,
						Mode:     player.Mode,
						Username: player.Username,
						Balance: engine.Money{
							Currency: player.Balance.Currency,
							Amount:   player.Balance.Amount,
						},
						NoOfFreeSpins:       player.NoOfFreeSpins,
						BetLimitSettingCode: player.BetLimitSettingCode,
					},
					GameStateStore{},
					nil
			}
		} else {
			logger.Warnf("NO PLAYER EXISTS")
			return PlayerStore{}, GameStateStore{}, nil
		}
	} else {
		return PlayerStore{}, GameStateStore{}, &Error{Code: ErrorCodeGeneralError, Message: "Unknown mode"}
	}
}

func (i *RemoteServiceImpl) request(apiType ApiType, body io.Reader) *http.Request {
	req, _ := http.NewRequest("POST", i.serverUrl+"/"+string(apiType), body)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(i.appId, i.appCredential)
	return req
}

func (i *RemoteServiceImpl) httpClient() http.Client {
	dashurClient, _ := remoteServiceImplHttpClient.ForRecipient("dashur")
	return *dashurClient
}

func (i *RemoteServiceImpl) demoToken() string {
	return i.demoTokenPrefix + ":" + i.demoCurrency + ":" + uuid.NewV4().String()
}

func (i *RemoteServiceImpl) errorJson(err error) *Error {
	if err != nil {
		return &Error{Code: ErrorCodeGeneralError, Message: "Unknown error in performing json conversion"}
	}
	return nil
}

func (i *RemoteServiceImpl) errorRest(err error) *Error {
	if err != nil {
		return &Error{Code: ErrorCodeGeneralError, Message: "Unknown error in performing rest call"}
	}
	return nil
}

func (i *RemoteServiceImpl) errorBase64(err error) *Error {
	if err != nil {
		return &Error{Code: ErrorCodeGeneralError, Message: "Error encoding/decoding base64 string"}
	}
	return nil
}

func (i *RemoteServiceImpl) errorHttpStatusCode(httpStatusCode int) *Error {
	if httpStatusCode != 200 {
		if httpStatusCode == 403 || httpStatusCode == 401 {
			return &Error{Code: ErrorCodeTokenExpired, Message: "Auth error in performing rest function"}
		} else if httpStatusCode == 404 {
			return &Error{Code: ErrorCodeEntityNotFound, Message: "Entity not found in performing rest function"}
		} else if httpStatusCode == 402 {
			return &Error{Code: ErrorCodeNotEnoughBalance, Message: "Not enough balance in performing rest function"}
		}
		return &Error{Code: ErrorCodeGeneralError, Message: "Unknown error in performing rest call"}
	}
	return nil
}

func (i *RemoteServiceImpl) errorResponseCode(responseCode string) *Error {
	if responseCode != string(ResponseCodeOk) {
		if responseCode == string(ResponseCodeDataError) {
			return &Error{Code: ErrorCodeGeneralError, Message: "Unable to perform rest function, found data input error"}
		} else if responseCode == string(ResponseCodeInsufficentBalance) {
			return &Error{Code: ErrorCodeNotEnoughBalance, Message: "Unable to perform rest function, not enough balance"}
		} else if responseCode == string(ResponseCodeSessionExpired) {
			return &Error{Code: ErrorCodeTokenExpired, Message: "Unable to perform rest function, token expired"}
		}
		return &Error{Code: ErrorCodeGeneralError, Message: "Unknown error in performing rest call"}
	}
	return nil
}

func (i *RemoteServiceImpl) PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, *Error) {
	if mode == ModeDemo {
		token = Token(i.demoToken())
	}

	authRq := restAuthenticateRequest{
		ReqId:    uuid.NewV4().String(),
		Token:    string(token),
		Game:     gameId,
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
		Language: i.defaultLanguage,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(authRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}

	req := i.request(ApiTypeAuth, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}

	var gameState []byte = nil
	authResp := i.restAuthenticateResponse(resp)
	logger.Debugf("response: %#v", authResp)
	finalErr = i.errorResponseCode(authResp.ResponseCode)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}

	if len(authResp.LastGameState) > 0 {
		gameState, err = base64.StdEncoding.DecodeString(authResp.LastGameState)

		finalErr = i.errorBase64(err)
		logger.Debugf("error: %v", finalErr)
		if finalErr != nil {
			return PlayerStore{}, GameStateStore{}, finalErr
		}
	}

	return PlayerStore{
			PlayerId: authResp.Id,
			Token:    Token(authResp.Token),
			Mode:     mode,
			Username: authResp.Username,
			Balance: engine.Money{
				Currency: authResp.Currency,
				Amount:   engine.Fixed(authResp.Balance * 10000),
			},
			NoOfFreeSpins:       authResp.FreeGames.NrGames,
			BetLimitSettingCode: authResp.BetLimit,
		},
		GameStateStore{GameState: gameState},
		nil
}

func (i *RemoteServiceImpl) restAuthenticateResponse(response *http.Response) restAuthenticateResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restAuthenticateResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) PlayerSave(token Token, mode Mode, player PlayerStore) (PlayerStore, *Error) {
	logger.Debugf("LocalServiceImpl.PlayerSave([%v], [%v], [%v])", token, mode, player)

	i.setToken(token, player.PlayerId)
	i.setPlayer(player.PlayerId, player)
	newToken := i.renewToken(token)

	return PlayerStore{
		PlayerId: player.PlayerId,
		Token:    newToken,
		Mode:     player.Mode,
		Username: player.Username,
		Balance: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   player.Balance.Amount,
		},
		NoOfFreeSpins:       player.NoOfFreeSpins,
		BetLimitSettingCode: player.BetLimitSettingCode,
	}, nil
}

func (i *LocalServiceImpl) renewToken(token Token) Token {
	playerId, _ := i.getToken(token)
	newToken := Token(rng.RandStringRunes(36))
	i.setToken(newToken, playerId)
	i.deleteToken(token)

	return newToken
}

func (i *LocalServiceImpl) BalanceByToken(token Token, mode Mode) (BalanceStore, *Error) {
	logger.Debugf("LocalServiceImpl.BalanceByToken([%v], [%v])", token, mode)

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)
	newToken := i.renewToken(token)

	return BalanceStore{
		PlayerId: playerId,
		Token:    newToken,
		Balance: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   player.Balance.Amount,
		},
	}, nil
}

func (i *RemoteServiceImpl) BalanceByToken(token Token, mode Mode) (BalanceStore, *Error) {
	balRq := restBalanceRequest{
		ReqId:    uuid.NewV4().String(),
		Token:    string(token),
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(balRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	req := i.request(ApiTypeBalance, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	balResp := i.restBalanceResponse(resp)

	finalErr = i.errorResponseCode(balResp.ResponseCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	return BalanceStore{
		PlayerId: balResp.PlayerId,
		Token:    Token(balResp.Token),
		Balance: engine.Money{
			Currency: balResp.Currency,
			Amount:   engine.Fixed(balResp.Balance * 10000),
		},
	}, nil
}

func (i *RemoteServiceImpl) restBalanceResponse(response *http.Response) restBalanceResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restBalanceResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, *Error) {
	logger.Debugf("LocalServiceImpl.Transaction([%v], [%v], [%v])", token, mode, transaction)

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)

	if transaction.Category == CategoryWager {
		player.Balance.Amount = player.Balance.Amount - transaction.Amount.Amount
	} else if transaction.Category == CategoryPayout {
		player.Balance.Amount = player.Balance.Amount + transaction.Amount.Amount
	} else if transaction.Category == CategoryRefund {
		parentTx, _ := i.getTransaction(transaction.ParentTransactionId)

		if parentTx.Category == CategoryWager {
			//refund wager
			player.Balance.Amount = player.Balance.Amount + transaction.Amount.Amount
		} else if parentTx.Category == CategoryPayout {
			//refund payout
			player.Balance.Amount = player.Balance.Amount - transaction.Amount.Amount
		}
	}
	if player.Balance.Amount < 0 {
		return BalanceStore{}, &Error{ErrorCodeNotEnoughBalance, "Low Balance"}
	}
	i.setTransaction(transaction.TransactionId, transaction)
	key := player.PlayerId + "::" + transaction.GameId
	i.setTransactionByPlayerGame(key, transaction)
	i.setPlayer(playerId, player)
	newToken := i.renewToken(token)

	return BalanceStore{
		PlayerId: player.PlayerId,
		Token:    newToken,
		Balance: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   player.Balance.Amount,
		},
	}, nil
}

func (i *RemoteServiceImpl) Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, *Error) {
	closeRound := false
	gameState := ""

	if RoundStatusClose == transaction.RoundStatus {
		closeRound = true
	}

	if transaction.GameState != nil {
		gameState = base64.StdEncoding.EncodeToString(transaction.GameState)
	}
	logger.Infof("SENDING TX RQ : %#v", transaction)

	txRq := restTransactionRequest{
		ReqId:       uuid.NewV4().String(),
		Token:       string(token),
		Game:        transaction.GameId,
		Platform:    i.defaultPlatform,
		Mode:        strings.ToLower(string(mode)),
		Session:     transaction.RoundId,
		Currency:    transaction.Amount.Currency,
		Amount:      int64(transaction.Amount.Amount / 10000), // Dashur expects amount in cents, transaction.Amount.Amount is type fixed (6decimals)
		BonusAmount: 0,
		JpAmount:    0,
		Category:    string(transaction.Category),
		CloseRound:  closeRound,
		GameState:   gameState,
		Round:       transaction.RoundId,
		TxRef:       transaction.TransactionId,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(txRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	req := i.request(ApiTypeTransaction, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	txResp := i.restTransactionResponse(resp)
	logger.Infof("TX RESPONSE: %#v", txResp)
	finalErr = i.errorResponseCode(txResp.ResponseCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}
	logger.Warnf("BALANCE: %v , div 10000: %v", txResp.Balance, engine.Fixed(txResp.Balance*10000))
	return BalanceStore{
		PlayerId: txResp.PlayerId,
		Token:    Token(txResp.Token),
		Balance: engine.Money{
			Currency: txResp.Currency,
			Amount:   engine.Fixed(txResp.Balance * 10000),
		},
	}, nil
}

func (i *RemoteServiceImpl) restTransactionResponse(response *http.Response) restTransactionResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restTransactionResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) GamestateById(gamestateId string) (GameStateStore, *Error) {
	logger.Debugf("LocalServiceImpl.GamestateById([%v])", gamestateId)
	transaction, ok := i.getTransaction(gamestateId)
	if !ok {
		return GameStateStore{}, &Error{ErrorCodeGeneralError, "bad gamestate ID"}
	}

	return GameStateStore{transaction.GameState}, nil
}

func (i *LocalServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error) {
	logger.Debugf("LocalServiceImpl.TransactionByGameId([%v], [%v], [%v])", token, mode, gameId)

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)
	key := player.PlayerId + "::" + gameId

	transaction, ok := i.getTransactionByPlayerGame(key)

	if !ok {
		return TransactionStore{}, &Error{ErrorCodeEntityNotFound, "No such transaction"}
	}

	return TransactionStore{
		TransactionId: transaction.TransactionId,
		Token:         token,
		Mode:          transaction.Mode,
		Category:      transaction.Category,
		RoundStatus:   transaction.RoundStatus,
		PlayerId:      transaction.PlayerId,
		GameId:        transaction.GameId,
		RoundId:       transaction.RoundId,
		Amount: engine.Money{
			Currency: transaction.Amount.Currency,
			Amount:   transaction.Amount.Amount,
		},
		ParentTransactionId: transaction.ParentTransactionId,
		TxTime:              transaction.TxTime,
		GameState:           transaction.GameState,
	}, nil
}

func (i *RemoteServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error) {
	queryRq := restQueryRequest{
		ReqId:    uuid.NewV4().String(),
		Token:    string(token),
		Game:     gameId,
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(queryRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return TransactionStore{}, finalErr
	}

	req := i.request(ApiTypeQuery, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return TransactionStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return TransactionStore{}, finalErr
	}

	var gameState []byte = nil
	queryResp := i.restQueryResponse(resp)

	finalErr = i.errorResponseCode(queryResp.ResponseCode)
	if finalErr != nil {

			// special handling for err does not exists
		if queryResp.ResponseCode == ResponseCodeUnknownError && strings.Contains(queryResp.Message, "E-CODE: [004:1003]") {
			return TransactionStore{},  &Error{ErrorCodeEntityNotFound, "Not Found"}
		} else {
			return TransactionStore{}, finalErr
		}
	}

	roundStatus := RoundStatusOpen

	if queryResp.CloseRound {
		roundStatus = RoundStatusClose
	}

	if len(queryResp.GameState) > 0 {
		gameState, err = base64.StdEncoding.DecodeString(queryResp.GameState)

		finalErr = i.errorBase64(err)
		if finalErr != nil {
			return TransactionStore{}, finalErr
		}
	}

	return TransactionStore{
		TransactionId: "", //TODO: fix this
		Token:         Token(queryResp.Token),
		Mode:          Mode(queryResp.Mode),
		Category:      Category(queryResp.Category),
		RoundStatus:   roundStatus,
		PlayerId:      "", //TODO: fix this
		GameId:        queryResp.Game,
		RoundId:       queryResp.Round,
		Amount: engine.Money{
			Currency: queryResp.Currency,
			Amount:   engine.Fixed(queryResp.Amount * 10000),
		},
		ParentTransactionId: "",         //TODO: fix this
		TxTime:              time.Now(), //TODO: fix this
		GameState:           gameState,
	}, nil
}

func (i *RemoteServiceImpl) restQueryResponse(response *http.Response) restQueryResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restQueryResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) GameStateByGameId(token Token, mode Mode, gameId string) (GameStateStore, *Error) {
	logger.Debugf("LocalServiceImpl.GameStateByGameId([%v], [%v], [%v])", token, mode, gameId)
	transaction, _ := i.TransactionByGameId(token, mode, gameId)

	return GameStateStore{
		GameState: transaction.GameState,
	}, nil

}

func (i *RemoteServiceImpl) GameStateByGameId(token Token, mode Mode, gameId string) (GameStateStore, *Error) {
	gameStateRq := restGameStateRequest{
		ReqId:    uuid.NewV4().String(),
		Token:    string(token),
		Game:     gameId,
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(gameStateRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return GameStateStore{}, finalErr
	}

	req := i.request(ApiTypeGameState, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return GameStateStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return GameStateStore{}, finalErr
	}

	var gameState []byte = nil
	gameStateResp := i.restGameStateResponse(resp)

	finalErr = i.errorResponseCode(gameStateResp.ResponseCode)
	if finalErr != nil {
		return GameStateStore{}, finalErr
	}

	if len(gameStateResp.GameState) > 0 {
		gameState, err = base64.StdEncoding.DecodeString(gameStateResp.GameState)

		finalErr = i.errorBase64(err)
		if finalErr != nil {
			return GameStateStore{}, finalErr
		}
	}

	return GameStateStore{
		GameState: gameState,
	}, nil
}

func (i *RemoteServiceImpl) restGameStateResponse(response *http.Response) restGameStateResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restGameStateResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error) {
	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)

	balance, err := i.Transaction(token, mode, TransactionStore{
		TransactionId: uuid.NewV4().String(),
		Token:         token,
		Mode:          mode,
		Category:      CategoryClose,
		RoundStatus:   RoundStatusClose,
		PlayerId:      playerId,
		GameId:        gameId,
		RoundId:       roundId,
		Amount: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   0,
		},
		ParentTransactionId: "",
		TxTime:              time.Now(),
		GameState:           gamestate,
	})

	if err != nil {
		return BalanceStore{}, err
	}
	// HACK to keep token from updating as client currently cannot handle token update on clientstate save call
	i.setToken(token, playerId)
	balance.Token = token
	return balance, nil
}

func (i *RemoteServiceImpl) CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error) {
	closeRound := true

	txRq := restTransactionRequest{
		ReqId:       uuid.NewV4().String(),
		Token:       string(token),
		Game:        gameId,
		Platform:    i.defaultPlatform,
		Mode:        strings.ToLower(string(mode)),
		Session:     roundId,
		BonusAmount: 0,
		JpAmount:    0,
		Category:    string(CategoryClose),
		CloseRound:  closeRound,
		Round:       roundId,
		TxRef:       roundId,
		GameState:   base64.StdEncoding.EncodeToString(gamestate),
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(txRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	req := i.request(ApiTypeTransaction, b)
	client := i.httpClient()
	resp, err := client.Do(req)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	txResp := i.restTransactionResponse(resp)

	finalErr = i.errorResponseCode(txResp.ResponseCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	return BalanceStore{
		PlayerId: txResp.PlayerId,
		Token:    Token(txResp.Token),
		Balance: engine.Money{
			Currency: txResp.Currency,
			Amount:   engine.Fixed(txResp.Balance * 10000),
		},
	}, nil
}

func (i *LocalServiceImpl) setToken(token Token, playerId string) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Token[token] = playerId
}

func (i *LocalServiceImpl) deleteToken(token Token) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()
	_, ok := ld.Token[token]
	// don't delete the token if it matches the player id
	if ok && string(token) != ld.Token[token] {
		delete(ld.Token, token)
	}
}

func (i *LocalServiceImpl) getToken(token Token) (string, bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	playerId, ok := ld.Token[token]

	return playerId, ok
}

func (i *LocalServiceImpl) setPlayer(playerId string, player PlayerStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Player[playerId] = player
}

func (i *LocalServiceImpl) getPlayer(playerId string) (PlayerStore, bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	player, ok := ld.Player[playerId]

	return player, ok
}

func (i *LocalServiceImpl) setTransaction(transactionId string, transaction TransactionStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Transaction[transactionId] = transaction
}

func (i *LocalServiceImpl) getTransaction(transactionId string) (TransactionStore, bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	tx, ok := ld.Transaction[transactionId]

	return tx, ok
}

func (i *LocalServiceImpl) setTransactionByPlayerGame(key string, transaction TransactionStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.TransactionByPlayerGame[key] = transaction
}

func (i *LocalServiceImpl) getTransactionByPlayerGame(key string) (TransactionStore, bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()
	tx, ok := ld.TransactionByPlayerGame[key]

	return tx, ok
}

func internalInit(c *config.Config) {
	logger.Infof("internal-init [DevMode: %v]", c.DevMode)

	if c.DevMode {
		if ld == nil {
			ld = new(LocalData)
			ld.Token = make(map[Token]string)
			ld.Player = make(map[string]PlayerStore)
			ld.Transaction = make(map[string]TransactionStore)
			ld.TransactionByPlayerGame = make(map[string]TransactionStore)
		}
	} else {
		remoteServiceImplHttpClient = &promhttp.Client{
			Client:     http.DefaultClient,
			Registerer: prometheus.DefaultRegisterer,
		}
	}
}

func internalCheck() error {
	if ld == nil {
		logger.Errorf("Local data is not initialized. Panic!!")
		return errors.New("Local data is not initalized")
	}

	if ld.Player == nil {
		logger.Errorf("Local data is not initialized. Panic!!")
		return errors.New("Local data is not initalized")
	}

	return nil
}

func New(c *config.Config) Service {
	internalInit(c)

	if c.DevMode {
		return &LocalServiceImpl{}
	}

	return &RemoteServiceImpl{
		serverUrl:       c.DashurConfig.StoreRemoteUrl,
		appId:           c.DashurConfig.StoreAppId,
		appCredential:   c.DashurConfig.StoreAppPass,
		defaultLanguage: c.DefaultLanguage,
		defaultPlatform: c.DefaultPlatform,
		demoTokenPrefix: c.DemoTokenPrefix,
		demoCurrency:    c.DemoCurrency,
	}
}

func NewLocal() LocalService {
	internalInit(&config.Config{
		DevMode: true,
	})

	return &LocalServiceImpl{}
}
