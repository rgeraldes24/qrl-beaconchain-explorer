package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/theQRL/qrl-beaconchain-explorer/cache"
	"github.com/theQRL/qrl-beaconchain-explorer/db"
	"github.com/theQRL/qrl-beaconchain-explorer/exporter"
	"github.com/theQRL/qrl-beaconchain-explorer/handlers"
	"github.com/theQRL/qrl-beaconchain-explorer/metrics"
	qrlclients "github.com/theQRL/qrl-beaconchain-explorer/qrlClients"
	"github.com/theQRL/qrl-beaconchain-explorer/ratelimit"
	"github.com/theQRL/qrl-beaconchain-explorer/rpc"
	"github.com/theQRL/qrl-beaconchain-explorer/services"
	"github.com/theQRL/qrl-beaconchain-explorer/static"
	"github.com/theQRL/qrl-beaconchain-explorer/types"
	"github.com/theQRL/qrl-beaconchain-explorer/utils"
	"github.com/theQRL/qrl-beaconchain-explorer/version"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

func init() {
	gob.Register(types.DataTableSaveState{})
}

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithFields(logrus.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}).Printf("starting")

	if utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 || utils.Config.Chain.ClConfig.SecondsPerSlot == 0 {
		utils.LogFatal(err, "invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file", 0)
	}

	if utils.Config.Pprof.Enabled {
		go func() {
			logrus.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			logrus.Info(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", utils.Config.Pprof.Port), nil))
		}()
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			SSL:          cfg.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitFrontendDB(&types.DatabaseConfig{
			Username:     cfg.Frontend.WriterDatabase.Username,
			Password:     cfg.Frontend.WriterDatabase.Password,
			Name:         cfg.Frontend.WriterDatabase.Name,
			Host:         cfg.Frontend.WriterDatabase.Host,
			Port:         cfg.Frontend.WriterDatabase.Port,
			MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.Frontend.ReaderDatabase.Username,
			Password:     cfg.Frontend.ReaderDatabase.Password,
			Name:         cfg.Frontend.ReaderDatabase.Name,
			Host:         cfg.Frontend.ReaderDatabase.Host,
			Port:         cfg.Frontend.ReaderDatabase.Port,
			MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		rpc.CurrentGzondClient, err = rpc.NewGzondClient(utils.Config.ELNodeEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing gzond client: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			logrus.Fatalf("error connecting to bigtable: %v", err)
		}
		db.BigtableClient = bt
	}()

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			logrus.Infof("tiered Cache initialized, latest finalized epoch: %v", services.LatestFinalizedEpoch())

		}()
	}

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("no cache provider set, please set TierdCacheProvider (example redis)")
	}

	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()
	defer db.BigtableClient.Close()

	if utils.Config.Metrics.Enabled {
		go metrics.MonitorDB(db.WriterDb)
		DBInfo := []string{
			cfg.WriterDatabase.Username,
			cfg.WriterDatabase.Password,
			cfg.WriterDatabase.Host,
			cfg.WriterDatabase.Port,
			cfg.WriterDatabase.Name}
		DBStr := strings.Join(DBInfo, "-")
		frontendDBInfo := []string{
			cfg.Frontend.WriterDatabase.Username,
			cfg.Frontend.WriterDatabase.Password,
			cfg.Frontend.WriterDatabase.Host,
			cfg.Frontend.WriterDatabase.Port,
			cfg.Frontend.WriterDatabase.Name}
		frontendDBStr := strings.Join(frontendDBInfo, "-")
		if DBStr != frontendDBStr {
			go metrics.MonitorDB(db.FrontendWriterDB)
		}
	}

	logrus.Infof("database connection established")

	if utils.Config.Indexer.Enabled {
		var rpcClient rpc.Client

		chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
		if utils.Config.Indexer.Node.Type == "qrysm" {
			rpcClient, err = rpc.NewQrysmClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainID)
			if err != nil {
				utils.LogFatal(err, "new explorer qrysm client error", 0)
			}
		} else {
			logrus.Fatalf("invalid note type %v specified. supported node types are: qrysm", utils.Config.Indexer.Node.Type)
		}

		go exporter.Start(rpcClient)
	}

	services.ReportStatus("frontend", "Running", nil)

	router := mux.NewRouter()

	router.HandleFunc("/api/healthz", handlers.ApiHealthz).Methods("GET", "HEAD")
	router.HandleFunc("/api/healthz-loadbalancer", handlers.ApiHealthzLoadbalancer).Methods("GET", "HEAD")

	if !utils.Config.Frontend.Debug {
		logrus.Infof("initializing qrlclients")
		qrlclients.Init()
		logrus.Infof("qrlclients initialized")
	}

	if cfg.Frontend.SessionSecret == "" {
		logrus.Fatal("session secret is empty, please provide a secure random string.")
		return
	}

	utils.InitSessionStore(cfg.Frontend.SessionSecret)

	if utils.Config.Frontend.SiteDomain == "" {
		utils.Config.Frontend.SiteDomain = "explorer.zond.theqrl.org"
	}

	router.HandleFunc("/", handlers.Index).Methods("GET")
	router.HandleFunc("/latestState", handlers.LatestState).Methods("GET")
	router.HandleFunc("/launchMetrics", handlers.SlotVizMetrics).Methods("GET")
	router.HandleFunc("/index/data", handlers.IndexPageData).Methods("GET")
	router.HandleFunc("/slot/{slotOrHash}", handlers.Slot).Methods("GET")
	router.HandleFunc("/slot/{slotOrHash}/deposits", handlers.SlotDepositData).Methods("GET")
	router.HandleFunc("/slot/{slotOrHash}/votes", handlers.SlotVoteData).Methods("GET")
	router.HandleFunc("/slot/{slot}/attestations", handlers.SlotAttestationsData).Methods("GET")
	router.HandleFunc("/slot/{slot}/withdrawals", handlers.SlotWithdrawalData).Methods("GET")
	router.HandleFunc("/slots/finder", handlers.SlotFinder).Methods("GET")
	router.HandleFunc("/slots", handlers.Slots).Methods("GET")
	router.HandleFunc("/slots/data", handlers.SlotsData).Methods("GET")
	router.HandleFunc("/blocks", handlers.ExecutionBlocks).Methods("GET")
	router.HandleFunc("/blocks/data", handlers.ExecutionBlocksData).Methods("GET")
	router.HandleFunc("/address/{address}", handlers.ExecutionAddress).Methods("GET")
	router.HandleFunc("/address/{address}/blocks", handlers.ExecutionAddressBlocksMined).Methods("GET")
	router.HandleFunc("/address/{address}/withdrawals", handlers.ExecutionAddressWithdrawals).Methods("GET")
	router.HandleFunc("/address/{address}/transactions", handlers.ExecutionAddressTransactions).Methods("GET")
	router.HandleFunc("/address/{address}/internalTxns", handlers.ExecutionAddressInternalTransactions).Methods("GET")
	router.HandleFunc("/address/{address}/sqrctf1", handlers.ExecutionAddressSqrcTf1Transactions).Methods("GET")
	router.HandleFunc("/address/{address}/sqrctn1", handlers.ExecutionAddressSqrcTn1Transactions).Methods("GET")
	router.HandleFunc("/address/{address}/sqrctb1", handlers.ExecutionAddressSqrcTb1Transactions).Methods("GET")
	router.HandleFunc("/token/{token}", handlers.ExecutionToken).Methods("GET")
	router.HandleFunc("/token/{token}/transfers", handlers.ExecutionTokenTransfers).Methods("GET")
	router.HandleFunc("/transactions", handlers.ExecutionTransactions).Methods("GET")
	router.HandleFunc("/transactions/data", handlers.ExecutionTransactionsData).Methods("GET")
	router.HandleFunc("/block/{block}", handlers.ExecutionBlock).Methods("GET")
	router.HandleFunc("/block/{block}/transactions", handlers.BlockTransactionsData).Methods("GET")
	router.HandleFunc("/tx/{hash}", handlers.ExecutionTransactionTx).Methods("GET")
	router.HandleFunc("/tx/{hash}/data", handlers.ExecutionTransactionTxData).Methods("GET")
	router.HandleFunc("/mempool", handlers.MempoolView).Methods("GET")
	router.HandleFunc("/burn", handlers.Burn).Methods("GET")
	router.HandleFunc("/burn/data", handlers.BurnPageData).Methods("GET")
	router.HandleFunc("/gasnow", handlers.GasNow).Methods("GET")
	router.HandleFunc("/gasnow/data", handlers.GasNowData).Methods("GET")

	router.HandleFunc("/vis", handlers.Vis).Methods("GET")
	router.HandleFunc("/charts", handlers.Charts).Methods("GET")
	router.HandleFunc("/charts/{chart}", handlers.Chart).Methods("GET")
	router.HandleFunc("/charts/{chart}/data", handlers.GenericChartData).Methods("GET")
	router.HandleFunc("/vis/blocks", handlers.VisBlocks).Methods("GET")
	router.HandleFunc("/epoch/{epoch}", handlers.Epoch).Methods("GET")
	router.HandleFunc("/epochs", handlers.Epochs).Methods("GET")
	router.HandleFunc("/epochs/data", handlers.EpochsData).Methods("GET")

	router.HandleFunc("/validator/{index}", handlers.Validator).Methods("GET")
	router.HandleFunc("/validator/{index}/proposedblocks", handlers.ValidatorProposedBlocks).Methods("GET")
	router.HandleFunc("/validator/{index}/attestations", handlers.ValidatorAttestations).Methods("GET")
	router.HandleFunc("/validator/{index}/withdrawals", handlers.ValidatorWithdrawals).Methods("GET")
	router.HandleFunc("/validator/{index}/sync", handlers.ValidatorSync).Methods("GET")
	router.HandleFunc("/validator/{index}/history", handlers.ValidatorHistory).Methods("GET")
	router.HandleFunc("/validator/{pubkey}/deposits", handlers.ValidatorDeposits).Methods("GET")
	router.HandleFunc("/validator/{index}/slashings", handlers.ValidatorSlashings).Methods("GET")
	router.HandleFunc("/validator/{index}/effectiveness", handlers.ValidatorAttestationInclusionEffectiveness).Methods("GET")
	router.HandleFunc("/validator/{index}/stats", handlers.ValidatorStatsTable).Methods("GET")
	router.HandleFunc("/validators", handlers.Validators).Methods("GET")
	router.HandleFunc("/validators/data", handlers.ValidatorsData).Methods("GET")
	router.HandleFunc("/validators/slashings", handlers.ValidatorsSlashings).Methods("GET")
	router.HandleFunc("/validators/slashings/data", handlers.ValidatorsSlashingsData).Methods("GET")
	router.HandleFunc("/validators/withdrawals", handlers.Withdrawals).Methods("GET")
	router.HandleFunc("/validators/withdrawals/data", handlers.WithdrawalsData).Methods("GET")
	router.HandleFunc("/validators/deposits", handlers.Deposits).Methods("GET")
	router.HandleFunc("/validators/initiated-deposits/data", handlers.ExecutionDepositsData).Methods("GET")
	router.HandleFunc("/validators/included-deposits/data", handlers.ConsensusDepositsData).Methods("GET")

	router.HandleFunc("/heatmap", handlers.Heatmap).Methods("GET")

	router.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")
	router.HandleFunc("/dashboard/data/allbalances", handlers.DashboardDataBalanceCombined).Methods("GET")
	router.HandleFunc("/dashboard/data/proposals", handlers.DashboardDataProposals).Methods("GET")
	router.HandleFunc("/dashboard/data/proposalshistory", handlers.DashboardDataProposalsHistory).Methods("GET")
	router.HandleFunc("/dashboard/data/validators", handlers.DashboardDataValidators).Methods("GET")
	router.HandleFunc("/dashboard/data/withdrawal", handlers.DashboardDataWithdrawals).Methods("GET")
	router.HandleFunc("/dashboard/data/effectiveness", handlers.DashboardDataEffectiveness).Methods("GET")
	router.HandleFunc("/dashboard/data/earnings", handlers.DashboardDataEarnings).Methods("GET")
	router.HandleFunc("/calculator", handlers.StakingCalculator).Methods("GET")
	router.HandleFunc("/search", handlers.Search).Methods("POST")
	router.HandleFunc("/search/{type}/{search}", handlers.SearchAhead).Methods("GET")
	router.HandleFunc("/imprint", handlers.Imprint).Methods("GET")
	router.HandleFunc("/tools/unitConverter", handlers.UnitConverter).Methods("GET")

	router.HandleFunc("/tables/{tableId}/state", handlers.GetDataTableStateChanges).Methods("GET")
	router.HandleFunc("/tables/{tableId}/state", handlers.SetDataTableStateChanges).Methods("PUT")

	// TODO(now.youtrack.cloud/issue/TZB-1)
	// router.HandleFunc("/qrns/{search}", handlers.QrnsSearch).Methods("GET")

	router.HandleFunc("/qrlClients", handlers.QRLClientsServices).Methods("GET")

	// TODO(now.youtrack.cloud/issue/TZB-13)
	// router.HandleFunc("/rewards", handlers.ValidatorRewards).Methods("GET")
	router.HandleFunc("/rewards/hist", handlers.RewardsHistoricalData).Methods("GET")
	router.HandleFunc("/rewards/hist/download", handlers.DownloadRewardsHistoricalData).Methods("GET")

	router.HandleFunc("/monitoring/{module}", handlers.Monitoring).Methods("GET", "OPTIONS")

	if utils.Config.Frontend.Debug {
		// serve files from local directory when debugging, instead of from go embed file
		templatesHandler := http.FileServer(http.Dir("templates"))
		router.PathPrefix("/templates").Handler(http.StripPrefix("/templates/", templatesHandler))

		cssHandler := http.FileServer(http.Dir("static/css"))
		router.PathPrefix("/css").Handler(http.StripPrefix("/css/", cssHandler))

		jsHandler := http.FileServer(http.Dir("static/js"))
		router.PathPrefix("/js").Handler(http.StripPrefix("/js/", jsHandler))
	}
	fileSys := http.FS(static.Files)
	router.PathPrefix("/").Handler(handlers.CustomFileServer(http.FileServer(fileSys), fileSys, handlers.NotFound))

	if utils.Config.Metrics.Enabled {
		router.Use(metrics.HttpMiddleware)
	}

	ratelimit.Init()
	router.Use(ratelimit.HttpMiddleware)

	n := negroni.New(negroni.NewRecovery())
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	pa := &proxyaddr.ProxyAddr{}
	pa.Init(proxyaddr.CIDRLoopback)
	n.Use(pa)

	n.UseHandler(utils.SessionStore.SCS.LoadAndSave(router))
	if utils.Config.Frontend.HttpWriteTimeout == 0 {
		utils.Config.Frontend.HttpWriteTimeout = time.Second * 15
	}
	if utils.Config.Frontend.HttpReadTimeout == 0 {
		utils.Config.Frontend.HttpReadTimeout = time.Second * 15
	}
	if utils.Config.Frontend.HttpIdleTimeout == 0 {
		utils.Config.Frontend.HttpIdleTimeout = time.Minute
	}
	srv := &http.Server{
		Addr:         cfg.Frontend.Server.Host + ":" + cfg.Frontend.Server.Port,
		WriteTimeout: utils.Config.Frontend.HttpWriteTimeout,
		ReadTimeout:  utils.Config.Frontend.HttpReadTimeout,
		IdleTimeout:  utils.Config.Frontend.HttpIdleTimeout,
		Handler:      n,
	}

	logrus.Printf("http server listening on %v", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.WithError(err).Fatal("Error serving frontend")
		}
	}()

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
