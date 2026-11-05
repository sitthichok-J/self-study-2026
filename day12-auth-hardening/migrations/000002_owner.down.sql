DROP INDEX claims_owner_created_idx;
ALTER TABLE claims DROP CONSTRAINT claims_owner_idempotency_key_uq;
ALTER TABLE claims ADD CONSTRAINT claims_idempotency_key_uq UNIQUE (idempotency_key);
ALTER TABLE claims DROP CONSTRAINT claims_owner_len;
ALTER TABLE claims DROP COLUMN owner_id;
