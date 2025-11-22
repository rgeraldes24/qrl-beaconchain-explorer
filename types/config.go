package types

import (
	"time"

	"github.com/theQRL/go-zond/params"
)

// Config is a struct to hold the configuration data
type Config struct {
	ReaderDatabase struct {
		Username     string `yaml:"user" envconfig:"READER_DB_USERNAME"`
		Password     string `yaml:"password" envconfig:"READER_DB_PASSWORD"`
		Name         string `yaml:"name" envconfig:"READER_DB_NAME"`
		Host         string `yaml:"host" envconfig:"READER_DB_HOST"`
		Port         string `yaml:"port" envconfig:"READER_DB_PORT"`
		MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"READER_DB_MAX_OPEN_CONNS"`
		MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"READER_DB_MAX_IDLE_CONNS"`
		SSL          bool   `yaml:"ssl" envconfig:"READER_DB_SSL"`
	} `yaml:"readerDatabase"`
	WriterDatabase struct {
		Username     string `yaml:"user" envconfig:"WRITER_DB_USERNAME"`
		Password     string `yaml:"password" envconfig:"WRITER_DB_PASSWORD"`
		Name         string `yaml:"name" envconfig:"WRITER_DB_NAME"`
		Host         string `yaml:"host" envconfig:"WRITER_DB_HOST"`
		Port         string `yaml:"port" envconfig:"WRITER_DB_PORT"`
		MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"WRITER_DB_MAX_OPEN_CONNS"`
		MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"WRITER_DB_MAX_IDLE_CONNS"`
		SSL          bool   `yaml:"ssl" envconfig:"WRITER_DB_SSL"`
	} `yaml:"writerDatabase"`
	Bigtable struct {
		Project      string `yaml:"project" envconfig:"BIGTABLE_PROJECT"`
		Instance     string `yaml:"instance" envconfig:"BIGTABLE_INSTANCE"`
		Emulator     bool   `yaml:"emulator" envconfig:"BIGTABLE_EMULATOR"`
		EmulatorPort int    `yaml:"emulatorPort" envconfig:"BIGTABLE_EMULATOR_PORT"`
		EmulatorHost string `yaml:"emulatorHost" envconfig:"BIGTABLE_EMULATOR_HOST"`
	} `yaml:"bigtable"`
	Chain struct {
		Name                  string `yaml:"name" envconfig:"CHAIN_NAME"`
		Id                    uint64 `yaml:"id" envconfig:"CHAIN_ID"`
		GenesisTimestamp      uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
		GenesisValidatorsRoot string `yaml:"genesisValidatorsRoot" envconfig:"CHAIN_GENESIS_VALIDATORS_ROOT"`
		// TODO(now.youtrack.cloud/issue/TZB-2)
		// DomainVoluntaryExit              string `yaml:"domainVoluntaryExit" envconfig:"CHAIN_DOMAIN_VOLUNTARY_EXIT"`
		ClConfigPath string `yaml:"clConfigPath" envconfig:"CHAIN_CL_CONFIG_PATH"`
		ClConfig     ClChainConfig
		ElConfig     *params.ChainConfig
	} `yaml:"chain"`
	ELNodeEndpoint string `yaml:"elNodeEndpoint" envconfig:"EL_NODE_ENDPOINT"`
	// TODO(now.youtrack.cloud/issue/TZB-5)
	// EtherscanAPIKey           string `yaml:"etherscanApiKey" envconfig:"ETHERSCAN_API_KEY"`
	// EtherscanAPIBaseURL       string `yaml:"etherscanApiBaseUrl" envconfig:"ETHERSCAN_API_BASEURL"`
	RedisCacheEndpoint        string `yaml:"redisCacheEndpoint" envconfig:"REDIS_CACHE_ENDPOINT"`
	RedisSessionStoreEndpoint string `yaml:"redisSessionStoreEndpoint" envconfig:"REDIS_SESSION_STORE_ENDPOINT"`
	TieredCacheProvider       string `yaml:"tieredCacheProvider" envconfig:"CACHE_PROVIDER"`
	ReportServiceStatus       bool   `yaml:"reportServiceStatus" envconfig:"REPORT_SERVICE_STATUS"`
	Indexer                   struct {
		Enabled bool `yaml:"enabled" envconfig:"INDEXER_ENABLED"`
		Node    struct {
			Port     string `yaml:"port" envconfig:"INDEXER_NODE_PORT"`
			Host     string `yaml:"host" envconfig:"INDEXER_NODE_HOST"`
			Type     string `yaml:"type" envconfig:"INDEXER_NODE_TYPE"`
			PageSize int32  `yaml:"pageSize" envconfig:"INDEXER_NODE_PAGE_SIZE"`
		} `yaml:"node"`
		DepositContractFirstBlock uint64 `yaml:"depositContractFirstBlock" envconfig:"INDEXER_DEPOSIT_CONTRACT_FIRST_BLOCK"`
		// TODO(now.youtrack.cloud/issue/TZB-1)
		// QrnsTransformer           struct {
		// 	ValidRegistrarContracts []string `yaml:"validRegistrarContracts" envconfig:"QRNS_VALID_REGISTRAR_CONTRACTS"`
		// } `yaml:"qrnsTransformer"`
	} `yaml:"indexer"`
	Frontend struct {
		Debug              bool   `yaml:"debug" envconfig:"FRONTEND_DEBUG"`
		DisableCharts      bool   `yaml:"disableCharts" envconfig:"disableCharts"`
		RecaptchaSiteKey   string `yaml:"recaptchaSiteKey" envconfig:"FRONTEND_RECAPTCHA_SITEKEY"`
		RecaptchaSecretKey string `yaml:"recaptchaSecretKey" envconfig:"FRONTEND_RECAPTCHA_SECRETKEY"`
		SiteBrand          string `yaml:"siteBrand" envconfig:"FRONTEND_SITE_BRAND"`
		Keywords           string `yaml:"keywords" envconfig:"FRONTEND_KEYWORDS"`
		// Imprint is deprdecated place imprint file into the legal directory
		Imprint string `yaml:"imprint" envconfig:"FRONTEND_IMPRINT"`
		Legal   struct {
			TermsOfServiceUrl string `yaml:"termsOfServiceUrl" envconfig:"FRONTEND_LEGAL_TERMS_OF_SERVICE_URL"`
			PrivacyPolicyUrl  string `yaml:"privacyPolicyUrl" envconfig:"FRONTEND_LEGAL_PRIVACY_POLICY_URL"`
			ImprintTemplate   string `yaml:"imprintTemplate" envconfig:"FRONTEND_LEGAL_IMPRINT_TEMPLATE"`
		} `yaml:"legal"`
		SiteDomain   string `yaml:"siteDomain" envconfig:"FRONTEND_SITE_DOMAIN"`
		SiteName     string `yaml:"siteName" envconfig:"FRONTEND_SITE_NAME"`
		SiteTitle    string `yaml:"siteTitle" envconfig:"FRONTEND_SITE_TITLE"`
		SiteSubtitle string `yaml:"siteSubtitle" envconfig:"FRONTEND_SITE_SUBTITLE"`
		Server       struct {
			Port string `yaml:"port" envconfig:"FRONTEND_SERVER_PORT"`
			Host string `yaml:"host" envconfig:"FRONTEND_SERVER_HOST"`
		} `yaml:"server"`
		ReaderDatabase struct {
			Username     string `yaml:"user" envconfig:"FRONTEND_READER_DB_USERNAME"`
			Password     string `yaml:"password" envconfig:"FRONTEND_READER_DB_PASSWORD"`
			Name         string `yaml:"name" envconfig:"FRONTEND_READER_DB_NAME"`
			Host         string `yaml:"host" envconfig:"FRONTEND_READER_DB_HOST"`
			Port         string `yaml:"port" envconfig:"FRONTEND_READER_DB_PORT"`
			MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"FRONTEND_READER_DB_MAX_OPEN_CONNS"`
			MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"FRONTEND_READER_DB_MAX_IDLE_CONNS"`
			SSL          bool   `yaml:"ssl" envconfig:"FRONTEND_READER_DB_SSL"`
		} `yaml:"readerDatabase"`
		WriterDatabase struct {
			Username     string `yaml:"user" envconfig:"FRONTEND_WRITER_DB_USERNAME"`
			Password     string `yaml:"password" envconfig:"FRONTEND_WRITER_DB_PASSWORD"`
			Name         string `yaml:"name" envconfig:"FRONTEND_WRITER_DB_NAME"`
			Host         string `yaml:"host" envconfig:"FRONTEND_WRITER_DB_HOST"`
			Port         string `yaml:"port" envconfig:"FRONTEND_WRITER_DB_PORT"`
			MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"FRONTEND_WRITER_DB_MAX_OPEN_CONNS"`
			MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"FRONTEND_WRITER_DB_MAX_IDLE_CONNS"`
			SSL          bool   `yaml:"ssl" envconfig:"FRONTEND_WRITER_DB_SSL"`
		} `yaml:"writerDatabase"`
		SessionSameSiteNone                  bool          `yaml:"sessionSameSiteNone" envconfig:"FRONTEND_SESSION_SAMESITE_NONE"`
		SessionSecret                        string        `yaml:"sessionSecret" envconfig:"FRONTEND_SESSION_SECRET"`
		SessionCookieDomain                  string        `yaml:"sessionCookieDomain" envconfig:"FRONTEND_SESSION_COOKIE_DOMAIN"`
		SessionCookieDeriveDomainFromRequest bool          `yaml:"sessionCookieDeriveDomainFromRequest" envconfig:"FRONTEND_SESSION_COOKIE_DERIVE_DOMAIN_FROM_REQUEST"`
		GATag                                string        `yaml:"gatag" envconfig:"GATAG"`
		HttpReadTimeout                      time.Duration `yaml:"httpReadTimeout" envconfig:"FRONTEND_HTTP_READ_TIMEOUT"`
		HttpWriteTimeout                     time.Duration `yaml:"httpWriteTimeout" envconfig:"FRONTEND_HTTP_WRITE_TIMEOUT"`
		HttpIdleTimeout                      time.Duration `yaml:"httpIdleTimeout" envconfig:"FRONTEND_HTTP_IDLE_TIMEOUT"`
	} `yaml:"frontend"`
	Metrics struct {
		Enabled bool   `yaml:"enabled" envconfig:"METRICS_ENABLED"`
		Address string `yaml:"address" envconfig:"METRICS_ADDRESS"`
		Pprof   bool   `yaml:"pprof" envconfig:"METRICS_PPROF"`
	} `yaml:"metrics"`
	Pprof struct {
		Enabled bool   `yaml:"enabled" envconfig:"PPROF_ENABLED"`
		Port    string `yaml:"port" envconfig:"PPROF_PORT"`
	} `yaml:"pprof"`
	// TODO(now.youtrack.cloud/issue/TZB-2)
	// NodeJobsProcessor struct {
	// 	ElEndpoint string `yaml:"elEndpoint" envconfig:"NODE_JOBS_PROCESSOR_EL_ENDPOINT"`
	// 	ClEndpoint string `yaml:"clEndpoint" envconfig:"NODE_JOBS_PROCESSOR_CL_ENDPOINT"`
	// } `yaml:"nodeJobsProcessor"`
	Monitoring struct {
		ServiceMonitoringConfigurations []ServiceMonitoringConfiguration `yaml:"serviceMonitoringConfigurations" envconfig:"SERVICE_MONITORING_CONFIGURATIONS"`
	} `yaml:"monitoring"`
	GithubApiHost string `yaml:"githubApiHost" envconfig:"GITHUB_API_HOST"`
}

