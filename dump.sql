-- USERS Table Definition ----------------------------------------------

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username character varying(50) NOT NULL UNIQUE,
    password character varying(100) NOT NULL
);

-- Indices --

CREATE UNIQUE INDEX user_pkey ON users(id int4_ops);
CREATE UNIQUE INDEX user_username_key ON users(username text_ops);

-- Table Definition ----------------------------------------------

CREATE TABLE persons (
    id SERIAL PRIMARY KEY UNIQUE
    firstname character varying(25) NOT NULL,
    lastname character varying(30) NOT NULL,
);

-- Indices --

CREATE UNIQUE INDEX persons_pkey ON persons(id int4_ops);
CREATE UNIQUE INDEX persons_id_key ON persons(id int4_ops);


-- ADDRESS Table Definition ----------------------------------------------

CREATE TABLE addresses (
    id integer DEFAULT nextval('addresses_id_seq1'::regclass) PRIMARY KEY UNIQUE,
    city character varying(25) NOT NULL,
    state character varying(25) NOT NULL,
    person_id integer REFERENCES persons(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Indices -------------------------------------------------------

CREATE UNIQUE INDEX addresses_pkey ON addresses(id int4_ops);
CREATE UNIQUE INDEX addresses_id_key ON addresses(id int4_ops);
