CREATE TABLE accounts (
    name "char" NOT NULL,
    kind "char" NOT NULL,
    id integer NOT NULL
);

CREATE TABLE transaction (
    id uuid NOT NULL,
    amount double precision NOT NULL,
    debit boolean NOT NULL,
    offset_account integer NOT NULL,
    account integer NOT NULL,
    date timestamp without time zone NOT NULL,
    description "char"
);

CREATE TABLE users (
    name "char" NOT NULL,
    password "char" NOT NULL,
    id integer
);

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (id);

ALTER TABLE ONLY transaction
    ADD CONSTRAINT "Account" FOREIGN KEY (account) REFERENCES accounts(id) NOT VALID;

ALTER TABLE ONLY transaction
    ADD CONSTRAINT "Offset" FOREIGN KEY (offset_account) REFERENCES accounts(id) NOT VALID;