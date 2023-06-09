package api

import (
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"

	//	"gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/forceTool"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/parameterSelector"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/metrics"
)

// APIVersion ...
const (
	APIVersion    = "v2"
	RegexGameSlug = "[A-Za-z0-9-]+"
	RegexWallet   = "[A-Za-z0-9-]+"
	RegexPlayerId = "[a-zA-Z0-9-_+]+"
	RegexId       = "[A-Za-z0-9-_+=.,:;/%]+"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()

	compressor := middleware.NewCompressor(flate.DefaultCompression)

	// See github.com/go-chi/chi/middleware for full
	// Explanation on each piece of middleware
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.RequestID,       // Attach unique ID to each Request
		middleware.RedirectSlashes, // Redirect any path ending with / to same path without / (ex: /v1/rgs/ to /v1/rgs)
		middleware.RealIP,          // Make the request IP the IP of the original client ip if X-Forwarded etc is sent
		//middleware.NoCache,         // Never cache RGS responses
		middleware.Logger, // Log requests
		compressor.Handler,
		//middleware.Recoverer,       // Make panics into 500 error
		Recovery, // Custom recovery middleware, Make panics into 500 error
	)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(8 * time.Second))

	m := metrics.NewHTTPMiddleware("RGS")
	router.Use(m)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/v2/rgs", func(r chi.Router) {

		// TODO: These endpoints will be deprecated with new client release
		r.Get("/init/{gameSlug:"+RegexGameSlug+"}/{wallet:"+RegexWallet+"}", func(w http.ResponseWriter, r *http.Request) {

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
			if strings.Contains(previousGamestate.Id, "GSinit") {
				previousGamestate = store.CreateInitGS(player, gameSlug)
			}
			balanceResponse := BalanceResponse{
				Currency:  player.Balance.Currency,
				Amount:    player.Balance.Amount,
				FreeGames: player.FreeGames.NoOfFreeSpins,
			}
			logger.Debugf("previous Gamestate: %#v", previousGamestate)

			var gamestateResponse GameplayResponse
			gamestateResponse = renderGamestate(r, previousGamestate, balanceResponse, engineConfig, player)

			engineDefs := engineConfig.EngineDefs
			stakeValues, defaultBet, _, _, err := parameterSelector.GetGameplayParameters(previousGamestate.BetPerLine, player.BetLimitSettingCode, gameSlug, player.BetSettingId)
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
				DefaultStake: defaultBet.ValueAsString(),
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
				Balance:    player.Balance.Amount.ValueAsString(),
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
		r.Post("/play/{gameSlug:"+RegexGameSlug+"}/{gamestateID:[A-Za-z0-9-_+=.,:;/]+}/{wallet:"+RegexWallet+"}", func(w http.ResponseWriter, r *http.Request) {
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
		r.Get("/gameplay/{gameplayID:"+RegexId+"}", func(w http.ResponseWriter, r *http.Request) {
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
		r.Get("/gamehashes", func(w http.ResponseWriter, r *http.Request) {
			gameinfos, err := getGameHashes(r)
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
			if err := render.Render(w, r, gameinfos); err != nil {
				logger.Errorf("Error rendering game info response %s", err)
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})
		r.Post("/init2", func(w http.ResponseWriter, r *http.Request) {
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
		r.Post("/play2", func(w http.ResponseWriter, r *http.Request) {
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
		r.Post("/playround", func(w http.ResponseWriter, r *http.Request) {
			round, err := playRound(r)
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
			if err := render.Render(w, r, round); err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})
		r.Post("/init3", func(w http.ResponseWriter, r *http.Request) {
			initResp, err := initV3(r)
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
		r.Post("/play3", func(w http.ResponseWriter, r *http.Request) {
			gamestate, err := playV3(r)
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
			if gamestate != nil {
				if err := render.Render(w, r, gamestate); err != nil {
					_ = render.Render(w, r, ErrRender(err))
					return
				}
			}
		})

		r.Put("/close", func(w http.ResponseWriter, r *http.Request) {
			err := CloseGS(r)
			if err != nil {
				logger.Debugf("error on round close: %v", err)
			}
			if err != nil {
				render.Render(w, r, ErrRender(err))
				w.WriteHeader(400)
				return
			}
			w.WriteHeader(200)
		})

		r.Put("/close3", func(w http.ResponseWriter, r *http.Request) {
			err := closeV3(r)
			if err != nil {
				logger.Debugf("error on round close: %v", err)
			}
			if err != nil {
				render.Render(w, r, ErrRender(err))
				w.WriteHeader(400)
				return
			}
			w.WriteHeader(200)
		})

		r.Put("/clientstate/{token:"+RegexId+"}/{gameSlug:"+RegexGameSlug+"}/{wallet:"+RegexWallet+"}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			gameSlug := chi.URLParam(r, "gameSlug")
			wallet := chi.URLParam(r, "wallet")
			var txStore store.TransactionStore
			var err rgserror.RGSErr
			switch wallet {
			case "demo":
				txStore, err = store.ServLocal.TransactionByGameId(store.Token(token), store.ModeDemo, gameSlug)
			case "dashur":
				txStore, err = store.Serv.TransactionByGameId(store.Token(token), store.ModeReal, gameSlug)
			}
			if txStore.WalletStatus != 1 {
				// if this is zero, the tx is pending and shouldn't be resent, if it is -1, the tx is failed and an error should be sent to reload the client
				logger.Debugf("STATUS: %v", txStore.WalletStatus)
				fmt.Fprint(w, []byte("ERROR"))
				return
			}
			gamestateUnmarshalled := store.DeserializeGamestateFromBytes(txStore.GameState)
			if len(gamestateUnmarshalled.NextActions) > 1 {
				// we should not be closing a gameround if the last gamestate has more actions to be completed
				fmt.Fprint(w, []byte("OK"))
				return
			}
			gamestateUnmarshalled.Closed = true
			roundId := gamestateUnmarshalled.RoundID
			if roundId == "" {
				roundId = gamestateUnmarshalled.Id
			}
			state := store.SerializeGamestateToBytes(gamestateUnmarshalled)
			ttl := gamestateUnmarshalled.GetTtl()
			switch wallet {
			case "demo":
				_, err = store.ServLocal.CloseRound(store.Token(token), store.ModeDemo, gameSlug, roundId, "", state, ttl, &store.TransactionHistory{})
			case "dashur":
				_, err = store.Serv.CloseRound(store.Token(token), store.ModeReal, gameSlug, roundId, txStore.FreeGames.CampaignRef, state, ttl, nil)
			}
			if err != nil {
				fmt.Fprint(w, []byte("ERROR"))
			}
			fmt.Fprint(w, []byte("OK"))
		})

		r.Get("/stopAuto/{playerId:"+RegexPlayerId+"}/{on:[a-z-_+]+}", func(w http.ResponseWriter, r *http.Request) {
			playerId := chi.URLParam(r, "playerId")
			on := chi.URLParam(r, "on")
			var err rgserror.RGSErr
			switch on {
			case "on":
				err = store.ServLocal.SetMessage(playerId, "stopAuto")
			default:
				err = store.ServLocal.SetMessage(playerId, "")
			}
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, err)
				return
			}
			w.WriteHeader(200)
			return
		})

		r.Get("/force", func(w http.ResponseWriter, r *http.Request) {
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
				e = errors.New("missing field: game name")
			} else if param.ForceID == "" {
				e = errors.New("missing field: force ID")
			}
			if e != nil {
				w.WriteHeader(400)
				fmt.Fprint(w, e)
				return
			}
			if err := forceTool.SetForce(param.GameSlug, param.ForceID, param.PlayerID); err != nil {
				w.WriteHeader(400)
				logger.Infof("Force Tool Active : %v, %v %v", param.PlayerID, param.GameSlug, param.ForceID)
				fmt.Fprint(w, err.Error())
				return
			} else {
				w.WriteHeader(200)
				fmt.Fprint(w, "OK")
				return
			}
		})
		r.Post("/setlastgs", func(w http.ResponseWriter, r *http.Request) {
			// force a specific gamestate
			token := r.FormValue("token")
			if config.GlobalConfig.DashurConfig.StoreRemoteUrl != "https://gnrc-api.dashur.io/v1/gnrc/maverick" {
				logger.Errorf("Can't force last gs on non-staging env")
				return
			}
			gs := r.FormValue("gamestate")
			wallet := r.FormValue("wallet")
			gsbytes, serr := base64.StdEncoding.DecodeString(gs)
			if serr != nil {
				logger.Errorf("error : %v", serr)
				return
			}
			gamestate := store.DeserializeGamestateFromBytes(gsbytes)
			player, _, err := store.ServLocal.PlayerByToken(store.Token(token), store.ModeDemo, gamestate.Game)
			if err != nil {
				logger.Errorf("error : %v", err)
				return
			}
			logger.Debugf("gamestate: %#v", gamestate)
			switch wallet {
			case "demo":
				_, err = store.ServLocal.Transaction(player.Token, store.ModeDemo, store.TransactionStore{
					TransactionId:       gamestate.Transactions[0].Id,
					Token:               "",
					Mode:                store.ModeDemo,
					Category:            store.CategoryPayout,
					RoundStatus:         store.RoundStatusOpen,
					PlayerId:            player.PlayerId,
					GameId:              gamestate.Game,
					RoundId:             gamestate.RoundID,
					Amount:              gamestate.Transactions[0].Amount,
					ParentTransactionId: "",
					TxTime:              time.Now(),
					BetLimitSettingCode: "",
					GameState:           gsbytes,
					FreeGames:           store.FreeGamesStore{},
					WalletStatus:        0,
					Ttl:                 gamestate.GetTtl(),
				})
			case "dashur":
				_, err = store.Serv.Transaction(player.Token, store.ModeReal, store.TransactionStore{
					TransactionId:       gamestate.Transactions[0].Id,
					Token:               "",
					Mode:                store.ModeReal,
					Category:            store.CategoryPayout,
					RoundStatus:         store.RoundStatusOpen,
					PlayerId:            player.PlayerId,
					GameId:              gamestate.Game,
					RoundId:             gamestate.RoundID,
					Amount:              gamestate.Transactions[0].Amount,
					ParentTransactionId: "",
					TxTime:              time.Now(),
					BetLimitSettingCode: "",
					GameState:           gsbytes,
					FreeGames:           store.FreeGamesStore{},
					WalletStatus:        0,
					Ttl:                 gamestate.GetTtl(),
				})
			}

			if err != nil {
				logger.Errorf("error : %v", err)
				return
			}
			logger.Infof("it worked")
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
				e = errors.New("missing field: game name")
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
		r.Get("/clearforce/{gameSlug:"+RegexGameSlug+"}/{playerID:"+RegexPlayerId+"}", func(w http.ResponseWriter, r *http.Request) {
			gameSlug := chi.URLParam(r, "gameSlug")
			playerID := chi.URLParam(r, "playerID")
			if err := forceTool.ClearForce(gameSlug, playerID); err != nil {
				logger.Warnf("Force Tool Cleared : %v, %v", gameSlug, playerID)
				fmt.Fprint(w, err.Error())
			} else {
				fmt.Fprint(w, "OK")
			}
		})
		r.Get("/playcheck/{gameplayID:"+RegexId+"}", func(w http.ResponseWriter, r *http.Request) {
			playcheck(r, w)
		})
		r.Post("/playcheckext", func(w http.ResponseWriter, r *http.Request) {
			var param PlayCheckExtParams
			err := json.NewDecoder(r.Body).Decode(&param)
			if err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
			playcheckExtResp, err := playcheckExt(r, w, param)
			if err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			if err := render.Render(w, r, playcheckExtResp); err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			return
		})
		r.Get("/balance/{wallet:"+RegexWallet+"}", func(w http.ResponseWriter, r *http.Request) {
			balResp, err := PlayerBalance(r)
			if err != nil {
				logger.Errorf("Error getting balance %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))
				}
				return
			}

			if err := render.Render(w, r, balResp); err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			return
		})
		r.Post("/setbalance/demo", func(w http.ResponseWriter, r *http.Request) {
			var param SetBalanceParams
			err := json.NewDecoder(r.Body).Decode(&param)
			if err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
			token, err := processAuthorization(r)
			if err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
			memID := parseMemID(token)
			err = store.SetPlayerBalance(memID, "demo", param.Balance)
			if err != nil {
				_ = render.Render(w, r, ErrRender(err))
				return
			}
		})
		r.Get("/stakes", func(w http.ResponseWriter, r *http.Request) {
			stakeInfo(r, w)
		})
		r.Post("/feed", func(w http.ResponseWriter, r *http.Request) {
			feedResp, err := Feed(r)
			if err != nil {
				logger.Errorf("Error getting feed %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))
				}
				return
			}
			if err := render.Render(w, r, feedResp); err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			return
		})
		r.Post("/feedround", func(w http.ResponseWriter, r *http.Request) {
			feedResp, err := FeedRound(r)
			if err != nil {
				logger.Errorf("Error getting feed %s", err.Error())

				switch t := err.(type) {
				default:
					_ = render.Render(w, r, ErrRender(err))
				case *rgserror.RGSError:
					logger.Debugf("%v", t)
					_ = render.Render(w, r, ErrBadRequestRender(err.(*rgserror.RGSError)))
				}
				return
			}
			if err := render.Render(w, r, feedResp); err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			return
		})

		if config.GlobalConfig.DevMode {
			r.Get("/debug/pprof/profile", pprof.Profile)
			r.Mount("/debug/pprof/heap", pprof.Handler("heap"))
			r.Mount("/debug/pprof/block", pprof.Handler("block"))
			r.Mount("/debug/pprof/mutex", pprof.Handler("mutex"))
			r.Mount("/debug/pprof/allocs", pprof.Handler("allocs"))
			r.Mount("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
			r.Mount("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		}

		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			versionFile, err := ioutil.ReadFile("version.txt")
			if err != nil {
				logger.Fatalf("Error reading version file: %v", err)
				_ = render.Render(w, r, ErrRender(err))
			}
			if err := render.Render(w, r, VersionResponse{Version: string(versionFile)}); err != nil {
				_ = render.Render(w, r, ErrRender(err))
			}
			return
		})
	})

	return router
}
