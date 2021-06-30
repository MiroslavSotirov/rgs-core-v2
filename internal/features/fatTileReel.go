package features

type FatTileReelData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	W      int `json:"w"`
	H      int `json:"h"`
	TileId int `json:"titleid"`
}

type FatTileReel struct {
	Id   int32           `json:"id"`
	Type string          `json:"type"`
	Data FatTileReelData `json:"data"`
}

func (f FatTileReel) GetId() int32 {
	return f.Id
}

func (f FatTileReel) GetType() string {
	return f.Type
}

func (f *FatTileReel) SetId(id int32) {
	f.Id = id
}

func (f *FatTileReel) SetType(typename string) {
	f.Type = typename
}

func (f *FatTileReel) DataPtr() interface{} {
	return &f.Data
}

func (f *FatTileReel) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f FatTileReel) forceActivateFeature(featurestate *FeatureState, x int, y int) {
	gridw, gridh := len(featurestate.SymbolGrid), len(featurestate.SymbolGrid[0])
	tilew, tileh := f.Data.W, f.Data.H
	for r := x; r < x+tilew && r < gridw; r++ {
		for s := y; s < y+tileh && s < gridh; s++ {
			featurestate.SymbolGrid[r][s] = 1
		}
	}
}

func (f FatTileReel) Trigger(featurestate FeatureState) []Feature {
	features := []Feature{}
	gridw, gridh := len(featurestate.SymbolGrid), len(featurestate.SymbolGrid[0])
	tilew, tileh := f.Data.W, f.Data.H
	tileid := f.Data.TileId

	for x := 0; x < gridw-tilew+1; x++ {
		for y := 0; y < gridh-tileh+1; y++ {
			found := func() bool {
				for r := x; r < x+tilew; r++ {
					for s := y; s < y+tileh; s++ {
						if featurestate.SymbolGrid[r][s] != tileid {
							return false
						}
					}
				}
				return true
			}()
			if found {
				features = append(features, &FatTileReel{
					Id:   f.Id,
					Type: f.Type,
					Data: FatTileReelData{
						X:      x,
						Y:      y,
						W:      tilew,
						H:      tileh,
						TileId: tileid,
					},
				})
			}
		}
	}
	return features
}

func (f *FatTileReel) Serialize() ([]byte, error) {
	return serializeFeatureToBytes(f)
}

func (f *FatTileReel) Deserialize(data []byte) (err error) {
	return deserializeFeatureFromBytes(f, data)
}
