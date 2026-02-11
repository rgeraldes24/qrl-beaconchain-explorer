package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image/color"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/theQRL/qrl-beaconchain-explorer/config"
	"github.com/theQRL/qrl-beaconchain-explorer/types"

	"github.com/carlmjohnson/requests"
	"github.com/kelseyhightower/envconfig"
	"github.com/mvdan/xurls"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	confusables "github.com/skygeario/go-confusable-homoglyphs"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	qrysm_params "github.com/theQRL/qrysm/config/params"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// Config is the globally accessible configuration
var Config *types.Config

var logger = logrus.New().WithField("module", "oauth")

var HashLikeRegex = regexp.MustCompile(`^[0-9a-fA-F]{0,96}$`)

// GetTemplateFuncs will get the template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"includeHTML":                             IncludeHTML,
		"includeSvg":                              IncludeSvg,
		"formatHTML":                              FormatMessageToHtml,
		"formatBalance":                           FormatBalance,
		"formatBalanceSql":                        FormatBalanceSql,
		"formatCurrentBalance":                    FormatCurrentBalance,
		"formatElCurrency":                        FormatElCurrency,
		"formatClCurrency":                        FormatClCurrency,
		"formatEffectiveBalance":                  FormatEffectiveBalance,
		"formatBlockStatus":                       FormatBlockStatus,
		"formatBlockSlot":                         FormatBlockSlot,
		"formatSlotToTimestamp":                   FormatSlotToTimestamp,
		"formatDepositAmount":                     FormatDepositAmount,
		"formatEpoch":                             FormatEpoch,
		"fixAddressCasing":                        FixAddressCasing,
		"formatAddressLong":                       FormatAddressLong,
		"formatHashLong":                          FormatHashLong,
		"formatExecutionBlock":                    FormatExecutionBlock,
		"formatExecutionBlockHash":                FormatExecutionBlockHash,
		"formatExecutionAddress":                  FormatExecutionAddress,
		"formatExecutionAddressStringLowerCase":   FormatExecutionAddressStringLowerCase,
		"formatExecutionTxHash":                   FormatExecutionTxHash,
		"formatGraffiti":                          FormatGraffiti,
		"formatHash":                              FormatHash,
		"formatWithdawalCredentials":              FormatWithdawalCredentials,
		"formatAddressToWithdrawalCredentials":    FormatAddressToWithdrawalCredentials,
		"formatBitlist":                           FormatBitlist,
		"formatBitvectorValidators":               formatBitvectorValidators,
		"formatParticipation":                     FormatParticipation,
		"formatIncome":                            FormatIncome,
		"formatIncomeSql":                         FormatIncomeSql,
		"formatSqlInt64":                          FormatSqlInt64,
		"formatValidator":                         FormatValidator,
		"formatValidatorWithName":                 FormatValidatorWithName,
		"formatValidatorInt64":                    FormatValidatorInt64,
		"formatValidatorStatus":                   FormatValidatorStatus,
		"formatPercentage":                        FormatPercentage,
		"formatPercentageWithPrecision":           FormatPercentageWithPrecision,
		"formatPercentageWithGPrecision":          FormatPercentageWithGPrecision,
		"formatPercentageColoredEmoji":            FormatPercentageColoredEmoji,
		"formatPublicKey":                         FormatPublicKey,
		"formatSlashedValidator":                  FormatSlashedValidator,
		"formatSlashedValidatorInt64":             FormatSlashedValidatorInt64,
		"formatTimestamp":                         FormatTimestamp,
		"formatTsWithoutTooltip":                  FormatTsWithoutTooltip,
		"formatValidatorName":                     FormatValidatorName,
		"formatAttestationInclusionEffectiveness": FormatAttestationInclusionEffectiveness,
		"formatValidatorTags":                     FormatValidatorTags,
		"formatValidatorTag":                      FormatValidatorTag,
		"formatQuanta":                            FormatQuanta,
		"formatFloat":                             FormatFloat,
		"formatAmount":                            FormatAmount,
		"formatBigAmount":                         FormatBigAmount,
		"formatBytesAmount":                       FormatBytesAmount,
		"formatYesNo":                             FormatYesNo,
		"formatAmountFormatted":                   FormatAmountFormatted,
		"formatAddressAsLink":                     FormatAddressAsLink,
		"config":                                  func() *types.Config { return Config },
		"epochOfSlot":                             EpochOfSlot,
		"dayToTime":                               DayToTime,
		"contains":                                strings.Contains,
		"bigIntCmp":                               func(i *big.Int, j int) int { return i.Cmp(big.NewInt(int64(j))) },
		"mod":                                     func(i, j int) bool { return i%j == 0 },
		"sub":                                     func(i, j int) int { return i - j },
		"subUI64":                                 func(i, j uint64) uint64 { return i - j },
		"add":                                     func(i, j int) int { return i + j },
		"addI64":                                  func(i, j int64) int64 { return i + j },
		"addUI64":                                 func(i, j uint64) uint64 { return i + j },
		"mulUI64":                                 func(i, j uint64) uint64 { return i * j },
		"addFloat64":                              func(i, j float64) float64 { return i + j },
		"addBigInt":                               func(i, j *big.Int) *big.Int { return new(big.Int).Add(i, j) },
		"mul":                                     func(i, j float64) float64 { return i * j },
		"div":                                     func(i, j float64) float64 { return i / j },
		"divUI64": func(i, j uint64) uint64 {
			if j == 0 {
				return 0
			}
			return i / j
		},
		"divInt":                                  func(i, j int) float64 { return float64(i) / float64(j) },
		"nef":                                     func(i, j float64) bool { return i != j },
		"gtf":                                     func(i, j float64) bool { return i > j },
		"ltf":                                     func(i, j float64) bool { return i < j },
		"round": func(i float64, n int) float64 {
			return math.Round(i*math.Pow10(n)) / math.Pow10(n)
		},
		"percent": func(i float64) float64 { return i * 100 },
		"formatThousands": func(i float64) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%.0f\n", i)
		},
		"formatThousandsFancy": func(i float64) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%v\n", i)
		},
		"formatThousandsInt": func(i int) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%d", i)
		},
		"formatStringThousands": FormatThousandsEnglish,
		"derefString":           DerefString,
		"firstCharToUpper":      func(s string) string { return cases.Title(language.English).String(s) },
		"eqsp": func(a, b *string) bool {
			if a != nil && b != nil {
				return *a == *b
			}
			return false
		},
		"stringsJoin":     strings.Join,
		"formatAddCommas": FormatAddCommas,
		"encodeToString":  hex.EncodeToString,

		"formatTokenBalance":         FormatTokenBalance,
		"formatAddressQuantaBalance": formatAddressQuantaBalance,
		"toBase64":                   ToBase64,
		"bytesToNumberString": func(input []byte) string {
			return new(big.Int).SetBytes(input).String()
		},
		"bigDecimalShift": func(num []byte, shift []byte) string {
			numDecimal := decimal.NewFromBigInt(new(big.Int).SetBytes(num), 0)
			denomDecimal := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), new(big.Int).SetBytes(shift), nil), 0)
			res := numDecimal.DivRound(denomDecimal, 18)
			return res.String()
		},
		"trimTrailingZero": func(num string) string {
			if strings.Contains(num, ".") {
				return strings.TrimRight(strings.TrimRight(num, "0"), ".")
			}
			return num
		},
		// Execution related formatting
		"formatExecutionTxStatus":    FormatExecutionTxStatus,
		"formatExecutionAddressFull": FormatExecutionAddressFull,
		"byteToString": func(num []byte) string {
			return string(num)
		},
		"bigToInt": func(val *hexutil.Big) *big.Int {
			if val != nil {
				return val.ToInt()
			}
			return nil
		},
		"formatBigNumberAddCommasFormated": FormatBigNumberAddCommasFormated,
		"formatTokenSymbolTitle":           FormatTokenSymbolTitle,
		"formatTokenSymbol":                FormatTokenSymbol,
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// IncludeHTML adds html to the page
func IncludeHTML(path string) template.HTML {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("includeHTML - error reading file: %v", err)
		return ""
	}
	return template.HTML(string(b))
}

