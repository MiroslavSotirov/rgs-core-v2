package store

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"

	"github.com/golang/protobuf/proto"
)

// set up local memcached server with:
// memcached -l 127.0.0.1 -m 64 -vv

//connect to memcache only for dev mode
var MC *memcache.Client
var ServLocal LocalService
var Serv Service

func Init() rgserror.IRGSError {

	//ServLocal = New(&config.GlobalConfig)
	ServLocal = NewLocal()
	Serv = New(&config.GlobalConfig)
	if config.GlobalConfig.DevMode {
		MC = memcache.New(config.GlobalConfig.MCRouter)
	}
	_, _, err := engine.GetHashes()
	if err != nil {
		hasherr := rgserror.ErrEngineHash
		hasherr.AppendErrorText(err.Error())
		logger.Errorf("could not generate hashes of engine files: %v", hasherr.Error())
	}

	return nil
}

func SerializeGamestateToString(deserialized engine.Gamestate) string {
	// serializes gamestate and transaction info to string with delimiter "::"
	// format: <Gamestate>::<TXWAGER>::<TXPAYOUT(optional)>::<TXENDROUND(optional)>

	deserializedPBType, deserializedTXPBType := deserialized.Convert()
	dataGS, err := proto.Marshal(&deserializedPBType)
	if err != nil {
		logger.Errorf("Error serializing Gamestate to String")
		return ""
	}
	returnStr := base64.StdEncoding.EncodeToString(dataGS)

	var dataTX []byte
	for _, tx := range deserializedTXPBType {
		dataTX, err = proto.Marshal(tx)
		if err != nil {
			logger.Errorf("Error serializing Gamestate to String")
			return ""
		}
		returnStr += fmt.Sprintf("::%v", base64.StdEncoding.EncodeToString(dataTX))
	}
	return returnStr
}

// todo: make these interface functions

func SerializeGamestateToBytes(deserialized engine.Gamestate) []byte {
	// turns session information into a byte slice
	if len(deserialized.Transactions) == 0 {
		logger.Errorf("ERROR: GAMESTATE HAS NO TX: %#v", deserialized)
		return []byte{}
	}
	deserializedPBType, deserializedPBTXType := deserialized.Convert()
	data, err := proto.Marshal(&deserializedPBType)
	if err != nil {
		logger.Errorf("Error serializing Gamestate to bytes")
		return []byte{}
	}
	var dataTx []byte
	delimiter := []byte("...")

	for i := 0; i < len(deserializedPBTXType); i++ {
		dataTx, err = proto.Marshal(deserializedPBTXType[i])
		data = append(data, delimiter...)
		data = append(data, dataTx...)
	}
	return data
}

func DeserializeGamestateFromBytes(serialized []byte) engine.Gamestate {
	// turns serialized session information into session struct
	var deserializedGS engine.GamestatePB
	data := bytes.Split(serialized, []byte("..."))
	err := proto.Unmarshal(data[0], &deserializedGS)
	// Decode (receive) the value.
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
	return deserializedGS.Convert(deserializedTX)
}
