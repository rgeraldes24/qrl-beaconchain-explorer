package exporter

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/theQRL/qrl-beaconchain-explorer/db"
	"github.com/theQRL/qrl-beaconchain-explorer/metrics"
	"github.com/theQRL/qrl-beaconchain-explorer/types"
	"github.com/theQRL/qrl-beaconchain-explorer/utils"

	"github.com/sirupsen/logrus"
	qrl "github.com/theQRL/go-zond"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	gzondTypes "github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/qrlclient"
	gzondRPC "github.com/theQRL/go-zond/rpc"
	"github.com/theQRL/qrysm/contracts/deposit"
	"github.com/theQRL/qrysm/crypto/hash"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrymspb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

var elLookBack = uint64(100)
var elMaxFetch = uint64(1000)
var elDepositEventSignature = hash.Keccak256([]byte("DepositEvent(bytes,bytes,bytes,bytes,bytes)"))
var depositContractFirstBlock uint64
var qrlDepositContractAddress common.Address
var elClient *qrlclient.Client
var elRPCClient *gzondRPC.Client
var gzondRequestEntityTooLargeRE = regexp.MustCompile("413 Request Entity Too Large")

// executionDepositsExporter regularly fetches the depositcontract-logs of the
// last 100 blocks and exports the deposits into the database.
// If a reorg of the execution-chain happened within these 100 blocks it will delete
// removed deposits.
func executionDepositsExporter() {
	var err error
	qrlDepositContractAddress, err = common.NewAddressFromString(utils.Config.Chain.ClConfig.DepositContractAddress)
	if err != nil {
		utils.LogFatal(err, "deposit contract address error", 0)
	}
	depositContractFirstBlock = utils.Config.Indexer.DepositContractFirstBlock

	rpcClient, err := gzondRPC.Dial(utils.Config.ELNodeEndpoint)
	if err != nil {
		utils.LogFatal(err, "new exporter gzond client error", 0)
	}
	elRPCClient = rpcClient
	client := qrlclient.NewClient(rpcClient)
	elClient = client

	lastFetchedBlock := uint64(0)

	for {
		t0 := time.Now()

		var lastDepositBlock uint64
		err = db.WriterDb.Get(&lastDepositBlock, "select coalesce(max(block_number),0) from execution_deposits")
		if err != nil {
			logger.WithError(err).Errorf("error retrieving highest block_number of execution-deposits from db")
			time.Sleep(time.Second * 5)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		header, err := elClient.HeaderByNumber(ctx, nil)
		if err != nil {
			logger.WithError(err).Errorf("error getting header from execution-client")
			cancel()
			time.Sleep(time.Second * 5)
			continue
		}
		cancel()

		blockHeight := header.Number.Uint64()

		fromBlock := lastDepositBlock + 1
		toBlock := blockHeight

		// start from the first block
		if fromBlock < depositContractFirstBlock {
			fromBlock = depositContractFirstBlock
		}
		// make sure we are progressing even if there are no deposits in the last batch
		if fromBlock < lastFetchedBlock+1 {
			fromBlock = lastFetchedBlock + 1
		}
		// if we are not synced to the head yet fetch missing blocks in batches of size 1000
		if toBlock > fromBlock+elMaxFetch {
			toBlock = fromBlock + elMaxFetch
		}
		if toBlock > blockHeight {
			toBlock = blockHeight
		}
		// if we are synced to the head look at the last 100 blocks
		if toBlock < fromBlock+elLookBack {
			if toBlock > elLookBack {
				fromBlock = toBlock - elLookBack
			} else {
				fromBlock = 0
			}
		}

		depositsToSave, err := fetchExecutionDeposits(fromBlock, toBlock)
		if err != nil {
			if gzondRequestEntityTooLargeRE.MatchString(err.Error()) {
				toBlock = fromBlock + 100
				if toBlock > blockHeight {
					toBlock = blockHeight
				}
				logger.Infof("limiting block-range to %v-%v when fetching execution-deposits due to too much results", fromBlock, toBlock)
				depositsToSave, err = fetchExecutionDeposits(fromBlock, toBlock)
			}
			if err != nil {
				logger.WithError(err).WithField("fromBlock", fromBlock).WithField("toBlock", toBlock).Errorf("error fetching execution-deposits")
				time.Sleep(time.Second * 5)
				continue
			}
		}

		err = saveExecutionDeposits(depositsToSave)
		if err != nil {
			logger.WithError(err).Errorf("error saving execution-deposits")
			time.Sleep(time.Second * 5)
			continue
		}

		if len(depositsToSave) > 0 {
			err = aggregateDeposits()
			if err != nil {
				logger.WithError(err).Errorf("error saving execution-deposits-leaderboard")
				time.Sleep(time.Second * 5)
				continue
			}
		}

		// make sure we are progressing even if there are no deposits in the last batch
		lastFetchedBlock = toBlock

		if len(depositsToSave) > 0 {
			logger.WithFields(logrus.Fields{
				"duration":      time.Since(t0),
				"blockHeight":   blockHeight,
				"fromBlock":     fromBlock,
				"toBlock":       toBlock,
				"depositsSaved": len(depositsToSave),
			}).Info("exported execution-deposits")
		}

		// progress faster if we are not synced to head yet
		if blockHeight != toBlock {
			time.Sleep(time.Second * 5)
			continue
		}

		time.Sleep(time.Minute)
	}
}

func fetchExecutionDeposits(fromBlock, toBlock uint64) (depositsToSave []*types.ExecutionDeposit, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	topic := common.BytesToHash(elDepositEventSignature[:])
	qry := qrl.FilterQuery{
		Addresses: []common.Address{
			qrlDepositContractAddress,
		},
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Topics:    [][]common.Hash{{topic}},
	}

	depositLogs, err := elClient.FilterLogs(ctx, qry)
	if err != nil {
		return depositsToSave, fmt.Errorf("error getting logs from execution-client: %w", err)
	}

	blocksToFetch := []uint64{}
	txsToFetch := []string{}

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return nil, err
	}

	for _, depositLog := range depositLogs {
		if depositLog.Topics[0] != elDepositEventSignature {
			continue
		}
		pubkey, withdrawalCredentials, amount, signature, merkletreeIndex, err := deposit.UnpackDepositLogData(depositLog.Data)
		if err != nil {
			return depositsToSave, fmt.Errorf("error unpacking execution-deposit-log: %x: %w", depositLog.Data, err)
		}
		err = deposit.VerifyDepositSignature(&qrymspb.Deposit_Data{
			PublicKey:             pubkey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                bytesutil.FromBytes8(amount),
			Signature:             signature,
		}, domain)
		validSignature := err == nil
		blocksToFetch = append(blocksToFetch, depositLog.BlockNumber)
		txsToFetch = append(txsToFetch, depositLog.TxHash.Hex())
		depositsToSave = append(depositsToSave, &types.ExecutionDeposit{
			TxHash:                depositLog.TxHash.Bytes(),
			TxIndex:               uint64(depositLog.TxIndex),
			BlockNumber:           depositLog.BlockNumber,
			PublicKey:             pubkey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                bytesutil.FromBytes8(amount),
			Signature:             signature,
			MerkletreeIndex:       merkletreeIndex,
			Removed:               depositLog.Removed,
			ValidSignature:        validSignature,
		})
	}

	headers, txs, err := executionBatchRequestHeadersAndTxs(blocksToFetch, txsToFetch)
	if err != nil {
		return depositsToSave, fmt.Errorf("error getting execution-blocks: %w\nblocks to fetch: %v\n tx to fetch: %v", err, blocksToFetch, txsToFetch)
	}

	for _, d := range depositsToSave {
		// get corresponding block (for the tx-time)
		b, exists := headers[d.BlockNumber]
		if !exists {
			return depositsToSave, fmt.Errorf("error getting block for execution-deposit: block does not exist in fetched map")
		}
		d.BlockTs = int64(b.Time)

		// get corresponding tx (for input and from-address)
		tx, exists := txs[fmt.Sprintf("0x%x", d.TxHash)]
		if !exists {
			return depositsToSave, fmt.Errorf("error getting tx for execution-deposit: tx does not exist in fetched map")
		}
		d.TxInput = tx.Data()
		chainID := tx.ChainId()
		if chainID == nil {
			return depositsToSave, fmt.Errorf("error getting tx-chainId for execution-deposit")
		}
		signer := gzondTypes.NewShanghaiSigner(chainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			return depositsToSave, fmt.Errorf("error getting sender for execution-deposit (txHash: %x, chainID: %v): %w", d.TxHash, chainID, err)
		}
		d.FromAddress = sender.Bytes()
	}

	return depositsToSave, nil
}

