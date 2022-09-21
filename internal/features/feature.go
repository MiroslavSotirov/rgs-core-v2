package features

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type FeatureDef struct {
	Id       int32         `yaml:"Id" json:"id"`
	Type     string        `yaml:"Type" json:"type"`
	Params   FeatureParams `yaml:"Params" json:"-"`
	Features []FeatureDef  `yaml:"Features" json:"-"`
}

type FeatureState struct {
	SourceGrid [][]int
	SymbolGrid [][]int
	Reels      [][]int
	StopList   []int
	Features   []Feature
	Wins       []FeatureWin
	TotalStake float64
	Stateful   *FeatureState
	Action     string
	PureWins   bool
	ReelsetId  string
}

func (fs *FeatureState) SetGrid(symbolgrid [][]int) {
	gridw, gridh := len(symbolgrid), len(symbolgrid[0])
	fs.SourceGrid = symbolgrid
	fs.SymbolGrid = make([][]int, gridw)
	grid := make([]int, gridw*gridh)
	for i := range fs.SymbolGrid {
		fs.SymbolGrid[i], grid = grid[:gridh], grid[gridh:]
		for j := range fs.SymbolGrid[i] {
			fs.SymbolGrid[i][j] = symbolgrid[i][j]
		}
	}
}

type FeatureWin struct {
	Index           string
	Multiplier      int
	Symbols         []int
	SymbolPositions []int
}

type InitConfig struct {
	FeatureDef
	Data FeatureParams `json:"data"`
}

