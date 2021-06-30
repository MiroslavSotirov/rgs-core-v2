package features

import "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"

type FatTileChanceData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	W      int `json:"w"`
	H      int `json:"h"`
	TileId int `json:"tileid"`
}

type FatTileChance struct {
	Id   int32             `json:"id"`
	Type string            `json:"type"`
	Data FatTileChanceData `json:"data"`
}

func (f FatTileChance) GetId() int32 {
	return f.Id
}

func (f FatTileChance) GetType() string {
	return f.Type
}

func (f *FatTileChance) SetId(id int32) {
	f.Id = id
}

func (f *FatTileChance) SetType(typename string) {
	f.Type = typename
}

func (f *FatTileChance) DataPtr() interface{} {
	return &f.Data
}

func (f *FatTileChance) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f FatTileChance) Trigger(featurestate FeatureState) []Feature {
	gridh := len(featurestate.SymbolGrid[0])
	random := rng.RandFromRange(15)
	h := random % 3
	x := random / 5
	y := 0
	bottom := (random/3)%2 > 0
	if bottom {
		y = gridh - h
	}
	return []Feature{
		&FatTileChance{
			Id:   f.Id,
			Type: f.Type,
			Data: FatTileChanceData{
				X:      x,
				Y:      y,
				W:      f.Data.W,
				H:      h,
				TileId: f.Data.TileId,
			},
		},
	}
}

func (f *FatTileChance) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *FatTileChance) Deserialize(data []byte) error {
	return deserializeFeatureFromBytes(f, data)
}
