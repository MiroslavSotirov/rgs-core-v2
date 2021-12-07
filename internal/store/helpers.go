package store

import (
	"bytes"
	"encoding/base64"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang/protobuf/proto"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

// set up local memcached server with:
// memcached -l 127.0.0.1 -m 64 -vv

//connect to memcache only for dev mode
var MC *memcache.Client
var ServLocal LocalService
var Serv Service

func Init(getHashes bool) rgse.RGSErr {

	//ServLocal = New(&config.GlobalConfig)
	ServLocal = NewLocal()
	Serv = New(&config.GlobalConfig)
	if config.GlobalConfig.DevMode {
		MC = memcache.New(config.GlobalConfig.MCRouter)
	}
	if getHashes {
		_, _, err := engine.GetHashes()
		if err != nil {
			return err
		}
	}

	return nil
}

var delimiter = []byte("...")

func SerializeGamestateToBytes(deserialized engine.Gamestate) []byte {
	// turns session information into a byte slice
	if len(deserialized.Transactions) == 0 {
		logger.Errorf("ERROR: GAMESTATE HAS NO TX: %#v", deserialized)
		return []byte{}
	}

	deserializedPB := deserialized.Convert()

	data, err := proto.Marshal(&deserializedPB)
	if err != nil {
		logger.Errorf("Error serializing Gamestate to bytes")
		return []byte{}
	}

	return data
}

func DeserializeGamestateFromBytes(serialized []byte) engine.Gamestate {
	// turns serialized session information into session struct
	var deserializedGS engine.GamestatePB
	err := proto.Unmarshal(serialized, &deserializedGS)
	if err != nil {
		logger.Warnf("Attempting old format deserialization")
		return DeserializeGamestateFromBytesLegacy(serialized)
	}
	gs := deserializedGS.Convert()

	if gs.Id == "" {
		logger.Warnf("Conversion failed, attempting old format deserialization")
		return DeserializeGamestateFromBytesLegacy(serialized)
	}

	return gs
}

func DeserializeGamestateFromBytesLegacy(serialized []byte) engine.Gamestate {
	// turns serialized session information into session struct
	var deserializedGS engine.GamestatePB

	data := bytes.Split(serialized, delimiter)
	if len(data) == 1 {
		logger.Warnf("Attempting to deserialize old format")
		data = bytes.Split(serialized, []byte("..."))
	}

	err := proto.Unmarshal(data[0], &deserializedGS)
	// Decode (receive) the value.
	logger.Debugf("GS %#v", deserializedGS)
	if err != nil {
		logger.Errorf("Error deserializing gamestate from bytes: %v", err)
		return engine.Gamestate{}
	}
	deserializedTX := make([]*engine.WalletTransactionPB, len(data)-1)

	for i := 1; i < len(data); i++ {
		var deserialized engine.WalletTransactionPB
		err := proto.Unmarshal(data[i], &deserialized)
		if err != nil {
			logger.Errorf("Error deserializing transaction from bytes: %v", err)
			return engine.Gamestate{}
		}
		deserializedTX[i-1] = &deserialized
	}
	if len(deserializedTX) == 0 {
		return engine.Gamestate{}
	}
	return deserializedGS.ConvertLegacy(deserializedTX)
}

func NewFeedRound(v restRounddata) (FeedRound, rgse.RGSErr) {
	gameState, errDecode := base64.StdEncoding.DecodeString(v.Metadata.Vendor.State)
	if errDecode != nil {
		return FeedRound{}, rgse.Create(rgse.B64Error)
	}
	gsDeserialized := DeserializeGamestateFromBytes(gameState)

	return FeedRound{
		Id:              v.Id,
		CurrencyUnit:    v.CurrencyUnit,
		ExternalRef:     v.ExternalRef,
		Status:          v.Status,
		TransactionIds:  v.TransactionIds,
		NumWager:        v.NumWager,
		SumWager:        v.SumWager,
		NumPayout:       v.NumPayout,
		SumPayout:       v.SumPayout,
		NumRefund:       v.NumRefund,
		SumRefundCredit: v.SumRefundCredit,
		SumRefundDebit:  v.SumRefundDebit,
		StartTime:       v.StartTime,
		CloseTime:       v.CloseTime,
		Metadata: FeedRoundMetadata{
			RoundId:   v.Metadata.RoundId,
			ExtItemId: v.Metadata.ExtItemId,
			ItemId:    v.Metadata.ItemId,
			Vendor: FeedRoundVendordata{
				State: gsDeserialized,
			},
		},
	}, nil
}

func NewFeedTransaction(v restTransactiondata) (FeedTransaction, rgse.RGSErr) {
	gameState, errDecode := base64.StdEncoding.DecodeString(v.Metadata.Vendor.State)
	if errDecode != nil {
		return FeedTransaction{}, rgse.Create(rgse.B64Error)
	}
	gsDeserialized := DeserializeGamestateFromBytes(gameState)

	return FeedTransaction{
		Id:           v.Id,
		Category:     v.Category,
		ExternalRef:  v.ExternalRef,
		CurrencyUnit: v.CurrencyUnit,
		Amount:       v.Amount,
		Metadata: FeedRoundMetadata{
			RoundId:   v.Metadata.RoundId,
			ExtItemId: v.Metadata.ExtItemId,
			ItemId:    v.Metadata.ItemId,
			Vendor: FeedRoundVendordata{
				State: gsDeserialized,
			},
		},
		TxTime: v.TxTime,
	}, nil
}

func GenerateToken() Token {
	bt, _ := time.Now().MarshalBinary()
	b64token := rng.RandStringRunes(16) + base64.StdEncoding.EncodeToString(bt)
	return Token(strings.ReplaceAll(b64token, "/", "-"))
}
