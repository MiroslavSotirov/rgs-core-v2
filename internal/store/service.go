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
		FreeGames           FreeGamesStore
	}

	GameStateStore struct {
		GameState []byte
		WalletInternalStatus int
	}

	BalanceStore struct {
		PlayerId string
		Token    Token
		Mode     Mode
		Balance  engine.Money
		FreeGames           FreeGamesStore
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
		BetLimitSettingCode string
		GameState           []byte
		FreeGames           FreeGamesStore
		WalletStatus        int
	}

	FreeGamesStore struct {
		NoOfFreeSpins int    `json:"count"`
		CampaignRef   string `json:"ref"`
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
		logAccount		string
	}

	// local service eq implemenation of service. so that unit test of services can be easily mocked.
	LocalService interface {
		PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, *Error)
		PlayerSave(token Token, mode Mode, player PlayerStore) (PlayerStore, *Error)
		BalanceByToken(token Token, mode Mode) (BalanceStore, *Error)
		Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, *Error)
		TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error)
		CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error)
		GamestateById(gamestateId string) (GameStateStore, *Error)
	}

	LocalServiceImpl struct{}

	restMetadata struct {
		ReqId          string `json:"req_id"`
		ProcessingTime int    `json:"processing_time"`
		VendorInfo      restVendorResponse `json:"vendor"`
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
		Metadata        restMetadata      `json:"metadata"`
		Token           string            `json:"token"`
		ResponseCode    string            `json:"code"`
		Message         string            `json:"message"`
		Id              string            `json:"id"`
		Username        string            `json:"username"`
		BetLimit        string            `json:"bet_limit"`
		FreeGames       restFreeGame      `json:"free_games"`
		Balance         int64             `json:"balance"`
		Currency        string            `json:"currency"`
		//LastGameState   string            `json:"last_game_state"`
		PlayerMessage   restPlayerMessage `json:"player_message"`
		Urls            map[string]string `json:"urls"`
		//LastGameStatus  int               `json:"last_game_status"`

	}
	restVendorResponse struct {
		LastAttemptedTx restTransactionRequest `json:"last_attempted_tx"`
		LastTx          restTransactionRequest `json:"last_tx"`
	}
	//restTxDetailResponse struct {
	//	//Complete bool `json:"complete"`//: true,
	//	//CompleteOK bool `json:"complete_ok"`//: true,
	//	//"req_id": "943bb0cb-7fb3-43e4-9a4c-fff51c82f2c8",
	//	Token string `json:"token"`
	//	//"game": "the-year-of-zhu",
	//	//"platform": "html5",
	//	//"mode": "real",
	//	//"session": "WAp15vFz",
	//	Currency string `json:"currency"`
	//	Amount int64 `json:"amount"`
	//	//"bonus_amount": 0,
	//	//"jp_amount": 0,
	//	Category string `json:"category"`
	//	//"campaign_ref": "",
	//	//"close_round": true,
	//	GameState string `json:"game_state"`
	//	GameRound string `json:"round"`
	//	TxRef string `json:"tx_ref"`
	//	//"description": "",
	//	InternalStatus int `json:"internal_status"`
	//}

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
		FreeGames     restFreeGame      `json:"free_games"`
	}

	restGameStateRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restGameStateResponse struct {
		Metadata       restMetadata `json:"metadata"`
		Token          string       `json:"token"`
		ResponseCode   string       `json:"code"`
		Message        string       `json:"message"`
		GameState      string       `json:"game_state"`
		InternalStatus int          `json:"internal_status"`
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
		InternalStatus int `json:"internal_status"`
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
		FreeGames     restFreeGame      `json:"free_games"`
	}

	restQueryRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restQueryResponse struct {
		Metadata       restMetadata `json:"metadata"`
		//Token          string       `json:"token"`
		ResponseCode   string       `json:"code"`
		Message        string       `json:"message"`
		ReqId          string       `json:"req_id"`
		//Game           string       `json:"game"`
		//Platform       string       `json:"platform"`
		//Mode           string       `json:"mode"`
		//Session        string       `json:"session"`
		//Currency       string       `json:"currency"`
		//Amount         int64        `json:"amount"`
		//BonusAmount    int64        `json:"bonus_amount"`
		//JpAmount       int64        `json:"jp_amount"`
		//Category       string       `json:"category"`
		//CampaignRef    string       `json:"campaign_ref"`
		//CloseRound     bool         `json:"close_round"`
		//GameState      string       `json:"game_state"`
		//Round          string       `json:"round"`
		//TxRef          string       `json:"tx_ref"`
		//Description    string       `json:"description"`
		BetLimit       string       `json:"bet_limit"`
		FreeGames      restFreeGame `json:"free_games"`
		//InternalStatus int          `json:"internal_status"`
		LastTx         restTransactionRequest `json:"last_tx"`
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
					FreeGames:       FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef},
					BetLimitSettingCode: player.BetLimitSettingCode,
				},
				GameStateStore{GameState: tx.GameState, WalletInternalStatus: 1},
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
					FreeGames:       FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef},
					BetLimitSettingCode: player.BetLimitSettingCode,
				},
				GameStateStore{},
				nil
		}
	} else {
		logger.Warnf("NO PLAYER EXISTS")
		return PlayerStore{}, GameStateStore{}, nil
	}

}

