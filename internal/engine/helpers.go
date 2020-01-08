package engine

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Gamification stage randomizer lower and upper bound
const (
	max = 70
	min = 50
)

func SelectFromWeightedOptions(options []int, weights []int) int {
	// This method is used for multipliers, which are always integers
	// Select from a list (options) with weights (weights)
	if len(options) == 0 {
		return 1
	} else if len(options) == 1 {
		return options[0]
	}

	return options[GetWeightedIndex(weights)]
}

func GetWeightedIndex(weights []int) int {
	// randomly selects an index given a list of weights
	var weightsSum int
	for _, weight := range weights {
		weightsSum += weight
	}
	random := rng.RandFromRange(weightsSum) + 1 // number is in range [1,weightsSum], inclusive
	var optionIndex int
	for p := weights[0]; p < random; p += weights[optionIndex] {
		optionIndex++
	}
	return optionIndex
}

func GetCurrentWinAndStake(gamestate Gamestate) (Fixed, Fixed) {

	// check for a payout or stake transaction
	currentWinnings := Fixed(0)
	currentStake := Fixed(0) // on bonus rounds, stake needs to be stake of initiating round, be sure to overwrite this in renderer
	for _, transaction := range gamestate.Transactions {
		if transaction.Type == "PAYOUT" {
			currentWinnings = transaction.Amount.Amount
		} else if transaction.Type == "WAGER" {
			currentStake = transaction.Amount.Amount
		}
	}
	return currentWinnings, currentStake
}

func (num Fixed) ValueAsFloat() float32 {
	// nb this should only be used to get a value for printing or transactional purposes, never for further calculations
	return float32(num) / float32(fixedExp)
}

func (num Fixed) ValueAsInt() int32 {
	// nb this truncates the value
	return int32(num / fixedExp)
}

func (num Fixed) ValueRaw() int64 {
	// this returns the value of the Fixed as its internal integer representation
	return int64(num)
}

func (num Fixed) Bytes() []byte {
	// todo improve this
	asStr := strconv.Itoa(int(num))
	return []byte(asStr)
}

func NewFromBytes(val []byte) Fixed {
	asInt, _ := strconv.Atoi(string(val))
	return Fixed(asInt)
}

func (num1 Fixed) Mul(num2 Fixed) Fixed {
	// multiply two fixed point numbers with e6 representation
	// num1 = realnum1 * 10^6
	// num1 / 10^6 = realnum1
	// num2 = realnum2 * 10^6
	// num2 / 10^6 = realnum2
	// res = realres * 10^6
	// realres = realnum1 * realnum2 = num1 / 10^6 * num2 / 10^6
	// res = num1 / 10^6 * num2 / 10^6 * 10^6 = num1 * num2 / 10^6
	return num1 * num2 / fixedExp
}

func (num1 Fixed) Div(num2 Fixed) Fixed {
	// divide two fixed point numbers with e6 representation
	// num1 = realnum1 * 10^6
	// num1 / 10^6 = realnum1
	// num2 = realnum2 * 10^6
	// num2 / 10^6 = realnum2
	// res = realres * 10^6
	// realres = realnum1 / realnum2 = num1 / 10^6 / num2 * 10^6
	// res = num1 / 10^6 / num2 * 10^6 * 10^6 = num1 / num2 * 10^6
	return num1 * fixedExp / num2
}

func (num1 Fixed) Add(num2 Fixed) Fixed {
	return num1 + num2
}

func (num1 Fixed) Sub(num2 Fixed) Fixed {
	return num1 - num2
}

func (num Fixed) Pow(exp int) Fixed {
	// num1 = realnum1 * 10^6
	// num1 / 10^6 = realnum1

	// res = realres * 10^6
	// realres = realnum1 * realnum1 * ... = num / 10^6 * num / 10^6 * ... = num ^ exp / (10^6 )^exp
	// res =  num ^ exp / (10^6 )^exp * 10^6 = num ^exp / (10^6)^(exp-1)
	res := num
	for i := 1; i < exp; i++ {
		res = res.Mul(num)
	}
	return res
}

func NewFixedFromFloat(num float32) Fixed {
	return Fixed(num * float32(fixedExp))
}

func NewFixedFromInt(num int) Fixed {
	return Fixed(int64(num) * int64(fixedExp))
}

type GameParams struct {
	Stake             Fixed     `json:"stake"`
	SelectedWinLines  []int     `json:"selectedWinLines"`
	Selection         string    `json:"selectedFeature"`
	RespinReel        int       `json:"respinReel"`
	Action            string    `json:"action"`
	previousGamestate Gamestate // this cannot be passed in
}

func GetGameIDFromPB(gameID string) string {
	// switch all uppercase to lowercase, all underscore to dash
	gameID = strings.ToLower(gameID)
	return strings.ReplaceAll(gameID, "_", "-")
}

func GetPBFromGameID(gameID string) string {
	// switch all lowercase to uppercase, all dash to underscore
	gameID = strings.ToUpper(gameID)
	return strings.ReplaceAll(gameID, "-", "_")
}

// get engine hashes
func GetHashes() ([]string, []string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var MD5Strings []string
	var SHA1Strings []string
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		return []string{}, []string{}, err
	}
	engineIDs, err := ioutil.ReadDir(filepath.Join(currentDir, "internal/engine/engineConfigs"))
	if err != nil {
		logger.Fatalf("Failed reading engineDefs")
		return []string{}, []string{}, err
	}
	for i := 0; i < len(engineIDs); i++ {
		logger.Infof("Generating checksums for file: %v", engineIDs[i].Name())
		filePath := filepath.Join(currentDir, "internal/engine/engineConfigs", engineIDs[i].Name())
		md5hash, sha1hash, err := GetHash(filePath)
		if err != nil {
			return []string{}, []string{}, err
		}
		MD5Strings = append(MD5Strings, md5hash)
		SHA1Strings = append(SHA1Strings, sha1hash)
	}
	// generate rng hashes
	logger.Infof("Generating checksums for rng")
	_, _, err = GetHash(filepath.Join(currentDir, "internal/rng/mt19937.go"))
	if err != nil {
		return []string{}, []string{}, err
	}
	return MD5Strings, SHA1Strings, nil

}

func GetHash(filePath string) (string, string, error) {

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}

	//Open a new hash interface to write to
	hash1 := md5.New()
	hash2 := sha1.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash1, file); err != nil {
		return "", "", err
	}
	if _, err := io.Copy(hash2, file); err != nil {
		return "", "", err
	}
	//Get the 16 bytes hash
	hashInBytes1 := hash1.Sum(nil)
	hashInBytes2 := hash2.Sum(nil)
	logger.Infof("MD5: %v", hex.EncodeToString(hashInBytes1))
	logger.Infof("SHA1: %v", hex.EncodeToString(hashInBytes2))

	//Convert the bytes to a string
	MD5String := hex.EncodeToString(hashInBytes1)
	SHA1String := hex.EncodeToString(hashInBytes2)
	err = file.Close()
	if err != nil {
		return "", "", err
	}
	return MD5String, SHA1String, nil
}

func randomRangeInt32() int32 {
	// cast to int32
	return int32(rng.RandFromRange(max-min+1) + min)
}