func GraffitiToString(graffiti []byte) string {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	s = strings.Replace(s, "\u0000", "", -1) // remove 0x00 bytes as it is not supported in postgres

	if !utf8.ValidString(s) {
		return "INVALID_UTF8_STRING"
	}

	return s
}

// FormatGraffitiString formats (and escapes) the graffiti
func FormatGraffitiString(graffiti string) string {
	return strings.Map(fixUtf, template.HTMLEscapeString(graffiti))
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}

func SyncPeriodOfEpoch(epoch uint64) uint64 {
	return epoch / Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod
}

// FirstEpochOfSyncPeriod returns the first epoch of a given sync period.
//
// For more information: https://eth2book.info/capella/annotated-spec/#sync-committee-updates
func FirstEpochOfSyncPeriod(syncPeriod uint64) uint64 {
	return syncPeriod * Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod
}

// EpochOfSlot returns the corresponding epoch of a slot
func EpochOfSlot(slot uint64) uint64 {
	return slot / Config.Chain.ClConfig.SlotsPerEpoch
}

// DayOfSlot returns the corresponding day of a slot
func DayOfSlot(slot uint64) uint64 {
	return Config.Chain.ClConfig.SecondsPerSlot * slot / (24 * 3600)
}

// SlotToTime returns a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.ClConfig.SecondsPerSlot), 0)
}

// TimeToSlot returns time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if Config.Chain.GenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.ClConfig.SecondsPerSlot
}