func (i *RemoteServiceImpl) request(apiType ApiType, body io.Reader) (resp *http.Response, err error) {
	req, _ := http.NewRequest("POST", i.serverUrl+"/"+string(apiType), body)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(i.appId, i.appCredential)
	start := time.Now()
	client := i.httpClient()
	resp, err = client.Do(req)
	logger.Debugf("%v request took %v", apiType, time.Now().Sub(start).String())
	return
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
	start := time.Now()
	resp, err := i.request(ApiTypeAuth, b)

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
	finalErr = i.errorResponseCode(authResp.ResponseCode)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}

	var lastTransaction restTransactionRequest
	var balance BalanceStore
	lastTransaction, balance, finalErr = i.getLastGamestate(authResp.Metadata.VendorInfo.LastTx, authResp.Metadata.VendorInfo.LastAttemptedTx)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}
	if balance.Token == "" {
		balance.Token = Token(authResp.Token)
		balance.FreeGames = FreeGamesStore{authResp.FreeGames.NrGames, authResp.FreeGames.CampaignRef}
		balance.Balance = engine.Money{
			Currency: authResp.Currency,
			Amount:   engine.Fixed(authResp.Balance * 10000),
		}
	}
	//lastTransaction.InternalStatus = authResp.Metadata.VendorInfo.LastAttemptedTx.InternalStatus
	if len(lastTransaction.GameState) > 0 {
		gameState, err = base64.StdEncoding.DecodeString(lastTransaction.GameState)

		finalErr = i.errorBase64(err)
		if finalErr != nil {
			return PlayerStore{}, GameStateStore{}, finalErr
		}
	}
	if authResp.Id == i.logAccount {
		logger.Infof("%v request took %v for account %v", ApiTypeAuth, time.Now().Sub(start).String(), authResp.Id)
	}

	return PlayerStore{
			PlayerId: authResp.Id,
			Token:    balance.Token,
			Mode:     mode,
			Username: authResp.Username,
			Balance:  balance.Balance,
			FreeGames:       balance.FreeGames,
			BetLimitSettingCode: authResp.BetLimit,
		},
		GameStateStore{GameState: gameState, WalletInternalStatus: lastTransaction.InternalStatus},
		nil
}

