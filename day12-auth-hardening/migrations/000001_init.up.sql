CREATE TABLE claims (
    id              uuid PRIMARY KEY,
    patient_ref     text        NOT NULL CHECK (patient_ref ~ '^FAKE-[A-Z0-9]{3,12}$'),
    amount          bigint      NOT NULL CHECK (amount > 0),
    status          text        NOT NULL CHECK (status IN ('PENDING','APPROVED','REJECTED')),
    decision_reason text        NOT NULL DEFAULT '',
    idempotency_key text        NOT NULL,
    request_hash    text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT claims_idempotency_key_uq UNIQUE (idempotency_key)
);

CREATE INDEX claims_status_created_idx ON claims (status, created_at DESC, id DESC);
CREATE INDEX claims_created_idx ON claims (created_at DESC, id DESC);

-- Transactional outbox: written in the same transaction as the claim.
CREATE TABLE outbox (
    id         bigserial PRIMARY KEY,
    topic      text        NOT NULL,
    key        text        NOT NULL,
    payload    jsonb       NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    sent_at    timestamptz
);

CREATE INDEX outbox_unsent_idx ON outbox (id) WHERE sent_at IS NULL;