func TimeToFirstSlotOfEpoch(timestamp uint64) uint64 {
	slot := TimeToSlot(timestamp)
	lastEpochOffset := slot % Config.Chain.ClConfig.SlotsPerEpoch
	slot = slot - lastEpochOffset
	return slot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.ClConfig.SecondsPerSlot*Config.Chain.ClConfig.SlotsPerEpoch), 0)
}

// TimeToDay will return a days since genesis for an timestamp
func TimeToDay(timestamp uint64) uint64 {
	const hoursInADay = float64(Day / time.Hour)
	return uint64(time.Unix(int64(timestamp), 0).Sub(time.Unix(int64(Config.Chain.GenesisTimestamp), 0)).Hours() / hoursInADay)
}

func DayToTime(day int64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp), 0).Add(Day * time.Duration(day))
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(Config.Chain.GenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.ClConfig.SecondsPerSlot) / int64(Config.Chain.ClConfig.SlotsPerEpoch)
}

func PlanckToQuanta(planck *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(planck, 0).DivRound(decimal.NewFromInt(params.Quanta), 18)
}

func PlanckBytesToQuanta(planck []byte) decimal.Decimal {
	return PlanckToQuanta(new(big.Int).SetBytes(planck))
}

// WaitForCtrlC will block/wait until a control-c is pressed
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// ReadConfig will process a configuration
func ReadConfig(cfg *types.Config, path string) error {
	configPathFromEnv := os.Getenv("BEACONCHAIN_CONFIG")

	if configPathFromEnv != "" { // allow the location of the config file to be passed via env args
		path = configPathFromEnv
	}
	if strings.HasPrefix(path, "projects/") {
		x, err := AccessSecretVersion(path)
		if err != nil {
			return fmt.Errorf("error getting config from secret store: %v", err)
		}
		err = yaml.Unmarshal([]byte(*x), cfg)
		if err != nil {
			return fmt.Errorf("error decoding config file %v: %v", path, err)
		}

		logger.Infof("seeded config file from secret store")
	} else {

		err := readConfigFile(cfg, path)
		if err != nil {
			return err
		}
	}

	readConfigEnv(cfg)
	err := readConfigSecrets(cfg)
	if err != nil {
		return err
	}

	if cfg.Frontend.SiteBrand == "" {
		cfg.Frontend.SiteBrand = "QRL Explorer"
	}

	if cfg.Chain.ClConfigPath == "" {
		// var qrysmParamsConfig *qrysmParams.BeaconChainConfig
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &cfg.Chain.ClConfig)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name: %v (path: %v)", cfg.Chain.Name, cfg.Chain.ClConfigPath)
		}
		if err != nil {
			return err
		}
		// err = qrysmParams.SetActive(qrysmParamsConfig)
		// if err != nil {
		// 	return fmt.Errorf("error setting chainConfig (%v) for qrysmParams: %w", cfg.Chain.Name, err)
		// }
	} else if cfg.Chain.ClConfigPath == "node" {
		nodeEndpoint := fmt.Sprintf("http://%s:%s", cfg.Indexer.Node.Host, cfg.Indexer.Node.Port)

		jr := &types.ConfigJsonResponse{}

		err := requests.
			URL(nodeEndpoint + "/qrl/v1/config/spec").
			ToJSON(jr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		chainCfg := types.ClChainConfig{
			PresetBase:                       jr.Data.PresetBase,
			ConfigName:                       jr.Data.ConfigName,
			MinGenesisActiveValidatorCount:   mustParseUint(jr.Data.MinGenesisActiveValidatorCount),
			MinGenesisTime:                   int64(mustParseUint(jr.Data.MinGenesisTime)),
			GenesisForkVersion:               jr.Data.GenesisForkVersion,
			GenesisDelay:                     mustParseUint(jr.Data.GenesisDelay),
			SecondsPerSlot:                   mustParseUint(jr.Data.SecondsPerSlot),
			SecondsPerExecutionBlock:         mustParseUint(jr.Data.SecondsPerExecutionBlock),
			MinValidatorWithdrawabilityDelay: mustParseUint(jr.Data.MinValidatorWithdrawabilityDelay),
			ShardCommitteePeriod:             mustParseUint(jr.Data.ShardCommitteePeriod),
			ExecutionFollowDistance:          mustParseUint(jr.Data.ExecutionFollowDistance),
			InactivityScoreBias:              mustParseUint(jr.Data.InactivityScoreBias),
			InactivityScoreRecoveryRate:      mustParseUint(jr.Data.InactivityScoreRecoveryRate),
			EjectionBalance:                  mustParseUint(jr.Data.EjectionBalance),
			MinPerEpochChurnLimit:            mustParseUint(jr.Data.MinPerEpochChurnLimit),
			ChurnLimitQuotient:               mustParseUint(jr.Data.ChurnLimitQuotient),
			MaxPerEpochActivationChurnLimit:  mustParseUint(jr.Data.MaxPerEpochActivationChurnLimit),
			ProposerScoreBoost:               mustParseUint(jr.Data.ProposerScoreBoost),
			DepositChainID:                   mustParseUint(jr.Data.DepositChainID),
			DepositNetworkID:                 mustParseUint(jr.Data.DepositNetworkID),
			DepositContractAddress:           jr.Data.DepositContractAddress,
			MaxCommitteesPerSlot:             mustParseUint(jr.Data.MaxCommitteesPerSlot),
			TargetCommitteeSize:              mustParseUint(jr.Data.TargetCommitteeSize),
			MaxValidatorsPerCommittee:        mustParseUint(jr.Data.TargetCommitteeSize),
			ShuffleRoundCount:                mustParseUint(jr.Data.ShuffleRoundCount),
			HysteresisQuotient:               mustParseUint(jr.Data.HysteresisQuotient),
			HysteresisDownwardMultiplier:     mustParseUint(jr.Data.HysteresisDownwardMultiplier),
			HysteresisUpwardMultiplier:       mustParseUint(jr.Data.HysteresisUpwardMultiplier),
			SafeSlotsToUpdateJustified:       mustParseUint(jr.Data.SafeSlotsToUpdateJustified),
			MinDepositAmount:                 mustParseUint(jr.Data.MinDepositAmount),
			MaxEffectiveBalance:              mustParseUint(jr.Data.MaxEffectiveBalance),
			EffectiveBalanceIncrement:        mustParseUint(jr.Data.EffectiveBalanceIncrement),
			MinAttestationInclusionDelay:     mustParseUint(jr.Data.MinAttestationInclusionDelay),
			SlotsPerEpoch:                    mustParseUint(jr.Data.SlotsPerEpoch),
			MinSeedLookahead:                 mustParseUint(jr.Data.MinSeedLookahead),
			MaxSeedLookahead:                 mustParseUint(jr.Data.MaxSeedLookahead),
			EpochsPerExecutionVotingPeriod:   mustParseUint(jr.Data.EpochsPerExecutionVotingPeriod),
			SlotsPerHistoricalRoot:           mustParseUint(jr.Data.SlotsPerHistoricalRoot),
			MinEpochsToInactivityPenalty:     mustParseUint(jr.Data.MinEpochsToInactivityPenalty),
			EpochsPerHistoricalVector:        mustParseUint(jr.Data.EpochsPerHistoricalVector),
			EpochsPerSlashingsVector:         mustParseUint(jr.Data.EpochsPerSlashingsVector),
			HistoricalRootsLimit:             mustParseUint(jr.Data.HistoricalRootsLimit),
			ValidatorRegistryLimit:           mustParseUint(jr.Data.ValidatorRegistryLimit),
			BaseRewardFactor:                 mustParseUint(jr.Data.BaseRewardFactor),
			WhistleblowerRewardQuotient:      mustParseUint(jr.Data.WhistleblowerRewardQuotient),
			ProposerRewardQuotient:           mustParseUint(jr.Data.ProposerRewardQuotient),
			InactivityPenaltyQuotient:        mustParseUint(jr.Data.InactivityPenaltyQuotient),
			MinSlashingPenaltyQuotient:       mustParseUint(jr.Data.MinSlashingPenaltyQuotient),
			ProportionalSlashingMultiplier:   mustParseUint(jr.Data.ProportionalSlashingMultiplier),
			MaxProposerSlashings:             mustParseUint(jr.Data.MaxProposerSlashings),
			MaxAttesterSlashings:             mustParseUint(jr.Data.MaxAttesterSlashings),
			MaxAttestations:                  mustParseUint(jr.Data.MaxAttestations),
			MaxDeposits:                      mustParseUint(jr.Data.MaxDeposits),
			MaxVoluntaryExits:                mustParseUint(jr.Data.MaxVoluntaryExits),
			SyncCommitteeSize:                mustParseUint(jr.Data.SyncCommitteeSize),
			EpochsPerSyncCommitteePeriod:     mustParseUint(jr.Data.EpochsPerSyncCommitteePeriod),
			MinSyncCommitteeParticipants:     mustParseUint(jr.Data.MinSyncCommitteeParticipants),
			MaxBytesPerTransaction:           mustParseUint(jr.Data.MaxBytesPerTransaction),
			MaxTransactionsPerPayload:        mustParseUint(jr.Data.MaxTransactionsPerPayload),
			BytesPerLogsBloom:                mustParseUint(jr.Data.BytesPerLogsBloom),
			MaxExtraDataBytes:                mustParseUint(jr.Data.MaxExtraDataBytes),
			MaxWithdrawalsPerPayload:         mustParseUint(jr.Data.MaxWithdrawalsPerPayload),
			MaxValidatorsPerWithdrawalSweep:  mustParseUint(jr.Data.MaxValidatorsPerWithdrawalsSweep),
		}

		cfg.Chain.ClConfig = chainCfg

		type GenesisResponse struct {
			Data struct {
				GenesisTime           string `json:"genesis_time"`
				GenesisValidatorsRoot string `json:"genesis_validators_root"`
				GenesisForkVersion    string `json:"genesis_fork_version"`
			} `json:"data"`
		}

		gtr := &GenesisResponse{}

		err = requests.
			URL(nodeEndpoint + "/qrl/v1/beacon/genesis").
			ToJSON(gtr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		cfg.Chain.GenesisTimestamp = mustParseUint(gtr.Data.GenesisTime)
		cfg.Chain.GenesisValidatorsRoot = gtr.Data.GenesisValidatorsRoot

		logger.Infof("loaded chain config from node with genesis time %s", gtr.Data.GenesisTime)

	} else {
		f, err := os.Open(cfg.Chain.ClConfigPath)
		if err != nil {
			return fmt.Errorf("error opening Chain Config file %v: %w", cfg.Chain.ClConfigPath, err)
		}
		var chainConfig *types.ClChainConfig
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding Chain Config file %v: %v", cfg.Chain.ClConfigPath, err)
		}
		cfg.Chain.ClConfig = *chainConfig
	}

	cfg.Chain.ElConfig = &params.ChainConfig{
		ChainID: big.NewInt(int64(cfg.Chain.Id)),
	}

	cfg.Chain.Name = cfg.Chain.ClConfig.ConfigName

	if cfg.Chain.GenesisTimestamp == 0 {
		switch cfg.Chain.Name {
		// TODO(now.youtrack.cloud/issue/TZB-9)
		// case "mainnet":
		// 	cfg.Chain.GenesisTimestamp = 1606824023
		default:
			return fmt.Errorf("tried to set known genesis-timestamp, but unknown chain-name")
		}
	}

	if cfg.Chain.GenesisValidatorsRoot == "" {
		switch cfg.Chain.Name {
		case "mainnet":
			cfg.Chain.GenesisValidatorsRoot = "0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95"
		default:
			return fmt.Errorf("tried to set known genesis-validators-root, but unknown chain-name")
		}
	}

	if cfg.Frontend.SiteTitle == "" {
		cfg.Frontend.SiteTitle = "QRL Explorer"
	}

	if cfg.Frontend.Keywords == "" {
		cfg.Frontend.Keywords = "qrl block explorer, qrl block explorer, beacon chain explorer, qrl blockchain explorer"
	}

	if cfg.Chain.Id != 0 {
		switch cfg.Chain.Name {
		case "mainnet", "qrl":
			cfg.Chain.Id = 1
		}
	}

	// we check for maching chain id just for safety
	if cfg.Chain.Id != 0 && cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID {
		logrus.Fatalf("cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID: %v != %v", cfg.Chain.Id, cfg.Chain.ClConfig.DepositChainID)
	}

	cfg.Chain.Id = cfg.Chain.ClConfig.DepositChainID

	if cfg.RedisSessionStoreEndpoint == "" && cfg.RedisCacheEndpoint != "" {
		logrus.Infof("using RedisCacheEndpoint %s as RedisSessionStoreEndpoint as no dedicated RedisSessionStoreEndpoint was provided", cfg.RedisCacheEndpoint)
		cfg.RedisSessionStoreEndpoint = cfg.RedisCacheEndpoint
	}

	logrus.WithFields(logrus.Fields{
		"genesisTimestamp":       cfg.Chain.GenesisTimestamp,
		"genesisValidatorsRoot":  cfg.Chain.GenesisValidatorsRoot,
		"configName":             cfg.Chain.ClConfig.ConfigName,
		"depositChainID":         cfg.Chain.ClConfig.DepositChainID,
		"depositNetworkID":       cfg.Chain.ClConfig.DepositNetworkID,
		"depositContractAddress": cfg.Chain.ClConfig.DepositContractAddress,
	}).Infof("did init config")

	return nil
}