type Feature interface {
	DefPtr() *FeatureDef
	DataPtr() interface{}
	Init(FeatureDef) error
	OnInit(*FeatureState)
	Trigger(*FeatureState, FeatureParams)
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// features must be included here to make them deserializable by the engine
type EnabledFeatureSet struct {
	_ FeatureNull
	_ ExpandingWild
	_ FatTile
	_ InstaWin
	_ ReplaceTile
	_ SetReels
	_ SetConditional
	_ StatefulMap
	_ Princess
	_ Respin
	_ TriggerFoxTale
	_ TriggerFoxTaleBonus
	_ TriggerFoxTaleWild
	_ TriggerSpiritHunters
	_ TriggerSpiritHuntersBonus
	_ TriggerSupaCrew
	_ TriggerSupaCrewActionSymbol
	_ TriggerSupaCrewSuperSymbol
	_ TriggerSupaCrewMultiSymbol
	_ TriggerWizardzWorld
	_ TriggerWizardzWorldBonus
	_ TriggerBattleOfMyths
	_ TriggerBattleOfMythsFreespin
	_ TriggerBattleOfMythsPrincess
	_ TriggerBattleOfMythsDragon
	_ TriggerBattleOfMythsTiger
	_ TriggerSwordKing
	_ TriggerSwordKingBonus
	_ TriggerSwordKingBonusScatter
	_ TriggerSwordKingFreespin
	_ TriggerSwordKingRandomWilds
	_ TriggerSwordKingRespin
	_ TriggerClashOfHeroes
	_ TriggerClashOfHeroesExpandingWilds
	_ TriggerClashOfHeroesRandomWilds
	_ TriggerClashOfHeroesSwapSymbols
	_ TriggerWeightedRandom
	_ TriggerWeightedPayout
	_ TriggerConditional
	_ TriggerFatTile
	_ TriggerRandom
}

func MakeFeature(typename string) Feature {
	featuretype, ok := enabledFeatureMap[typename]
	if !ok {
		return nil
	}
	feature, ok := reflect.New(featuretype).Interface().(Feature)
	if !ok {
		return nil
	}
	return feature
}

func FindFeature(typename string, features []Feature) Feature {
	for _, f := range features {
		if f.DefPtr().Type == typename {
			return f
		}
	}
	return nil
}

func ActivateFeatures(def FeatureDef, state *FeatureState, params FeatureParams) {
	activateFeatures(def, state, params)
}

func InitFeatures(def FeatureDef, state *FeatureState) {
	initFeatures(def, state)
}

//func GetTypeName(f Feature) string {
//	return reflect.TypeOf(f).Name()
//}

var enabledFeatureMap map[string]reflect.Type = buildFeatureMap(EnabledFeatureSet{})

func buildFeatureMap(featureset EnabledFeatureSet) map[string]reflect.Type {
	featuresetval := reflect.ValueOf(featureset)
	numfields := featuresetval.NumField()
	featuremap := make(map[string]reflect.Type)
	for i := 0; i < numfields; i++ {
		featureval := featuresetval.Field(i)
		featuretype := featureval.Type()
		featureptrtype := reflect.PtrTo(featuretype)
		if featureptrtype.Implements(reflect.TypeOf((*Feature)(nil)).Elem()) {
			featuremap[featuretype.Name()] = featuretype
		} else {
			panic(fmt.Sprintf("EnabledFeatureSet contains %s that doesn't implement Feature",
				featuretype.Name()))
		}
	}
	return featuremap
}

func mergeParams(p1 FeatureParams, p2 FeatureParams) (p FeatureParams) {
	p = make(FeatureParams, len(p1)+len(p2))
	for k, v := range p1 {
		p[k] = v
	}
	for k, v := range p2 {
		p[k] = v
	}
	return
}

func collateParams(p1 FeatureParams, p2 FeatureParams) (p FeatureParams) {
	for k, v := range mergeParams(p2, p1) {
		p2[k] = v
	}
	return p2
}

type FilterFunc func(int, FeatureDef, *FeatureState, FeatureParams) bool

func activateFeatures(def FeatureDef, state *FeatureState, params FeatureParams) {
	all := func(i int, d FeatureDef, s *FeatureState, p FeatureParams) bool { return true }
	activateFilteredFeatures(def, state, params, all)
}

func activateFilteredFeatures(def FeatureDef, state *FeatureState, params FeatureParams, filter FilterFunc) {
	collate := params.HasKey("Collated") && params.GetBool("Collated")
	for i, featuredef := range def.Features {
		if filter(i, featuredef, state, params) {
			logger.Debugf("activate feature %s collated %v", featuredef.Type, collate)
			feature := MakeFeature(featuredef.Type)
			if feature == nil {
				panic(fmt.Sprintf("feature %s is not registred", featuredef.Type))
				continue
			}
			feature.Init(featuredef)
			if collate {
				feature.Trigger(state, collateParams(featuredef.Params, params))
			} else {
				feature.Trigger(state, mergeParams(featuredef.Params, params))
			}
		}
	}
}

func initFeatures(def FeatureDef, state *FeatureState) {
	for _, featuredef := range def.Features {
		feature := MakeFeature(featuredef.Type)
		if feature == nil {
			logger.Errorf("feature %s is not registred", featuredef.Type)
			continue
		}
		feature.Init(featuredef)
		feature.OnInit(state)
	}
}

func deserializeFeatureDef(f Feature, def FeatureDef) error {
	//	f.SetId(def.Id)
	//	f.SetType(def.Type)

	f.DefPtr().Id = def.Id
	f.DefPtr().Type = def.Type
	//	err := DeserializeStruct(f.DataPtr(), def.Params)
	//	if err != nil {
	//		logger.Errorf("err: %s", err.Error())
	//		return err
	//	}
	//	var dataptr interface{} = f.DataPtr()
	//	logger.Debugf("FeatureDef created: %v from params %v\n", dataptr, def.Params)
	// return err
	f.DefPtr().Params = def.Params
	f.DefPtr().Features = def.Features
	return nil
}

func deserializeFeatureFromBytes(f Feature, b []byte) error {
	return json.Unmarshal(b, f)
}

func serializeFeatureToBytes(f Feature) ([]byte, error) {
	return json.Marshal(f)
}

func deserializeTriggerFromBytes(f Feature, b []byte) error {
	return fmt.Errorf("trying to deserialize feature trigger %s from bytes", f.DefPtr().Type)
}

func serializeTriggerToBytes(f Feature) ([]byte, error) {
	return []byte{}, fmt.Errorf("trying to serialize feature trigger %s to bytes", f.DefPtr().Type)
}
