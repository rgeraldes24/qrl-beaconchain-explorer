package types

import (
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// ChainHead is a struct to hold chain head data
type ChainHead struct {
	HeadSlot                   uint64
	HeadEpoch                  uint64
	HeadBlockRoot              []byte
	FinalizedSlot              uint64
	FinalizedEpoch             uint64
	FinalizedBlockRoot         []byte
	JustifiedSlot              uint64
	JustifiedEpoch             uint64
	JustifiedBlockRoot         []byte
	PreviousJustifiedSlot      uint64
	PreviousJustifiedEpoch     uint64
	PreviousJustifiedBlockRoot []byte
}

type FinalityCheckpoints struct {
	PreviousJustified struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"previous_justified"`
	CurrentJustified struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"current_justified"`
	Finalized struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"finalized"`
}

type Slot uint64
type Epoch uint64
type ValidatorIndex uint64
type SyncCommitteePeriod uint64
type CommitteeIndex uint64

// EpochData is a struct to hold epoch data
type EpochData struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorAssignmentes   *EpochAssignments
	Blocks                  map[uint64]map[string]*Block
	FutureBlocks            map[uint64]map[string]*Block
	EpochParticipationStats *ValidatorParticipation
	AttestationDuties       map[Slot]map[ValidatorIndex][]Slot
	SyncDuties              map[Slot]map[ValidatorIndex]bool
	Finalized               bool
}

// ValidatorParticipation is a struct to hold validator participation data
type ValidatorParticipation struct {
	Epoch                   uint64
	GlobalParticipationRate float32
	VotedQuanta             uint64
	EligibleQuanta          uint64
	Finalized               bool
}

// Validator is a struct to hold validator data
type Validator struct {
	Index                      uint64 `db:"validatorindex"`
	PublicKey                  []byte `db:"pubkey"`
	PublicKeyHex               string `db:"pubkeyhex"`
	Balance                    uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	WithdrawalCredentials      []byte `db:"withdrawalcredentials"`

	BalanceActivation sql.NullInt64 `db:"balanceactivation"`
	Status            string        `db:"status"`

	LastAttestationSlot sql.NullInt64 `db:"lastattestationslot"`
}

// ValidatorQueue is a struct to hold validator queue data
type ValidatorQueue struct {
	Activating uint64
	Exiting    uint64
}

type SyncAggregate struct {
	SyncCommitteeValidators    []uint64
	SyncCommitteeBits          []byte
	SyncCommitteeSignatures    [][]byte
	SyncAggregateParticipation float64
}

// Block is a struct to hold block data
type Block struct {
	Status            uint64
	Proposer          uint64
	BlockRoot         []byte
	Slot              uint64
	ParentRoot        []byte
	StateRoot         []byte
	Signature         []byte
	RandaoReveal      []byte
	Graffiti          []byte
	ExecutionData     *ExecutionData
	BodyRoot          []byte
	ProposerSlashings []*ProposerSlashing
	AttesterSlashings []*AttesterSlashing
	Attestations      []*Attestation
	Deposits          []*Deposit
	VoluntaryExits    []*VoluntaryExit
	SyncAggregate     *SyncAggregate
	ExecutionPayload  *ExecutionPayload
	AttestationDuties map[ValidatorIndex][]Slot
	SyncDuties        map[ValidatorIndex]bool
	Finalized         bool
	EpochAssignments  *EpochAssignments
	Validators        []*Validator
}

type Transaction struct {
	Raw []byte
	// Note: below values may be nil/0 if Raw fails to decode into a valid transaction
	TxHash       []byte
	AccountNonce uint64
	// big endian
	Price     []byte
	GasLimit  uint64
	Sender    []byte
	Recipient []byte
	// big endian
	Amount  []byte
	Payload []byte

	MaxPriorityFeePerGas uint64
	MaxFeePerGas         uint64
}

type ExecutionPayload struct {
	ParentHash    []byte
	FeeRecipient  []byte
	StateRoot     []byte
	ReceiptsRoot  []byte
	LogsBloom     []byte
	Random        []byte
	BlockNumber   uint64
	GasLimit      uint64
	GasUsed       uint64
	Timestamp     uint64
	ExtraData     []byte
	BaseFeePerGas uint64
	BlockHash     []byte
	Transactions  []*Transaction
	Withdrawals   []*Withdrawals
}

type Withdrawals struct {
	Slot           uint64 `json:"slot,omitempty"`
	BlockRoot      []byte `json:"blockroot,omitempty"`
	Index          uint64 `json:"index"`
	ValidatorIndex uint64 `json:"validatorindex"`
	Address        []byte `json:"address"`
	Amount         uint64 `json:"amount"`
}

type WithdrawalsByEpoch struct {
	Epoch          uint64
	ValidatorIndex uint64
	Amount         uint64
}

type WithdrawalsNotification struct {
	Slot           uint64 `json:"slot,omitempty"`
	Index          uint64 `json:"index"`
	ValidatorIndex uint64 `json:"validatorindex"`
	Address        []byte `json:"address"`
	Amount         uint64 `json:"amount"`
	Pubkey         []byte `json:"pubkey"`
}

// ExecutionData is a struct to hold the Execution data
type ExecutionData struct {
	DepositRoot  []byte
	DepositCount uint64
	BlockHash    []byte
}

// ProposerSlashing is a struct to hold proposer slashing data
type ProposerSlashing struct {
	ProposerIndex uint64
	Header1       *Block
	Header2       *Block
}

// AttesterSlashing is a struct to hold attester slashing
type AttesterSlashing struct {
	Attestation1 *IndexedAttestation
	Attestation2 *IndexedAttestation
}

// IndexedAttestation is a struct to hold indexed attestation data
type IndexedAttestation struct {
	Data             *AttestationData
	AttestingIndices []uint64
	Signatures       [][]byte
}

// Attestation is a struct to hold attestation header data
type Attestation struct {
	AggregationBits []byte
	Attesters       []uint64
	Data            *AttestationData
	Signatures      [][]byte
}

// AttestationData to hold attestation detail data
type AttestationData struct {
	Slot            uint64
	CommitteeIndex  uint64
	BeaconBlockRoot []byte
	Source          *Checkpoint
	Target          *Checkpoint
}

// Checkpoint is a struct to hold checkpoint data
type Checkpoint struct {
	Epoch uint64
	Root  []byte
}

// Deposit is a struct to hold deposit data
type Deposit struct {
	Proof                 [][]byte
	PublicKey             []byte
	WithdrawalCredentials []byte
	Amount                uint64
	Signature             []byte
}

// VoluntaryExit is a struct to hold voluntary exit data
type VoluntaryExit struct {
	Epoch          uint64
	ValidatorIndex uint64
	Signature      []byte
}

// MinimalBlock is a struct to hold minimal block data
type MinimalBlock struct {
	Epoch      uint64 `db:"epoch"`
	Slot       uint64 `db:"slot"`
	BlockRoot  []byte `db:"blockroot"`
	ParentRoot []byte `db:"parentroot"`
	Canonical  bool   `db:"-"`
}

// CanonBlock is a struct to hold canon block data
type CanonBlock struct {
	BlockRoot []byte `db:"blockroot"`
	Slot      uint64 `db:"slot"`
	Canonical bool   `db:"-"`
}

// EpochAssignments is a struct to hold epoch assignment data
type EpochAssignments struct {
	ProposerAssignments map[uint64]uint64
	AttestorAssignments map[string]uint64
	SyncAssignments     []uint64
}

// ExecutionDeposit is a struct to hold execution-deposit data
type ExecutionDeposit struct {
	TxHash                []byte `db:"tx_hash"`
	TxInput               []byte `db:"tx_input"`
	TxIndex               uint64 `db:"tx_index"`
	BlockNumber           uint64 `db:"block_number"`
	BlockTs               int64  `db:"block_ts"`
	FromAddress           []byte `db:"from_address"`
	FromName              string
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawal_credentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
	MerkletreeIndex       []byte `db:"merkletree_index"`
	Removed               bool   `db:"removed"`
	ValidSignature        bool   `db:"valid_signature"`
}

// ConsensusDeposit is a struct to hold consensus-deposit data
type ConsensusDeposit struct {
	BlockSlot             uint64 `db:"block_slot"`
	BlockIndex            uint64 `db:"block_index"`
	BlockRoot             []byte `db:"block_root"`
	Proof                 []byte `db:"proof"`
	Publickey             []byte `db:"publickey"`
	Withdrawalcredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

type TagMetadata struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	PublicLink  string `json:"public_url"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type TagMetadataSlice []TagMetadata

func (s *TagMetadataSlice) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		err := json.Unmarshal(v, s)
		if err != nil {
			return err
		}
		// if no tags were found we will get back an empty TagMetadata, we don't want that
		if len(*s) == 1 && (*s)[0].Name == "" {
			*s = nil
		}
		return nil
	case string:
		return json.Unmarshal([]byte(v), s)
	}
	return errors.New("type assertion failed")
}

