-- +goose NO TRANSACTION

-- +goose Up

-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_search;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS idx_blocks_withdrawals_address_text;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE blocks_withdrawals DROP COLUMN IF EXISTS address_text;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address ON blocks_withdrawals (address);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
ALTER TABLE blocks_withdrawals ADD COLUMN IF NOT EXISTS address_text TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blocks_withdrawals_address_text ON blocks_withdrawals USING gin (address_text gin_trgm_ops);
-- +goose StatementEnd
