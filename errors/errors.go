package rgserror

import (
	"fmt"
	"strings"
)

// TODO: And more specific error codes
// Error code constants
const (
	rgsInitError        = 1
	storeInitError      = 2
	badConfigError      = 10
	engineHashError     = 11
	engineConfigError   = 12
	genericEngineError  = 100
	spinSequenceError   = 101
	engineNotFoundError = 102
	betConfigError      = 103
	InvalidStakeError   = 104

	// Gamestate JSON marshalling
	gamestateStringSerializerError   = 110
	gamestateStringDeserializerError = 111
	gamestateByteSerializerError     = 112
	gamestateByteDeserializerError   = 114
	//Memcached gamestate store/retrieve error
	gamestateCacheStoreError    = 200
	gamestateCacheRetrieveError = 201

	invalidContentTypeError    = 300
	contentTypeNotAllowedError = 301
	// Dashur Internal Error
	dasInvalidTokenError     = 401
	dasInsufficientFundError = 402
	DASHostError             = 403
	// Wallet Error
	invalidCredentials    = 420
	invalidWallet         = 421
	invalidWalletCurrency = 422
	balanceStoreError     = 423
	insufficientFundError = 450
	genericWalletError    = 451
	peviousTXPendingError = 452

	// Wallet & Operator
	badOperatorConfig = 600
	// System Error
	internalServerError = 500
	// Session Error
	createSessionError  = 700
	updateSessionError  = 701
	deletetSessionError = 702
	fetchSessionError   = 703

	createDemoSessionError = 704
	// Generic errors
)

// ErrMsg Error message key value map
var ErrMsg = map[int]string{

	badConfigError:                   "Bad configuration",
	engineHashError:                  "Could not generate hashes of engine files",
	engineConfigError:                "No game config",
	genericEngineError:               "Engine error",
	spinSequenceError:                "Spin request out of sequence, please reload",
	engineNotFoundError:              "Engine not found",
	betConfigError:                   "Bet configuration error",
	InvalidStakeError:                "Invalid stake error",
	rgsInitError:                     "RGS Initialization error",
	storeInitError:                   "RGS Storage Initialization error",
	gamestateStringSerializerError:   "Failure serializing Gamestate to string",
	gamestateByteSerializerError:     "Failure serializing Gamestate to bytes",
	gamestateStringDeserializerError: "Failure deserializing Gamestate from string",
	gamestateByteDeserializerError:   "Failure deserializing Gamestate from Errorbytes",
	gamestateCacheStoreError:         "Failure storing gamestate to memcached",
	gamestateCacheRetrieveError:      "Failure retrieving gamestate from memcached",
	invalidContentTypeError:          "Invalid Content-Type",
	contentTypeNotAllowedError:       "Content-Type not allowed",
	dasInvalidTokenError:             "Invalid Access Token",
	dasInsufficientFundError:         "Insufficient Fund",
	insufficientFundError:            "Insufficient Fund",
	invalidCredentials:               "Invalid Credentials",
	invalidWallet:                    "Invalid Wallet",
	invalidWalletCurrency:            "Transaction currency does not match wallet",
	balanceStoreError:                "Failed to store balance",
	createSessionError:               "Failure creating new Session",
	updateSessionError:               "Failure updating session",
	deletetSessionError:              "Failure deleting session",
	fetchSessionError:                "Failure fetching session",
	createDemoSessionError:           "Error setting demo session ",
	badOperatorConfig:                "Bad Operator configuration",
	internalServerError:              "System Error",
	genericWalletError:               "Generic wallet error",
	peviousTXPendingError:            "Previous transaction still pending, please try again",
}

type IRGSError interface {
	Error() string
	Init(int, ...string)
	AppendErrorText(string)
	SetErrorTextByCode(int)
}

// RGSError Generic RGS Error
type RGSError struct {
	ErrCode          int    `json:"err_code"`          // numeric error code
	DefaultErrorText string `json:"-"`                 // application-level error message
	ErrorText        string `json:"err_msg,omitempty"` // application-level error message
}

