DROP DATABASE IF EXISTS bondsocial CASCADE;
CREATE DATABASE IF NOT EXISTS bondsocial;
SET DATABASE = bondsocial;

DROP TABLE timeline;
DROP TABLE posts;
DROP TABLE users;

-- Users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    email VARCHAR NOT NULL UNIQUE,
    karma INTEGER NOT NULL DEFAULT 0,
    password VARCHAR(255) UNIQUE NOT NULL
);

-- Posts
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

-- Timeline
CREATE TABLE IF NOT EXISTS timeline (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users ON DELETE CASCADE,
    post_id INT NOT NULL REFERENCES posts ON DELETE CASCADE    
);

CREATE UNIQUE INDEX IF NOT EXISTS timeline_unique ON timeline(user_id, post_id);

-- Post Votes
CREATE TABLE IF NOT EXISTS post_votes (
    user_id INT NOT NULL REFERENCES users ON DELETE CASCADE,
    post_id INT NOT NULL REFERENCES posts ON DELETE CASCADE,
    vote_type VARCHAR NOT NULL,
    PRIMARY KEY(user_id, post_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS vote_unique_idx ON post_votes(user_id, post_id);

-- Comments
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE post_comments (
  id SERIAL PRIMARY KEY,
  post_id INT NOT NULL REFERENCES posts ON DELETE CASCADE,
  parent_id INTEGER REFERENCES post_comments(id) ON DELETE CASCADE,
  path ltree,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX comments_path_idx ON post_comments (path);

CREATE OR REPLACE FUNCTION comments_path_trigger()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.parent_id IS NOT NULL THEN
    NEW.path = (SELECT path FROM post_comments WHERE id = NEW.parent_id) || NEW.id::TEXT;
  ELSE
    NEW.path = NEW.id::TEXT;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_comments_path
BEFORE INSERT OR UPDATE ON post_comments
FOR EACH ROW
EXECUTE FUNCTION comments_path_trigger();