func (i *RemoteServiceImpl) getLastGamestate(lastTx restTransactionRequest, lastAttemptedTx restTransactionRequest) (lastTxInfo restTransactionRequest, balance BalanceStore, err *Error) {
	// if last_tx and last_attempted_tx are the same, doesn't matter
	logger.Debugf("Determining last GS : Last Tx = %#v, Last Attempted Tx = %#v", lastTx, lastAttemptedTx)
	if lastAttemptedTx.TxRef == lastTx.TxRef {
		lastTxInfo = lastTx
		return
	}

	// otherwise, check which tx it was that failed
	// we assume there is never more than one pending/failed tx on top of the last successful tx
	// because in freespins the PAYOUT tx is treated the same way as the WAGER in normal play, we need to check if this is the first tx of a gamestate
	if len(lastAttemptedTx.GameState) == 0 {
		err = &Error{ErrorCodeGeneralError, "No gamestate attached to previous tx"}
		return
	}
	gameState, errDecode := base64.StdEncoding.DecodeString(lastAttemptedTx.GameState)

	err = i.errorBase64(errDecode)
	if err != nil {
		return
	}
	gsDeserialized := DeserializeGamestateFromBytes(gameState)

	if lastAttemptedTx.TxRef == gsDeserialized.Transactions[0].Id {
		// this failed tx was the first tx of the round, so we return the previous successful gamestate which should be the gamestate attached to the previous round
		lastTxInfo = lastTx
		// set internalstatus to -1 so that gamestate calculated on top of this previous gamestate gets a suffix
		lastTxInfo.InternalStatus = -1
		return
	}

	// the first tx has been processed, so we should try to close out the round
	switch Category(lastAttemptedTx.Category) {
	case CategoryWager:
		// we do not expect this case, a new wager should be associated with a new gamestate
		err = &Error{ErrorCodeGeneralError, "Got WAGER as non-primary tx of gamestate"}
		return
	case CategoryPayout:
		// the wager was successful but payout failed, so retry the payout one time with same ID
		balance, err = i.txSend(lastAttemptedTx)
		// if no error, return this transaction and the updated balance info
		if err == nil {
			lastTxInfo = lastAttemptedTx
			lastTxInfo.InternalStatus = 1
			return
		}
		logger.Warnf("Failed PAYOUT TX retry, round %v should be settled manually before play can continue", gsDeserialized.RoundID)
		//if the error persists, return an error, payout issue needs to be solved manually
		return
	case CategoryClose:
		// endround tx was attempted and failed. try to settle this but if the error persists, allow gameplay to continue
		balance, err = i.txSend(lastAttemptedTx)
		if err != nil {
			logger.Warnf("Failed ENDROUND TX retry, continuing with game, round %v should be settled manually", gsDeserialized.RoundID)
		}
		lastTxInfo = lastAttemptedTx
		lastTxInfo.InternalStatus = 1
		return
	}
	err = &Error{ErrorCodeGeneralError, "Unrecognized previous transaction configuration"}
	return
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
		FreeGames:       FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef},
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
		FreeGames:       FreeGamesStore{0, ""},
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
	start := time.Now()
	resp, err := i.request(ApiTypeBalance, b)

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
	if balResp.PlayerId == i.logAccount {
		logger.Infof("%v request took %v for account %v", ApiTypeBalance, time.Now().Sub(start).String(), balResp.PlayerId)
	}
	return BalanceStore{
		PlayerId: balResp.PlayerId,
		Token:    Token(balResp.Token),
		Balance: engine.Money{
			Currency: balResp.Currency,
			Amount:   engine.Fixed(balResp.Balance * 10000),
		},
		FreeGames:       FreeGamesStore{balResp.FreeGames.NrGames, balResp.FreeGames.CampaignRef},
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
		// process free game
		if transaction.FreeGames.CampaignRef != "" && player.FreeGames.CampaignRef == transaction.FreeGames.CampaignRef && player.FreeGames.NoOfFreeSpins > 0{
			player.FreeGames.NoOfFreeSpins -= 1
			logger.Warnf("DEBITING FREE SPIN")
		} else {
			player.Balance.Amount = player.Balance.Amount - transaction.Amount.Amount
		}
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
		FreeGames:       FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef},
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
		CampaignRef: transaction.FreeGames.CampaignRef,
	}

	return i.txSend(txRq)
}

func (i *RemoteServiceImpl) txSend(txRq restTransactionRequest) (BalanceStore, *Error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(txRq)

	finalErr := i.errorJson(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}
	start := time.Now()
	resp, err := i.request(ApiTypeTransaction, b)
	finalErr = i.errorRest(err)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	finalErr = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	txResp := i.restTransactionResponse(resp)
	logger.Debugf("TX RESP: %#v", txResp)

	finalErr = i.errorResponseCode(txResp.ResponseCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}
	if txResp.PlayerId == i.logAccount {
		logger.Infof("%v request took %v for account %v", ApiTypeTransaction, time.Now().Sub(start).String(), txResp.PlayerId)
	}
	return BalanceStore{
		PlayerId: txResp.PlayerId,
		Token:    Token(txResp.Token),
		Balance: engine.Money{
			Currency: txResp.Currency,
			Amount:   engine.Fixed(txResp.Balance * 10000),
		},
		FreeGames:       FreeGamesStore{txResp.FreeGames.NrGames, txResp.FreeGames.CampaignRef},
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

	return GameStateStore{GameState: transaction.GameState}, nil
}

func (i *LocalServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error) {
	// Used at beginning of play() func to get previous gamestate, betlimit settings code, and free games info
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
		BetLimitSettingCode: player.BetLimitSettingCode,
		FreeGames:      player.FreeGames,
		WalletStatus: 1,
	}, nil
}

