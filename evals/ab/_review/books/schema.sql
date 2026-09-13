-- Schema of the ledger database. cmd/migrate applies it; the Go code in this
-- package reads and writes these two tables and nothing else.

CREATE TABLE accounts (
    id            BIGSERIAL PRIMARY KEY,
    owner         TEXT        NOT NULL,
    currency      CHAR(3)     NOT NULL,
    balance_cents BIGINT      NOT NULL DEFAULT 0,
    frozen        BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE TABLE entries (
    id           BIGSERIAL PRIMARY KEY,
    account_id   BIGINT      NOT NULL REFERENCES accounts (id),
    debit_cents  BIGINT      NOT NULL DEFAULT 0,
    credit_cents BIGINT      NOT NULL DEFAULT 0,
    memo         TEXT,                                -- NULL when the transfer carried none
    posted_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX entries_by_account ON entries (account_id, id);
