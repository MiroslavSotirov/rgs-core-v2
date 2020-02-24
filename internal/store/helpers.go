package store

import (
	"bytes"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang/protobuf/proto"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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

var delimiter = []byte("...")

func SerializeGamestateToBytes(deserialized engine.Gamestate) []byte {
	// turns session information into a byte slice
	if len(deserialized.Transactions) == 0 {
		logger.Errorf("ERROR: GAMESTATE HAS NO TX: %#v", deserialized)
		return []byte{}
	}
	if deserialized.Gamification == nil {logger.Errorf("NO GAMIFICATION")}
	deserializedPB := deserialized.Convert()
	//deserializedPBType, deserializedPBTXType := deserialized.ConvertLegacy()
	data, err := proto.Marshal(&deserializedPB)
	if err != nil {
		logger.Errorf("Error serializing Gamestate to bytes")
		return []byte{}
	}
	//var dataTx []byte

	//for i := 0; i < len(deserializedPBTXType); i++ {
	//	dataTx, err = proto.Marshal(deserializedPBTXType[i])
	//	data = append(data, delimiter...)
	//	data = append(data, dataTx...)
	//}
	//
	//logger.Warnf("delim bytes: %v", delimiter)
	//logger.Warnf("Serialized GS : %#v", deserializedPBType)
	logger.Warnf("serialized: %v", data)
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
	//err := deserializedGS.XXX_Unmarshal(data[0])
	err := proto.Unmarshal(data[0], &deserializedGS)
	// Decode (receive) the value.
	logger.Debugf("GS %#v", deserializedGS)
	if err != nil {

		//err = proto.Unmarshal(data[0], &deserializedGS)
		//if err != nil {
		logger.Errorf("Error deserializing gamestate from bytes: %v", err)
		return engine.Gamestate{}
		//}

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
	return deserializedGS.ConvertLegacy(deserializedTX)
}