func CreateRGSErr(code int) *RGSError {
	return &RGSError{ErrCode: code, DefaultErrorText: ErrMsg[code]}
}

func (e *RGSError) Error() string {
	if e.ErrorText == "" {
		return fmt.Sprintf("Error %d, %s", e.ErrCode, e.DefaultErrorText)
	}
	return fmt.Sprintf("Error %d, %s - %s", e.ErrCode, e.DefaultErrorText, e.ErrorText)
}

func (e *RGSError) Init(code int, msgs ...string) {
	e.ErrCode = code
	e.ErrorText = ErrMsg[code]
	e.ErrorText = fmt.Sprintf("%s %s", e.ErrorText, strings.Join(msgs, " "))
}

//AppendErrorText append custom error message
func (e *RGSError) AppendErrorText(text string) {
	e.ErrorText = text
}

//SetErrorTextByCode set custom error message
func (e *RGSError) SetErrorTextByCode(code int) {
	e.ErrorText = ErrMsg[code]
}

// Pre-defined errors
// Add custom errors here
var (
	ErrRGSInit      = CreateRGSErr(rgsInitError)
	ErrStoreInit    = CreateRGSErr(storeInitError)
	ErrBadConfig    = CreateRGSErr(badConfigError)
	ErrEngine       = CreateRGSErr(genericEngineError)
	ErrEngineHash   = CreateRGSErr(engineHashError)
	ErrEngineConfig = CreateRGSErr(engineConfigError)

	ErrEngineNotFound = CreateRGSErr(engineNotFoundError)
	ErrBetConfig      = CreateRGSErr(betConfigError)
	ErrSpinSequence   = CreateRGSErr(spinSequenceError)
	ErrInvalidStake   = CreateRGSErr(InvalidStakeError)

	ErrGamestateStringDeserializer = CreateRGSErr(gamestateStringDeserializerError)
	ErrGamestateStringSerializer   = CreateRGSErr(gamestateStringSerializerError)
	ErrGamestateByteSerializer     = CreateRGSErr(gamestateByteSerializerError)
	ErrGamestateByteDeserializer   = CreateRGSErr(gamestateByteDeserializerError)

	ErrGamestateStore = CreateRGSErr(gamestateCacheStoreError)
	ErrPreviousTXPending = CreateRGSErr(peviousTXPendingError)

	ErrGamestateRetrieve  = CreateRGSErr(gamestateCacheRetrieveError)
	ErrInvalidContentType = CreateRGSErr(invalidContentTypeError)

	ErrContentTypeNotAllowed = CreateRGSErr(contentTypeNotAllowedError)

	ErrDasInvalidTokenError     = CreateRGSErr(dasInvalidTokenError)
	ErrDasInsufficientFundError = CreateRGSErr(dasInsufficientFundError)

	ErrInsufficientFundError = CreateRGSErr(insufficientFundError)
	ErrInvalidCredentials    = CreateRGSErr(invalidCredentials)
	ErrInvalidWallet         = CreateRGSErr(invalidWallet)
	ErrInvalidWalletCurrency = CreateRGSErr(invalidWalletCurrency)
	ErrBalanceStoreError     = CreateRGSErr(balanceStoreError)
	ErrGenericWalletErr      = CreateRGSErr(genericWalletError)

	ErrCreateSession = CreateRGSErr(createSessionError)
	ErrUpdateSession = CreateRGSErr(updateSessionError)
	ErrFetchSession  = CreateRGSErr(fetchSessionError)
	ErrDeleteSession = CreateRGSErr(deletetSessionError)

	ErrSetDemoSession = CreateRGSErr(createDemoSessionError)

	ErrBadOperatorConfig   = CreateRGSErr(badOperatorConfig)
	ErrInternalServerError = CreateRGSErr(internalServerError)
)
