package features

type FatTileData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	W      int `json:"w"`
	H      int `json:"h"`
	TileId int `json:"titleid"`
}

type FatTile struct {
	FeatureDef
	Data FatTileData `json:"data"`
}

func (f *FatTile) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *FatTile) DataPtr() interface{} {
	return &f.Data
}

func (f *FatTile) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f FatTile) forceActivateFeature(featurestate *FeatureState, x int, y int, tileid int) {
	gridw, gridh := len(featurestate.SymbolGrid), len(featurestate.SymbolGrid[0])
	tilew, tileh := f.Data.W, f.Data.H
	for r := x; r < x+tilew && r < gridw; r++ {
		for s := y; s < y+tileh && s < gridh; s++ {
			featurestate.SymbolGrid[r][s] = tileid
		}
	}
}

func (f FatTile) Trigger(featurestate *FeatureState, params FeatureParams) {
	x := params.GetInt("X")
	y := params.GetInt("Y")
	w := params.GetInt("W")
	h := params.GetInt("H")
	tileid := params.GetInt("TileId")
	for r := x; r < x+w; r++ {
		for s := y; s < y+h; s++ {
			featurestate.SymbolGrid[r][s] = tileid
		}
	}
	featurestate.Features = append(featurestate.Features,
		&FatTile{
			FeatureDef: *f.DefPtr(),
			Data: FatTileData{
				X:      x,
				Y:      y,
				W:      w,
				H:      h,
				TileId: tileid,
			},
		})
}

func (f *FatTile) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *FatTile) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
