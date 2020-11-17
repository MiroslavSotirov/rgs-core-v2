package volumeTester

import (
	"encoding/csv"
	"fmt"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type defInfo struct {
	totalWin   engine.Fixed
	totalPlays int
	prizes     map[string]*prizeInfo
	maxWin     engine.Fixed
}

type prizeInfo struct {
	totalWin int
	hits     int
	maxWin   int
}

func (d *defInfo) addPlay(win engine.Fixed) bool {
	d.totalWin += win
	d.totalPlays++
	if win > d.maxWin {
		d.maxWin = win
		return true
	}
	return false
}

func (d *defInfo) addWins(p []engine.Prize, mul int) {
	if d.prizes == nil {
		d.prizes = make(map[string]*prizeInfo)
	}

	for i := 0; i < len(p); i++ {
		if d.prizes[p[i].Index] == nil {
			d.prizes[p[i].Index] = &prizeInfo{}
		}
		win := p[i].Payout.Multiplier * p[i].Multiplier
		d.prizes[p[i].Index].addWin(win * mul)
	}
}

func (p *prizeInfo) addWin(win int) {
	p.hits++
	p.totalWin += win
	if win > p.maxWin {
		p.maxWin = win
	}
}

func GetMaxes(engineID string) {
	refTime := time.Now()
	engineConf := engine.BuildEngineDefs(engineID)
	logger.Infof("finding max values for engine %v", engineID)
	engine.GetMaxWin(engineConf)
	logger.Infof("took %v", time.Now().Sub(refTime))
}

func VolumeTestEngine(engineID string, numPlays int, chunks int, perSpin bool) ([]string, bool) {
	refTime := time.Now()
	failed := true
	totalWin := engine.Fixed(0)
	totalBet := engine.Fixed(0)
	var writer *csv.Writer
	if perSpin == true {
		outFile := fmt.Sprintf("%v_%v.csv", engineID, time.Now().Format("2006-01-02"))
		file, err := os.Create(outFile)
		if err != nil {
			perSpin = false
			logger.Errorf("Could not open file %v -- not saving perSpin results", outFile)
		} else {
			defer file.Close()
			writer = csv.NewWriter(file)
		}
		defer writer.Flush()
	}
	engineConf := engine.BuildEngineDefs(engineID)
	if numPlays == 0 {
		numPlays = engineConf.NumSpinsStat()
	}

	s2 := variance{xBar: float64(engineConf.RTP), betMult: float64(engineConf.EngineDefs[0].StakeDivisor)}

	logger.Infof("Running %v spins for engine %v", numPlays, engineID)

	chunkSize := numPlays / chunks
	infoStructs := make([]defInfo, len(engineConf.EngineDefs)*2)

	initString := fmt.Sprintf("Running %v spins in %v chunks for %v \n Expected RTP: %v \n Volatility: %v\n", numPlays, chunks, engineID, engineConf.RTP, engineConf.Volatility)
	vtInfo := []string{initString, "Chunk || RTP || RTP Feature || RTP base \n"}
	featureWin := engine.Fixed(0)
	respinWin := engine.Fixed(0)
	respinBet := engine.Fixed(0)
	featureHits := 0
	ctCascades := 0
	previousGamestate := engine.Gamestate{NextActions: []string{"finish"}, Game: getMatchingGame(engineID), DefID: 0, NextGamestate: "FirstSpinVT" + engineID}
	for i := 0; i < chunks; i++ {

		for j := 0; j < chunkSize; j++ {
			var params engine.GameParams
			//if strings.Contains(previousGamestate.Action,"freespin") || (j == 0 && i == 0) {
				params.Action = "base" // change this to maxBase or to any other special function for a particular wallet to see special RTP
			//} else {
			//	params.Action = "respin"
			//	params.RespinReel = rng.RandFromRange(5)
			//	logger.Warnf("Need to uncomment the next line and the SetPG function for this to work for respin engine")
			//	//params.SetPG(previousGamestate)
			//}
			if previousGamestate.NextActions[0] == "cascade" {
				params.Action = "cascade"
			}
			if previousGamestate.NextActions[0] == "pickSpins" {
				// user action is required (we are assuming here this is engine II, update later if more choice engines added)
				params.Selection = []string{"freespin25:25", "freespin10:10", "freespin5:5"}[engine.SelectFromWeightedOptions([]int{0, 1, 2}, []int{1, 1, 1})]
				// we do not add any selected win lines, always assume all lines. NB: ENGINE X has variable RTP based on selected win lines
			}
			gamestate, _ := engine.Play(previousGamestate, engine.Fixed(1000), "BTC", params)
			currentWinnings, currentStake := engine.GetCurrentWinAndStake(gamestate)
			//if strings.Contains(gamestate.Action, "freespin") {
			//	logger.Warnf("winnings: %v, gamesate: %v", currentWinnings, gamestate)
			//
			//}
			totalWin += currentWinnings
			totalBet += currentStake
			s2.addSample(float64(currentWinnings.ValueAsFloat()))
			// compile hit frequencies
			defID := gamestate.DefID
			if gamestate.Action == "cascade" {
				defID += len(engineConf.EngineDefs)
			}
			newmax := infoStructs[defID].addPlay(currentWinnings)
			if newmax {
				logger.Debugf("new max for engine %v: %#v", defID, gamestate)
			}
			infoStructs[defID].addWins(gamestate.Prizes, gamestate.Multiplier)

			if gamestate.Action != "base" {
				if gamestate.Action == "cascade" {
					ctCascades++
				}
				featureWin = featureWin.Add(currentWinnings)
				if gamestate.Action == "respin" {
					respinBet += currentStake
					respinWin += currentWinnings
				}
			}
			if perSpin == true {
				err := writer.Write([]string{fmt.Sprintf("%v,%v,%v,%v,%v,%v", defID, gamestate.Action, time.Now().Format("02 Jan 06 15:04 MST"), currentStake, currentWinnings, gamestate.StopList)})
				if err != nil {
					logger.Errorf("error writing to csv: %v", err)
					perSpin = false
					logger.Errorf("stoping perspin results")
				}
				//if len(gamestate.RecoveryGrid) > 0 {
				//	for r:=0; r<len(gamestate.RecoveryGrid); r++{
				//
				//		err = writer.Write([]string{fmt.Sprintf("%v,%v,%v,%v -- DISCARDED", defID, gamestate.Action, time.Now().Format("02 Jan 06 15:04 MST"), gamestate.RecoveryGrid[r])})
				//		if err != nil {
				//			logger.Errorf("error writing to csv: %v", err)
				//			perSpin = false
				//			logger.Errorf("stoping perspin results")
				//		}
				//	}
				//}
			}

			previousGamestate = gamestate
		}
		RTP := totalWin.Div(totalBet)
		RTPBase := totalWin.Sub(featureWin).Div(totalBet)
		RTPFeature := featureWin.Div(totalBet)
		//RTPRespin := respinWin.Div(respinBet)
		// fsPct := float64(fsTriggers) / (float64(chunkSize) * float64(i+1))
		ftInfo := ""
		//logger.Infof("avg feature multiplier: %v%%", float64(featureMultiplier)/float64(ftTriggers[1]["rounds"])*100)
		for rsID, reelsetInfo := range infoStructs {
			//logger.Infof("info: %v", reelsetInfo)
			if reelsetInfo.totalPlays == 0 {
				continue
			}
			defID := rsID
			if rsID >= len(engineConf.EngineDefs) {
				defID = rsID - len(engineConf.EngineDefs)
			}
			ftInfo += fmt.Sprintf("\n\nEngine: %v | Expected RTP: %.2f | Actual RTP: %f | Payout per round: %.2f || Max payout: %v || Rounds: %v \n", rsID, engineConf.EngineDefs[defID].RTP*100, reelsetInfo.totalWin.Div(totalBet).ValueAsFloat()*100, reelsetInfo.totalWin.Div(engine.Fixed(reelsetInfo.totalPlays)).ValueAsFloat(), reelsetInfo.maxWin, reelsetInfo.totalPlays)
			for x, prizeInfo := range reelsetInfo.prizes {
				ftInfo += fmt.Sprintf("%v ==  %.2f%% | RTP %.2f%% | max win: %v\n", x, float64(prizeInfo.hits)/float64(reelsetInfo.totalPlays)*100, engine.Fixed(prizeInfo.totalWin).Div(totalBet).ValueAsFloat()*100, prizeInfo.maxWin)
			}
		}
		logger.Infof("FS HIT RATE: %v in %v", featureHits, infoStructs[0].totalPlays)

		chunkInfo := fmt.Sprintf(" %v | RTP: %v%% | Feature: %v%% | Base: %v%% | Variance: %.5f | \n %v \n", i+1, RTP.ValueAsFloat()*100., RTPFeature.ValueAsFloat()*100., RTPBase.ValueAsFloat()*100., s2.getS2(), ftInfo)
		vtInfo = append(vtInfo, chunkInfo)
		//float64(featureMultiplier)/float64(ftTriggersFeature["rounds"]), float64(wildCounts)/float64(ftTriggersFeature["rounds"])

		if math.Abs(float64(RTP.Sub(engine.NewFixedFromFloat(engineConf.RTP)).ValueAsFloat())) > 0.01 {
			logger.Warnf("WARNING : RTP DEVIANT (%.2f%%)", RTP.ValueAsFloat()*100)
			logger.Infof(chunkInfo)
		} else {
			failed = false
			if i == chunks-1 {
				logger.Infof(chunkInfo)
			}
		}
		//logger.Warnf("Respin RTP: %v", RTPRespin.ValueAsFloat()*100)

		//logger.Warnf("%v cascades total", ctCascades)
		logger.Infof("Chunk %v done in %v", i+1, time.Now().Sub(refTime))
		refTime = time.Now()
	}
	return vtInfo, failed
}

func getMatchingGame(engineID string) string {
	// function to get a game name that matches the given engine
	for i := 0; i < len(config.GlobalGameConfig); i++ {
		if config.GlobalGameConfig[i].EngineID == engineID {
			return config.GlobalGameConfig[i].Games[0]
		}
	}
	return ""
}

type variance struct {
	xBar    float64
	sumDist float64
	n       int
	betMult float64
}

func (v *variance) addSample(val float64) {
	val = val / v.betMult
	v.sumDist += math.Pow(val-v.xBar, 2)
	v.n++
}

func (v *variance) getS2() float64 {
	if v.n < 2 {
		return 0
	}
	return v.sumDist / float64(v.n)
}

func RunVT(engineID string, spins int, chunks int, perSpin bool, maxes bool) (failed bool) {
	// Run VT from command line

	var results []string
	failed = false
	if engineID != "" {
		// run VT on one engine
		if engineID == "RNG" {
			err := TestRNG()
			if err != nil {
				logger.Errorf("Error running RNG output: %v", err)
				return true
			}
			return false
		}
		if maxes {
			GetMaxes(engineID)
		} else {
			results, failed = VolumeTestEngine(engineID, spins, chunks, perSpin)
		}
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			logger.Fatalf("Failed opening current directory")
			return true
		}
		engineIDs, err := ioutil.ReadDir(filepath.Join(currentDir, "internal/engine/engineConfigs"))
		if err != nil {
			logger.Fatalf("Failed reading engineDefs")
			return true
		}
		for _, engines := range engineIDs {
			if strings.Split(engines.Name(), ".")[0] == "mvgEngineXI" {
				continue
			}
			if maxes {
				GetMaxes(strings.Split(engines.Name(), ".")[0])
			} else {
				newResults, newFail := VolumeTestEngine(strings.Split(engines.Name(), ".")[0], spins, chunks, perSpin)
				failed = failed || newFail
				results = append(results, newResults...)
			}

		}
	}
	//fmt.Print(results)
	return
}

