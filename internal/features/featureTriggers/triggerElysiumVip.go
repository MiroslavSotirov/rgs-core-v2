package featureTriggers

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/featureProducts"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	FEATURE_ID_TRIGGER_ELYSIUM_VIP = "TriggerElysiumVip"

	PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID = "WildId"

	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL     = "level"
	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS   = "inserts"
	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS = "originals"
	STATEFUL_ID_TRIGGER_ELYSIUM_VIP_EMPLACED  = "emplaced"
)

var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_TRIGGER_ELYSIUM_VIP, func() feature.Feature { return new(TriggerElysiumVip) })

type TriggerElysiumVip struct {
	feature.Base
}

func (f TriggerElysiumVip) Trigger(state *feature.FeatureState, params feature.FeatureParams) {

	if f.ForceTrigger(state, params) {
		logger.Debugf("force %s was applied so no more features will be executed", params.GetString("force"))
		return
	}

	level := 0
	inserts := []int{}
	originals := []int{0, 1, 2}
	emplaced := []int{0, 1, 2}
	if state.Action != "base" {
		statefulStake := feature.GetStatefulStakeMap(*state)
		logger.Debugf("statefulStake: %#v", statefulStake)
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL) {
			level = statefulStake.GetInt(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL)
		}
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS) {
			inserts = feature.ConvertIntSlice(statefulStake.GetSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS))
		}
		if statefulStake.HasKey(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_EMPLACED) {
			emplaced = statefulStake.GetIntSlice(STATEFUL_ID_TRIGGER_ELYSIUM_VIP_EMPLACED)
			originals = make([]int, len(emplaced))
			for i, v := range emplaced {
				originals[i] = v
			}
		}
		if len(state.SymbolGrid) != len(state.Stateful.SymbolGrid)+len(inserts) {
			panic(fmt.Sprintf("number of reels %d is not last spin num %d plus num inserts %d",
				len(state.SymbolGrid), len(state.Stateful.SymbolGrid), len(inserts)))
		}

		logger.Debugf("copying wilds from last spin")
		wildId := params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID)
		ireel := 0
		for ilreel, lreel := range state.Stateful.SymbolGrid {
			if func() bool {
				for _, ins := range inserts {
					if ins == ilreel {
						return true
					}
				}
				return false
			}() {
				ireel++
			}
			for isym, lsym := range lreel {
				if lsym == wildId {
					logger.Debugf("setting reel %d row %d to %d", ireel, isym, wildId)
					state.SymbolGrid[ireel][isym] = wildId
				}
			}
			ireel++
		}
		logger.Debugf("%v", state.SymbolGrid)
		//		inserts := []int{}
	}

	feature.SetStatefulStakeMap(*state, feature.FeatureParams{
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL:     level,
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS:   inserts,
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS: originals,
		STATEFUL_ID_TRIGGER_ELYSIUM_VIP_EMPLACED:  emplaced},
		params)

	feature.ActivateFeatures(f.FeatureDef, state, params)
	return
}

