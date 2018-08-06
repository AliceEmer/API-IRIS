-- USERS Table Definition ----------------------------------------------

CREATE TABLE if not exists users (
    id SERIAL PRIMARY KEY UNIQUE,
    username character varying(50) NOT NULL UNIQUE,
    password character varying(100) NOT NULL,
    email text NOT NULL UNIQUE,
    role integer DEFAULT 0,
    uuid text,
    email_validated boolean DEFAULT false,
    twofa_activated boolean DEFAULT false
);

-- Indices --

CREATE UNIQUE INDEX user_id_key ON users(id);
CREATE UNIQUE INDEX user_username_key ON users(username);
CREATE UNIQUE INDEX users_email_key ON users(email text_ops);


-- Persons Table Definition ----------------------------------------------

CREATE TABLE if not exists persons (
    id SERIAL PRIMARY KEY UNIQUE
    firstname character varying(25) NOT NULL,
    lastname character varying(30) NOT NULL,
);

-- Indices --

CREATE UNIQUE INDEX persons_id_key ON persons(id);


-- ADDRESS Table Definition ----------------------------------------------

CREATE TABLE if not exists addresses (
    id SERIAL PRIMARY KEY UNIQUE
    city character varying(25) NOT NULL,
    state character varying(25) NOT NULL,
    person_id integer REFERENCES persons(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Indices --

CREATE UNIQUE INDEX addresses_id_key ON addresses(id);
