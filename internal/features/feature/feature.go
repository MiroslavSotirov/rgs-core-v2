package feature

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
	SourceGrid       [][]int
	SymbolGrid       [][]int
	Reels            [][]int
	StopList         []int
	Features         []Feature
	Wins             []FeatureWin
	TotalStake       float64
	Stateful         *FeatureState
	Stateless        *FeatureState
	Action           string
	CalculateWins    func([][]int, []FeaturePayout) []FeatureWin
	ReelsetId        string
	CascadePositions []int
	Multiplier       int
	NextReplay       *FeatureState
	Replay           bool
	ReplayTries      int
	ReplayParams     FeatureParams
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

type FeaturePayout struct {
	Symbol     int
	Count      int
	Multiplier int
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

func DeserializeFeatureDef(f Feature, def FeatureDef) error {
	f.DefPtr().Id = def.Id
	f.DefPtr().Type = def.Type
	f.DefPtr().Params = def.Params
	f.DefPtr().Features = def.Features
	return nil
}

func DeserializeFeatureFromBytes(f Feature, b []byte) error {
	return json.Unmarshal(b, f)
}

func SerializeFeatureToBytes(f Feature) ([]byte, error) {
	return json.Marshal(f)
}

func DeserializeTriggerFromBytes(f Feature, b []byte) error {
	return fmt.Errorf("trying to deserialize feature trigger %s from bytes", f.DefPtr().Type)
}

func SerializeTriggerToBytes(f Feature) ([]byte, error) {
	return []byte{}, fmt.Errorf("trying to serialize feature trigger %s to bytes", f.DefPtr().Type)
}

type Factory func() Feature

var enabledFeatures map[string]Factory = map[string]Factory{}

func RegisterFeature(typename string, factory Factory) Factory {
	_, exists := enabledFeatures[typename]
	if exists {
		panic(fmt.Sprintf("feature %s already registred", typename))
	}
	enabledFeatures[typename] = factory
	return factory
}

func MakeFeature(typename string) Feature {
	factory, ok := enabledFeatures[typename]
	if !ok {
		return nil
	}
	feature := factory()
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

func InitFeatures(def FeatureDef, state *FeatureState) {
	initFeatures(def, state)
}

func GetTypeName(f Feature) string {
	return reflect.TypeOf(f).Name()
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

func ActivateFeatures(def FeatureDef, state *FeatureState, params FeatureParams) {
	all := func(i int, d FeatureDef, s *FeatureState, p FeatureParams) bool { return true }
	ActivateFilteredFeatures(def, state, params, all)
}

func ActivateFilteredFeatures(def FeatureDef, state *FeatureState, params FeatureParams, filter FilterFunc) {
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