func GetVTInfo() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed opening current directory")
	}
	engineIDs, err := ioutil.ReadDir(filepath.Join(currentDir, "internal/engine/engineConfigs"))
	if err != nil {
		log.Fatal("Failed reading engineDefs")
	}
	for _, engines := range engineIDs {
		engineConf := engine.BuildEngineDefs(strings.Split(engines.Name(), ".")[0])
		//glog.Infof("EngineConf: %v", engineConf)
		logger.Infof("EngineID: %v", strings.Split(engines.Name(), ".")[0])
		logger.Infof("NumSpins: %v", engineConf.NumSpinsStat())
		logger.Infof("Volatility: %v", engineConf.Volatility)

	}

}

func TestRNG() error {
	//Generate 3 x 3m rows of each
	ranges := []int{33, 66, 99, 500, 999, 5, 36, 51}
	for i := 0; i < len(ranges); i++ {
		err := Gen3x3MData(ranges[i])
		if err != nil {
			return err
		}
	}
	err := Generate3x3MHex()
	if err != nil {
		return err
	}
	return Generate3x3MDecks()
}

func Gen3x3MData(max int) error {
	logger.Infof("Generating three times three million rows of random data from 0 to %v", max)
	// Generate 3 x 3m rows of randomly drawn data from 0 to max, inclusive
	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("RNG_Output_0_%v_%v__%v.csv", max, time.Now().Format("2006-01-02"), i)
		file, err := os.Create(fileName)
		if err != nil {
			logger.Errorf("Could not open file %v", fileName)
			return fmt.Errorf("RNG Output error: %v", err)
		}
		writer := csv.NewWriter(file)

		for j := 1; j <= 3000000; j++ {
			// add one line to the file
			err = writer.Write([]string{fmt.Sprintf("%v", rng.RandFromRange(max+1))})
			if err != nil {
				logger.Errorf("error writing to csv: %v", err)
				return err
			}
		}
		writer.Flush()
		err = file.Close()
		if err != nil {
			logger.Errorf("error writing to csv: %v", err)
			return err
		}
	}
	return nil
}

