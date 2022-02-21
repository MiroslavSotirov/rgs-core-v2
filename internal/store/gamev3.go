package store

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"io"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type IGameV3 interface {
	Base() *GameV3
	InitState() engine.IGameStateV3
	SerializeState(engine.IGameStateV3) []byte
	DeserializeState([]byte) (engine.IGameStateV3, rgse.RGSErr)
}

type GameV3 struct {
	Game       string
	EngineId   string
	Wallet     string
	Currency   string
	Token      Token
	EngineConf engine.EngineConfig
}

func (g *GameV3) Base() *GameV3 {
	return g
}

func (g GameV3) InitState() engine.IGameStateV3 {
	return nil
}

func (g GameV3) SerializeState(_ engine.IGameStateV3) []byte {
	panic("GameV3 SerializeState is stub. Use game specific implementation")
	return []byte{}
}

func (g GameV3) DeserializeState(serialized []byte) (state engine.IGameStateV3, rgserr rgse.RGSErr) {
	var stateV3 engine.GameStateV3
	var uncompressed []byte
	uncompressed, rgserr = DecompressState(serialized)
	if rgserr != nil {
		return
	}
	rgserr = stateV3.Deserialize(uncompressed)
	state = &stateV3
	return
}

const (
	// compression header (must be 4 bytes long)
	COMPRESSION_LZW  = "lzw "
	COMPRESSION_ZLIB = "zlib"
)

func CompressState(serialized []byte, algo string) []byte {
	var compressed []byte

	startTime := time.Now()
	switch algo {
	case COMPRESSION_LZW:
		compressed = CompressStateLzw(serialized)
	case COMPRESSION_ZLIB:
		compressed = CompressStateZlib(serialized)
	default:
		panic("CompressState unknown algorithm: " + algo)
	}
	duration := time.Now().Sub(startTime)
	logger.Debugf("CompressState algo= %s input len= %d output len= %d in %.4fms", algo, len(serialized), len(compressed), duration.Seconds()*1000.0)
	return compressed
}

func DecompressState(compressed []byte) ([]byte, rgse.RGSErr) {
	var serialized []byte
	var err rgse.RGSErr

	startTime := time.Now()
	if len(compressed) <= 4 {
		logger.Errorf("compressed buffer is too small to be valid")
		return []byte{}, rgse.Create(rgse.GamestateByteDecompressAlgoError)
	}
	algo := string(compressed[0:4])
	switch algo {
	case COMPRESSION_LZW:
		serialized, err = DecompressStateLzw(compressed[4:])
	case COMPRESSION_ZLIB:
		serialized, err = DecompressStateZlib(compressed[4:])
	default:
		logger.Errorf("compressed buffer does not encode an implemented algorithm")
		err = rgse.Create(rgse.GamestateByteDecompressAlgoError)
	}
	if err != nil {
		return []byte{}, err
	}
	duration := time.Now().Sub(startTime)
	logger.Debugf("DecompressState algo= %s input len= %d output len= %d in %.4fms", algo, len(compressed), len(serialized), duration.Seconds()*1000.0)
	return serialized, err
}

func CompressStateZlib(serialized []byte) (compressed []byte) {
	var b bytes.Buffer
	b.WriteString(COMPRESSION_ZLIB)
	w := zlib.NewWriter(&b)
	w.Write(serialized)
	w.Close()
	compressed = b.Bytes()
	return
}

func DecompressStateZlib(compressed []byte) (serialized []byte, rgserr rgse.RGSErr) {
	bin := bytes.NewBuffer(compressed)
	r, err := zlib.NewReader(bin)
	if err != nil {
		logger.Errorf("could not create zlib reader")
		rgserr = rgse.Create(rgse.GamestateByteDecompressAlgoError)
		return
	}
	var bout bytes.Buffer
	nb, err := io.Copy(&bout, r)
	if err != nil {
		logger.Errorf("could not decompress using zlib reader. input len= %dnb decompressed len=%dnb", len(compressed), nb)
		rgserr = rgse.Create(rgse.GamestateByteDecompressError)
		return
	}
	serialized = bout.Bytes()
	r.Close()
	return
}

func CompressStateLzw(serialized []byte) (compressed []byte) {
	var b bytes.Buffer
	b.WriteString(COMPRESSION_LZW)
	w := lzw.NewWriter(&b, lzw.LSB, 8)
	w.Write(serialized)
	w.Close()
	compressed = b.Bytes()
	return
}

func DecompressStateLzw(compressed []byte) (serialized []byte, rgserr rgse.RGSErr) {
	bin := bytes.NewBuffer(compressed)
	r := lzw.NewReader(bin, lzw.LSB, 8)
	var bout bytes.Buffer
	nb, err := io.Copy(&bout, r)
	if err != nil {
		logger.Errorf("could not decompress using lzw reader. input len= %dnb decompressed len=%dnb", len(compressed), nb)
		rgserr = rgse.Create(rgse.GamestateByteDecompressAlgoError)
		return
	}
	serialized = bout.Bytes()
	r.Close()
	return
}

func (g *GameV3) Init(token Token, wallet string, currency string) {
	g.Token = token
	g.Wallet = wallet
	g.Currency = currency
	g.EngineConf = engine.BuildEngineDefs(g.EngineId)
}

func CreateGameV3FromEngine(engineId string) (IGameV3, rgse.RGSErr) {
	switch engineId {
	case "mvgEngineRoulette1":
		return &GameRouletteV3{
			GameV3: GameV3{
				EngineId: engineId,
			},
		}, nil
	}
	return nil, rgse.Create(rgse.EngineNotFoundError)
}

func CreateGameV3(game string) (IGameV3, rgse.RGSErr) {
	engineId, rgserr := config.GetEngineFromGame(game)
	if rgserr != nil {
		return nil, rgserr
	}
	gameV3, rgserr := CreateGameV3FromEngine(engineId)
	if rgserr != nil {
		return nil, rgserr
	}
	gameV3.Base().Game = game
	return gameV3, nil
}
