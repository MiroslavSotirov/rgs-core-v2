package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerBattleOfMythsFreespin struct {
	FeatureDef
}

func (f *TriggerBattleOfMythsFreespin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerBattleOfMythsFreespin) DataPtr() interface{} {
	return nil
}

func (f *TriggerBattleOfMythsFreespin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerBattleOfMythsFreespin) OnInit(state *FeatureState) {
	state.Features = append(state.Features,
		&Config{
			FeatureDef: *f.DefPtr(),
			Data:       f.DefPtr().Params,
		})
}

func (f TriggerBattleOfMythsFreespin) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	var counter int
	var fstype string
	var scatterType int
	statefulStake := GetStatefulStakeMap(*state)
	if statefulStake.HasKey("counter") {
		counter = statefulStake.GetInt("counter")
		fstype = statefulStake.GetString("fstype")
	}

	scatterInc := params.GetInt("ScatterInc")
	scatterDec := params.GetInt("ScatterDec")
	scatterMin := params.GetInt("ScatterMin")
	scatterMax := params.GetInt("ScatterMax")
	numFreespins := params.GetInt("NumFreespins")

	var scatterTile int
	positions := []int{}

	if state.Action == "base" {
		// reset freespin after a freespin sequence
		fstype = ""
	}

	if state.Action == "cascade" {
		logger.Debugf("skipping placing scatters due to cascade action")
	} else if state.PureWins {
		logger.Debugf("skipping placing scatters due to wins")
	} else {

		if counter >= scatterMax || counter <= scatterMin {
			counter = 0
		}

		if params.HasKey("NumScatters") {
			numScatters := params.GetIntSlice("NumScatters")[WeightedRandomIndex(params.GetIntSlice("NumProbabilities"))]

			if (params.HasKey("RunTiger") && params.GetBool("RunTiger")) ||
				(params.HasKey("RunDragon") && params.GetBool("RunDragon")) ||
				(params.HasKey("RunPrincess") && params.GetBool("RunPrincess")) {
				logger.Debugf("skipping placing scatters due to conflicting features")
			} else {
				if rng.RandFromRange(10000) > params.GetInt("ScatterProbability") {
					logger.Debugf("skipping placing scatters dues to activation probability")
				} else {
					Abs := func(x int) int {
						if x < 0 {
							return -x
						}
						return x
					}

					if counter < 0 {
						scatterType = 0
					} else {
						scatterType = 1
					}

					sameTypeProbs := params.GetIntSlice("SameTypeProbabilities")
					absCounter := Abs(counter)
					if absCounter < len(sameTypeProbs) {
						sameProb := sameTypeProbs[absCounter]
						if rng.RandFromRange(10000) > sameProb {
							scatterType = scatterType ^ 1
						}
					}
					if scatterType == 0 {
						scatterTile = scatterDec
						counter -= numScatters
					} else {
						scatterTile = scatterInc
						counter += numScatters
					}

					var avoidTiles []int
					if params.HasKey("AvoidTiles") {
						avoidTiles = params.GetIntSlice("AvoidTiles")
					} else {
						avoidTiles = []int{}
					}
					gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])
					for sc := 0; sc < numScatters; sc++ {
						var reel, symb, pos int
						cont := true
						for cont {
							reel = rng.RandFromRange(gridw)
							symb = rng.RandFromRange(gridh)
							pos = reel*gridh + symb
							tile := state.SymbolGrid[reel][symb]
							cont = func() bool {
								for _, avoid := range avoidTiles {
									if tile == avoid {
										return true
									}
								}
								return false
							}()
							if !cont {
								break
							}
							cont = func() bool {
								for _, p := range positions {
									if p == pos {
										return true
									}
								}
								return false
							}()
						}
						logger.Debugf("adding scatter tile %d on reel %d symbol %d pos %d", scatterTile, reel, symb, pos)
						state.SymbolGrid[reel][symb] = scatterTile
						positions = append(positions, pos)
					}
				}
			}
		}

		if counter >= scatterMax || counter <= scatterMin {
			// add freespin win to state.Wins

			if numFreespins > 0 {
				fstype = "freespinE1"
				if counter >= scatterMax {
					fstype = "freespinE2"
				}

				state.Wins = append(state.Wins, FeatureWin{
					Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
					SymbolPositions: positions,
				})
				logger.Debugf("Trigger %d freespins of type %s", numFreespins, fstype)
			}
		}
	}

	SetStatefulStakeMap(*state, FeatureParams{
		"counter": counter,
		"fstype":  fstype,
	}, params)

	logger.Debugf("SetStatefulMap")

	if len(positions) > 0 {
		params["Positions"] = positions
		params["ReplaceWithId"] = scatterTile
		params["TileId"] = 0 // ignored
		activateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f TriggerBattleOfMythsFreespin) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerBattleOfMythsFreespin) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsFreespin) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
