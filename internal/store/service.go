package store

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/travelaudience/go-promhttp"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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
	ApiTypeFeed        ApiType = "feed"

	ModeDemo Mode = "DEMO"
	ModeReal Mode = "REAL"

	RoundStatusOpen  RoundStatus = "OPEN"
	RoundStatusClose RoundStatus = "CLOSE"
	//
	//ErrorCodeGeneralError     ErrorCode = "ERR-001" // general error.
	//ErrorCodeNotEnoughBalance ErrorCode = "ERR-002"
	//ErrorCodeTokenExpired     ErrorCode = "ERR-003"
	//ErrorCodeEntityNotFound   ErrorCode = "ERR-004"

	ResponseCodeOk                 ResponseCode = "0"
	ResponseCodeInsufficentBalance ResponseCode = "1"
	ResponseCodeDataError          ResponseCode = "2"
	ResponseCodeSessionExpired     ResponseCode = "3"
	ResponseCodeUnknownError                    = "4"
	ResponseCodeCustomError                     = "6"

	ErrorCodeCustomError            = "CUSTOM_ERROR"
	ErrorCodeSpendingBudgetExceeded = "SPENDING_BUDGET_EXCEEDED"
	ErrorCodeBlockedFromProduct     = "BLOCKED_FROM_PRODUCT"
	ErrorCodeIpBlocked              = "IP_BLOCKED"
	ErrorCodeMonthlyTimeLimit       = "MONTHLY_TIME_LIMIT"
	ErrorCodeWeeklyTimeLimit        = "WEEKLY_TIME_LIMIT"
	ErrorCodeDailyTimeLimit         = "DAILY_TIME_LIMIT"
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
		CompanyId           string
		BetSettingId        string
		FreeGames           FreeGamesStore
	}

	GameStateStore struct {
		GameState            []byte
		WalletInternalStatus int
	}

	BalanceStore struct {
		PlayerId  string
		Message   string
		Token     Token
		Mode      Mode
		Balance   engine.Money
		FreeGames FreeGamesStore
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
		CompanyId           string
		BetSettingId        string
		GameState           []byte
		FreeGames           FreeGamesStore
		WalletStatus        int
		Ttl                 int64
		History             TransactionHistory
	}

	TransactionHistory struct {
		NumWager        int
		SumWager        float64
		NumPayout       int
		SumPayout       float64
		NumRefund       int
		SumRefundCredit float64
		SumRefundDebit  float64
	}

	FreeGamesStore struct {
		NoOfFreeSpins int          `json:"count"`
		CampaignRef   string       `json:"ref"`
		TotalWagerAmt engine.Fixed `json:"wager_amount"`
	}

	FeedRound struct {
		Id              int64             `json:"id"`
		CurrencyUnit    string            `json:"currency_unit"`
		ExternalRef     string            `json:"external_ref"`
		Status          string            `json:"status"`
		TransactionIds  []int64           `json:"transaction_ids"`
		NumWager        int               `json:"num_of_wager"`
		SumWager        float64           `json:"sum_of_wager"`
		NumPayout       int               `json:"num_of_payout"`
		SumPayout       float64           `json:"sum_of_payout"`
		NumRefund       int               `json:"num_of_refund"`
		SumRefundCredit float64           `json:"sum_of_refund_credit"`
		SumRefundDebit  float64           `json:"sum_of_refund_debit"`
		StartTime       string            `json:"start_time"`
		CloseTime       string            `json:"close_time"`
		Metadata        FeedRoundMetadata `json:"meta_data"`
	}

	FeedRoundMetadata struct {
		RoundId   string              `json:"round_id"`
		ExtItemId string              `json:"ext_item_id"`
		ItemId    int64               `json:"item_id"`
		Vendor    FeedRoundVendordata `json:"vendor"`
	}

	FeedRoundVendordata struct {
		State   engine.Gamestate    `json:"state"`
		StateV3 engine.IGameStateV3 `json:"statev3,omitempty"`
	}

	FeedTransaction struct {
		Id           int64             `json:"id"`
		Category     string            `json:"category"`
		ExternalRef  string            `json:"external_ref"`
		CurrencyUnit string            `json:"currency_unit"`
		Amount       float64           `json:"amount"`
		Metadata     FeedRoundMetadata `json:"meta_data"`
		TxTime       string            `json:"transaction_time"`
	}

	//Error struct {
	//	Code    ErrorCode
	//	Message string
	//}

	LocalData struct {
		Token                   map[Token]gcstring
		Player                  map[string]gcPlayerStore
		Transaction             map[string]gcTransactionStore
		TransactionByPlayerGame map[string]gcTransactionStore
		Message                 map[string]gcstring
		Lock                    sync.RWMutex
	}

	Service interface {
		// authenticate token, given the game id, it will also retrieve the latest gamestate from latest transaction.
		PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, rgse.RGSErr)

		// retrieve latest balance by token
		BalanceByToken(token Token, mode Mode) (BalanceStore, rgse.RGSErr)

		// create transaction.
		Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, rgse.RGSErr)

		// request multiple transactions
		MultiTransaction(token Token, mode Mode, transaction []TransactionStore) (BalanceStore, rgse.RGSErr)

		// retrieve latest transcation by player and by game id
		TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, rgse.RGSErr)

		// close round.
		CloseRound(token Token, mode Mode, gameId string, roundId string, campaignRef string, gamestate []byte, ttl int64, history *TransactionHistory) (BalanceStore, rgse.RGSErr)

		//// gamestate by id
		//GamestateById(gamestateId string) (GameStateStore, *Error)

		// retrieve transaction feed
		Feed(token Token, mode Mode, gameId string, startTime string, endTime string, pageSize int, page int) ([]FeedRound, int, rgse.RGSErr)
		FeedRound(token Token, mode Mode, gameId string, roundId int64) ([]FeedTransaction, rgse.RGSErr)
		//		FeedV3(token Token, mode Mode, gameId string, startTime string, endTime string, pageSize int, page int) ([]FeedRound, int, rgse.RGSErr)
		//		FeedRoundV3(token Token, mode Mode, gameId string, roundId int64) ([]FeedTransaction, rgse.RGSErr)
	}

	RemoteServiceImpl struct {
		serverUrl       string
		appId           string
		appCredential   string
		defaultPlatform string
		defaultLanguage string
		demoTokenPrefix string
		demoCurrency    string
		logAccount      string
		dataLimit       int
		maxRetries      int
		timeoutMs       int64
	}

	// local service eq implemenation of service. so that unit test of services can be easily mocked.
	LocalService interface {
		PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, rgse.RGSErr)
		PlayerSave(token Token, mode Mode, player PlayerStore) (PlayerStore, rgse.RGSErr)
		BalanceByToken(token Token, mode Mode) (BalanceStore, rgse.RGSErr)
		Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, rgse.RGSErr)
		MultiTransaction(token Token, mode Mode, transaction []TransactionStore) (BalanceStore, rgse.RGSErr)
		TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, rgse.RGSErr)
		CloseRound(token Token, mode Mode, gameId string, roundId string, campaignRef string, gamestate []byte, ttl int64, history *TransactionHistory) (BalanceStore, rgse.RGSErr)
		GamestateById(gamestateId string) (GameStateStore, rgse.RGSErr)
		SetMessage(playerId string, message string) rgse.RGSErr
		SetBalance(token Token, amount engine.Money) rgse.RGSErr
		Feed(token Token, mode Mode, gameId string, startTime string, endTime string, pageSize int, page int) ([]FeedRound, int, rgse.RGSErr)
		FeedRound(token Token, mode Mode, gameId string, roundId int64) ([]FeedTransaction, rgse.RGSErr)
		//		FeedV3(token Token, mode Mode, gameId string, startTime string, endTime string, pageSize int, page int) ([]FeedRound, int, rgse.RGSErr)
		//		FeedRoundV3(token Token, mode Mode, gameId string, roundId int64) ([]FeedTransaction, rgse.RGSErr)
	}

	LocalServiceImpl struct{}

	restMetadata struct {
		ReqId          string             `json:"req_id"`
		ProcessingTime int                `json:"processing_time"`
		VendorInfo     restVendorResponse `json:"vendor"`
	}

	restErrorResponse struct {
		ResponseCode string `json:"code"`
		Message      string `json:"message"`
		ErrorCode    string `json:"err_code"`
		ErrorDesc    string `json:"err_desc"`
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
		WagerAmt    int64  `json:"wager_amount"`
	}

	restPlayerMessage struct {
		Title    string `json:"title"`
		Link     string `json:"link"`
		Message  string `json:"message"`
		Location int    `json:"location"`
	}

	restAuthenticateResponse struct {
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		ResponseCode string       `json:"code"`
		//		Message      string       `json:"message"`
		Id           string       `json:"id"`
		Username     string       `json:"username"`
		BetLimit     string       `json:"bet_limit"`
		CompanyId    int64        `json:"company_id"`
		BetSettingId string       `json:"bet_setting_id"`
		FreeGames    restFreeGame `json:"free_games"`
		Balance      int64        `json:"balance"`
		Currency     string       `json:"currency"`
		//LastGameState   string            `json:"last_game_state"`
		PlayerMessage restPlayerMessage `json:"player_message"`
		Urls          map[string]string `json:"urls"`
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
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		ResponseCode  string            `json:"code"`
		//		Message       string            `json:"message"`
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
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		ResponseCode   string       `json:"code"`
		//		Message        string       `json:"message"`
		GameState      string `json:"game_state"`
		InternalStatus int    `json:"internal_status"`
	}

	restVersionRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		Mode     string `json:"mode"`
	}

	restVersionResponse struct {
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		ResponseCode string       `json:"code"`
		//		Message      string       `json:"message"`
		Version string `json:"version"`
	}

	restTransactionRequest struct {
		ReqId          string `json:"req_id"`
		Token          string `json:"token"`
		Game           string `json:"game"`
		Platform       string `json:"platform"`
		Mode           string `json:"mode"`
		Session        string `json:"session"`
		Currency       string `json:"currency"`
		Round          string `json:"round"`
		Description    string `json:"description"`
		InternalStatus int    `json:"internal_status"`
		Ttl            int64  `json:"ttl"`
		TtlStamp       int64  `json:"ttlstamp"`
		restTransactionDesc
	}

	restMultiTransactionRequest struct {
		ReqId          string                `json:"req_id"`
		Token          string                `json:"token"`
		Game           string                `json:"game"`
		Platform       string                `json:"platform"`
		Mode           string                `json:"mode"`
		Session        string                `json:"session"`
		Currency       string                `json:"currency"`
		Round          string                `json:"round"`
		Description    string                `json:"description"`
		InternalStatus int                   `json:"internal_status"`
		Ttl            int64                 `json:"ttl"`
		TtlStamp       int64                 `json:"ttlstamp"`
		MultiTxes      []restTransactionDesc `json:"multi_txes,omitempty"`
	}

	restTransactionDesc struct {
		Amount      int64  `json:"amount"`
		BonusAmount int64  `json:"bonus_amount"`
		JpAmount    int64  `json:"jp_amount"`
		Category    string `json:"category"`
		CampaignRef string `json:"campaign_ref"`
		CloseRound  bool   `json:"close_round"`
		TxRef       string `json:"tx_ref"`
		GameState   string `json:"game_state"`
	}

	restTransactionResponse struct {
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		ResponseCode  string            `json:"code"`
		//		Message       string            `json:"message"`
		BetLimit      string            `json:"bet_limit"`
		CompanyId     int64             `json:"company_id"`
		BetSettingId  string            `json:"bet_setting_id"`
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
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		//Token          string       `json:"token"`
		//		ResponseCode string `json:"code"`
		//		Message      string `json:"message"`
		ReqId string `json:"req_id"`
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
		BetLimit     string       `json:"bet_limit"`
		CompanyId    int64        `json:"company_id"`
		BetSettingId string       `json:"bet_setting_id"`
		FreeGames    restFreeGame `json:"free_games"`
		//InternalStatus int          `json:"internal_status"`
		LastTx restTransactionRequest `json:"last_tx"`
	}

	RestFeedRequest struct {
		ReqId     string `json:"req_id"`
		Token     string `json:"token"`
		Game      string `json:"game"`
		Platform  string `json:"platform"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		PageSize  int    `json:"page_size"`
		Page      int    `json:"page"`
	}

	RestFeedRoundRequest struct {
		ReqId    string `json:"req_id"`
		Token    string `json:"token"`
		Game     string `json:"game"`
		Platform string `json:"platform"`
		RoundId  int64  `json:"round_id"`
	}

	RestFeedResponse struct {
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		Code     string          `json:"code"`
		Rounds   []RestRounddata `json:"rounds"`
		NextPage int             `json:"next_page"`
	}

	RestFeedRoundResponse struct {
		Metadata restMetadata `json:"metadata"`
		restErrorResponse
		Token string `json:"token"`
		//		Code     string                `json:"code"`
		Feeds []RestTransactiondata `json:"feeds"`
	}

	RestRoundVendordata struct {
		State string `json:"state"`
	}

	RestRoundMetadata struct {
		RoundId   string              `json:"round_id"`
		ExtItemId string              `json:"ext_item_id"`
		ItemId    int64               `json:"item_id"`
		Vendor    RestRoundVendordata `json:"vendor"`
	}

	RestRounddata struct {
		Id              int64             `json:"id"`
		CurrencyUnit    string            `json:"currency_unit"`
		ExternalRef     string            `json:"external_ref"`
		Status          string            `json:"status"`
		TransactionIds  []int64           `json:"transaction_ids"`
		NumWager        int               `json:"num_of_wager"`
		SumWager        float64           `json:"sum_of_wager"`
		NumPayout       int               `json:"num_of_payout"`
		SumPayout       float64           `json:"sum_of_payout"`
		NumRefund       int               `json:"num_of_refund"`
		SumRefundCredit float64           `json:"sum_of_refund_credit"`
		SumRefundDebit  float64           `json:"sum_of_refund_debit"`
		StartTime       string            `json:"start_time"`
		CloseTime       string            `json:"close_time"`
		Metadata        RestRoundMetadata `json:"meta_data"`
	}

	RestTransactiondata struct {
		Id           int64             `json:"id"`
		Category     string            `json:"category"`
		ExternalRef  string            `json:"external_ref"`
		CurrencyUnit string            `json:"currency_unit"`
		Amount       float64           `json:"amount"`
		Metadata     RestRoundMetadata `json:"meta_data"`
		TxTime       string            `json:"transaction_time"`
	}
)

//
//// Error functions to allow them to be implemented as RGSError interface
//func (e *Error) Error() string {
//	if e.Message == "" {
//		return fmt.Sprintf("Error %d, Service error", e.Code)
//	}
//	return fmt.Sprintf("Error %d, Service error - %s", e.Code, e.Message)
//}
//
//func (e *Error) Init(code int, msgs ...string) {
//	e.Code = ErrorCode(fmt.Sprintf("ERR-%03d", code))
//	e.Message = strings.Join(msgs, " ")
//}
//
//func (e *Error) AppendErrorText(text string) {
//	e.Message = e.Message + text
//}
//
//func (e *Error) SetErrorTextByCode(code int) {
//	e.Message = fmt.Sprintf("Error %d, Service error", e.Code)
//}

//
//// UnmarshalJSON implements the json.Unmarshaler interface, which
//// allows us to ingest values of any json type as a string and run our custom conversion
//
//func (t *Token) UnmarshalJSON(b []byte) error {
//	var s string
//	if err := json.Unmarshal(b, &s); err != nil {
//		return err
//	}
//	*t = Token(s)
//	return nil
//}
//
//func (t Token) MarshalJSON() ([]byte, error) {
//	s := string(t)
//	return json.Marshal(&s)
//}

var ld *LocalData
var remoteServiceImplHttpClient *promhttp.Client

func (i *LocalServiceImpl) PlayerByToken(token Token, mode Mode, gameId string) (player PlayerStore, gs GameStateStore, err rgse.RGSErr) {
	logger.Debugf("LocalServiceImpl.PlayerByToken([%v], [%v])", token, mode)

	err = internalCheck()
	if err != nil {
		return
	}

	storePlayerId, playerIdExists := i.getToken(token)
	logger.Debugf("player id: %v", storePlayerId)
	if playerIdExists {
		player, _ = i.getPlayer(storePlayerId)
		newToken := i.renewToken(token)
		key := player.PlayerId + "::" + gameId
		tx, txExists := i.getTransactionByPlayerGame(key)

		if txExists && tx.GameState != nil && len(tx.GameState) > 0 {
			player = PlayerStore{
				PlayerId: player.PlayerId,
				Token:    newToken,
				Mode:     player.Mode,
				Username: player.Username,
				Balance: engine.Money{
					Currency: player.Balance.Currency,
					Amount:   player.Balance.Amount,
				},
				FreeGames:           FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef, player.FreeGames.TotalWagerAmt},
				BetLimitSettingCode: player.BetLimitSettingCode,
				CompanyId:           player.CompanyId,
				BetSettingId:        player.BetSettingId,
			}
			gs = GameStateStore{GameState: tx.GameState, WalletInternalStatus: 1}
			return
		} else {
			// if in V1 api, this is likely an error, if player exists, there should be a previous gameplay unless init was called and never spun, which will throw an error
			logger.Warnf("DEMO WALLET PLAYER EXISTS BUT NO PREVIOUS TX")
			player = PlayerStore{
				PlayerId: player.PlayerId,
				Token:    newToken,
				Mode:     player.Mode,
				Username: player.Username,
				Balance: engine.Money{
					Currency: player.Balance.Currency,
					Amount:   player.Balance.Amount,
				},
				FreeGames:           FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef, player.FreeGames.TotalWagerAmt},
				BetLimitSettingCode: player.BetLimitSettingCode,
				CompanyId:           player.CompanyId,
				BetSettingId:        player.BetSettingId,
			}
			return
		}
	} else {
		err = rgse.Create(rgse.NoSuchPlayer)
		logger.Warnf("NO PLAYER EXISTS")
		return
	}

}

func (i *RemoteServiceImpl) request(apiType ApiType, body *bytes.Buffer) (resp *http.Response, err error) {
	logger.Debugf("%s request: %s", apiType, body.String())
	req, _ := http.NewRequest("POST", i.serverUrl+"/"+string(apiType), body)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(i.appId, i.appCredential)
	start := time.Now()
	client := i.httpClient()
	resp, err = client.Do(req)
	logger.Debugf("%v request to %s took %v", apiType, req.URL, time.Now().Sub(start).String())
	return
}

func (i *RemoteServiceImpl) httpClient() http.Client {
	dashurClient, _ := remoteServiceImplHttpClient.ForRecipient("dashur")
	return *dashurClient
}

func (i *RemoteServiceImpl) demoToken() string {
	return i.demoTokenPrefix + ":" + i.demoCurrency + ":" + rng.Uuid()
}

func (i *RemoteServiceImpl) errorJson(err error) rgse.RGSErr {
	if err != nil {
		return rgse.Create(rgse.JsonError)
	}
	return nil
}

func (i *RemoteServiceImpl) errorRest(err error) rgse.RGSErr {
	if err != nil {
		return rgse.Create(rgse.RestError)
	}
	return nil
}

func (i *RemoteServiceImpl) errorBase64(err error) rgse.RGSErr {
	if err != nil {
		return rgse.Create(rgse.B64Error)
	}
	return nil
}

func (i *RemoteServiceImpl) errorHttpStatusCode(httpStatusCode int) (rgse.RGSErr, bool) {
	if httpStatusCode != 200 {
		logger.Debugf("handling http status code %d", httpStatusCode)
		if httpStatusCode == 403 || httpStatusCode == 401 {
			return rgse.Create(rgse.TokenExpired), false
		} else if httpStatusCode == 404 {
			return rgse.Create(rgse.EntityNotFound), false
		} else if httpStatusCode == 402 {
			return rgse.Create(rgse.InsufficientFundError), false
		} else if httpStatusCode == 408 || httpStatusCode == 504 {
			return rgse.Create(rgse.RequestTimeout), true
		}
		return rgse.Create(rgse.GenericWalletError), true
	}
	return nil, false
}

func (i *RemoteServiceImpl) errorResponseCode(errorResponse restErrorResponse) (rgse.RGSErr, bool) {
	responseCode := errorResponse.ResponseCode
	if responseCode != string(ResponseCodeOk) {
		logger.Debugf("handling response code %s", responseCode)
		if responseCode == string(ResponseCodeDataError) {
			return rgse.Create(rgse.BadRequest), false
		} else if responseCode == string(ResponseCodeInsufficentBalance) {
			return rgse.Create(rgse.InsufficientFundError), false
		} else if responseCode == string(ResponseCodeSessionExpired) {
			return rgse.Create(rgse.TokenExpired), false
		} else if responseCode == string(ResponseCodeCustomError) {
			var customErr rgse.RGSErr
			switch errorResponse.ErrorCode {
			case ErrorCodeCustomError:
				customErr = rgse.Create(rgse.CustomOperatorError)
				customErr.AppendErrorDesc(errorResponse.ErrorDesc)
			case ErrorCodeSpendingBudgetExceeded:
				customErr = rgse.Create(rgse.SpendingBudgetExceeded)
			case ErrorCodeBlockedFromProduct:
				customErr = rgse.Create(rgse.BlockedFromProduct)
			case ErrorCodeIpBlocked:
				customErr = rgse.Create(rgse.IpBlocked)
			case ErrorCodeMonthlyTimeLimit:
				customErr = rgse.Create(rgse.MontlyTimeLimit)
			case ErrorCodeWeeklyTimeLimit:
				customErr = rgse.Create(rgse.WeeklyTimeLimit)
			case ErrorCodeDailyTimeLimit:
				customErr = rgse.Create(rgse.DailyTimeLimit)
			}
			return customErr, false
		}
		// (don't retry until tested)
		return rgse.Create(rgse.GenericWalletError), false // true
	}
	return nil, false
}

func (i *RemoteServiceImpl) PlayerByToken(token Token, mode Mode, gameId string) (PlayerStore, GameStateStore, rgse.RGSErr) {
	logger.Debugf("RemoteServiceImpl.PlayerByToken([%v], [%v])", token, mode)
	if mode == ModeDemo {
		token = Token(i.demoToken())
	}

	authRq := restAuthenticateRequest{
		ReqId:    rng.Uuid(),
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

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return PlayerStore{}, GameStateStore{}, finalErr
	}
	var gameState []byte = nil
	authResp := i.restAuthenticateResponse(resp)
	finalErr, _ = i.errorResponseCode(authResp.restErrorResponse)
	if finalErr != nil {
		finalErr.AppendErrorText(authResp.Message)
		return PlayerStore{}, GameStateStore{}, finalErr
	}
	var lastTransaction restTransactionRequest
	var balance BalanceStore
	// we don't care about this error
	lastTransaction, balance, _ = i.getLastGamestate(authResp.Metadata.VendorInfo.LastTx, authResp.Metadata.VendorInfo.LastAttemptedTx)

	if balance.Token == "" {
		balance.Token = Token(authResp.Token)
		balance.FreeGames = FreeGamesStore{authResp.FreeGames.NrGames, authResp.FreeGames.CampaignRef, engine.Fixed(authResp.FreeGames.WagerAmt * 10000)}
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
			PlayerId:            authResp.Id,
			Token:               balance.Token,
			Mode:                mode,
			Username:            authResp.Username,
			Balance:             balance.Balance,
			FreeGames:           balance.FreeGames,
			BetLimitSettingCode: authResp.BetLimit,
			CompanyId:           fmt.Sprintf("%v", authResp.CompanyId),
			BetSettingId:        authResp.BetSettingId,
		},
		GameStateStore{GameState: gameState, WalletInternalStatus: lastTransaction.InternalStatus},
		nil
}

func (i *RemoteServiceImpl) getLastGamestate(lastTx restTransactionRequest, lastAttemptedTx restTransactionRequest) (lastTxInfo restTransactionRequest, balance BalanceStore, err rgse.RGSErr) {
	// if last_tx and last_attempted_tx are the same, doesn't matter
	logger.Debugf("Determining last GS : Last Tx = {%v}, Last Attempted Tx = {%v}", lastTx, lastAttemptedTx)
	if lastAttemptedTx.TxRef == lastTx.TxRef {
		lastTxInfo = lastTx
		return
	}

	// if no gamestate has been processed but there has been an error on attempting to process the first state:
	if len(lastTx.GameState) == 0 {
		err = rgse.Create(rgse.EntityNotFound)
		return
		//if strings.Contains(lastAttemptedTx.Round, "GSinit") {
		//	// the init gs was failed, a new init round will be generated
		//	return
		//}
		//err = rgse.Create(rgse.NoTxHistory)
		//return
	}
	// otherwise, check which tx it was that failed
	// we assume there is never more than one pending/failed tx on top of the last successful tx
	// because in freespins the PAYOUT tx is treated the same way as the WAGER in normal play, we need to check if this is the first tx of a gamestate
	//if len(lastAttemptedTx.GameState) == 0 || len(lastTx.GameState) == 0 {
	//
	//}

	gameState, errDecode := base64.StdEncoding.DecodeString(lastAttemptedTx.GameState)

	err = i.errorBase64(errDecode)
	if err != nil {
		return
	}

	gsDeserialized, _, serializeErr := DeserializeGamestate(lastAttemptedTx.Game, gameState)
	if serializeErr != nil {
		logger.Debugf("failed to deserialize lastAttemptedTx")
		return
	}

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
		// todo: this would be ok in a respin case
		err = rgse.Create(rgse.UnexpectedTx)
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
	err = rgse.Create(rgse.UnexpectedTx)
	return
}

func (i *RemoteServiceImpl) restAuthenticateResponse(response *http.Response) restAuthenticateResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("auth response: %s", string(body))
	var data restAuthenticateResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) PlayerSave(token Token, mode Mode, player PlayerStore) (PlayerStore, rgse.RGSErr) {
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
		FreeGames:           FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef, player.FreeGames.TotalWagerAmt},
		BetLimitSettingCode: player.BetLimitSettingCode,
		CompanyId:           player.CompanyId,
		BetSettingId:        player.BetSettingId,
	}, nil
}

func (i *LocalServiceImpl) renewToken(token Token) Token {
	playerId, _ := i.getToken(token)
	newToken := GenerateToken()
	i.setToken(newToken, playerId)
	i.deleteToken(token)

	return newToken
}

func (i *LocalServiceImpl) BalanceByToken(token Token, mode Mode) (BalanceStore, rgse.RGSErr) {
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
		FreeGames: FreeGamesStore{0, "", 0},
	}, nil
}

func (i *RemoteServiceImpl) BalanceByToken(token Token, mode Mode) (BalanceStore, rgse.RGSErr) {
	balRq := restBalanceRequest{
		ReqId:    rng.Uuid(),
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

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	balResp := i.restBalanceResponse(resp)

	finalErr, _ = i.errorResponseCode(balResp.restErrorResponse)
	if finalErr != nil {
		finalErr.AppendErrorText(balResp.Message)
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
		FreeGames: FreeGamesStore{balResp.FreeGames.NrGames, balResp.FreeGames.CampaignRef, engine.Fixed(balResp.FreeGames.WagerAmt * 10000)},
	}, nil
}

func (i *RemoteServiceImpl) restBalanceResponse(response *http.Response) restBalanceResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("balance response: %s", string(body))
	var data restBalanceResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, rgse.RGSErr) {
	logger.Debugf("LocalServiceImpl.Transaction([%v], [%v], [%v])", token, mode, transaction.TransactionId)

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)

	if transaction.Category == CategoryWager {
		// process free game
		if transaction.FreeGames.CampaignRef != "" && player.FreeGames.CampaignRef == transaction.FreeGames.CampaignRef && player.FreeGames.NoOfFreeSpins > 0 && transaction.RoundId == transaction.TransactionId {
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
		logger.Debugf("insufficient funds: %#v, %#v", transaction, player.Balance)
		return BalanceStore{}, rgse.Create(rgse.InsufficientFundError)
	}

	if RoundStatusClose == transaction.RoundStatus {
		logger.Debugf("close due to RoundStatusClose")
		// HACK to keep token from updating as client currently cannot handle token update on clientstate save call
		// i.setToken(token, playerId)
		// balance.Token = token
		//		return i.CloseRound(token, mode, transaction.GameId, transaction.RoundId, transaction.GameState, transaction.Ttl, &transaction.History)
	}

	i.setTransaction(transaction.TransactionId, transaction)
	key := player.PlayerId + "::" + transaction.GameId
	i.setTransactionByPlayerGame(key, transaction)
	i.setPlayer(playerId, player)
	newToken := i.renewToken(token)
	message, _ := i.getMessage(playerId)
	return BalanceStore{
		PlayerId: player.PlayerId,
		Token:    newToken,
		Message:  message,
		Balance: engine.Money{
			Currency: player.Balance.Currency,
			Amount:   player.Balance.Amount,
		},
		FreeGames: FreeGamesStore{player.FreeGames.NoOfFreeSpins, player.FreeGames.CampaignRef, player.FreeGames.TotalWagerAmt},
	}, nil
}

func (i *LocalServiceImpl) MultiTransaction(token Token, mode Mode, transactions []TransactionStore) (BalanceStore, rgse.RGSErr) {
	logger.Debugf("LocalServiceImpl.MultiTransactions([%v], [%v], [%v])", token, mode, len(transactions))
	var balance BalanceStore
	for _, t := range transactions {
		bs, err := i.Transaction(token, mode, t)
		if err != nil {
			return balance, err
		}
		token = bs.Token
		balance = bs
	}
	return balance, nil
}

func (i *RemoteServiceImpl) Transaction(token Token, mode Mode, transaction TransactionStore) (BalanceStore, rgse.RGSErr) {
	closeRound := false
	gameState := ""

	if RoundStatusClose == transaction.RoundStatus {
		closeRound = true
		logger.Debugf("close due to RoundStatusClose")
	}

	if transaction.GameState != nil {
		if len(transaction.GameState) > i.dataLimit {
			sentry.CaptureMessage(fmt.Sprintf("gamestate size exceeds store data limit of %d bytes", i.dataLimit))
		}
		gameState = base64.StdEncoding.EncodeToString(transaction.GameState)
	}

	txRq := restTransactionRequest{
		ReqId:    rng.Uuid(),
		Token:    string(token),
		Game:     transaction.GameId,
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
		Session:  transaction.RoundId,
		Currency: transaction.Amount.Currency,
		restTransactionDesc: restTransactionDesc{
			Amount:      int64(transaction.Amount.Amount / 10000), // Dashur expects amount in cents, transaction.Amount.Amount is type fixed (6decimals)
			BonusAmount: 0,
			JpAmount:    0,
			Category:    string(transaction.Category),
			CampaignRef: transaction.FreeGames.CampaignRef,
			CloseRound:  closeRound,
			TxRef:       transaction.TransactionId,
			GameState:   gameState,
		},
		Round:    transaction.RoundId,
		Ttl:      transaction.Ttl,
		TtlStamp: transaction.TxTime.Unix() + transaction.Ttl,
	}

	return i.txSend(txRq)
}

func (i *RemoteServiceImpl) MultiTransaction(token Token, mode Mode, transactions []TransactionStore) (BalanceStore, rgse.RGSErr) {
	if len(transactions) == 0 {
		panic("no transaction to send")
	}

	descs := make([]restTransactionDesc, len(transactions))
	for i_tx, transaction := range transactions {
		gamestate := ""
		if transaction.GameState != nil {
			if len(transaction.GameState) > i.dataLimit {
				sentry.CaptureMessage(fmt.Sprintf("gamestate size exceeds store data limit of %d bytes", i.dataLimit))
			}
			gamestate = base64.StdEncoding.EncodeToString(transaction.GameState)
		}
		descs[i_tx] = restTransactionDesc{
			Amount:      int64(transaction.Amount.Amount / 10000), // Dashur expects amount in cents, transaction.Amount.Amount is type fixed (6decimals)
			BonusAmount: 0,
			JpAmount:    0,
			Category:    string(transaction.Category),
			CampaignRef: transaction.FreeGames.CampaignRef,
			CloseRound:  RoundStatusClose == transaction.RoundStatus,
			TxRef:       transaction.TransactionId,
			GameState:   gamestate,
		}
	}

	txRq := restMultiTransactionRequest{
		ReqId:     rng.Uuid(),
		Token:     string(token),
		Game:      transactions[0].GameId,
		Platform:  i.defaultPlatform,
		Mode:      strings.ToLower(string(mode)),
		MultiTxes: descs,
		Session:   transactions[0].RoundId,
		Currency:  transactions[0].Amount.Currency,
		Round:     transactions[0].RoundId,
		Ttl:       transactions[0].Ttl,
		TtlStamp:  transactions[0].TxTime.Unix() + transactions[0].Ttl,
	}

	return i.txSendMulti(txRq)
}

func (i *RemoteServiceImpl) txSend(txRq restTransactionRequest) (BalanceStore, rgse.RGSErr) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(txRq)
	jsonErr := i.errorJson(err)
	if err != nil {
		return BalanceStore{}, jsonErr
	}
	return i.Send(b)
}

func (i *RemoteServiceImpl) txSendMulti(txRq restMultiTransactionRequest) (BalanceStore, rgse.RGSErr) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(txRq)
	jsonErr := i.errorJson(err)
	if err != nil {
		return BalanceStore{}, jsonErr
	}
	return i.Send(b)
}

func (i *RemoteServiceImpl) Send(b *bytes.Buffer) (BalanceStore, rgse.RGSErr) {
	start := time.Now()
	now := start
	var try int
	var retry bool
	var finalErr rgse.RGSErr
	for try = 0; try <= i.maxRetries; try, now = try+1, time.Now() {
		if i.timeoutMs > 0 && now.Sub(start).Milliseconds() > i.timeoutMs {
			logger.Errorf("transaction attempts exceeded timout after %v seconds: %s", now.Sub(start).String(), b.String())
			break
		}
		resp, err := i.request(ApiTypeTransaction, b)
		finalErr = i.errorRest(err)
		if finalErr != nil {
			continue
		}

		finalErr, retry = i.errorHttpStatusCode(resp.StatusCode)
		if finalErr != nil {
			if retry {
				continue
			}
			break
		}

		txResp := i.restTransactionResponse(resp)

		finalErr, retry = i.errorResponseCode(txResp.restErrorResponse)
		if finalErr != nil {
			if retry {
				continue
			}
			finalErr.AppendErrorText(txResp.Message)
			break
		}
		if txResp.PlayerId == i.logAccount {
			logger.Infof("transaction try %v took %v for account %v: %s", try, now.Sub(start).String(), txResp.PlayerId, b.String())
		}
		return BalanceStore{
			PlayerId: txResp.PlayerId,
			Token:    Token(txResp.Token),
			Balance: engine.Money{
				Currency: txResp.Currency,
				Amount:   engine.Fixed(txResp.Balance * 10000),
			},
			FreeGames: FreeGamesStore{txResp.FreeGames.NrGames, txResp.FreeGames.CampaignRef, engine.Fixed(txResp.FreeGames.WagerAmt * 10000)},
		}, nil
	}
	if try > i.maxRetries {
		logger.Errorf("transaction exceeded retry limit of %v: %s", i.maxRetries, b.String())
	}
	if finalErr != nil {
		logger.Errorf("transaction failed with error %v after %v tries: %s", finalErr.Error(), try, b.String())
	}
	return BalanceStore{}, finalErr
}

func (i *RemoteServiceImpl) restTransactionResponse(response *http.Response) restTransactionResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("transaction response: %s", string(body))
	var data restTransactionResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) GamestateById(gamestateId string) (GameStateStore, rgse.RGSErr) {
	logger.Debugf("LocalServiceImpl.GamestateById([%v])", gamestateId)
	transaction, ok := i.getTransaction(gamestateId)
	if !ok {
		return GameStateStore{}, rgse.Create(rgse.EntityNotFound)
	}

	return GameStateStore{GameState: transaction.GameState}, nil
}

func (i *LocalServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, rgse.RGSErr) {
	// Used at beginning of play() func to get previous gamestate, betlimit settings code, and free games info
	logger.Debugf("LocalServiceImpl.TransactionByGameId(token=%v, mode=%v, game=%v)", token, mode, gameId)

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)
	key := player.PlayerId + "::" + gameId

	transaction, ok := i.getTransactionByPlayerGame(key)

	if !ok {
		return TransactionStore{}, rgse.Create(rgse.EntityNotFound)
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
		CompanyId:           player.CompanyId,
		BetSettingId:        player.BetSettingId,
		FreeGames:           player.FreeGames,
		WalletStatus:        1,
		Ttl:                 transaction.Ttl,
		History:             transaction.History,
	}, nil
}

func (i *RemoteServiceImpl) TransactionByGameId(token Token, mode Mode, gameId string) (TransactionStore, rgse.RGSErr) {
	// Used at beginning of play() func to get previous gamestate, betlimit settings code, and free games info

	queryRq := restQueryRequest{
		ReqId:    rng.Uuid(),
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

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return TransactionStore{}, finalErr
	}

	var gameState []byte = nil
	queryResp := i.restQueryResponse(resp)

	finalErr, _ = i.errorResponseCode(queryResp.restErrorResponse)
	if finalErr != nil {
		// special handling for err does not exists
		if queryResp.ResponseCode == ResponseCodeUnknownError && strings.Contains(queryResp.Message, "E-CODE: [004:1003]") {
			return TransactionStore{}, rgse.Create(rgse.EntityNotFound)
		} else {
			finalErr.AppendErrorText(queryResp.Message)
			return TransactionStore{}, finalErr
		}
	}
	var lastTx restTransactionRequest
	var balance BalanceStore
	lastTx, balance, finalErr = i.getLastGamestate(queryResp.LastTx, queryResp.Metadata.VendorInfo.LastAttemptedTx)

	if finalErr != nil {
		return TransactionStore{}, finalErr
	}
	if balance.Token == "" {
		campaignRef := queryResp.FreeGames.CampaignRef
		// FreeGames is not included after the last wager of a campaign
		if campaignRef == "" {
			campaignRef = lastTx.CampaignRef
		}
		balance.Token = Token(lastTx.Token)
		balance.FreeGames = FreeGamesStore{queryResp.FreeGames.NrGames, campaignRef, engine.Fixed(queryResp.FreeGames.WagerAmt * 10000)}
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
		TransactionId:       lastTx.TxRef,
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
		CompanyId:           fmt.Sprintf("%v", queryResp.CompanyId),
		BetSettingId:        queryResp.BetSettingId,
		FreeGames:           balance.FreeGames,
		WalletStatus:        lastTx.InternalStatus,
		Ttl:                 lastTx.Ttl,
	}, nil
}

func (i *RemoteServiceImpl) restQueryResponse(response *http.Response) restQueryResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("query response: %s", string(body))
	var data restQueryResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *RemoteServiceImpl) restGameStateResponse(response *http.Response) restGameStateResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("auth response: %s", string(body))
	var data restGameStateResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *RemoteServiceImpl) restFeedResponse(response *http.Response) RestFeedResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("feed response: %s", string(body))
	var data RestFeedResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *RemoteServiceImpl) restFeedRoundResponse(response *http.Response) RestFeedRoundResponse {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	logger.Debugf("feed round response: %s", string(body))
	var data RestFeedRoundResponse
	json.Unmarshal(body, &data)
	return data
}

func (i *LocalServiceImpl) CloseRound(token Token, mode Mode, gameId string, roundId string, campaignRef string, gamestate []byte, ttl int64, history *TransactionHistory) (BalanceStore, rgse.RGSErr) {
	// Used in clientstate call
	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)

	balance, err := i.Transaction(token, mode, TransactionStore{
		TransactionId: rng.Uuid(),
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
		CompanyId:           player.CompanyId,
		BetSettingId:        player.BetSettingId,
		FreeGames:           player.FreeGames,
		Ttl:                 ttl,
		History:             *history,
	})

	if err != nil {
		return BalanceStore{}, err
	}
	// HACK to keep token from updating as client currently cannot handle token update on clientstate save call
	i.setToken(token, playerId)
	balance.Token = token
	return balance, nil
}

func (i *RemoteServiceImpl) CloseRound(token Token, mode Mode, gameId string, roundId string, campaignRef string, gamestate []byte, ttl int64, _ *TransactionHistory) (BalanceStore, rgse.RGSErr) {
	// Used in clientstate call
	closeRound := true

	if len(gamestate) > i.dataLimit {
		sentry.CaptureMessage(fmt.Sprintf("gamestate size exceeds store data limit of %d bytes", i.dataLimit))
	}

	txRq := restTransactionRequest{
		ReqId:    rng.Uuid(),
		Token:    string(token),
		Game:     gameId,
		Platform: i.defaultPlatform,
		Mode:     strings.ToLower(string(mode)),
		Session:  roundId,
		restTransactionDesc: restTransactionDesc{
			BonusAmount: 0,
			JpAmount:    0,
			Category:    string(CategoryClose),
			CampaignRef: campaignRef,
			CloseRound:  closeRound,
			TxRef:       roundId,
			GameState:   base64.StdEncoding.EncodeToString(gamestate),
		},
		Round:    roundId,
		Ttl:      ttl,
		TtlStamp: time.Now().Unix() + ttl,
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

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return BalanceStore{}, finalErr
	}

	txResp := i.restTransactionResponse(resp)

	finalErr, _ = i.errorResponseCode(txResp.restErrorResponse)
	if finalErr != nil {
		finalErr.AppendErrorText(txResp.Message)
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
		FreeGames: FreeGamesStore{txResp.FreeGames.NrGames, txResp.FreeGames.CampaignRef, engine.Fixed(txResp.FreeGames.WagerAmt * 10000)},
	}, nil
}

func (i *LocalServiceImpl) setToken(token Token, playerId string) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Token[token] = NewGcString(playerId, 3600)
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
	if ok && string(token) != ld.Token[token].str {
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
	if ok {
		return playerId.str, true
	}
	return "", false
}

func (i *LocalServiceImpl) setPlayer(playerId string, player PlayerStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Player[playerId] = NewGcPlayerStore(player, 3600)
}

func (i *LocalServiceImpl) getPlayer(playerId string) (ps PlayerStore, ok bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	var player gcPlayerStore
	player, ok = ld.Player[playerId]
	if ok {
		ps = player.ps
	}

	return
}

func (i *LocalServiceImpl) SetMessage(playerId string, message string) rgse.RGSErr {
	// this is used
	err := internalCheck()
	if err != nil {
		return rgse.Create(rgse.InternalServerError)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Message[playerId] = NewGcString(message, 3600)
	return nil
}

func (i *LocalServiceImpl) getMessage(playerId string) (msg string, ok bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	var message gcstring
	message, ok = ld.Message[playerId]
	if ok {
		msg = message.str
	}

	return
}

func (i *LocalServiceImpl) setTransaction(transactionId string, transaction TransactionStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.Transaction[transactionId] = NewGcTransactionStore(transaction, transaction.Ttl)
}

func (i *LocalServiceImpl) getTransaction(transactionId string) (tx TransactionStore, ok bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	var transaction gcTransactionStore
	transaction, ok = ld.Transaction[transactionId]
	if ok {
		tx = transaction.ts
	}

	return
}

func (i *LocalServiceImpl) setTransactionByPlayerGame(key string, transaction TransactionStore) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.Lock()
	defer ld.Lock.Unlock()

	ld.TransactionByPlayerGame[key] = NewGcTransactionStore(transaction, transaction.Ttl)
}

func (i *LocalServiceImpl) getTransactionByPlayerGame(key string) (tx TransactionStore, ok bool) {
	err := internalCheck()
	if err != nil {
		panic(err)
	}

	ld.Lock.RLock()
	defer ld.Lock.RUnlock()

	var transaction gcTransactionStore
	transaction, ok = ld.TransactionByPlayerGame[key]
	if ok {
		tx = transaction.ts
	}
	return
}

func (i *LocalServiceImpl) SetBalance(token Token, balance engine.Money) rgse.RGSErr {
	err := internalCheck()
	if err != nil {
		return rgse.Create(rgse.InternalServerError)
	}

	playerId, _ := i.getToken(token)
	player, _ := i.getPlayer(playerId)
	player.Balance = balance
	logger.Debugf("Setting playerId %s balance to %s in currency %s", playerId, player.Balance.Amount.ValueAsString(), player.Balance.Currency)
	i.setPlayer(playerId, player)
	return nil
}

func hashString(s string) int64 {
	crcTable := crc32.MakeTable(crc32.IEEE) // ISO)
	buf := bytes.NewBufferString(s)
	return int64(crc32.Checksum(buf.Bytes(), crcTable))
}

func DeserializeGamestate(gameId string, serialized []byte) (state engine.Gamestate, stateV3 engine.IGameStateV3, rgserr rgse.RGSErr) {
	var gameV3 IGameV3
	gameV3, rgserr = CreateGameV3(gameId)
	if rgserr != nil {
		logger.Debugf("deserialize V2 gamestate")
		state = DeserializeGamestateFromBytes(serialized)
		rgserr = nil
		return
	}
	logger.Debugf("deserialize V3 gamestate")
	stateV3, rgserr = gameV3.DeserializeState(serialized)
	if rgserr != nil {
		return
	}
	base := stateV3.Base()
	state = engine.Gamestate{
		Id:                base.Id,
		Game:              base.Game,
		Transactions:      base.Transactions,
		PreviousGamestate: base.PreviousGamestate,
		NextGamestate:     base.NextGamestate,
		Closed:            base.Closed,
		RoundID:           base.RoundId,
		Features:          base.Features,
	}
	return
}

func (i *LocalServiceImpl) Feed(token Token, mode Mode, gameId, startTime string, endTime string, pageSize int, page int) (rounds []FeedRound, nextPage int, finalErr rgse.RGSErr) {
	rounds = []FeedRound{}
	var ts TransactionStore
	var err rgse.RGSErr
	ts, err = i.TransactionByGameId(token, mode, gameId)
	if err != nil {
		nextPage = 1
		return
	}
	const timeLayout string = "2006-01-02 15:04:05.000"
	tstart, terr1 := time.Parse(timeLayout, startTime)
	tend, terr2 := time.Parse(timeLayout, endTime)
	if terr1 != nil || terr2 != nil {
		if terr1 != nil {
			logger.Debugf("could not parse startTime %s", startTime)
		}
		if terr2 != nil {
			logger.Debugf("could not parse endTime %s", startTime)
		}
		finalErr = rgse.Create(rgse.JsonError)
		return
	}
	idx := 0
	pageidx := 0
	round := FeedRound{Metadata: FeedRoundMetadata{}}
	addRound := func() {
		if round.Metadata.RoundId != "" {
			idx++
			for idx > pageidx*pageSize {
				pageidx++
			}
			if pageidx == page {
				rounds = append(rounds, round)
				round.Metadata.RoundId = ""
			}
		}
	}
	for true {
		state, stateV3, _ := DeserializeGamestate(gameId, ts.GameState)
		if ts.TxTime.After(tstart) && ts.TxTime.Before(tend) {
			tids := make([]int64, len(state.Transactions))
			for i, t := range state.Transactions {
				tids[i] = hashString(t.Id)
			}
			if state.RoundID == round.Metadata.RoundId {
				round.TransactionIds = append(round.TransactionIds, tids...)
			} else {
				addRound()
				round = FeedRound{
					Id:             hashString(ts.TransactionId),
					CurrencyUnit:   ts.Amount.Currency,
					ExternalRef:    ts.TransactionId,
					Status:         string(ts.RoundStatus),
					TransactionIds: tids,
					StartTime:      ts.TxTime.UTC().Format("2006-01-02 15:04:05.000"),
					Metadata: FeedRoundMetadata{
						RoundId:   state.RoundID,
						ExtItemId: gameId,
						ItemId:    0,
						Vendor: FeedRoundVendordata{
							State:   state,
							StateV3: stateV3,
						},
					},
					NumWager:  ts.History.NumWager,
					SumWager:  ts.History.SumWager,
					NumPayout: ts.History.NumPayout,
					SumPayout: ts.History.SumPayout,
				}
			}
		}

		var ok bool
		ts, ok = i.getTransaction(state.PreviousGamestate)
		if !ok {
			addRound()
			break
		}
	}
	nextPage = page
	if pageidx > page {
		nextPage++
	}
	return
}

func feedRoundId(i *LocalServiceImpl, token Token, mode Mode, gameId string, transactionId int64) (string, rgse.RGSErr) {
	var ts TransactionStore
	var err rgse.RGSErr
	ts, err = i.TransactionByGameId(token, mode, gameId)
	if err != nil {
		return "", nil
	}

	for true {
		state, _, _ := DeserializeGamestate(gameId, ts.GameState)
		tid := hashString(ts.TransactionId)
		if tid == transactionId {
			return state.RoundID, nil
		}

		var ok bool
		ts, ok = i.getTransaction(state.PreviousGamestate)
		if !ok {
			break
		}
	}
	return "", nil
}

func (i *LocalServiceImpl) FeedRound(token Token, mode Mode, gameId string, transactionId int64) (feeds []FeedTransaction, finalErr rgse.RGSErr) {
	feeds = []FeedTransaction{}

	roundId, _ := feedRoundId(i, token, mode, gameId, transactionId)
	if roundId == "" {
		logger.Errorf("could not find any transaction matching id %d", transactionId)
	} else {
		logger.Debugf("looking for transactions with roundId %s", roundId)
	}

	var ts TransactionStore
	var err rgse.RGSErr
	ts, err = i.TransactionByGameId(token, mode, gameId)
	if err != nil {
		return
	}

	found := false
	for true {
		state, stateV3, _ := DeserializeGamestate(gameId, ts.GameState)

		if state.RoundID == roundId {
			amount := ts.Amount.Amount.ValueAsFloat64()
			feed := FeedTransaction{
				Id:           hashString(ts.TransactionId),
				Category:     string(ts.Category),
				ExternalRef:  ts.TransactionId,
				CurrencyUnit: ts.Amount.Currency,
				Amount:       amount,
				Metadata: FeedRoundMetadata{
					RoundId:   state.RoundID,
					ExtItemId: gameId,
					ItemId:    0,
					Vendor: FeedRoundVendordata{
						State:   state,
						StateV3: stateV3,
					},
				},
				TxTime: ts.TxTime.UTC().Format("2006-01-02 15:04:05.000"),
			}
			feeds = append(feeds, feed)
		} else if found {
			break
		}

		var ok bool
		ts, ok = i.getTransaction(state.PreviousGamestate)
		if !ok {
			break
		}
	}

	return
}

func (i *RemoteServiceImpl) Feed(token Token, mode Mode, gameId, startTime string, endTime string, pageSize int, page int) (rounds []FeedRound, nextPage int, finalErr rgse.RGSErr) {
	feedRq := RestFeedRequest{
		ReqId:     rng.Uuid(),
		Token:     string(token),
		Game:      gameId,
		Platform:  i.defaultPlatform,
		StartTime: startTime,
		EndTime:   endTime,
		PageSize:  pageSize,
		Page:      page,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(feedRq)

	logger.Debugf("Feed request: %s", b.String())

	finalErr = i.errorJson(err)
	if finalErr != nil {
		return
	}
	//start := time.Now()
	resp, err := i.request(ApiTypeFeed, b)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return
	}

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return
	}

	feedResp := i.restFeedResponse(resp)
	bfeedresp, _ := json.Marshal(feedResp)

	finalErr, _ = i.errorResponseCode(feedResp.restErrorResponse)
	if finalErr != nil {
		logger.Errorf("feed response error code. feedresponse: %s", bfeedresp)
		return
	}
	nextPage = feedResp.NextPage
	rounds = make([]FeedRound, len(feedResp.Rounds))
	for i, v := range feedResp.Rounds {
		rounds[i], finalErr = NewFeedRound(v, gameId)
		if finalErr != nil {
			return
		}
	}

	finalErr = nil
	return
}

func (i *RemoteServiceImpl) FeedRound(token Token, mode Mode, gameId string, roundId int64) (feeds []FeedTransaction, finalErr rgse.RGSErr) {
	feedRq := RestFeedRoundRequest{
		ReqId:    rng.Uuid(),
		Token:    string(token),
		Game:     gameId,
		Platform: i.defaultPlatform,
		RoundId:  roundId,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(feedRq)

	logger.Debugf("Feed round request: %s", b.String())

	finalErr = i.errorJson(err)
	if finalErr != nil {
		return
	}
	//start := time.Now()
	resp, err := i.request(ApiTypeFeed, b)

	finalErr = i.errorRest(err)
	if finalErr != nil {
		return
	}

	finalErr, _ = i.errorHttpStatusCode(resp.StatusCode)
	if finalErr != nil {
		return
	}

	feedResp := i.restFeedRoundResponse(resp)
	bfeedresp, _ := json.Marshal(feedResp)

	finalErr, _ = i.errorResponseCode(feedResp.restErrorResponse)
	if finalErr != nil {
		logger.Errorf("feed round response error code. feedresponse: %s", bfeedresp)
		return
	}
	feeds = make([]FeedTransaction, len(feedResp.Feeds))
	for i, v := range feedResp.Feeds {
		feeds[i], finalErr = NewFeedTransaction(v, gameId)
		if finalErr != nil {
			return
		}
	}

	finalErr = nil
	return
}

func internalInit(c *config.Config) {
	logger.Infof("internal-init [DevMode: %v]", c.DevMode)

	if c.DevMode {
		if ld == nil {
			ld = new(LocalData)
			ld.Token = make(map[Token]gcstring)
			ld.Player = make(map[string]gcPlayerStore)
			ld.Message = make(map[string]gcstring)
			ld.Transaction = make(map[string]gcTransactionStore)
			ld.TransactionByPlayerGame = make(map[string]gcTransactionStore)

			go garbageCollector()
		}
	} else {
		remoteServiceImplHttpClient = &promhttp.Client{
			Client:     http.DefaultClient,
			Registerer: prometheus.DefaultRegisterer,
		}
	}
}

func internalCheck() rgse.RGSErr {
	if ld == nil {
		logger.Errorf("Local data is not initialized. Panic!!")
		return rgse.Create(rgse.StoreInitError)
	}

	if ld.Player == nil {
		logger.Errorf("Local data is not initialized. Panic!!")
		return rgse.Create(rgse.StoreInitError)
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
		maxRetries:      c.DashurConfig.StoreMaxRetries,
		timeoutMs:       c.DashurConfig.StoreTimeoutMs,
		defaultLanguage: c.DefaultLanguage,
		defaultPlatform: c.DefaultPlatform,
		demoTokenPrefix: c.DemoTokenPrefix,
		demoCurrency:    c.DemoCurrency,
		logAccount:      c.LogAccount,
		dataLimit:       c.DataLimit,
	}
}

func NewLocal() LocalService {
	internalInit(&config.Config{
		DevMode: true,
	})

	return &LocalServiceImpl{}
}