func mustParseUint(str string) uint64 {
	if str == "" {
		return 0
	}

	nbr, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logrus.Fatalf("fatal error parsing uint %s: %v", str, err)
	}

	return nbr
}

func readConfigFile(cfg *types.Config, path string) error {
	if path == "" {
		return yaml.Unmarshal([]byte(config.DefaultConfigYml), cfg)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file %v: %v", path, err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error decoding config file %v: %v", path, err)
	}

	return nil
}

func readConfigEnv(cfg *types.Config) error {
	return envconfig.Process("", cfg)
}

func readConfigSecrets(cfg *types.Config) error {
	return ProcessSecrets(cfg)
}

// MustParseHex will parse a string into hex
func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func IsApiRequest(r *http.Request) bool {
	query, ok := r.URL.Query()["format"]
	return ok && len(query) > 0 && query[0] == "json"
}

var qrlAddressRE = regexp.MustCompile("^Q[0-9a-fA-F]{40}$")
var withdrawalCredentialsAddressRE = regexp.MustCompile("^(0x)?" + BeginningOfSetWithdrawalCredentials + "[0-9a-fA-F]{40}$")
var txHashRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{64}$")
var zeroHashRE = regexp.MustCompile("^(0x)?0+$")
var hashRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{96}$")

// IsAddress verifies whether a string represents a qrl address.
func IsAddress(s string) bool {
	return qrlAddressRE.MatchString(s)
}

