-- Day 10: every claim has an owner; idempotency keys are unique per owner.
ALTER TABLE claims ADD COLUMN owner_id text NOT NULL DEFAULT 'legacy';
ALTER TABLE claims ALTER COLUMN owner_id DROP DEFAULT;
ALTER TABLE claims ADD CONSTRAINT claims_owner_len CHECK (char_length(owner_id) BETWEEN 1 AND 64);
ALTER TABLE claims DROP CONSTRAINT claims_idempotency_key_uq;
ALTER TABLE claims ADD CONSTRAINT claims_owner_idempotency_key_uq UNIQUE (owner_id, idempotency_key);
CREATE INDEX claims_owner_created_idx ON claims (owner_id, created_at DESC, id DESC);
