package models

import (
	"fmt"
	"log"

	"github.com/jackc/pgx"
)

var conn *pgx.Conn

func init() {
	config := pgx.ConnConfig{
		Host:     "docker",
		User:     "docker",
		Password: "docker",
		Database: "docker",
	}
	var err error
	conn, err = pgx.Connect(config)
	if err != nil {
		log.Fatalf("cant connest to db: %v", err)
	}
	log.Println("base up")
	Clear()
}

func GetInfo() (info *DBInfo) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	info = &DBInfo{}
	row := tx.QueryRow(`SELECT forums, posts, threads, users FROM info LIMIT 1;`)
	err := row.Scan(&info.Forums, &info.Posts, &info.Threads, &info.Users)
	if err != nil {
		fmt.Println("get info err:", err)
		return nil
	}
	return
}

func Clear() {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(clearTpl)
	if err != nil {
		fmt.Println("clear err:", err)
		return
	}
	tx.Commit()
}

const clearTpl = `DROP  TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
  id bigserial NOT NULL PRIMARY KEY,
  about varchar(1024) ,
  email CITEXT NOT NULL UNIQUE,
  fullname varchar(128) NOT NULL,
  nickname CITEXT NOT NULL UNIQUE
);

DROP TABLE IF EXISTS forums CASCADE ;
CREATE TABLE forums (
  id bigserial NOT NULL PRIMARY KEY,
  posts INTEGER NOT NULL DEFAULT 0,
  threads INTEGER NOT NULL DEFAULT 0,
  slug CITEXT NOT NULL UNIQUE,
  title varchar(128) NOT NULL,
  author CITEXT NOT NULL REFERENCES users(nickname) ON DELETE CASCADE
);


DROP TABLE IF EXISTS threads CASCADE ;
CREATE TABLE threads (
  author CITEXT NOT NULL REFERENCES users(nickname) ON DELETE CASCADE,
  created TIMESTAMP NOT NULL,
  forum CITEXT NOT NULL REFERENCES forums(slug) ON DELETE CASCADE,
  id bigserial NOT NULL PRIMARY KEY,
  message varchar(8192),
  slug CITEXT UNIQUE,
  title varchar(128) NOT NULL,
  votes INTEGER NOT NULL DEFAULT 0
);

DROP TABLE IF EXISTS users_threads CASCADE;
CREATE TABLE users_threads(
  nickname CITEXT NOT NULL REFERENCES users(nickname) ON DELETE CASCADE ,
  thread_id INTEGER NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
  rate INTEGER NOT NULL,
  CONSTRAINT nick_thread_pkey PRIMARY KEY (nickname, thread_id)
);


DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
  author CITEXT NOT NULL REFERENCES users(nickname) ON DELETE CASCADE,
  created TIMESTAMP NOT NULL DEFAULT NOW(),
  forum CITEXT NOT NULL REFERENCES forums(slug) ON DELETE CASCADE,
  id bigserial NOT NULL PRIMARY KEY,
  isEdited BOOLEAN NOT NULL DEFAULT false,
  message varchar(8192),
  parent INTEGER REFERENCES posts(id) ON DELETE CASCADE,
  thread bigserial NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
  path integer[] NOT NULL
);




DROP TABLE IF EXISTS users_forums CASCADE;
CREATE TABLE users_forums(
  nickname CITEXT NOT NULL REFERENCES users(nickname) ON DELETE CASCADE ,
  forum CITEXT NOT NULL REFERENCES forums(slug) ON DELETE CASCADE,
  CONSTRAINT nick_forum_pkey PRIMARY KEY (nickname, forum)
);


DROP TABLE IF EXISTS info CASCADE;
CREATE TABLE info(
  forums BIGINT DEFAULT 0,
  posts BIGINT DEFAULT 0,
  threads BIGINT DEFAULT 0,
  users BIGINT DEFAULT 0
);
INSERT INTO info(forums, posts, threads, users) VALUES (0,0,0,0);


/*triggers*/
BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS parent_post ON posts;
DROP FUNCTION IF EXISTS parent_post();
CREATE OR REPLACE FUNCTION parent_post() RETURNS trigger AS $$
BEGIN
    IF new.parent IS NULL THEN
      return NULL;
    END IF;
    if NOT EXISTS(SELECT * FROM posts WHERE id = new.parent AND thread = new.thread) THEN
      RAISE 'bred';
    end if;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER parent_post AFTER INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE parent_post();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS thread_inc ON threads;
DROP FUNCTION IF EXISTS thread_inc();
CREATE OR REPLACE FUNCTION thread_inc() RETURNS trigger AS $$
BEGIN
    UPDATE forums
      SET threads = threads +1
    WHERE slug = NEW.forum;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER thread_inc AFTER INSERT ON threads
    FOR EACH ROW EXECUTE PROCEDURE thread_inc();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS post_inc ON posts;
DROP FUNCTION IF EXISTS post_inc();
CREATE OR REPLACE FUNCTION post_inc() RETURNS trigger AS $$
BEGIN
    UPDATE forums
      SET posts = posts +1
    WHERE slug = NEW.forum;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER post_inc AFTER INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE post_inc();
COMMIT;



BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS votes_create ON users_threads;
DROP FUNCTION IF EXISTS votes_create();
CREATE OR REPLACE FUNCTION votes_create() RETURNS trigger AS $$
BEGIN
    UPDATE threads
      SET votes = votes + NEW.rate
    WHERE id = NEW.thread_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER votes_create AFTER INSERT ON users_threads
    FOR EACH ROW EXECUTE PROCEDURE votes_create();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS votes_update ON users_threads;
DROP FUNCTION IF EXISTS votes_update();
CREATE OR REPLACE FUNCTION votes_update() RETURNS trigger AS $$
BEGIN
    UPDATE threads
      SET votes = votes - OLD.rate + NEW.rate
    WHERE id = NEW.thread_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER votes_update AFTER UPDATE ON users_threads
    FOR EACH ROW EXECUTE PROCEDURE votes_update();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS user_forum_thread ON threads;
DROP FUNCTION IF EXISTS user_forum_thread();
CREATE OR REPLACE FUNCTION user_forum_thread() RETURNS trigger AS $$
BEGIN
   INSERT INTO users_forums(nickname, forum) VALUES
   (NEW.author, NEW.forum)
   ON CONFLICT DO NOTHING;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_forum_thread AFTER INSERT ON threads
    FOR EACH ROW EXECUTE PROCEDURE user_forum_thread();
COMMIT;

BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS user_forum_posts ON posts;
DROP FUNCTION IF EXISTS user_forum_posts();
CREATE OR REPLACE FUNCTION user_forum_posts() RETURNS trigger AS $$
BEGIN
      INSERT INTO users_forums(nickname, forum) VALUES
      (NEW.author, NEW.forum)
   ON CONFLICT DO NOTHING;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_forum_posts AFTER INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE user_forum_posts();
COMMIT;



BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS info_forums_inc ON info;
DROP FUNCTION IF EXISTS info_forums_inc();
CREATE OR REPLACE FUNCTION info_forums_inc() RETURNS trigger AS $$
BEGIN
    UPDATE info
      SET forums = forums + 1;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER info_forums_inc AFTER INSERT ON forums
    FOR EACH ROW EXECUTE PROCEDURE info_forums_inc();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS info_posts_inc ON info;
DROP FUNCTION IF EXISTS info_posts_inc();
CREATE OR REPLACE FUNCTION info_posts_inc() RETURNS trigger AS $$
BEGIN
    UPDATE info
      SET posts = posts + 1;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER info_posts_inc AFTER INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE info_posts_inc();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS info_threads_inc ON info;
DROP FUNCTION IF EXISTS info_threads_inc();
CREATE OR REPLACE FUNCTION info_threads_inc() RETURNS trigger AS $$
BEGIN
    UPDATE info
      SET threads = threads + 1;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER info_threads_inc AFTER INSERT ON threads
    FOR EACH ROW EXECUTE PROCEDURE info_threads_inc();
COMMIT;


BEGIN TRANSACTION;
DROP TRIGGER IF EXISTS info_users_inc ON info;
DROP FUNCTION IF EXISTS info_users_inc();
CREATE OR REPLACE FUNCTION info_users_inc() RETURNS trigger AS $$
BEGIN
    UPDATE info
      SET users = users + 1;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER info_users_inc AFTER INSERT ON users
    FOR EACH ROW EXECUTE PROCEDURE info_users_inc();
COMMIT;
`
