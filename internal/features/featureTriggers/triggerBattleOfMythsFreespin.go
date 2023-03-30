package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN = "TriggerBattleOfMythsFreespin"

	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_INC             = "ScatterInc"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_DEC             = "ScatterDec"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_MIN             = "ScatterMin"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_MAX             = "ScatterMax"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_FREESPINS           = "NumFreespins"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_SCATTERS            = "NumScatters"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_PROBABILITIES       = "NumProbabilities"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SAME_TYPE_PROBABILITIES = "SameTypeProbabilities"
	PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_AVOID_TILES             = "AvoidTiles"

	STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_COUNTER        = "counter"
	STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_FSTYPE         = "fstype"
	STATEFUL_VALUE_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_E1 = "freespinE1"
	STATEFUL_VALUE_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_E2 = "freespinE2"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN, func() feature.Feature { return new(TriggerBattleOfMythsFreespin) })

type TriggerBattleOfMythsFreespin struct {
	feature.Base
}

func (f *TriggerBattleOfMythsFreespin) OnInit(state *feature.FeatureState) {
	state.Features = append(state.Features,
		&feature.Config{
			Base: feature.Base{FeatureDef: *f.DefPtr()},
			Data: f.DefPtr().Params,
		})
}

func (f TriggerBattleOfMythsFreespin) Trigger(state *feature.FeatureState, params feature.FeatureParams) {
	var counter int
	var fstype string
	var scatterType int
	statefulStake := feature.GetStatefulStakeMap(*state)
	if statefulStake.HasKey(STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_COUNTER) {
		counter = statefulStake.GetInt(STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_COUNTER)
		fstype = statefulStake.GetString(STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_FSTYPE)
	}

	scatterInc := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_INC)
	scatterDec := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_DEC)
	scatterMin := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_MIN)
	scatterMax := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SCATTER_MAX)
	numFreespins := params.GetInt(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_FREESPINS)

	var scatterTile int
	positions := []int{}

	if state.Action == "base" {
		// reset freespin after a freespin sequence
		fstype = ""
	}

	if state.Action == "cascade" {
		// logger.Debugf("skipping placing scatters due to cascade action")
	} else if params.HasKey("PureWins") {
		// logger.Debugf("skipping placing scatters due to wins")
	} else {

		if counter >= scatterMax || counter <= scatterMin {
			counter = 0
		}

		if params.HasKey(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_SCATTERS) {
			numScatters := params.GetIntSlice(
				PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_SCATTERS)[feature.WeightedRandomIndex(
				params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_NUM_PROBABILITIES))]

			if (params.HasKey(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_TIGER) && params.GetBool(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_TIGER)) ||
				(params.HasKey(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_DRAGON) && params.GetBool(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_DRAGON)) ||
				(params.HasKey(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_PRINCESS) && params.GetBool(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_RUN_PRINCESS)) {
				// logger.Debugf("skipping placing scatters due to conflicting features")
			} else {
				if rng.RandFromRangePool(10000) > params.GetInt("ScatterProbability") {
					// logger.Debugf("skipping placing scatters dues to activation probability")
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

					sameTypeProbs := params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_SAME_TYPE_PROBABILITIES)
					absCounter := Abs(counter)
					if absCounter < len(sameTypeProbs) {
						sameProb := sameTypeProbs[absCounter]
						if rng.RandFromRangePool(10000) > sameProb {
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
					if params.HasKey(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_AVOID_TILES) {
						avoidTiles = params.GetIntSlice(PARAM_ID_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_AVOID_TILES)
					} else {
						avoidTiles = []int{}
					}
					gridw, gridh := len(state.SymbolGrid), len(state.SymbolGrid[0])
					for sc := 0; sc < numScatters; sc++ {
						var reel, symb, pos int
						cont := true
						for cont {
							reel = rng.RandFromRangePool(gridw)
							symb = rng.RandFromRangePool(gridh)
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
						state.SymbolGrid[reel][symb] = scatterTile
						positions = append(positions, pos)
					}
				}
			}
		}

		if counter >= scatterMax || counter <= scatterMin {
			// add freespin win to state.Wins

			if numFreespins > 0 {
				fstype = STATEFUL_VALUE_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_E1
				if counter >= scatterMax {
					fstype = STATEFUL_VALUE_TRIGGER_BATTLE_OF_MYTHS_FREESPIN_E2
				}

				state.Wins = append(state.Wins, feature.FeatureWin{
					Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
					SymbolPositions: positions,
				})
				logger.Debugf("Trigger %d freespins of type %s", numFreespins, fstype)
			}
		}
	}

	feature.SetStatefulStakeMap(*state, feature.FeatureParams{
		STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_COUNTER: counter,
		STATEFUL_ID_TRIGGER_BATTLE_OF_MYTHS_FSTYPE:  fstype,
	}, params)

	if len(positions) > 0 {
		params[featureProducts.PARAM_ID_REPLACE_TILE_POSITIONS] = positions
		params[featureProducts.PARAM_ID_REPLACE_TILE_REPLACE_WITH_ID] = scatterTile
		params[featureProducts.PARAM_ID_REPLACE_TILE_TILE_ID] = 0 // ignored
		feature.ActivateFeatures(f.FeatureDef, state, params)
	}
	return
}

func (f *TriggerBattleOfMythsFreespin) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerBattleOfMythsFreespin) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
