package main

import (
	"flag"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/api"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/volumeTester"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"os"
	"strconv"
)

var (
	runVT     bool
	engineID  string
	spins     int
	chunks    int
	perSpin   bool
	maxes     bool
	getHashes bool
)

func main() {
	// load configs (i.e. dashur api url, memcached address, etc.
	flag.BoolVar(&runVT, "vt", false, "Run volume tester? if true, will not start server unless all engines pass VT")
	flag.StringVar(&engineID, "engine", "", "which engine to test (tests all if blank)")
	flag.IntVar(&spins, "spins", 0, "number of spins to run, defaults to number to reach < 1% deviation from RTP based on engine volatility")
	flag.IntVar(&chunks, "chunks", 10, "number of chunks to run (default 10)")
	flag.BoolVar(&perSpin, "perspin", false, "show results per spin")
	flag.BoolVar(&maxes, "maxes", false, "get max theoretical values per engine")
	flag.BoolVar(&getHashes, "gethashes", true, "get hashes of engine files")

	config.InitConfig()
	initerr := store.Init(getHashes)
	if initerr != nil {
		logger.Errorf("Error initializing store %s", initerr)
		os.Exit(3)
	}

	rng.Init()
	logger.Infof("API INIT: OK")
	//
	// initial serve web
	if runVT == true {
		logger.Errorf("Running VT : spins %v  chunks %v engine %v", spins, chunks, engineID)
		failed := volumeTester.RunVT(engineID, spins, chunks, perSpin, maxes)
		if failed == true {
			logger.Errorf("VT Failed, not starting server")
			os.Exit(5)
		}
		//volumeTester.GetVTInfo()
	}
	// Setup routes
	router := api.Routes()
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Infof("%s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		logger.Errorf("Error walking routes: %s\n", err.Error())
	}
	cwd, _ := os.Getwd()
	logger.Infof("Work dir: %s", cwd)
	// Start Server
	port, err := strconv.Atoi(config.GlobalConfig.Server.Port)
	if err != nil {
		logger.Errorf("Config error: %s\n", err.Error())
	}

	err = sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: config.GlobalConfig.SentryDsn,
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug:        false,
		Environment:  config.GlobalConfig.Environment,
		IgnoreErrors: []string{"Insufficient Fund", "No force matching that code", "No player found"},
	})
	if err != nil {
		logger.Errorf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	//defer sentry.Flush(20 * time.Millisecond)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: router}
	logger.Infof("Starting RGS Core on port: %d", port)
	logger.Fatalf("%v", srv.ListenAndServe())
}
