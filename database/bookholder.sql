CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE accounts (
    id integer NOT NULL,
    name character varying NOT NULL,
    kind character varying NOT NULL
);

CREATE TABLE transactions (
    id serial NOT NULL PRIMARY KEY,
    amount double precision NOT NULL,
    debit boolean NOT NULL,
    offset_account integer NOT NULL,
    account integer NOT NULL,
    date timestamp without time zone NOT NULL,
    description character varying
);

CREATE TABLE users (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name character varying  UNIQUE NOT NULL,
    password character varying NOT NULL
);

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY transactions
    ADD CONSTRAINT "Account" FOREIGN KEY (account) REFERENCES accounts(id) NOT VALID;

ALTER TABLE ONLY transactions
    ADD CONSTRAINT "Offset" FOREIGN KEY (offset_account) REFERENCES accounts(id) NOT VALID;