// IsValidTxHash verifies whether a string represents a valid qrl tx-hash.
func IsValidTxHash(s string) bool {
	return !zeroHashRE.MatchString(s) && txHashRE.MatchString(s)
}

// IsTxHash verifies whether a string represents a qrl tx-hash.
// In contrast to IsValidTxHash, this also returns true for the 0x0 address
func IsTxHash(s string) bool {
	return txHashRE.MatchString(s)
}

// IsHash verifies whether a string represents a qrl hash.
func IsHash(s string) bool {
	return hashRE.MatchString(s)
}

// IsValidWithdrawalCredentials verifies whether a string represents valid withdrawal credentials.
func IsValidWithdrawalCredentials(s string) bool {
	return withdrawalCredentialsAddressRE.MatchString(s)
}

// Glob walks through a directory and returns files with a given extension
func Glob(dir string, ext string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// ValidateReCAPTCHA validates a ReCaptcha server side
func ValidateReCAPTCHA(recaptchaResponse string) (bool, error) {
	// Check this URL verification details from Google
	// https://developers.google.com/recaptcha/docs/verify
	req, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {Config.Frontend.RecaptchaSecretKey},
		"response": {recaptchaResponse},
	})
	if err != nil { // Handle error from HTTP POST to Google reCAPTCHA verify server
		return false, err
	}
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body) // Read the response from Google
	if err != nil {
		return false, err
	}

	var googleResponse types.GoogleRecaptchaResponse
	err = json.Unmarshal(body, &googleResponse) // Parse the JSON response from Google
	if err != nil {
		return false, err
	}
	if len(googleResponse.ErrorCodes) > 0 {
		err = fmt.Errorf("error validating ReCaptcha %v", googleResponse.ErrorCodes)
	} else {
		err = nil
	}

	if googleResponse.Score > 0.5 {
		return true, err
	}

	return false, fmt.Errorf("score too low threshold not reached, Score: %v - Required >0.5; %v", googleResponse.Score, err)
}

