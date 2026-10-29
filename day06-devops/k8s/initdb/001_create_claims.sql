CREATE TABLE claims (
    id            BIGSERIAL PRIMARY KEY,
    patient_ref   VARCHAR(32)  NOT NULL,
    hospital      VARCHAR(100) NOT NULL,
    amount_satang BIGINT       NOT NULL CHECK (amount_satang > 0),
    status        VARCHAR(16)  NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    created_at    TIMESTAMPTZ  NOT NULL,
    updated_at    TIMESTAMPTZ  NOT NULL
);

CREATE INDEX idx_claims_status ON claims (status);
