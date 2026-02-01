-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

/* trigram extension for faster text-search */
/*
This TABLE IS used to store the current state (latest exported epoch) of ALL validators
It also acts AS a lookup-table to store the index-pubkey association
IN ORDER to save db space we only use the UNIQUE validator INDEX IN ALL other tables
IN the future it IS better to REPLACE this TABLE WITH an IN memory cache (redis)
 */
CREATE TABLE IF NOT EXISTS
    validators (
        validatorindex INT NOT NULL,
        pubkey bytea NOT NULL,
        pubkeyhex TEXT NOT NULL DEFAULT '',
        withdrawableepoch BIGINT NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        balance BIGINT NOT NULL,
        balanceactivation BIGINT,
        effectivebalance BIGINT NOT NULL,
        slashed bool NOT NULL,
        activationeligibilityepoch BIGINT NOT NULL,
        activationepoch BIGINT NOT NULL,
        exitepoch BIGINT NOT NULL,
        lastattestationslot BIGINT,
        status VARCHAR(20) NOT NULL DEFAULT '',
        PRIMARY KEY (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validators_pubkey ON validators (pubkey);

-- TODO(now.youtrack.cloud/issue/TZB-9)
-- CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex ON validators (pubkeyhex);
-- CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex_pattern_pos ON validators (pubkeyhex varchar_pattern_ops);


CREATE INDEX IF NOT EXISTS idx_validators_status ON validators (status);

CREATE INDEX IF NOT EXISTS idx_validators_balanceactivation ON validators (balanceactivation);

CREATE INDEX IF NOT EXISTS idx_validators_activationepoch ON validators (activationepoch);

CREATE INDEX IF NOT EXISTS validators_is_offline_vali_idx ON validators (validatorindex, lastattestationslot, pubkey);

CREATE INDEX IF NOT EXISTS idx_validators_withdrawalcredentials ON validators (withdrawalcredentials, validatorindex);

CREATE TABLE IF NOT EXISTS
    validator_pool (
        publickey bytea NOT NULL,
        pool VARCHAR(40),
        PRIMARY KEY (publickey)
    );

CREATE TABLE IF NOT EXISTS
    validator_names (
        publickey bytea NOT NULL,
        NAME VARCHAR(40),
        PRIMARY KEY (publickey)
    );

CREATE INDEX IF NOT EXISTS idx_validator_names_publickey ON validator_names (publickey);

CREATE INDEX IF NOT EXISTS idx_validator_names_name ON validator_names (NAME);

CREATE TABLE IF NOT EXISTS
    validator_set (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        withdrawableepoch BIGINT NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        effectivebalance BIGINT NOT NULL,
        slashed bool NOT NULL,
        activationeligibilityepoch BIGINT NOT NULL,
        activationepoch BIGINT NOT NULL,
        exitepoch BIGINT NOT NULL,
        PRIMARY KEY (validatorindex, epoch)
    );

CREATE TABLE IF NOT EXISTS
    validator_performance (
        validatorindex INT NOT NULL,
        balance BIGINT NOT NULL,
        cl_performance_1d BIGINT NOT NULL,
        cl_performance_7d BIGINT NOT NULL,
        cl_performance_31d BIGINT NOT NULL,
        cl_performance_365d BIGINT NOT NULL,
        cl_performance_total BIGINT NOT NULL,
        el_performance_1d BIGINT NOT NULL,
        el_performance_7d BIGINT NOT NULL,
        el_performance_31d BIGINT NOT NULL,
        el_performance_365d BIGINT NOT NULL,
        el_performance_total BIGINT NOT NULL,
        rank7d INT NOT NULL,
        PRIMARY KEY (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_performance_balance ON validator_performance (balance);

CREATE INDEX IF NOT EXISTS idx_validator_performance_rank7d ON validator_performance (rank7d);

CREATE TABLE IF NOT EXISTS
    proposal_assignments (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        proposerslot INT NOT NULL,
        status INT NOT NULL,
        /* Can be 0 = scheduled, 1 executed, 2 missed */
        PRIMARY KEY (epoch, validatorindex, proposerslot)
    );

CREATE INDEX IF NOT EXISTS idx_proposal_assignments_epoch ON proposal_assignments (epoch);

CREATE TABLE IF NOT EXISTS
    sync_committees (
        period INT NOT NULL,
        validatorindex INT NOT NULL,
        committeeindex INT NOT NULL,
        PRIMARY KEY (period, validatorindex, committeeindex)
    );

CREATE TABLE IF NOT EXISTS
    sync_committees_count_per_validator (
        period INT NOT NULL UNIQUE,
        count_so_far float8 NOT NULL,
        PRIMARY KEY (period)
    );

CREATE TABLE IF NOT EXISTS
    validator_balances_recent (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        balance BIGINT NOT NULL,
        PRIMARY KEY (epoch, validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_epoch ON validator_balances_recent (epoch);

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_validatorindex ON validator_balances_recent (validatorindex);

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_balance ON validator_balances_recent (balance);

CREATE TABLE IF NOT EXISTS
    validator_stats (
        validatorindex INT NOT NULL,
        DAY INT NOT NULL,
        start_balance BIGINT,
        end_balance BIGINT,
        min_balance BIGINT,
        max_balance BIGINT,
        start_effective_balance BIGINT,
        end_effective_balance BIGINT,
        min_effective_balance BIGINT,
        max_effective_balance BIGINT,
        missed_attestations INT,
        orphaned_attestations INT,
        participated_sync INT,
        missed_sync INT,
        orphaned_sync INT,
        proposed_blocks INT,
        missed_blocks INT,
        orphaned_blocks INT,
        attester_slashings INT,
        proposer_slashings INT,
        deposits INT,
        deposits_amount BIGINT,
        withdrawals INT,
        withdrawals_amount BIGINT,
        cl_rewards_shor BIGINT,
        cl_rewards_shor_total BIGINT,
        el_rewards_planck DECIMAL,
        el_rewards_planck_total DECIMAL,
        PRIMARY KEY (validatorindex, DAY)
    );

CREATE INDEX IF NOT EXISTS idx_validator_stats_day ON validator_stats (DAY);

CREATE TABLE IF NOT EXISTS
    validator_stats_status (
        DAY INT NOT NULL,
        status BOOLEAN NOT NULL,
        income_exported BOOLEAN NOT NULL DEFAULT FALSE,
        PRIMARY KEY (DAY)
    );

CREATE TABLE IF NOT EXISTS
    queue (
        ts TIMESTAMP WITHOUT TIME ZONE,
        entering_validators_count INT NOT NULL,
        exiting_validators_count INT NOT NULL,
        PRIMARY KEY (ts)
    );

CREATE TABLE IF NOT EXISTS
    validatorqueue_activation (
        INDEX INT NOT NULL,
        publickey bytea NOT NULL,
        PRIMARY KEY (INDEX, publickey)
    );

CREATE TABLE IF NOT EXISTS
    validatorqueue_exit (
        INDEX INT NOT NULL,
        publickey bytea NOT NULL,
        PRIMARY KEY (INDEX, publickey)
    );

CREATE TABLE IF NOT EXISTS
    epochs_notified (
        epoch INT NOT NULL PRIMARY KEY,
        sentOn TIMESTAMP NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    epochs (
        epoch INT NOT NULL,
        blockscount INT NOT NULL DEFAULT 0,
        proposerslashingscount INT NOT NULL,
        attesterslashingscount INT NOT NULL,
        attestationscount INT NOT NULL,
        depositscount INT NOT NULL,
        withdrawalcount INT NOT NULL DEFAULT 0,
        voluntaryexitscount INT NOT NULL,
        validatorscount INT NOT NULL,
        averagevalidatorbalance BIGINT NOT NULL,
        totalvalidatorbalance BIGINT NOT NULL,
        finalized bool,
        eligiblequanta BIGINT,
        globalparticipationrate FLOAT,
        votedquanta BIGINT,
        rewards_exported bool NOT NULL DEFAULT FALSE,
        PRIMARY KEY (epoch)
    );

CREATE TABLE IF NOT EXISTS
    blocks (
        epoch INT NOT NULL,
        slot INT NOT NULL,
        blockroot bytea NOT NULL,
        parentroot bytea NOT NULL,
        stateroot bytea NOT NULL,
        signature bytea NOT NULL,
        randaoreveal bytea,
        graffiti bytea,
        graffiti_text TEXT NULL,
        executiondata_depositroot bytea,
        executiondata_depositcount INT NOT NULL,
        executiondata_blockhash bytea,
        syncaggregate_bits bytea,
        syncaggregate_signatures bytea[],
        syncaggregate_participation FLOAT NOT NULL DEFAULT 0,
        proposerslashingscount INT NOT NULL,
        attesterslashingscount INT NOT NULL,
        attestationscount INT NOT NULL,
        depositscount INT NOT NULL,
        withdrawalcount INT NOT NULL DEFAULT 0,
        voluntaryexitscount INT NOT NULL,
        proposer INT NOT NULL,
        status TEXT NOT NULL,
        /* Can be 0 = scheduled, 1 proposed, 2 missed, 3 orphaned */
        -- https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
        -- https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
        exec_parent_hash bytea,
        exec_fee_recipient bytea,
        exec_state_root bytea,
        exec_receipts_root bytea,
        exec_logs_bloom bytea,
        exec_random bytea,
        exec_block_number INT,
        exec_gas_limit INT,
        exec_gas_used INT,
        exec_timestamp INT,
        exec_extra_data bytea,
        exec_base_fee_per_gas BIGINT,
        exec_block_hash bytea,
        exec_transactions_count INT NOT NULL DEFAULT 0,
        PRIMARY KEY (slot, blockroot)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_proposer ON blocks (proposer);

CREATE INDEX IF NOT EXISTS idx_blocks_epoch ON blocks (epoch);

CREATE INDEX IF NOT EXISTS idx_blocks_graffiti_text ON blocks USING gin (graffiti_text gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_blocks_blockrootstatus ON blocks (blockroot, status);

CREATE INDEX IF NOT EXISTS idx_blocks_exec_block_number ON blocks (exec_block_number);

CREATE TABLE IF NOT EXISTS
    blocks_withdrawals (
        block_slot INT NOT NULL,
        block_root bytea NOT NULL,
        withdrawalindex INT NOT NULL,
        validatorindex INT NOT NULL,
        address bytea NOT NULL,
        amount BIGINT NOT NULL,
        -- in Shor
        PRIMARY KEY (block_slot, block_root, withdrawalindex)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_recipient ON blocks_withdrawals (address);

CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_validatorindex ON blocks_withdrawals (validatorindex);

CREATE TABLE IF NOT EXISTS
    blocks_transactions (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        raw bytea NOT NULL,
        txhash bytea NOT NULL,
        nonce INT NOT NULL,
        gas_price bytea NOT NULL,
        gas_limit BIGINT NOT NULL,
        sender bytea NOT NULL,
        recipient bytea NOT NULL,
        amount bytea NOT NULL,
        payload bytea NOT NULL,
        max_priority_fee_per_gas BIGINT,
        max_fee_per_gas BIGINT,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_proposerslashings (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        proposerindex INT NOT NULL,
        header1_slot BIGINT NOT NULL,
        header1_parentroot bytea NOT NULL,
        header1_stateroot bytea NOT NULL,
        header1_bodyroot bytea NOT NULL,
        header1_signature bytea NOT NULL,
        header2_slot BIGINT NOT NULL,
        header2_parentroot bytea NOT NULL,
        header2_stateroot bytea NOT NULL,
        header2_bodyroot bytea NOT NULL,
        header2_signature bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_attesterslashings (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        attestation1_indices INTEGER[] NOT NULL,
        attestation1_signatures bytea[] NOT NULL,
        attestation1_slot BIGINT NOT NULL,
        attestation1_index INT NOT NULL,
        attestation1_beaconblockroot bytea NOT NULL,
        attestation1_source_epoch INT NOT NULL,
        attestation1_source_root bytea NOT NULL,
        attestation1_target_epoch INT NOT NULL,
        attestation1_target_root bytea NOT NULL,
        attestation2_indices INTEGER[] NOT NULL,
        attestation2_signatures bytea[] NOT NULL,
        attestation2_slot BIGINT NOT NULL,
        attestation2_index INT NOT NULL,
        attestation2_beaconblockroot bytea NOT NULL,
        attestation2_source_epoch INT NOT NULL,
        attestation2_source_root bytea NOT NULL,
        attestation2_target_epoch INT NOT NULL,
        attestation2_target_root bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_attestations (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        aggregationbits bytea NOT NULL,
        validators INT[] NOT NULL,
        signatures bytea[] NOT NULL,
        slot INT NOT NULL,
        committeeindex INT NOT NULL,
        beaconblockroot bytea NOT NULL,
        source_epoch INT NOT NULL,
        source_root bytea NOT NULL,
        target_epoch INT NOT NULL,
        target_root bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_beaconblockroot ON blocks_attestations (beaconblockroot);

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_source_root ON blocks_attestations (source_root);

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_target_root ON blocks_attestations (target_root);

CREATE TABLE IF NOT EXISTS
    blocks_deposits (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        proof bytea[],
        publickey bytea NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        amount BIGINT NOT NULL,
        signature bytea NOT NULL,
        valid_signature bool NOT NULL DEFAULT TRUE,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_deposits_publickey ON blocks_deposits (publickey);

CREATE TABLE IF NOT EXISTS
    blocks_voluntaryexits (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        signature bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    network_liveness (
        ts TIMESTAMP WITHOUT TIME ZONE,
        headepoch INT NOT NULL,
        finalizedepoch INT NOT NULL,
        justifiedepoch INT NOT NULL,
        previousjustifiedepoch INT NOT NULL,
        PRIMARY KEY (ts)
    );
    
CREATE TABLE IF NOT EXISTS
    execution_deposits (
        tx_hash bytea NOT NULL,
        tx_input bytea NOT NULL,
        tx_index INT NOT NULL,
        block_number INT NOT NULL,
        block_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        from_address bytea NOT NULL,
        publickey bytea NOT NULL,
        withdrawal_credentials bytea NOT NULL,
        amount BIGINT NOT NULL,
        signature bytea NOT NULL,
        merkletree_index bytea NOT NULL,
        removed bool NOT NULL,
        valid_signature bool NOT NULL,
        PRIMARY KEY (tx_hash, merkletree_index)
    );

CREATE INDEX IF NOT EXISTS idx_execution_deposits ON execution_deposits (publickey);

CREATE INDEX IF NOT EXISTS idx_execution_deposits_from_address ON execution_deposits (from_address);

CREATE TABLE IF NOT EXISTS
    execution_deposits_aggregated (
        from_address bytea NOT NULL,
        amount BIGINT NOT NULL,
        validcount INT NOT NULL,
        invalidcount INT NOT NULL,
        slashedcount INT NOT NULL,
        totalcount INT NOT NULL,
        activecount INT NOT NULL,
        pendingcount INT NOT NULL,
        voluntary_exit_count INT NOT NULL,
        PRIMARY KEY (from_address)
    );

CREATE TABLE IF NOT EXISTS
    validator_tags (
        publickey bytea NOT NULL,
        tag CHARACTER VARYING(100) NOT NULL,
        PRIMARY KEY (publickey, tag)
    );

CREATE TABLE IF NOT EXISTS
    chart_images (
        NAME VARCHAR(100) NOT NULL PRIMARY KEY,
        image bytea NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    finality_checkpoints (
        head_epoch INT NOT NULL,
        head_root bytea NOT NULL,
        current_justified_epoch INT NOT NULL,
        current_justified_root bytea NOT NULL,
        previous_justified_epoch INT NOT NULL,
        previous_justified_root bytea NOT NULL,
        finalized_epoch INT NOT NULL,
        finalized_root bytea NOT NULL,
        PRIMARY KEY (head_epoch, head_root)
    );

CREATE TABLE IF NOT EXISTS
    historical_pool_performance (
        DAY INT NOT NULL,
        pool VARCHAR(40) NOT NULL,
        validators INT NOT NULL,
        effective_balances_sum_planck NUMERIC NOT NULL,
        start_balances_sum_planck NUMERIC NOT NULL,
        end_balances_sum_planck NUMERIC NOT NULL,
        deposits_sum_planck NUMERIC NOT NULL,
        tx_fees_sum_planck NUMERIC NOT NULL,
        consensus_rewards_sum_planck NUMERIC NOT NULL,
        total_rewards_planck NUMERIC NOT NULL,
        apr FLOAT NOT NULL,
        PRIMARY KEY (DAY, pool)
    );

CREATE INDEX IF NOT EXISTS idx_historical_pool_performance_pool ON historical_pool_performance (pool, DAY DESC);

CREATE TABLE IF NOT EXISTS
    tags (
        id VARCHAR NOT NULL,
        metadata jsonb NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS
    blocks_tags (
        slot int4 NOT NULL,
        blockroot bytea NOT NULL,
        tag_id VARCHAR NOT NULL,
        PRIMARY KEY (slot, blockroot, tag_id),
        FOREIGN KEY (slot, blockroot) REFERENCES blocks (slot, blockroot),
        FOREIGN KEY (tag_id) REFERENCES tags (id)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_tags_slot ON blocks_tags (slot);

CREATE INDEX IF NOT EXISTS idx_blocks_tags_tag_id ON blocks_tags (tag_id);

CREATE TABLE IF NOT EXISTS
    validator_queue_deposits (
        validatorindex int4 NOT NULL,
        block_slot int4 NULL,
        block_index int4 NULL,
        CONSTRAINT validator_queue_deposits_fk FOREIGN KEY (block_slot, block_index) REFERENCES blocks_deposits (block_slot, block_index),
        CONSTRAINT validator_queue_deposits_fk_validators FOREIGN KEY (validatorindex) REFERENCES validators (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_queue_deposits_block_slot ON validator_queue_deposits USING btree (block_slot);

CREATE UNIQUE INDEX IF NOT EXISTS idx_validator_queue_deposits_validatorindex ON validator_queue_deposits USING btree (validatorindex);

CREATE TABLE IF NOT EXISTS
    service_status (
        NAME TEXT NOT NULL,
        executable_name TEXT NOT NULL,
        VERSION TEXT NOT NULL,
        pid INT NOT NULL,
        status TEXT NOT NULL,
        metadata jsonb,
        last_update TIMESTAMP NOT NULL,
        PRIMARY KEY (NAME, executable_name, VERSION, pid)
    );

CREATE TABLE IF NOT EXISTS
    chart_series (
        "time" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        indicator CHARACTER VARYING(50) NOT NULL,
        VALUE NUMERIC NOT NULL,
        PRIMARY KEY ("time", indicator)
    );

CREATE TABLE IF NOT EXISTS
    chart_series_status (
        DAY INT NOT NULL,
        status BOOLEAN NOT NULL,
        PRIMARY KEY (DAY)
    );

CREATE TABLE IF NOT EXISTS
    global_notifications (
        target VARCHAR(20) NOT NULL PRIMARY KEY,
        CONTENT TEXT NOT NULL,
        enabled bool NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    explorer_configurations (
        category VARCHAR(40),
        -- the category is used to group settings together
        KEY VARCHAR(40),
        -- identifies the config
        VALUE TEXT,
        -- holds the value of the config
        data_type VARCHAR(40),
        -- defines to which data type the value is mapped
        PRIMARY KEY (category, KEY)
    );

CREATE INDEX IF NOT EXISTS idx_explorer_configurations ON explorer_configurations (category, KEY);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'NOT SUPPORTED';
-- +goose StatementEnd
