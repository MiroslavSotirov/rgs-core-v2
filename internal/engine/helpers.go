package engine

import (
	// "crypto/md5"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
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

func (num Fixed) ValueAsFloat64() float64 {
	v, err := strconv.ParseFloat(num.ValueAsString(), 64)
	if err != nil {
		panic(fmt.Sprintf("could not parse currency value %#v", num))
	}
	return v
}

func (num Fixed) ValueAsString() string {
	// prints number with max 3 decimal places
	// this value is chosen due to the minimum currency value we support (3 decimal places)
	n := num.Abs()
	s := fmt.Sprintf("%d", n/fixedExp)
	d := fmt.Sprintf(".%06d", n%fixedExp)
	o := s + d[:4]
	if num.IsNeg() {
		return "-" + o
	}
	return o
}

func (num Fixed) StringFmt(p int) string {
	if p > 6 {
		p = 6
	}
	// prints as a string with p decimal places
	n := num.Abs()
	s := fmt.Sprintf("%d", n/fixedExp)
	d := fmt.Sprintf(".%06d", n%fixedExp)
	o := s + d[:p+1]
	if num.IsNeg() {
		return "-" + o
	}
	return o

}

func (num Fixed) IsNeg() bool {
	return int64(num) < 0
}

func (num Fixed) Abs() Fixed {
	if int64(num) < 0 {
		return Fixed(-int64(num))
	}
	return num
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
	// return num1 * num2 / fixedExp

	// E = 10^6
	// num1 = aE + b
	// num2 = cE + d
	// num1*num2/E = (acE^2 + (ad + bc)E + bd) / E = acE + ad + bc + bd/E = a(cE + d) + bc + (bd)/E
	var a = num1 / fixedExp
	var b = num1 - a*fixedExp
	var c = num2 / fixedExp
	var d = num2 - c*fixedExp
	return a*(c*fixedExp+d) + b*c + b*d/fixedExp
}

func (num1 Fixed) MulFloat(num2 Fixed) Fixed {
	return Fixed(float64(num1) / float64(fixedExp) * float64(num2))
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
	return Fixed(math.Round(float64(num) * float64(fixedExp)))
}

func NewFixedFromFloat64(num float64) Fixed {
	return Fixed(math.Round(num * float64(fixedExp)))
}

func NewFixedFromInt(num int) Fixed {
	return Fixed(int64(num) * int64(fixedExp))
}

type GameParams struct {
	Stake             Fixed     `json:"stake"`
	TotalStake        Fixed     `json:"totalStake"`
	SelectedWinLines  []int     `json:"selectedWinLines"`
	Selection         string    `json:"selectedFeature"`
	RespinReel        int       `json:"respinReel"`
	Action            string    `json:"action"`
	Game              string    `json:"game"`
	Wallet            string    `json:"wallet"`
	PreviousID        string    `json:"previousID"`
	Force             string    `json:"force"`
	previousGamestate Gamestate // this cannot be passed in
	//stopPostitions    []int     // this can also not be passed in from outside the package (only for testing)
}

//
//func (params *GameParams) SetPG(pg Gamestate) {
//// used for VT
//	params.previousGamestate = pg
//	return
//}

func (gp *GameParams) Decode(request *http.Request) (rgserr rgse.RGSErr) {
	decoder := json.NewDecoder(request.Body)
	decoderror := decoder.Decode(gp)

	if decoderror != nil {
		rgserr = rgse.Create(rgse.JsonError)
		rgserr.AppendErrorText(decoderror.Error())
	}
	return
}

// if this is the action, a wager must be charged
var paidActions = []string{"base", "maxBase", "respin", "gamble"}

func (p GameParams) Validate() (err rgse.RGSErr) {
	if p.Game == "" || p.Action == "" {
		return rgse.Create(rgse.BadConfigError)
	}
	return nil
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
func GetHashes() (IDStrings []string, MD5Strings []string, SHA1Strings []string, rgserr rgse.RGSErr) {
	//Initialize variable returnMD5String now in case an error has to be returned
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		rgserr = rgse.Create(rgse.EngineHashError)
		return
	}
	engineIDs, err := ioutil.ReadDir(filepath.Join(currentDir, "internal/engine/engineConfigs"))
	if err != nil {
		logger.Fatalf("Failed reading engineDefs")
		rgserr = rgse.Create(rgse.EngineHashError)
		return
	}
	for i := 0; i < len(engineIDs); i++ {
		if !strings.Contains(engineIDs[i].Name(), "mvgEngine") {
			logger.Infof("skipping checksum for file %v", engineIDs[i].Name())
			continue
		}
		logger.Infof("Generating checksums for file: %v", engineIDs[i].Name())
		filePath := filepath.Join(currentDir, "internal/engine/engineConfigs", engineIDs[i].Name())
		md5hash, sha1hash, err := GetHash(filePath)
		if err != nil {
			rgserr = rgse.Create(rgse.EngineHashError)
			return
		}
		logger.Debugf("MD5: %v SHA1: %v", md5hash, sha1hash)
		IDStrings = append(IDStrings, engineIDs[i].Name())
		MD5Strings = append(MD5Strings, md5hash)
		SHA1Strings = append(SHA1Strings, sha1hash)
	}
	// generate rng hashes
	if config.GlobalConfig.Local {
		logger.Infof("Generating checksums for rng")
		_, _, err = GetHash(filepath.Join(currentDir, "internal/rng/mt19937.go"))
		if err != nil {
			logger.Errorf("error generating checksum for rng: %v", err)
			rgserr = rgse.Create(rgse.EngineHashError)
			return
		}
	}
	return
}

func GetHash(filePath string) (string, string, error) {
	MD5Hash, err := GetHashFile(filePath, GetHashMd5)
	SHA1Hash, err := GetHashFile(filePath, GetHashSha1)
	return MD5Hash, SHA1Hash, err
}

func GetHashFile(filePath string, hashfunc func(file io.Reader) (string, error)) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return hashfunc(file)
}