type DatabaseConfig struct {
	Username     string
	Password     string
	Name         string
	Host         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
	SSL          bool
}

type ServiceMonitoringConfiguration struct {
	Name     string        `yaml:"name" envconfig:"NAME"`
	Duration time.Duration `yaml:"duration" envconfig:"DURATION"`
}

type ConfigJsonResponse struct {
	Data struct {
		ConfigName                           string `json:"CONFIG_NAME"`
		PresetBase                           string `json:"PRESET_BASE"`
		SafeSlotsToImportOptimistically      string `json:"SAFE_SLOTS_TO_IMPORT_OPTIMISTICALLY"`
		MinGenesisActiveValidatorCount       string `json:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
		MinGenesisTime                       string `json:"MIN_GENESIS_TIME"`
		GenesisForkVersion                   string `json:"GENESIS_FORK_VERSION"`
		GenesisDelay                         string `json:"GENESIS_DELAY"`
		SecondsPerSlot                       string `json:"SECONDS_PER_SLOT"`
		SecondsPerExecutionBlock             string `json:"SECONDS_PER_EXECUTION_BLOCK"`
		MinValidatorWithdrawabilityDelay     string `json:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"`
		ShardCommitteePeriod                 string `json:"SHARD_COMMITTEE_PERIOD"`
		ExecutionFollowDistance              string `json:"EXECUTION_FOLLOW_DISTANCE"`
		SubnetsPerNode                       string `json:"SUBNETS_PER_NODE"`
		InactivityScoreBias                  string `json:"INACTIVITY_SCORE_BIAS"`
		InactivityScoreRecoveryRate          string `json:"INACTIVITY_SCORE_RECOVERY_RATE"`
		EjectionBalance                      string `json:"EJECTION_BALANCE"`
		MinPerEpochChurnLimit                string `json:"MIN_PER_EPOCH_CHURN_LIMIT"`
		ChurnLimitQuotient                   string `json:"CHURN_LIMIT_QUOTIENT"`
		MaxPerEpochActivationChurnLimit      string `json:"MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT"`
		ProposerScoreBoost                   string `json:"PROPOSER_SCORE_BOOST"`
		DepositChainID                       string `json:"DEPOSIT_CHAIN_ID"`
		DepositNetworkID                     string `json:"DEPOSIT_NETWORK_ID"`
		DepositContractAddress               string `json:"DEPOSIT_CONTRACT_ADDRESS"`
		MaxCommitteesPerSlot                 string `json:"MAX_COMMITTEES_PER_SLOT"`
		TargetCommitteeSize                  string `json:"TARGET_COMMITTEE_SIZE"`
		MaxValidatorsPerCommittee            string `json:"MAX_VALIDATORS_PER_COMMITTEE"`
		ShuffleRoundCount                    string `json:"SHUFFLE_ROUND_COUNT"`
		HysteresisQuotient                   string `json:"HYSTERESIS_QUOTIENT"`
		HysteresisDownwardMultiplier         string `json:"HYSTERESIS_DOWNWARD_MULTIPLIER"`
		HysteresisUpwardMultiplier           string `json:"HYSTERESIS_UPWARD_MULTIPLIER"`
		SafeSlotsToUpdateJustified           string `json:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED"`
		MinDepositAmount                     string `json:"MIN_DEPOSIT_AMOUNT"`
		MaxEffectiveBalance                  string `json:"MAX_EFFECTIVE_BALANCE"`
		EffectiveBalanceIncrement            string `json:"EFFECTIVE_BALANCE_INCREMENT"`
		MinAttestationInclusionDelay         string `json:"MIN_ATTESTATION_INCLUSION_DELAY"`
		SlotsPerEpoch                        string `json:"SLOTS_PER_EPOCH"`
		MinSeedLookahead                     string `json:"MIN_SEED_LOOKAHEAD"`
		MaxSeedLookahead                     string `json:"MAX_SEED_LOOKAHEAD"`
		EpochsPerExecutionVotingPeriod       string `json:"EPOCHS_PER_EXECUTION_VOTING_PERIOD"`
		SlotsPerHistoricalRoot               string `json:"SLOTS_PER_HISTORICAL_ROOT"`
		MinEpochsToInactivityPenalty         string `json:"MIN_EPOCHS_TO_INACTIVITY_PENALTY"`
		EpochsPerHistoricalVector            string `json:"EPOCHS_PER_HISTORICAL_VECTOR"`
		EpochsPerSlashingsVector             string `json:"EPOCHS_PER_SLASHINGS_VECTOR"`
		HistoricalRootsLimit                 string `json:"HISTORICAL_ROOTS_LIMIT"`
		ValidatorRegistryLimit               string `json:"VALIDATOR_REGISTRY_LIMIT"`
		BaseRewardFactor                     string `json:"BASE_REWARD_FACTOR"`
		WhistleblowerRewardQuotient          string `json:"WHISTLEBLOWER_REWARD_QUOTIENT"`
		ProposerRewardQuotient               string `json:"PROPOSER_REWARD_QUOTIENT"`
		InactivityPenaltyQuotient            string `json:"INACTIVITY_PENALTY_QUOTIENT"`
		MinSlashingPenaltyQuotient           string `json:"MIN_SLASHING_PENALTY_QUOTIENT"`
		ProportionalSlashingMultiplier       string `json:"PROPORTIONAL_SLASHING_MULTIPLIER"`
		MaxProposerSlashings                 string `json:"MAX_PROPOSER_SLASHINGS"`
		MaxAttesterSlashings                 string `json:"MAX_ATTESTER_SLASHINGS"`
		MaxAttestations                      string `json:"MAX_ATTESTATIONS"`
		MaxDeposits                          string `json:"MAX_DEPOSITS"`
		MaxVoluntaryExits                    string `json:"MAX_VOLUNTARY_EXITS"`
		SyncCommitteeSize                    string `json:"SYNC_COMMITTEE_SIZE"`
		EpochsPerSyncCommitteePeriod         string `json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`
		MinSyncCommitteeParticipants         string `json:"MIN_SYNC_COMMITTEE_PARTICIPANTS"`
		MaxBytesPerTransaction               string `json:"MAX_BYTES_PER_TRANSACTION"`
		MaxTransactionsPerPayload            string `json:"MAX_TRANSACTIONS_PER_PAYLOAD"`
		BytesPerLogsBloom                    string `json:"BYTES_PER_LOGS_BLOOM"`
		MaxExtraDataBytes                    string `json:"MAX_EXTRA_DATA_BYTES"`
		MaxWithdrawalsPerPayload             string `json:"MAX_WITHDRAWALS_PER_PAYLOAD"`
		MaxValidatorsPerWithdrawalsSweep     string `json:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP"`
		DomainAggregateAndProof              string `json:"DOMAIN_AGGREGATE_AND_PROOF"`
		TargetAggregatorsPerSyncSubcommittee string `json:"TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE"`
		SyncCommitteeSubnetCount             string `json:"SYNC_COMMITTEE_SUBNET_COUNT"`
		DomainRandao                         string `json:"DOMAIN_RANDAO"`
		DomainVoluntaryExit                  string `json:"DOMAIN_VOLUNTARY_EXIT"`
		DomainSyncCommitteeSelectionProof    string `json:"DOMAIN_SYNC_COMMITTEE_SELECTION_PROOF"`
		DomainBeaconAttester                 string `json:"DOMAIN_BEACON_ATTESTER"`
		DomainBeaconProposer                 string `json:"DOMAIN_BEACON_PROPOSER"`
		DomainDeposit                        string `json:"DOMAIN_DEPOSIT"`
		DomainSelectionProof                 string `json:"DOMAIN_SELECTION_PROOF"`
		DomainSyncCommittee                  string `json:"DOMAIN_SYNC_COMMITTEE"`
		TargetAggregatorsPerCommittee        string `json:"TARGET_AGGREGATORS_PER_COMMITTEE"`
		DomainContributionAndProof           string `json:"DOMAIN_CONTRIBUTION_AND_PROOF"`
		DomainApplicationMask                string `json:"DOMAIN_APPLICATION_MASK"`
	} `json:"data"`
}
