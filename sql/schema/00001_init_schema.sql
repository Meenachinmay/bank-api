-- +goose Up
CREATE TABLE "accounts" (
                            "id"         bigserial PRIMARY KEY,
                            "owner"      varchar     NOT NULL,
                            "email" VARCHAR(255) NOT NULL UNIQUE,
                            "extra_interest" FLOAT DEFAULT 0,
                            "extra_interest_start_date" DATE,
                            "extra_interest_duration" INT NOT NULL DEFAULT 9,
                            "interest" FLOAT DEFAULT 4.5 NOT NULL,
                            "balance"    bigint      NOT NULL,
                            "currency"   varchar     NOT NULL,
                            "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "entries" (
                           "id" bigserial PRIMARY KEY,
                           "account_id" bigint NOT NULL,
                           "amount" bigint NOT NULL,
                           "created_at" timestamptz NOT NULL DEFAULT (now()));

CREATE TABLE "transfers" (
                             "id"              bigserial PRIMARY KEY,
                             "from_account_id" bigint  NOT NULL,
                             "to_account_id"   bigint  NOT NULL,
                             "amount"          bigint  NOT NULL,
                             "created_at"      timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE referral_codes (
    id bigserial PRIMARY KEY,
    referral_code VARCHAR(255) NOT NULL UNIQUE,
    referrer_account_id bigint NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP
);

CREATE TABLE referral_history (
    id bigserial PRIMARY KEY,
    referrer_account_id bigint NOT NULL,
    referred_account_id bigint NOT NULL,
    referral_code_id bigint NOT NULL,
    referral_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "referral_codes" ADD FOREIGN KEY ("referrer_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "referral_history" ADD FOREIGN KEY ("referrer_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "referral_history" ADD FOREIGN KEY ("referred_account_id") REFERENCES "accounts" ("id");
ALTER TABLE "referral_history" ADD FOREIGN KEY ("referral_code_id") REFERENCES "referral_codes" ("id");

CREATE INDEX ON "accounts" ("owner");

CREATE INDEX ON "entries" ("account_id");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON COLUMN "entries". "amount" IS 'can be negative or positive';

COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';

-- +goose Down
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS referral_history;
DROP TABLE IF EXISTS referral_codes;
DROP TABLE IF EXISTS accounts;
