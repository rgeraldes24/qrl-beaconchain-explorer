package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/theQRL/qrl-beaconchain-explorer/cache"
	"github.com/theQRL/qrl-beaconchain-explorer/db"
	"github.com/theQRL/qrl-beaconchain-explorer/types"
	"github.com/theQRL/qrl-beaconchain-explorer/utils"
)

func statsUpdater(wg *sync.WaitGroup) {
	sleepDuration := time.Duration(time.Duration(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second)

	logger.Infof("sleep duration is %v", sleepDuration)
	firstrun := true
	for {
		latestEpoch := LatestEpoch()

		now := time.Now()
		statResult, err := calculateStats()
		if err != nil {
			logger.WithField("epoch", latestEpoch).Errorf("error updating stats: %v", err)
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(now)).Info("stats update completed")

		cacheKey := fmt.Sprintf("%d:frontend:latestStats", utils.Config.Chain.ClConfig.DepositChainID)
		err = cache.TieredCache.Set(cacheKey, statResult, utils.Day)
		if err != nil {
			logger.Errorf("error caching latestStats: %v", err)
		}
		if firstrun {
			wg.Done()
			firstrun = false
		}
		ReportStatus("statsUpdater", "Running", nil)
		time.Sleep(sleepDuration)
	}
}

func calculateStats() (*types.Stats, error) {
	stats := types.Stats{}

	topDeposits, err := executionTopDepositers()
	if err != nil {
		return nil, err
	}
	stats.TopDepositors = topDeposits
	invalidCount, err := executionInvalidDeposits()
	if err != nil {
		return nil, err
	}
	stats.InvalidDepositCount = invalidCount

	uniqueValidatorCount, err := executionUniqueValidatorsCount()
	if err != nil {
		return nil, err
	}
	stats.UniqueValidatorCount = uniqueValidatorCount

	totalValidatorCount, err := db.GetTotalValidatorsCount()
	if err != nil {
		logger.WithError(err).Error("error getting total validator count")
	}
	stats.TotalValidatorCount = &totalValidatorCount

	activeValidatorCount, err := db.GetActiveValidatorCount()
	if err != nil {
		logger.WithError(err).Error("error getting active validator count")
	}

	stats.ActiveValidatorCount = &activeValidatorCount

	pendingValidatorCount, err := db.GetPendingValidatorCount()
	if err != nil {
		logger.WithError(err).Error("error getting pending validator count")
	}

	stats.PendingValidatorCount = &pendingValidatorCount

	validatorChurnLimit, err := getValidatorChurnLimit(activeValidatorCount)
	if err != nil {
		logger.WithError(err).Error("error getting total validator churn limit")
	}

	stats.ValidatorChurnLimit = &validatorChurnLimit

	epoch := LatestEpoch()
	validatorActivationChurnLimit, err := getValidatorActivationChurnLimit(activeValidatorCount, epoch)
	if err != nil {
		logger.WithError(err).Error("error getting total validator churn limit")
	}

	stats.ValidatorActivationChurnLimit = &validatorActivationChurnLimit

	LatestValidatorWithdrawalIndex, err := db.GetMostRecentWithdrawalValidator()
	if err != nil {
		logger.WithError(err).Error("error getting most recent withdrawal validator index")
	}

	stats.LatestValidatorWithdrawalIndex = &LatestValidatorWithdrawalIndex

	WithdrawableValidatorCount, err := db.GetWithdrawableValidatorCount(epoch)
	if err != nil {
		logger.WithError(err).Error("error getting withdrawable validator count")
	}

	stats.WithdrawableValidatorCount = &WithdrawableValidatorCount

	TotalAmountWithdrawn, WithdrawalCount, err := db.GetTotalAmountWithdrawn()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn")
	}
	stats.TotalAmountWithdrawn = &TotalAmountWithdrawn
	stats.WithdrawalCount = &WithdrawalCount

	TotalAmountDeposited, err := db.GetTotalAmountDeposited()
	if err != nil {
		logger.WithError(err).Error("error getting total deposited")
	}

	stats.TotalAmountDeposited = &TotalAmountDeposited

	return &stats, nil
}

func executionTopDepositers() (*[]types.StatsTopDepositors, error) {
	topDepositors := []types.StatsTopDepositors{}

	err := db.WriterDb.Select(&topDepositors, `
	SELECT 
		ENCODE(from_address::bytea, 'hex') as from_address, 
		count(from_address) as count 
	FROM execution_deposits
	where valid_signature = true 
	GROUP BY 
		from_address
	ORDER BY count DESC LIMIT 5;
	`)
	if err != nil {
		return nil, err
	}

	return &topDepositors, nil
}

func executionInvalidDeposits() (*uint64, error) {
	count := uint64(0)

	err := db.WriterDb.Get(&count, `
	SELECT 
		count(*) as count
	FROM execution_deposits
	WHERE
	  valid_signature = false
	`)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

func executionUniqueValidatorsCount() (*uint64, error) {
	count := uint64(0)

	err := db.WriterDb.Get(&count, `
	SELECT 
		count(*) as count
	FROM 
	(
		SELECT 
			publickey, 
			sum(amount) 
		FROM 
			execution_deposits 
		WHERE 
			valid_signature = true 
		GROUP BY 
			publickey 
		HAVING sum(amount) >= 40000e9
	) as q;
	`)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// getValidatorActivationChurnLimit returns the rate at which validators can enter the system, see https://eips.ethereum.org/EIPS/eip-7514
func getValidatorActivationChurnLimit(validatorCount, epoch uint64) (uint64, error) {
	vcl, err := getValidatorChurnLimit(validatorCount)
	if err != nil {
		return 0, err
	}
	if vcl > utils.Config.Chain.ClConfig.MaxPerEpochActivationChurnLimit {
		return utils.Config.Chain.ClConfig.MaxPerEpochActivationChurnLimit, nil
	}
	return vcl, nil
}

// getValidatorChurnLimit returns the rate at which validators can leave the system
func getValidatorChurnLimit(validatorCount uint64) (uint64, error) {
	min := utils.Config.Chain.ClConfig.MinPerEpochChurnLimit

	adaptable := uint64(0)
	if validatorCount > 0 {
		adaptable = validatorCount / utils.Config.Chain.ClConfig.ChurnLimitQuotient
	}

	if min > adaptable {
		return min, nil
	}

	return adaptable, nil
}
