package api

import (
	ft "gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

type GameFields struct {
	ID     string
	Name   string
	Engine string
}

type ForceValuesFields struct {
	ID     string
	Name   string
	Engine string
}

type ForceToolData struct {
	PlayerID     string
	SelectedGame string
	Games        []GameFields
	ForceValues  []ForceValuesFields
}

func removeDuplicates(elements []ForceValuesFields) []ForceValuesFields {
	// Use map to record duplicates as we find them.
	encountered := map[ForceValuesFields]bool{}
	result := []ForceValuesFields{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func listForceTools(r *http.Request, w http.ResponseWriter) {
	tpl, err := template.ParseFiles("templates/api/forcetool/forcetool.html")
	if err != nil {
		logger.Errorf("template parsing error: ", err)
		return
	}
	gameName := r.URL.Query().Get("gameSlug")
	logger.Debugf("GameName: %s", gameName)

	playerID := r.URL.Query().Get("playerID")
	logger.Debugf("PlayerID: %s", playerID)

	forceGP := ft.ReadForcedGameplays(gameName)
	ge := ft.ReadGamesEngines()

	var games []GameFields
	var forceValuesList []ForceValuesFields

	for _, g := range ge {
		for _, f := range forceGP {
			engine := strings.TrimSuffix(f.Engine, ".yml")
			if g.Engine == engine {
				canonicalName := func(n string) string {
					return strings.Title(strings.ReplaceAll(n, "-", " "))
				}
				games = append(games, GameFields{ID: g.GameName, Name: canonicalName(g.GameName), Engine: engine})
			}
			for _, i := range f.Forces {
				appendOK := true
				// For engine VII, listing retriggers may confuse the user that i'ts supposed to be used sequentially.
				// Instead we'll list only a single retrigger and let force tool automatically choose what retrigger to appy
				// don't add retriggers from Engine VII forcetool config, we'll add the separately
				if engine == "mvgEngineVII" {
					if strings.HasPrefix(i.ID, "retrigger") || strings.HasPrefix(i.ID, "FS") {
						appendOK = false
					}
				}
				if appendOK {
					forceValuesList = append(forceValuesList, ForceValuesFields{ID: i.ID, Name: i.ID, Engine: engine})
				}
			}
			// lets add Engine VII retrigger here
			if engine == "mvgEngineVII" {
				insertIndex := 3
				forceValuesList = append(forceValuesList, ForceValuesFields{})
				copy(forceValuesList[insertIndex+1:], forceValuesList[insertIndex:])
				forceValuesList[insertIndex] = ForceValuesFields{ID: "retrigger", Name: "Retrigger", Engine: engine}
			}

		}

	}

	sort.Slice(games, func(i, j int) bool {
		return games[i].ID < games[j].ID
	})

	forceValuesList = removeDuplicates(forceValuesList)
	data := ForceToolData{
		PlayerID:     playerID,
		SelectedGame: gameName,
		Games:        games,
		ForceValues:  forceValuesList,
	}
	tpl.Execute(w, data)
}