func (f TriggerElysiumVip) ForceTrigger(state *feature.FeatureState, params feature.FeatureParams) bool {

	setGrids := func(grid [][]int) {
		setGrid(state.SymbolGrid, grid)
		setGrid(state.SourceGrid, grid)
	}
	replaceTiles := func(pos []int) {
		replaceTile := feature.MakeFeature(featureProducts.FEATURE_ID_REPLACE_TILE)
		logger.Debugf("REPLACETILE: %#v", replaceTile)

		state.Features = append(state.Features, &featureProducts.ReplaceTile{
			Base: feature.Base{FeatureDef: feature.FeatureDef{Type: featureProducts.FEATURE_ID_REPLACE_TILE}},
			Data: featureProducts.ReplaceTileData{
				Positions:     pos,
				TileId:        params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID),
				ReplaceWithId: params.GetInt(PARAM_ID_TRIGGER_ELYSIUM_VIP_WILD_ID),
			},
		})
	}
	respin := func(action string) {
		state.Wins = append(state.Wins, feature.FeatureWin{
			Index: fmt.Sprintf("%s:%d", action, 1),
		})
	}
	setState := func(level int, inserts []int, originals []int, emplaced []int) {
		feature.SetStatefulStakeMap(*state, feature.FeatureParams{
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_LEVEL:     level,
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_INSERTS:   inserts,
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_ORIGINALS: originals,
			STATEFUL_ID_TRIGGER_ELYSIUM_VIP_EMPLACED:  emplaced},
			params)

		state.Features = append(state.Features, &feature.StatefulMap{
			Base: feature.Base{FeatureDef: feature.FeatureDef{Type: feature.FEATURE_ID_STATEFUL_MAP}},
			Data: feature.StatefulMapData{Map: params.GetParams("StatefulMap")},
		})
	}

	if config.GlobalConfig.DevMode && params.HasKey("force") {
		switch {
		case params.HasForce("maxforce1"):
			setGrids([][]int{
				{8, 2, 6},
				{8, 6, 2},
				{3, 2, 7},
			})
			replaceTiles([]int{0, 3})
			respin("respinall2")
			setState(1, []int{1}, []int{0, 1, 2}, []int{0, 2, 3})
			return true
		case params.HasForce("maxforce2"):
			setGrids([][]int{
				{8, 1, 1},
				{8, 6, 6},
				{8, 5, 4},
				{6, 6, 3},
			})
			replaceTiles([]int{3})
			respin("respinall2")
			setState(2, []int{}, []int{0, 2, 3}, []int{0, 2, 3})
			return true
		case params.HasForce("maxforce3"):
			setGrids([][]int{
				{8, 2, 7},
				{8, 2, 5},
				{8, 7, 4},
				{8, 2, 1},
			})
			replaceTiles([]int{9})
			respin("respinall3")
			setState(3, []int{3}, []int{0, 2, 3}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce4"):
			setGrids([][]int{
				{8, 1, 2},
				{8, 7, 7},
				{8, 5, 1},
				{8, 2, 3},
				{8, 6, 1},
			})
			replaceTiles([]int{9})
			respin("respinall3")
			setState(4, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce5"):
			setGrids([][]int{
				{8, 6, 3},
				{8, 2, 5},
				{8, 6, 3},
				{8, 4, 2},
				{8, 4, 8},
			})
			replaceTiles([]int{14})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce6"):
			setGrids([][]int{
				{8, 2, 2},
				{8, 7, 5},
				{8, 3, 3},
				{8, 7, 8},
				{8, 6, 8},
			})
			replaceTiles([]int{11})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce7"):
			setGrids([][]int{
				{8, 4, 5},
				{8, 4, 8},
				{8, 2, 2},
				{8, 5, 8},
				{8, 6, 8},
			})
			replaceTiles([]int{5})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce8"):
			setGrids([][]int{
				{8, 2, 3},
				{8, 6, 8},
				{8, 5, 8},
				{8, 2, 8},
				{8, 4, 8},
			})
			replaceTiles([]int{8})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce9"):
			setGrids([][]int{
				{8, 7, 8},
				{8, 1, 8},
				{8, 1, 8},
				{8, 1, 8},
				{8, 4, 8},
			})
			replaceTiles([]int{2})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce10"):
			setGrids([][]int{
				{8, 7, 8},
				{8, 7, 8},
				{8, 5, 8},
				{8, 8, 8},
				{8, 1, 8},
			})
			replaceTiles([]int{10})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce11"):
			setGrids([][]int{
				{8, 2, 8},
				{8, 4, 8},
				{8, 8, 8},
				{8, 8, 8},
				{8, 6, 8},
			})
			replaceTiles([]int{7})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce12"):
			setGrids([][]int{
				{8, 6, 8},
				{8, 7, 8},
				{8, 8, 8},
				{8, 8, 8},
				{8, 8, 8},
			})
			replaceTiles([]int{13})
			respin("respinall3")
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		case params.HasForce("maxforce13"):
			setGrids([][]int{
				{8, 8, 8},
				{8, 8, 8},
				{8, 8, 8},
				{8, 8, 8},
				{8, 8, 8},
			})
			replaceTiles([]int{2, 4})
			setState(5, []int{}, []int{0, 2, 4}, []int{0, 2, 4})
			return true
		}
	}
	return false
}

func (f *TriggerElysiumVip) Serialize() ([]byte, error) {
	return feature.SerializeTriggerToBytes(f)
}

func (f *TriggerElysiumVip) Deserialize(data []byte) (err error) {
	return feature.DeserializeTriggerFromBytes(f, data)
}
