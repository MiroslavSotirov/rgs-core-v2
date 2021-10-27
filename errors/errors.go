package rgserror

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

// Error code constants
const (
	RgsInitError         = 1
	StoreInitError       = 2
	BadConfigError       = 10
	EngineHashError      = 11
	EngineConfigError    = 12
	GenericEngineError   = 100
	SpinSequenceError    = 101
	EngineNotFoundError  = 102
	BetConfigError       = 103
	InvalidStakeError    = 104
	IncompleteRoundError = 105

	// Gamestate JSON marshalling
	GamestateStringSerializerError   = 110
	GamestateStringDeserializerError = 111
	GamestateByteSerializerError     = 112
	GamestateByteDeserializerError   = 114

	//Memcached gamestate store/retrieve error
	GamestateCacheStoreError    = 200
	GamestateCacheRetrieveError = 201

	InvalidContentTypeError    = 300
	ContentTypeNotAllowedError = 301
	// Dashur Internal Error
	DasInvalidTokenError     = 401
	DasInsufficientFundError = 402
	DASHostError             = 403

	// Wallet Error
	InvalidCredentials     = 420
	InvalidWallet          = 421
	InvalidWalletCurrency  = 422
	BalanceStoreError      = 423
	InsufficientFundError  = 450
	GenericWalletError     = 451
	PeviousTXPendingError  = 452
	NoSuchPlayer           = 453 // demo wallet only
	JsonError              = 454
	RestError              = 455
	B64Error               = 456
	TokenExpired           = 457
	EntityNotFound         = 458
	BadRequest             = 459
	NoTxHistory            = 460
	UnexpectedTx           = 461
	UnexpectedWalletStatus = 462
	YamlError              = 463
	RequestTimeout         = 464

	// Wallet & Operator
	BadOperatorConfig = 600
	BadFSWagerAmt     = 601
	// System Error
	InternalServerError = 500
	// Session Error
	CreateSessionError  = 700
	UpdateSessionError  = 701
	DeletetSessionError = 702
	FetchSessionError   = 703

	CreateDemoSessionError = 704
	// Generic errors

	// forceTool
	NoForceError    = 800
	ForceProhibited = 801
	Forcing         = 802
)

var sentryIgnoreList = []int{
	NoForceError,
	ForceProhibited,
	Forcing,
	EntityNotFound,     // this error happens a lot
	GenericWalletError, // this error happens a lot
	SpinSequenceError,
	InvalidStakeError,
	TokenExpired,
}

// ErrMsg Error message key value map
var ErrMsg = map[int]string{

	BadConfigError:                   "Bad configuration",
	EngineHashError:                  "Could not generate hashes of engine files",
	EngineConfigError:                "No game config",
	GenericEngineError:               "Engine error",
	SpinSequenceError:                "Spin request out of sequence, please reload",
	EngineNotFoundError:              "Engine not found",
	BetConfigError:                   "Bet configuration error",
	InvalidStakeError:                "Invalid stake error",
	RgsInitError:                     "RGS Initialization error",
	StoreInitError:                   "RGS Storage Initialization error",
	GamestateStringSerializerError:   "Failure serializing Gamestate to string",
	GamestateByteSerializerError:     "Failure serializing Gamestate to bytes",
	GamestateStringDeserializerError: "Failure deserializing Gamestate from string",
	GamestateByteDeserializerError:   "Failure deserializing Gamestate from Errorbytes",
	GamestateCacheStoreError:         "Failure storing gamestate to memcached",
	GamestateCacheRetrieveError:      "Failure retrieving gamestate from memcached",
	InvalidContentTypeError:          "Invalid Content-Type",
	ContentTypeNotAllowedError:       "Content-Type not allowed",
	DasInvalidTokenError:             "Invalid Access Token",
	DasInsufficientFundError:         "Insufficient Fund",
	InsufficientFundError:            "Insufficient Fund",
	InvalidCredentials:               "Invalid Credentials",
	InvalidWallet:                    "Invalid Wallet",
	InvalidWalletCurrency:            "Transaction currency does not match wallet",
	BalanceStoreError:                "Failed to store balance",
	CreateSessionError:               "Failure creating new Session",
	UpdateSessionError:               "Failure updating session",
	DeletetSessionError:              "Failure deleting session",
	FetchSessionError:                "Failure fetching session",
	CreateDemoSessionError:           "Error setting demo session ",
	BadOperatorConfig:                "Bad Operator configuration",
	InternalServerError:              "System Error",
	GenericWalletError:               "Generic wallet error",
	PeviousTXPendingError:            "Previous transaction still pending, please try again",
	IncompleteRoundError:             "Not the final state in round, can't be closed",
	NoForceError:                     "No force matching that code",
	ForceProhibited:                  "Force prohibited for this gamestate",
	Forcing:                          "FORCING GAMESTATE",
	NoSuchPlayer:                     "No player found",
	JsonError:                        "Failure encoding/decoding json",
	RestError:                        "REST error",
	B64Error:                         "Failure encoding/decoding base64",
	TokenExpired:                     "Token expired",
	EntityNotFound:                   "Entity not found",
	BadRequest:                       "Unable to perform rest function, found data input error",
	NoTxHistory:                      "No transaction history",
	UnexpectedTx:                     "Got unexpected WAGER tx",
	UnexpectedWalletStatus:           "Unexpected Wallet status",
	YamlError:                        "Error encoding/decoding yaml",
	BadFSWagerAmt:                    "Bad freespin wager amount",
	RequestTimeout:                   "Request took too long",
}

type RGSErr interface {
	Error() string
	//Init(int, ...string)
	AppendErrorText(string)
	//SetErrorTextByCode(int)
}

// RGSError Generic RGS Error
type RGSError struct {
	ErrCode          int    `json:"err_code"`          // numeric error code
	DefaultErrorText string `json:"-"`                 // application-level error message
	ErrorText        string `json:"err_msg,omitempty"` // application-level error message
}

func Create(code int) *RGSError {
	e := &RGSError{ErrCode: code, DefaultErrorText: ErrMsg[code]}
	for i := 0; i < len(sentryIgnoreList); i++ {
		if e.ErrCode == sentryIgnoreList[i] {
			return e
		}
	}
	sentry.CaptureException(e)
	return e
}

func (e *RGSError) Error() (errorMsg string) {
	//sentry.CaptureException(e)
	//sentry.Flush(10*time.Millisecond)
	if e.ErrorText == "" {
		return fmt.Sprintf("Error %d, %s", e.ErrCode, e.DefaultErrorText)
	}
	return fmt.Sprintf("Error %d, %s - %s", e.ErrCode, e.DefaultErrorText, e.ErrorText)
}

//
//func (e *RGSError) Init(code int, msgs ...string) {
//	// this isn't actually used anywhere
//	e.ErrCode = code
//	e.DefaultErrorText = ErrMsg[code]
//	e.ErrorText = fmt.Sprintf("%s %s", e.DefaultErrorText, strings.Join(msgs, " "))
//}

//AppendErrorText append custom error message
func (e *RGSError) AppendErrorText(text string) {
	e.ErrorText = text
}

//
////SetErrorTextByCode set custom error message
//func (e *RGSError) SetErrorTextByCode(code int) {
//	e.ErrorText = ErrMsg[code]
//}