func BitAtVector(b []byte, i int) bool {
	bb := b[i/8]
	return (bb & (1 << uint(i%8))) > 0
}

func FormatThousandsEnglish(number string) string {
	runes := []rune(number)
	cnt := 0
	for _, rune := range runes {
		if rune == '.' {
			break
		}
		cnt += 1
	}
	amt := cnt / 3
	rem := cnt % 3

	if rem == 0 {
		amt -= 1
	}

	res := make([]rune, 0, amt+rem)
	if amt <= 0 {
		return number
	}
	for i := 0; i < len(runes); i++ {
		if i != 0 && i == rem {
			res = append(res, ',')
			amt -= 1
		}

		if amt > 0 && i > rem && ((i-rem)%3) == 0 {
			res = append(res, ',')
			amt -= 1
		}

		res = append(res, runes[i])
	}

	return string(res)
}

// Generates a QR code for an address
// returns two transparent base64 encoded img strings for dark and light theme
// the first has a black QR code the second a white QR code
func GenerateQRCodeForAddress(address []byte) (string, string, error) {
	q, err := qrcode.New(FixAddressCasing(fmt.Sprintf("Q%x", address)), qrcode.Medium)
	if err != nil {
		return "", "", err
	}

	q.BackgroundColor = color.Transparent
	q.ForegroundColor = color.Black

	png, err := q.PNG(320)
	if err != nil {
		return "", "", err
	}

	q.ForegroundColor = color.White

	pngInverse, err := q.PNG(320)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(png), base64.StdEncoding.EncodeToString(pngInverse), nil
}

func FormatTokenSymbolTitle(symbol string) string {
	if isMaliciousToken(symbol) {
		return fmt.Sprintf("The token symbol (%s) has been hidden because it contains a URL or a confusable character", symbol)
	}
	return ""
}

func FormatTokenSymbol(symbol string) string {
	if isMaliciousToken(symbol) {
		return "[hidden-symbol] ⚠️"
	}
	return symbol
}

func isMaliciousToken(symbol string) bool {
	containsUrls := len(xurls.Relaxed.FindAllString(symbol, -1)) > 0
	isConfusable := len(confusables.IsConfusable(symbol, false, []string{"LATIN", "COMMON"})) > 0
	isMixedScript := confusables.IsMixedScript(symbol, nil)
	return containsUrls || isConfusable || isMixedScript || strings.ToUpper(symbol) == "QRL"
}

func ReverseSlice[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func AddBigInts(a, b []byte) []byte {
	return new(big.Int).Add(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b)).Bytes()
}

