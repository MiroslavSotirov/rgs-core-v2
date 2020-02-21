package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/metrics"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIVersion ...
var APIVersion = "v2"

func Routes() *chi.Mux {
	router := chi.NewRouter()
	// See github.com/go-chi/chi/middleware for full
	// Explanation on each piece of middleware
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.RequestID,       // Attach unique ID to each Request
		middleware.RedirectSlashes, // Redirect any path ending with / to same path without / (ex: /v1/rgs/ to /v1/rgs)
		middleware.RealIP,          // Make the request IP the IP of the original client ip if X-Forwarded etc is sent
		//middleware.NoCache,         // Never cache RGS responses
		middleware.Logger,          // Log requests
		middleware.DefaultCompress, // Compress requests
		//middleware.Recoverer,       // Make panics into 500 error
		Recovery, // Custom recovery middleware, Make panics into 500 error
	)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(1 * time.Second))

	m := metrics.NewHTTPMiddleware("RGS")
	router.Use(m)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/v2/rgs", func(r chi.Router) {

		// TODO: These endpoints will be deprecated with new client release
		r.Get("/init/{gameSlug:[0-9a-z-]+}/{wallet:[a-z-]+}", func(w http.ResponseWriter, r *http.Request) {

			gameSlug := chi.URLParam(r, "gameSlug")
			engineID, err := config.GetEngineFromGame(gameSlug)

			player, engineConfig, previousGamestate, err := initGame(r)
			if err != nil {
				logger.Errorf("Error initializing game %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))

				}
				return
			}
			balanceResponse := BalanceResponse{
				Currency: player.Balance.Currency,
				Amount:   player.Balance.Amount.ValueAsString(),
			}
			logger.Debugf("lastgamestate nextactions: %v", previousGamestate.NextActions)
			var gamestateResponse GameplayResponse
			gamestateResponse = renderGamestate(r, previousGamestate, balanceResponse, engineConfig, player)

			engineDefs := engineConfig.EngineDefs
			stakeValues, defaultBet, err := parameterSelector.GetGameplayParameters(previousGamestate.BetPerLine.Amount, player, gameSlug)
			if err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}

			stakeValuesString := fmt.Sprintf("%v", stakeValues[0].ValueAsString())
			for i := 1; i < len(stakeValues); i++ {
				stakeValuesString += fmt.Sprintf(" %v", stakeValues[i].ValueAsString())
			}

			parameters := ParameterResponse{
				StakeValues:  stakeValuesString,
				DefaultStake: defaultBet.ValueAsFloat(),
				SessionID:    player.Token,
			}

			// craft windefs
			winDefs := make(map[string]WinResponse, len(engineDefs[0].Payouts)+len(engineDefs[0].SpecialPayouts)) // todo: Find a smart way to make this the length of all prizes, this may fail if freespin defines prizes not defined in main game
			for _, def := range engineDefs {
				var winType string
				if strings.Contains(def.Function, "Ways") {
					winType = "way"
				} else {
					winType = "line"
				}
				for _, p := range def.Payouts {
					var winDefID string
					if engineID == "mvgEngineX" {
						// Engine X WinDefID format <symbol:count:multipier>
						for i := 0; i <= 3; i++ { //multiplier
							winDefID = strconv.Itoa(p.Symbol) + ":" + strconv.Itoa(p.Count) + ":" + strconv.Itoa(i)
							if i == 0 {
								winDefs[winDefID] = WinResponse{
									PayoutFactor: p.Multiplier, //don't multiply
									Type:         winType,
									Symbol:       p.Symbol,
									StakeIndex:   strconv.Itoa(i),
									Frequency:    strconv.Itoa(p.Count),
								}
							} else {
								winDefs[winDefID] = WinResponse{
									PayoutFactor: p.Multiplier * i, //multiply
									Type:         winType,
									Symbol:       p.Symbol,
									StakeIndex:   strconv.Itoa(i),
									Frequency:    strconv.Itoa(p.Count),
								}
							}

						}

					} else {
						winDefID = strconv.Itoa(p.Symbol) + ":" + strconv.Itoa(p.Count)
						winDefs[winDefID] = WinResponse{
							PayoutFactor: p.Multiplier,
							Type:         winType,
							Symbol:       p.Symbol,
						}
					}

				}
				for _, p := range def.SpecialPayouts {
					winDefs[p.Index] = WinResponse{
						PayoutFactor: p.Payout.Multiplier,
						Type:         "scatter", //TODO: extend for other bonus types?
						Symbol:       p.Payout.Symbol,
					}
				}
			}

			//// translate winlines for old game clients: WARNING: this only works for 5 reel games of consistent reel sizes
			translatedWinLines := make([][]int, len(engineDefs[0].WinLines))
			reelMult := 5
			if len(engineDefs[0].ViewSize) == 3 {
				reelMult = 3
			}
			for i, line := range engineDefs[0].WinLines {
				translatedWinLine := make([]int, len(line))
				for j, pos := range line {
					translatedWinLine[j] = pos*reelMult + j
				}
				translatedWinLines[i] = translatedWinLine
			}

			// add engineDefinitions
			engineDefResponse := GetEngineDefResponse(engineConfig, engineID)
			prizeDefResponse := GetPrizeDefResponse(engineConfig, engineID)
			symbolDefResponses := make(map[int]SymbolDefResponse)

			for i := 0; i < len(engineConfig.EngineDefs[0].Payouts); i++ {
				symbolDefResponses[engineConfig.EngineDefs[0].Payouts[i].Symbol] = SymbolDefResponse{IsWild: false, DisplayName: fmt.Sprintf("%v", engineConfig.EngineDefs[0].Payouts[i].Symbol)}
			}
			// do wilds second in case a payout is listed for a wild symbol
			for j := 0; j < len(engineConfig.EngineDefs); j++ {
				for i := 0; i < len(engineConfig.EngineDefs[j].Wilds); i++ {
					symbolDefResponses[engineConfig.EngineDefs[j].Wilds[i].Symbol] = SymbolDefResponse{IsWild: true, DisplayName: fmt.Sprintf("%v", engineConfig.EngineDefs[j].Wilds[i].Symbol)}
				}
			}
			reelsInit := [][]int{{3, 5, 6, 1, 8}, {3, 4, 6, 5, 1}, {2, 5, 4, 1, 5}, {3, 2, 6, 0, 1}, {3, 6, 2, 0, 7}}
			if engineID == "mvgEngineIX" {
				reelsInit = [][]int{{3, 4, 3, 2}, {2, 2, 3, 3}, {4, 2, 4, 3}}
			}
			if len(engineConfig.EngineDefs[0].SpecialPayouts) > 0 {
				symbolDefResponses[engineConfig.EngineDefs[0].SpecialPayouts[0].Payout.Symbol] = SymbolDefResponse{IsWild: false, DisplayName: fmt.Sprintf("%v", engineConfig.EngineDefs[0].SpecialPayouts[0].Payout.Symbol)}
			}
			definition := GameDefinitionResponse{WinLineDefs: translatedWinLines, ReelSetDefs: engineDefResponse, WinDefs: winDefs, ReelsInit: reelsInit, PrizeDefs: prizeDefResponse, SymbolDefs: symbolDefResponses} //, ViewDefs:viewDef}

			// todo: send link for freeplay if recovering in fp mode

			links := gamestateResponse.Links
			if strings.Contains(previousGamestate.Id, "GSinit") == false {
				logger.Infof("prev gamestate id: %v", previousGamestate.Id)
				href := fmt.Sprintf("%s%s/%s/rgs/playcheck/%s", GetURLScheme(r), r.Host, APIVersion, previousGamestate.Id)
				latestGameplayLink := LinkResponse{
					Href:   href,
					Method: "GET",
					Rel:    "latest-gameplay",
					Type:   "",
				}
				links = append(links, latestGameplayLink)
			}
			//if gamestateResponse.Links[1].Rel == "option" {
			//	links = append(links, gamestateResponse.Links[1])
			//}

			forms := BuildForm(previousGamestate, engineID, false)
			gameInit := GameInitResponse{
				Schema:     DefaultSchema,
				Balance:    player.Balance.Amount.ValueAsFloat(),
				Currency:   player.Balance.Currency,
				Parameters: parameters,
				Player:     gamestateResponse.Player,
				Definition: definition,
				Links:      links,
				Forms:      forms,
				Gameplay:   gamestateResponse,
			}
			if err := render.Render(w, r, gameInit); err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})

		r.Post("/play/{gameSlug:[A-Za-z0-9-]+}/{gamestateID:[A-Za-z0-9-_]+}/{wallet:[A-Za-z0-9-]+}", func(w http.ResponseWriter, r *http.Request) {
			gameplay, err := renderNextGamestate(r)

			if err != nil {
				logger.Errorf("Error initializing game %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))

				}
				return
			}

			if err := render.Render(w, r, gameplay); err != nil {
				logger.Errorf("Error rendering gameplay %s", err)
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})

		r.Post("/state", func(w http.ResponseWriter, r *http.Request) {
			state := StateResponse{State: []byte("state")}
			if err := render.Render(w, r, state); err != nil {
				logger.Errorf("Error rendering state response %s", err)
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})
		r.Get("/gameplay/{gameplayID:[A-Za-z-]+}", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, []byte("OK"))
		})

		// TODO: These endpoints will remain

		r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/game", func(w http.ResponseWriter, r *http.Request) {
			link := getGameLink(r)
			if err := render.Render(w, r, link); err != nil {
				logger.Errorf("Error rendering game info response %s", err)
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})

		r.Get("/init2", func(w http.ResponseWriter, r *http.Request) {
			initResp, err := initV2(r)
			if err != nil {
				logger.Errorf("Error initializing game %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))

				}
				return
			}

			if err := render.Render(w, r, initResp); err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})
		r.Post("/play2/{gameSlug:[A-Za-z0-9-]+}", func(w http.ResponseWriter, r *http.Request) {
			gamestate, err := playV2(r)
			if err != nil {
				logger.Errorf("Error initializing game %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))

				}
				return
			}
			if err := render.Render(w, r, gamestate); err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})

		r.Put("/clientstate/{gamestateID:[A-Za-z0-9-_]+}/{token:[A-Za-z0-9-_.:,]+}/{gameSlug:[A-Za-z0-9-]+}/{wallet:[A-Za-z0-9-_]+}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			gameSlug := chi.URLParam(r, "gameSlug")
			wallet := chi.URLParam(r, "wallet")
			gamestateID := chi.URLParam(r, "gamestateID")
			var gamestate store.GameStateStore
			var err *store.Error
			switch wallet {
			case "demo":
				gamestate, err = store.ServLocal.GameStateByGameId(store.Token(token), store.ModeDemo, gameSlug)
			case "dashur":
				gamestate, err = store.Serv.GameStateByGameId(store.Token(token), store.ModeReal, gameSlug)

			}

			switch wallet {
			case "demo":
				_, err = store.ServLocal.CloseRound(store.Token(token), store.ModeDemo, gameSlug, gamestateID, gamestate.GameState)
			case "dashur":
				_, err = store.Serv.CloseRound(store.Token(token), store.ModeReal, gameSlug, gamestateID, gamestate.GameState)
			}
			if err != nil {
				fmt.Fprint(w, []byte("ERROR"))
			}
			fmt.Fprint(w, []byte("OK"))
		})
		r.Get("/force", func(w http.ResponseWriter, r *http.Request){
			if config.GlobalConfig.DevMode == true {
				listForceTools(r, w)
			}
		})
		r.Post("/force", func(w http.ResponseWriter, r *http.Request) {
			var param forceTool.ForceToolParams
			err := json.NewDecoder(r.Body).Decode(&param)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, errors.New("failed to decode request"))
				return
			}
			logger.Debugf("Game:%v ForceID:%v playerID: %v", param.GameSlug, param.ForceID, param.PlayerID)
			var e error
			if param.PlayerID == "" {
				e = errors.New("missing field: player ID")
			} else if param.GameSlug == "" {
					e =  errors.New("missing field: game name")
			} else if param.ForceID == "" {
				e =  errors.New("missing field: force ID")
			}
			if e != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, e)
				return
			}
			if err := forceTool.SetForce(param.GameSlug, param.ForceID, param.PlayerID); err != nil {
				w.WriteHeader(400)
				logger.Infof("Force Tool Active : %v, %v %v", param.PlayerID, param.GameSlug, param.ForceID )
				fmt.Fprint(w, err.Error())
				return
			} else {
				w.WriteHeader(200)
				fmt.Fprint(w, "OK")
				return
			}
		})
		r.Delete("/force", func(w http.ResponseWriter, r *http.Request) {
			var param forceTool.ForceToolParams
			err := json.NewDecoder(r.Body).Decode(&param)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, errors.New("failed to decode request"))
				return
			}
			var e error
			if param.PlayerID == "" {
				e = errors.New("missing field: player ID")
			} else if param.GameSlug == "" {
				e =  errors.New("missing field: game name")
			}
			if e != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, e)
				return
			}
			if err := forceTool.ClearForce(param.GameSlug, param.PlayerID); err != nil {
				w.WriteHeader(400)
				logger.Infof("Force Tool cleared : %v, %v %v", param.PlayerID, param.GameSlug)
				fmt.Fprint(w, err.Error())
				return
			} else {
				w.WriteHeader(200)
				fmt.Fprint(w, "OK")
				return
			}
		})
		r.Get("/clearforce/{gameSlug:[A-Za-z-]+}/{playerID:[A-Za-z0-9-_]+}", func(w http.ResponseWriter, r *http.Request) {
			gameSlug := chi.URLParam(r, "gameSlug")
			playerID := chi.URLParam(r, "playerID")
			if err := forceTool.ClearForce(gameSlug, playerID); err != nil {
				logger.Warnf("Force Tool Cleared : %v, %v", gameSlug, playerID)
				fmt.Fprint(w, err.Error())
			} else {
				fmt.Fprint(w, "OK")
			}
		})
		r.Get("/playcheck/{gameplayID:[A-Za-z0-9-:]+}", func(w http.ResponseWriter, r *http.Request) {
			playcheck(r, w)
		})

	})

	return router
}