func GetHashSha1(file io.Reader) (string, error) {
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

func GetHashMd5(file io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

func randomRangeInt32(min, max int) int32 {
	// cast to int32
	return int32(rng.RandFromRange(max-min+1) + min)
}

func (gamestate Gamestate) isFreespin() bool {
	def, err := gamestate.EngineDef()
	if err != nil {
		return false
	}
	if strings.Contains(def.ID, "freespin") {
		return true
	}
	return false
}

func GetMaxWin(e EngineConfig) {
	for i := 0; i < len(e.EngineDefs); i++ {
		logger.Infof("finding max win for def %v", i)
		ed := e.EngineDefs[i]
		if len(ed.Reels) != 5 {
			logger.Errorf("engine def %v does not have 5 reels", ed.ID)
			continue
		}
		var winLines []int
		for x := 0; x < len(ed.WinLines); x++ {
			winLines = append(winLines, x)
		}
		// set all multipliers to max

		// wilds
		var wilds []wild
		for j := 0; j < len(ed.Wilds); j++ {
			var max = 1
			for k := 0; k < len(ed.Wilds[j].Multiplier.Multipliers); k++ {
				if ed.Wilds[j].Multiplier.Multipliers[k] > max {
					max = ed.Wilds[j].Multiplier.Multipliers[k]
				}
			}
			wilds = append(wilds, wild{
				Symbol: ed.Wilds[j].Symbol,
				Multiplier: weightedMultiplier{
					Multipliers:   []int{max},
					Probabilities: []int{1},
				},
			})
		}
		ed.Wilds = wilds

		// multiplier
		var max = 1
		for j := 0; j < len(ed.Multiplier.Multipliers); j++ {
			if ed.Multiplier.Multipliers[j] > max {
				max = ed.Multiplier.Multipliers[j]
			}
		}
		ed.Multiplier = weightedMultiplier{
			Multipliers:   []int{max},
			Probabilities: []int{1},
		}

		parameters := GameParams{
			SelectedWinLines: winLines,
			Selection:        "freespin5:5",
		}

		//var maxCascades = 0
		//var maxCascadesStops []int
		var maxPayout = 0
		var maxGS Gamestate
		var ok bool
		// iterate over all possible stop positions
		// assume 5 reels
		var stopList = make([]int, 5)
		for j1 := 0; j1 < len(ed.Reels[0]); j1++ {
			stopList[0] = j1
			for j2 := 0; j2 < len(ed.Reels[1]); j2++ {
				stopList[1] = j2
				for j3 := 0; j3 < len(ed.Reels[2]); j3++ {
					stopList[2] = j3
					for j4 := 0; j4 < len(ed.Reels[3]); j4++ {
						stopList[3] = j4
						for j5 := 0; j5 < len(ed.Reels[4]); j5++ {
							stopList[4] = j5
							var lastGS Gamestate
							//var triggeringGS Gamestate
							var relativePayout int
							ed, _ = ed.SetForce(stopList)
							parameters.Action = ""
							call := reflect.ValueOf(ed).MethodByName(ed.Function)
							gamestateAndNextActions := call.Call([]reflect.Value{reflect.ValueOf(parameters)})
							lastGS, ok = gamestateAndNextActions[0].Interface().(Gamestate)
							if !ok {
								panic("value not a gamestate")
							}
							//triggeringGS = lastGS
							relativePayout += lastGS.RelativePayout * lastGS.Multiplier
							//cascades := 0
							//for len(lastGS.NextActions) > 0 && lastGS.NextActions[0] == "cascade"{
							//	ed.SetForce([]int{})
							//	//call := reflect.ValueOf(ed).MethodByName(ed.Function)
							//	parameters.Action = "cascade"
							//	cascades ++
							//	parameters.previousGamestate = lastGS
							//	gamestateAndNextActions := call.Call([]reflect.Value{reflect.ValueOf(parameters)})
							//	lastGS, ok = gamestateAndNextActions[0].Interface().(Gamestate)
							//	if !ok {
							//		panic("value not a gamestate")
							//	}
							//	relativePayout += lastGS.RelativePayout * lastGS.Multiplier
							//
							//}
							//if cascades > maxCascades {
							//	maxCascadesStops = stopList
							//	maxCascades = cascades
							//	logger.Warnf("cascaded %v times w/ stoplist %v", cascades, stopList)
							//}
							if relativePayout > maxPayout {
								maxPayout = relativePayout
								maxGS = lastGS //Gamestate{StopList:stopList}
								logger.Warnf("new max payout: %#v", lastGS)
							}

						}
					}
				}

			}
		}

		logger.Infof("Found max relative multiplier %v, %#v", maxPayout, maxGS) //,\nMax cascades %v: %v maxCascades, maxCascadesStops)

	}
}

func GetDefaultView(gameName string) (symbolGrid [][]int, features []features.Feature, defId int, reelsetId string) {
	logger.Debugf("GetDefaultView(%s)", gameName)
	e, err := GetEngineDefFromGame(gameName)
	if err != nil {
		return
	}
	method, id, err := e.getEngineAndMethodInternal("init", false)
	if err == nil {
		symbolGrid, features = GetDefaultViewFromFunction(method)
		defId, reelsetId = id, e.EngineDefs[id].ReelsetId
		return
	}
	logger.Debugf("could not find a config for \"init\" use reels from first defined config")

	for i := 0; i < len(e.EngineDefs[0].ViewSize); i++ {
		row := []int{}
		for j := 0; j < e.EngineDefs[0].ViewSize[i]; j++ {
			row = append(row, e.EngineDefs[0].Reels[i][j])
		}
		symbolGrid = append(symbolGrid, row)
	}
	return
}

func GetDefaultViewFromFunction(method reflect.Value) (symbolGrid [][]int, features []features.Feature) {
	logger.Debugf("call configured init function to generate default view")
	var params GameParams
	objs := method.Call([]reflect.Value{reflect.ValueOf(params)})
	gamestate, ok := objs[0].Interface().(Gamestate)
	if !ok {
		panic("value not a gamestate")
	}
	logger.Debugf("default symbolGrid= %v", gamestate.SymbolGrid)
	logger.Debugf("default features= %v", gamestate.Features)
	symbolGrid = gamestate.SymbolGrid
	features = gamestate.Features
	return
}

var MilliUnitCCies = []Ccy{Ccy_BTC, Ccy_EGP, Ccy_TND, Ccy_KWD, Ccy_BHD, Ccy_IQD, Ccy_OMR, Ccy_XBT, Ccy_ETH, Ccy_LTC, Ccy_BCH}

func RoundUpToNearestCCYUnit(in Money) (out Money) {
	// this function is basically a ceiling function, but if the value is zero then it will also be incremented to the smallest allowed currency value
	d := 10000
	out.Currency = in.Currency
	for ccy := 0; ccy < len(MilliUnitCCies); ccy++ {
		if in.Currency == MilliUnitCCies[ccy].String() {
			d /= 10
		}
	}

	// if any digit after the final keeper digit is not zero, round up
	r := int(in.Amount) % d
	if r == 0 && in.Amount != Fixed(0) {
		//this commented would make it so that the value 0.01 is rounded up to 0.02
		out.Amount = in.Amount
		return
	}

	out.Amount = in.Amount.Add(Fixed(d - r))
	return
}

func (gamestate Gamestate) GetChoices() (choices []string) {
	// returns allowed selections for a pickspins round
	if len(gamestate.NextActions) < 1 || gamestate.NextActions[0] != "pickSpins" {
		return
	}
	EC, err := gamestate.Engine()
	if err != nil {
		return
	}
	ED := EC.EngineDefs[EC.DefIdByName("pickSpins")]

	for i := 0; i < len(ED.SpecialPayouts); i++ {
		choices = append(choices, ED.SpecialPayouts[i].Index)
	}
	return
}

func (gamestate Gamestate) GetTtl() int64 {
	// returns number of seconds a completed round should stay open depending on features
	for _, action := range gamestate.NextActions {
		if action == "base" || action == "finish" {
			continue
		}
		_, ok := GamestatePB_Action_value[action]
		if ok {
			return 3600 * 24 * 2
		}
	}
	return 3600
}
