DROP DATABASE IF EXISTS bondsocial CASCADE;
CREATE DATABASE IF NOT EXISTS bondsocial;
SET DATABASE = bondsocial;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    email VARCHAR NOT NULL UNIQUE,
    followers_count INT NOT NULL DEFAULT 0 CHECK (followers_count >= 0),
    followees_count INT NOT NULL DEFAULT 0 CHECK (followees_count >= 0),
);

CREATE TABLE IF NOT EXISTS follows (
    -- follower_id is the user who starts the follow
    follower_id INT NOT NULL, 
    -- followee is the user target
    followee_id INT NOT NULL,
    PRIMARY KEY(follower_id, followee_id)
);

INSERT INTO users (id, email, username) VALUES 
    (1, "john@example.com", "john"),
    (2, "jane@example.com", "jane");

