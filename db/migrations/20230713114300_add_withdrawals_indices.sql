-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - add column for blocks_withdrawals';
ALTER TABLE blocks_withdrawals ADD COLUMN IF NOT EXISTS address_text TEXT NOT NULL DEFAULT '';

SELECT 'populate new blocks_withdrawals column'; 
UPDATE blocks_withdrawals SET address_text = ENCODE(address, 'hex') WHERE address_text = '';

SELECT 'create new blocks_withdrawals indices'; 
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals (address_text);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_block_slot ON blocks_withdrawals (block_slot);
CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_search ON blocks_withdrawals (validatorindex, block_slot, address_text);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - drop blocks_withdrawals indices'; 
DROP INDEX IF EXISTS idx_blocks_withdrawals_address_text;
DROP INDEX IF EXISTS idx_blocks_withdrawals_block_slot;
DROP INDEX IF EXISTS idx_blocks_withdrawals_search;

SELECT 'drop column for blocks_withdrawals';
ALTER TABLE blocks_withdrawals DROP COLUMN IF EXISTS address_text;
-- +goose StatementEnd
