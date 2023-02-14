DROP DATABASE IF EXISTS bondsocial CASCADE;
CREATE DATABASE IF NOT EXISTS bondsocial;
SET DATABASE = bondsocial;


DROP TABLE timeline;
DROP TABLE posts;
DROP TABLE users;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    email VARCHAR NOT NULL UNIQUE,
    password VARCHAR(255) UNIQUE NOT NULL,
    karma INT NOT NULL DEFAULT 0,
    interacted_posts JSONB
);

CREATE TABLE IF NOT EXISTS posts (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users, 

    upvotes INTEGER NOT NULL DEFAULT 0,
    downvotes INTEGER NOT NULL DEFAULT 0,
    views BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP with TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    title TEXT,
    body TEXT,
    link TEXT,

    album JSONB,
    poll JSONB
);

CREATE INDEX IF NOT EXISTS sorted_posts ON posts(created_at DESC);

CREATE TABLE IF NOT EXISTS timeline (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users,
    post_id INT NOT NULL REFERENCES posts    
);

CREATE UNIQUE INDEX IF NOT EXISTS timeline_unique ON timeline(user_id, post_id);

CREATE TABLE IF NOT EXISTS post_votes (
    user_id INT NOT NULL REFERENCES users,
    post_id INT NOT NULL REFERENCES posts,
    vote_type VARCHAR NOT NULL,
    PRIMARY KEY(user_id, post_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS post_votes_unique ON post_votes(user_id, post_id);