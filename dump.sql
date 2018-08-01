-- USERS Table Definition ----------------------------------------------

CREATE TABLE if not exists users (
    id SERIAL PRIMARY KEY,
    username character varying(50) NOT NULL,
    password character varying(100) NOT NULL
);

-- Indices --

CREATE UNIQUE INDEX user_id_key ON users(id);
CREATE UNIQUE INDEX user_username_key ON users(username);

-- Table Definition ----------------------------------------------

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