func ShuffleDeck() (shuffledDeck []string) {
	var deck []string
	// shuffles and returns a standard deck of 56 cards
	suits := []string{"h", "d", "s", "c"}
	values := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	for i := 0; i < len(suits); i++ {
		for j := 0; j < len(values); j++ {
			deck = append(deck, fmt.Sprintf("%v%v", values[j], suits[i]))
		}
	}
	for i := 0; i < 52; i++ {
		// choose randomly a card from the remaining cards and add to the new deck
		chosenCard := rng.RandFromRange(len(deck))
		shuffledDeck = append(shuffledDeck, deck[chosenCard])
		logger.Debugf("shuffledDeck: %v", shuffledDeck)
		// remove from the old deck
		if chosenCard == len(deck) {
			deck = deck[:chosenCard]
		} else {
			deck = append(deck[:chosenCard], deck[chosenCard+1:]...)
		}
		logger.Debugf("deck: %v", deck)

	}
	return
}

func Generate3x3MDecks() error {
	// Generate 3 x 3m rows of randomly shuffled decks of 56 cards
	logger.Infof("Generating 3 x 3 million shuffled decks")
	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("RNG_Output_Decks_%v__%v.csv", time.Now().Format("2006-01-02"), i)
		file, err := os.Create(fileName)
		if err != nil {
			logger.Errorf("Could not open file %v", fileName)
			return fmt.Errorf("RNG Output error: %v", err)
		}
		writer := csv.NewWriter(file)

		for j := 1; j <= 3000000; j++ {
			// add one line to the file
			err = writer.Write(ShuffleDeck())
			if err != nil {
				logger.Errorf("error writing to csv: %v", err)
				return err
			}
		}
		writer.Flush()
		err = file.Close()
		if err != nil {
			logger.Errorf("error writing to csv: %v", err)
			return err
		}
	}
	return nil
}

func Generate3x3MHex() error {
	// generates 3 files, each file containing 3m rows of data, 16 values each row
	logger.Infof("Generating 3 x 3 million rows of hex values 0-255")
	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("RNG_Output_Hex_%v__%v.csv", time.Now().Format("2006-01-02"), i)
		file, err := os.Create(fileName)
		if err != nil {
			logger.Errorf("Could not open file %v", fileName)
			return fmt.Errorf("RNG Output error: %v", err)
		}
		writer := csv.NewWriter(file)

		for j := 1; j <= 3000000; j++ {
			// add one line to the file
			hexes := []string{}
			for k := 0; k < 16; k++ {
				hexes = append(hexes, fmt.Sprintf("%02x", rng.RandFromRange(256)))
			}
			err = writer.Write(hexes)
			if err != nil {
				logger.Errorf("error writing to csv: %v", err)
				return err
			}
		}
		writer.Flush()
		err = file.Close()
		if err != nil {
			logger.Errorf("error writing to csv: %v", err)
			return err
		}
	}
	return nil
}