// GetTimeToNextWithdrawal calculates the time it takes for the validators next withdrawal to be processed.
func GetTimeToNextWithdrawal(distance uint64) time.Time {
	minTimeToWithdrawal := time.Now().Add(time.Second * time.Duration((distance/Config.Chain.ClConfig.MaxValidatorsPerWithdrawalSweep)*Config.Chain.ClConfig.SecondsPerSlot))
	timeToWithdrawal := time.Now().Add(time.Second * time.Duration((float64(distance)/float64(Config.Chain.ClConfig.MaxWithdrawalsPerPayload))*float64(Config.Chain.ClConfig.SecondsPerSlot)))

	if timeToWithdrawal.Before(minTimeToWithdrawal) {
		return minTimeToWithdrawal
	}

	return timeToWithdrawal
}

func EpochsPerDay() uint64 {
	return (uint64(Day.Seconds()) / Config.Chain.ClConfig.SlotsPerEpoch) / Config.Chain.ClConfig.SecondsPerSlot
}

func GetFirstAndLastEpochForDay(day uint64) (firstEpoch uint64, lastEpoch uint64) {
	firstEpoch = day * EpochsPerDay()
	lastEpoch = firstEpoch + EpochsPerDay() - 1
	return firstEpoch, lastEpoch
}

func GetLastBalanceInfoSlotForDay(day uint64) uint64 {
	return ((day+1)*EpochsPerDay() - 1) * Config.Chain.ClConfig.SlotsPerEpoch
}

// LogFatal logs a fatal error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogFatal is called.
func LogFatal(err error, errorMsg interface{}, callerSkip int, additionalInfos ...map[string]interface{}) {
	logErrorInfo(err, callerSkip, additionalInfos...).Fatal(errorMsg)
}

// LogError logs an error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogError is called.
func LogError(err error, errorMsg interface{}, callerSkip int, additionalInfos ...map[string]interface{}) {
	logErrorInfo(err, callerSkip, additionalInfos...).Error(errorMsg)
}

// LogError logs a warning with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogError is called.
func LogWarn(err error, errorMsg interface{}, callerSkip int, additionalInfos ...map[string]interface{}) {
	logErrorInfo(err, callerSkip, additionalInfos...).Warn(errorMsg)
}

func logErrorInfo(err error, callerSkip int, additionalInfos ...map[string]interface{}) *logrus.Entry {
	logFields := logrus.NewEntry(logrus.New())

	pc, fullFilePath, line, ok := runtime.Caller(callerSkip + 2)
	if ok {
		logFields = logFields.WithFields(logrus.Fields{
			"_file":     filepath.Base(fullFilePath),
			"_function": runtime.FuncForPC(pc).Name(),
			"_line":     line,
		})
	} else {
		logFields = logFields.WithField("runtime", "Callstack cannot be read")
	}

	errColl := []string{}
	for {
		errColl = append(errColl, fmt.Sprint(err))
		nextErr := errors.Unwrap(err)
		if nextErr != nil {
			err = nextErr
		} else {
			break
		}
	}

	errMarkSign := "~"
	for idx := 0; idx < (len(errColl) - 1); idx++ {
		errInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx, errMarkSign)
		nextErrInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx+1, errMarkSign)
		if idx == (len(errColl) - 2) {
			nextErrInfoText = fmt.Sprintf("%serror%s", errMarkSign, errMarkSign)
		}

		// Replace the last occurrence of the next error in the current error
		lastIdx := strings.LastIndex(errColl[idx], errColl[idx+1])
		if lastIdx != -1 {
			errColl[idx] = errColl[idx][:lastIdx] + nextErrInfoText + errColl[idx][lastIdx+len(errColl[idx+1]):]
		}

		errInfoText = strings.ReplaceAll(errInfoText, errMarkSign, "")
		logFields = logFields.WithField(errInfoText, errColl[idx])
	}

	if err != nil {
		logFields = logFields.WithField("errType", fmt.Sprintf("%T", err)).WithError(err)
	}

	for _, infoMap := range additionalInfos {
		for name, info := range infoMap {
			logFields = logFields.WithField(name, info)
		}
	}

	return logFields
}