func saveExecutionDeposits(depositsToSave []*types.ExecutionDeposit) error {
	tx, err := db.WriterDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertDepositStmt, err := tx.Prepare(`
		INSERT INTO execution_deposits (
			tx_hash,
			tx_input,
			tx_index,
			block_number,
			block_ts,
			from_address,
			from_address_text,
			publickey,
			withdrawal_credentials,
			amount,
			signature,
			merkletree_index,
			removed,
			valid_signature
		)
		VALUES ($1, $2, $3, $4, TO_TIMESTAMP($5), $6, ENCODE($7, 'hex'), $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (merkletree_index) DO UPDATE SET
			tx_input               = EXCLUDED.tx_input,
			tx_index               = EXCLUDED.tx_index,
			block_number           = EXCLUDED.block_number,
			block_ts               = EXCLUDED.block_ts,
			from_address           = EXCLUDED.from_address,
			from_address_text      = EXCLUDED.from_address_text,
			publickey              = EXCLUDED.publickey,
			withdrawal_credentials = EXCLUDED.withdrawal_credentials,
			amount                 = EXCLUDED.amount,
			signature              = EXCLUDED.signature,
			merkletree_index       = EXCLUDED.merkletree_index,
			removed                = EXCLUDED.removed,
			valid_signature        = EXCLUDED.valid_signature`)
	if err != nil {
		return err
	}
	defer insertDepositStmt.Close()

	for _, d := range depositsToSave {
		_, err := insertDepositStmt.Exec(d.TxHash, d.TxInput, d.TxIndex, d.BlockNumber, d.BlockTs, d.FromAddress, d.FromAddress, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, d.MerkletreeIndex, d.Removed, d.ValidSignature)
		if err != nil {
			return fmt.Errorf("error saving execution-deposit to db: %v: %w", fmt.Sprintf("%x", d.TxHash), err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx for execution-deposits: %w", err)
	}

	return nil
}

// executionBatchRequestHeadersAndTxs requests the block range specified in the arguments.
// Instead of requesting each block in one call, it batches all requests into a single rpc call.
// This code is shamelessly stolen and adapted from https://github.com/prysmaticlabs/prysm/blob/2eac24c/beacon-chain/powchain/service.go#L473
func executionBatchRequestHeadersAndTxs(blocksToFetch []uint64, txsToFetch []string) (map[uint64]*gzondTypes.Header, map[string]*gzondTypes.Transaction, error) {
	elems := make([]gzondRPC.BatchElem, 0, len(blocksToFetch)+len(txsToFetch))
	headers := make(map[uint64]*gzondTypes.Header, len(blocksToFetch))
	txs := make(map[string]*gzondTypes.Transaction, len(txsToFetch))
	errors := make([]error, 0, len(blocksToFetch)+len(txsToFetch))

	for _, b := range blocksToFetch {
		header := &gzondTypes.Header{}
		err := error(nil)
		elems = append(elems, gzondRPC.BatchElem{
			Method: "qrl_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(b))), false},
			Result: header,
			Error:  err,
		})
		headers[b] = header
		errors = append(errors, err)
	}

	for _, txHashHex := range txsToFetch {
		tx := &gzondTypes.Transaction{}
		err := error(nil)
		elems = append(elems, gzondRPC.BatchElem{
			Method: "qrl_getTransactionByHash",
			Args:   []interface{}{txHashHex},
			Result: tx,
			Error:  err,
		})
		txs[txHashHex] = tx
		errors = append(errors, err)
	}

	lenElems := len(elems)

	if lenElems == 0 {
		return headers, txs, nil
	}

	for i := 0; (i * 100) < lenElems; i++ {
		start := (i * 100)
		end := start + 100

		if end > lenElems {
			end = lenElems
		}

		ioErr := elRPCClient.BatchCall(elems[start:end])
		if ioErr != nil {
			return nil, nil, ioErr
		}
	}

	for _, e := range errors {
		if e != nil {
			return nil, nil, e
		}
	}

	return headers, txs, nil
}

func aggregateDeposits() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("exporter_aggregate_execution_deposits").Observe(time.Since(start).Seconds())
	}()
	_, err := db.WriterDb.Exec(`
		INSERT INTO execution_deposits_aggregated (from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count)
		SELECT
			execution.from_address,
			SUM(execution.amount) as amount,
			SUM(execution.validcount) AS validcount,
			SUM(execution.invalidcount) AS invalidcount,
			COUNT(CASE WHEN v.status = 'slashed' THEN 1 END) AS slashedcount,
			COUNT(v.pubkey) AS totalcount,
			COUNT(CASE WHEN v.status = 'active_online' OR v.status = 'active_offline' THEN 1 END) as activecount,
			COUNT(CASE WHEN v.status = 'deposited' THEN 1 END) AS pendingcount,
			COUNT(CASE WHEN v.status = 'exited' THEN 1 END) AS voluntary_exit_count
		FROM (
			SELECT 
				from_address,
				publickey,
				SUM(amount) AS amount,
				COUNT(CASE WHEN valid_signature = 't' THEN 1 END) AS validcount,
				COUNT(CASE WHEN valid_signature = 'f' THEN 1 END) AS invalidcount
			FROM execution_deposits
			GROUP BY from_address, publickey
		) execution
		LEFT JOIN (SELECT pubkey, status FROM validators) v ON v.pubkey = execution.publickey
		GROUP BY execution.from_address
		ON CONFLICT (from_address) DO UPDATE SET
			amount               = excluded.amount,
			validcount           = excluded.validcount,
			invalidcount         = excluded.invalidcount,
			slashedcount         = excluded.slashedcount,
			totalcount           = excluded.totalcount,
			activecount          = excluded.activecount,
			pendingcount         = excluded.pendingcount,
			voluntary_exit_count = excluded.voluntary_exit_count`)
	if err != nil && err != sql.ErrNoRows {
		return nil
	}
	return err
}