func (i *RemoteServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, *Error) {
	// Used at beginning of play() func to get previous gamestate, betlimit settings code, and free games info

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
	//start := time.Now()
	resp, err := i.request(ApiTypeQuery, b)

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
	var lastTx restTransactionRequest
	var balance BalanceStore
	lastTx, balance, finalErr = i.getLastGamestate(queryResp.LastTx, queryResp.Metadata.VendorInfo.LastAttemptedTx)

	if balance.Token == "" {
		balance.Token = Token(lastTx.Token)
		balance.FreeGames = FreeGamesStore{queryResp.FreeGames.NrGames, queryResp.FreeGames.CampaignRef}
		balance.Balance = engine.Money{
			Currency: lastTx.Currency,
			Amount:   engine.Fixed(lastTx.Amount * 10000),
		}
	}
	roundStatus := RoundStatusOpen

	if lastTx.CloseRound {
		roundStatus = RoundStatusClose
	}

	if len(lastTx.GameState) > 0 {
		gameState, err = base64.StdEncoding.DecodeString(lastTx.GameState)

		finalErr = i.errorBase64(err)
		if finalErr != nil {
			return TransactionStore{}, finalErr
		}
	}
	//if queryResp.PlayerId == i.logAccount {
	//	logger.Infof("%v request took %v for account %v", ApiTypeBalance, time.Now().Sub(start).String(), balResp.PlayerId)
	//}
	return TransactionStore{
		TransactionId: lastTx.TxRef,
		Token:               balance.Token, // the token returned in the queryResp is the token used to make the tx call, not a new token
		Mode:                mode,
		Category:            Category(lastTx.Category),
		RoundStatus:         roundStatus,
		PlayerId:            "", //TODO: fix this
		GameId:              lastTx.Game,
		RoundId:             lastTx.Round,
		Amount:              balance.Balance,
		ParentTransactionId: "",         //TODO: fix this
		TxTime:              time.Now(), //TODO: fix this
		GameState:           gameState,
		BetLimitSettingCode: queryResp.BetLimit,
		FreeGames:           balance.FreeGames,
		WalletStatus:        lastTx.InternalStatus,
	}, nil
}


func (i *RemoteServiceImpl) restQueryResponse(response *http.Response) restQueryResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restQueryResponse
	json.Unmarshal(body, &data)
	return data
}


func (i *RemoteServiceImpl) restGameStateResponse(response *http.Response) restGameStateResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var data restGameStateResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) CloseRound(token Token, mode Mode, gameId string, roundId string, gamestate []byte) (BalanceStore, *Error) {
	// Used in clientstate call
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
		FreeGames:     player.FreeGames,
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
	// Used in clientstate call
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
	start := time.Now()
	resp, err := i.request(ApiTypeTransaction, b)

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
	if txResp.PlayerId == i.logAccount {
			logger.Infof("%v request took %v for account %v", ApiTypeTransaction, time.Now().Sub(start).String(), txResp.PlayerId)
	}
	return BalanceStore{
		PlayerId: txResp.PlayerId,
		Token:    Token(txResp.Token),
		Balance: engine.Money{
			Currency: txResp.Currency,
			Amount:   engine.Fixed(txResp.Balance * 10000),
		},
		FreeGames: FreeGamesStore{txResp.FreeGames.NrGames, txResp.FreeGames.CampaignRef},
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
		logAccount:		 c.LogAccount,
	}
}

func NewLocal() LocalService {
	internalInit(&config.Config{
		DevMode: true,
	})

	return &LocalServiceImpl{}
}