func GetSigningDomain() ([]byte, error) {
	beaconConfig := qrysm_params.BeaconConfig()
	genForkVersion, err := hex.DecodeString(strings.Replace(Config.Chain.ClConfig.GenesisForkVersion, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	domain, err := signing.ComputeDomain(
		beaconConfig.DomainDeposit,
		genForkVersion,
		beaconConfig.ZeroHash[:],
	)

	if err != nil {
		return nil, err
	}

	return domain, err
}

// SlotsPerSyncCommittee returns the count of slots per sync committee period
// (might be wrong for the first sync period at atlair which might be shorter, see https://eth2book.info/capella/annotated-spec/#sync-committee-updates)
func SlotsPerSyncCommittee() uint64 {
	return Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod * Config.Chain.ClConfig.SlotsPerEpoch
}

// GetRemainingScheduledSyncDuties returns the remaining count of scheduled slots given the stats of the current period, while also accounting for exported slots.
//
// Parameters:
//   - validatorCount: the count of validators associated with the stats.
//   - stats: the current sync committee stats of the validators
//   - lastExportedEpoch: the last epoch that was exported into the validator_stats table
//   - firstEpochOfPeriod: the first epoch of the current sync committee period
func GetRemainingScheduledSyncDuties(validatorCount int, stats types.SyncCommitteesStats, lastExportedEpoch, firstEpochOfPeriod uint64) uint64 {
	// check how many sync duties remain in the current sync committee based on firstEpochOfPeriod
	slotsPerSyncCommittee := SlotsPerSyncCommittee()
	dutiesPerSyncCommittee := slotsPerSyncCommittee * uint64(validatorCount)

	// check how many duties are already exported
	exportedEpochs := uint64(0)
	if lastExportedEpoch >= firstEpochOfPeriod {
		exportedEpochs = lastExportedEpoch - firstEpochOfPeriod + 1
	}
	exportedDuties := exportedEpochs * Config.Chain.ClConfig.SlotsPerEpoch * uint64(validatorCount)

	// calculate how many duties are remaining i.e. are scheduled
	totalStats := stats.MissedSlots + stats.ParticipatedSlots + stats.ScheduledSlots
	return (dutiesPerSyncCommittee - ((exportedDuties + totalStats) % dutiesPerSyncCommittee)) % dutiesPerSyncCommittee
}

// AddSyncStats adds the sync stats of a set of validators from a given syncDutiesHistory to the given stats, if stats is nil a new stats object is created.
// Parameters:
//   - validators: the validators to add the stats for
//   - syncDutiesHistory: the sync duties history of all queried validators
//   - stats: the stats object to add the stats to, if nil a new stats object is created
func AddSyncStats(validators []uint64, syncDutiesHistory map[uint64]map[uint64]*types.ValidatorSyncParticipation, stats *types.SyncCommitteesStats) types.SyncCommitteesStats {
	if stats == nil {
		stats = &types.SyncCommitteesStats{}
	}
	for _, validator := range validators {
		v := syncDutiesHistory[validator]
		for _, r := range v {
			slotTime := SlotToTime(r.Slot)
			if r.Status == 0 && time.Since(slotTime) <= time.Minute {
				r.Status = 2
			}
			switch r.Status {
			case 0:
				stats.MissedSlots++
			case 1:
				stats.ParticipatedSlots++
			case 2:
				stats.ScheduledSlots++
			}
		}
	}
	return *stats
}

// To remove all round brackets (including its content) from a string
func RemoveRoundBracketsIncludingContent(input string) string {
	openCount := 0
	result := ""
	for {
		if len(input) == 0 {
			break
		}
		openIndex := strings.Index(input, "(")
		closeIndex := strings.Index(input, ")")
		if openIndex == -1 && closeIndex == -1 {
			if openCount == 0 {
				result += input
			}
			break
		} else if openIndex != -1 && (openIndex < closeIndex || closeIndex == -1) {
			openCount++
			if openCount == 1 {
				result += input[:openIndex]
			}
			input = input[openIndex+1:]
		} else {
			if openCount > 0 {
				openCount--
			} else if openIndex == -1 && len(result) == 0 {
				result += input[:closeIndex]
			}
			input = input[closeIndex+1:]
		}
	}
	return result
}

// Prompt asks for a string value using the label. For comand line interactions.
func CmdPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func SortedUniqueUint64(arr []uint64) []uint64 {
	if len(arr) <= 1 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	result := make([]uint64, 1, len(arr))
	result[0] = arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i-1] != arr[i] {
			result = append(result, arr[i])
		}
	}

	return result
}

type HttpReqHttpError struct {
	StatusCode int
	Url        string
	Body       []byte
}

func (err *HttpReqHttpError) Error() string {
	return fmt.Sprintf("error response: url: %s, status: %d, body: %s", err.Url, err.StatusCode, err.Body)
}

func HttpReq(ctx context.Context, method, url string, params, result interface{}) error {
	var err error
	var req *http.Request
	if params != nil {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("error marshaling params for request: %w, url: %v", err, url)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(paramsJSON))
		if err != nil {
			return fmt.Errorf("error creating request with params: %w, url: %v", err, url)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w, url: %v", err, url)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: time.Minute}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return &HttpReqHttpError{
			StatusCode: res.StatusCode,
			Url:        url,
			Body:       body,
		}
	}
	if result != nil {
		err = json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error unmarshaling response: %w, url: %v", err, url)
		}
	}
	return nil
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func GetParentFuncName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

func GetMaxAllowedDayRangeValidatorStats(validatorAmount int) int {
	if validatorAmount > 100000 {
		return 0 // exact day only
	} else if validatorAmount > 10000 {
		return 3
	} else if validatorAmount > 1000 {
		return 10
	} else {
		return math.MaxInt
	}
}

func FixAddressCasing(add string) string {
	addr, _ := common.NewAddressFromString(add)
	return addr.Hex()
}