type ValidatorStatsTableDbRow struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	Day            int64  `db:"day"`

	StartBalance          int64 `db:"start_balance"`
	EndBalance            int64 `db:"end_balance"`
	MinBalance            int64 `db:"min_balance"`
	MaxBalance            int64 `db:"max_balance"`
	StartEffectiveBalance int64 `db:"start_effective_balance"`
	EndEffectiveBalance   int64 `db:"end_effective_balance"`
	MinEffectiveBalance   int64 `db:"min_effective_balance"`
	MaxEffectiveBalance   int64 `db:"max_effective_balance"`

	MissedAttestations      int64 `db:"missed_attestations"`
	MissedAttestationsTotal int64 `db:"missed_attestations_total"`
	OrphanedAttestations    int64 `db:"orphaned_attestations"`

	ParticipatedSync      int64 `db:"participated_sync"`
	ParticipatedSyncTotal int64 `db:"participated_sync_total"`
	MissedSync            int64 `db:"missed_sync"`
	MissedSyncTotal       int64 `db:"missed_sync_total"`
	OrphanedSync          int64 `db:"orphaned_sync"`
	OrphanedSyncTotal     int64 `db:"orphaned_sync_total"`

	ProposedBlocks int64 `db:"proposed_blocks"`
	MissedBlocks   int64 `db:"missed_blocks"`
	OrphanedBlocks int64 `db:"orphaned_blocks"`

	AttesterSlashings int64 `db:"attester_slashings"`
	ProposerSlashing  int64 `db:"proposer_slashings"`

	Deposits            int64 `db:"deposits"`
	DepositsTotal       int64 `db:"deposits_total"`
	DepositsAmount      int64 `db:"deposits_amount"`
	DepositsAmountTotal int64 `db:"deposits_amount_total"`

	Withdrawals            int64 `db:"withdrawals"`
	WithdrawalsTotal       int64 `db:"withdrawals_total"`
	WithdrawalsAmount      int64 `db:"withdrawals_amount"`
	WithdrawalsAmountTotal int64 `db:"withdrawals_amount_total"`

	ClRewardsShor      int64 `db:"cl_rewards_shor"`
	ClRewardsShorTotal int64 `db:"cl_rewards_shor_total"`

	ClPerformance1d   int64 `db:"-"`
	ClPerformance7d   int64 `db:"-"`
	ClPerformance31d  int64 `db:"-"`
	ClPerformance365d int64 `db:"-"`

	ElRewardsPlanck      decimal.Decimal `db:"el_rewards_planck"`
	ElRewardsPlanckTotal decimal.Decimal `db:"el_rewards_planck_total"`

	ElPerformance1d   decimal.Decimal `db:"-"`
	ElPerformance7d   decimal.Decimal `db:"-"`
	ElPerformance31d  decimal.Decimal `db:"-"`
	ElPerformance365d decimal.Decimal `db:"-"`
}
