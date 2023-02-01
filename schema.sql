DROP DATABASE IF EXISTS bondsocial CASCADE;
CREATE DATABASE IF NOT EXISTS bondsocial;
SET DATABASE = bondsocial;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    email VARCHAR NOT NULL UNIQUE
);