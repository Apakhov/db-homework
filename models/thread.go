package models

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx"
)

const createThreadTpl = `
INSERT INTO threads (forum, title, created, message, author, slug) VALUES
($1, $2, $3, $4, $5, $6)
RETURNING *;`

func CreateThread(tdescr *ThreadDescr) (nameMiss bool, slugMiss bool, th Thread, ok bool) {
	tx, err := conn.Begin()
	fmt.Println("ne beginknulos:", err)
	defer tx.Rollback()

	//checking existense
	row := tx.QueryRow("SELECT nickname FROM users WHERE nickname = $1;", tdescr.Author)
	err = row.Scan(&tdescr.Author)
	if err == pgx.ErrNoRows {
		nameMiss = true
		return
	}
	fmt.Println("possible err:", err)
	row = tx.QueryRow("SELECT slug FROM forums WHERE slug = $1;", tdescr.Forum)
	err = row.Scan(&tdescr.Forum)
	if err == pgx.ErrNoRows {
		slugMiss = true
		return
	}
	fmt.Println("possible err:", err)

	row = tx.QueryRow(createThreadTpl, tdescr.Forum, tdescr.Title, tdescr.Created, tdescr.Message, tdescr.Author, tdescr.Slug)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	fmt.Println("thread creation err: ", err)
	if err == nil {
		tx.Commit()
		ok = true
		return
	}
	tx.Commit()
	tx, _ = conn.Begin()
	row = tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE slug = $1;", tdescr.Slug)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	fmt.Println("lol here")
	if err != nil {
		fmt.Println("thred conf err: ", err)
	}
	return

}

const getThreadsByForumSlugTpl = `
SELECT * FROM threads
WHERE forum = '`

func GetThreadsByForumSlug(slug *string, limit *int, since *time.Time, desc bool) (ths []Thread, forumConf, ok bool) {

	tx, _ := conn.Begin()
	defer tx.Rollback()

	rows, _ := conn.Query(getForumTpl, slug)
	if !rows.Next() {
		rows.Close()
		forumConf = true
		return
	}
	rows.Close()

	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(getThreadsByForumSlugTpl)
	queryBuffer.WriteString(*slug)
	if since.IsZero() {
		queryBuffer.WriteString(`' `)
	} else if desc {
		queryBuffer.WriteString(`' AND created <= $1 `)
	} else {
		queryBuffer.WriteString(`' AND created >= $1 `)

	}
	if desc {
		queryBuffer.WriteString(` ORDER BY created DESC `)
	} else {
		queryBuffer.WriteString(` ORDER BY created ASC `)
	}
	if *limit != -1 {
		queryBuffer.WriteString(` LIMIT `)
		queryBuffer.WriteString(strconv.Itoa(*limit))
	}
	queryBuffer.WriteString(";")

	var err error
	if since.IsZero() {
		rows, err = tx.Query(queryBuffer.String())
	} else {
		rows, err = tx.Query(queryBuffer.String(), since)

	}
	defer rows.Close()
	if err != nil {
		fmt.Println("thread find query err: ", err)
		return
	}
	ths = make([]Thread, 0, 0)
	ok = true
	for rows.Next() {
		th := Thread{}
		err = rows.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
		fmt.Println("thread find slug err: ", err)
		ths = append(ths, th)
	}

	return
}

const voteIDTpl = `
INSERT INTO users_threads (nickname, thread_id, rate) VALUES
($1, $2, $3)
ON CONFLICT (nickname, thread_id)
DO
 UPDATE
   SET rate = EXCLUDED.rate;`

func VoteID(id int, vote *Vote) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(voteIDTpl, vote.Nickname, id, vote.Voice)
	if err != nil {
		fmt.Println("voting id insert err: ", err)
		return
	}

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE id = $1;", id)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("voting id get thread err: ", err)
	}

	tx.Commit()

	return
}

const voteSlugTpl = `
INSERT INTO users_threads (nickname, thread_id, rate) VALUES
($1, (SELECT id FROM threads WHERE slug = $2), $3)
ON CONFLICT (nickname, thread_id)
DO
 UPDATE
   SET rate = EXCLUDED.rate;`

func VoteSlug(slug *string, vote *Vote) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(voteSlugTpl, vote.Nickname, slug, vote.Voice)
	if err != nil {
		fmt.Println("voting slug insert err: ", err)
		return
	}

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE slug = $1;", slug)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("voting slug get thread err: ", err)
	}
	tx.Commit()
	return
}

const UpdateThreadSlugTpl = `
UPDATE threads
SET message = CASE WHEN LENGTH($1) > 0 THEN $1 ELSE message END,
    title = CASE WHEN LENGTH($2) > 0 THEN $2 ELSE title END
WHERE slug = $3;`

func UpdateThreadSlug(slug *string, vote *ThreadUPD) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(UpdateThreadSlugTpl, vote.Message, vote.Title, slug)
	if err != nil {
		fmt.Println("upd thread slug insert err: ", err)
		return
	}

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE slug = $1;", slug)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("upd thread slug get thread err: ", err)
		return nil
	}
	tx.Commit()

	return
}

const UpdateThreadIDTpl = `
UPDATE threads
SET message = CASE WHEN LENGTH($1) > 0 THEN $1 ELSE message END,
    title = CASE WHEN LENGTH($2) > 0 THEN $2 ELSE title END
WHERE id = $3;`

func UpdateThreadID(id *int, vote *ThreadUPD) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(UpdateThreadIDTpl, vote.Message, vote.Title, id)
	if err != nil {
		fmt.Println("upd thread id insert err: ", err)
		return
	}

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE id = $1;", id)
	err = row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("upd thread id get thread err: ", err)
		return nil
	}
	tx.Commit()

	return
}

func GetThreadSlug(slug *string) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE slug = $1;", slug)
	err := row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("get thread slug get thread err: ", err)
		return nil
	}

	return
}

func GetThreadID(id *int) (th *Thread) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	th = &Thread{}
	row := tx.QueryRow("SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE id = $1;", id)
	err := row.Scan(&th.Author, &th.Created, &th.Forum, &th.ID, &th.Message, &th.Slug, &th.Title, &th.Votes)
	if err != nil {
		fmt.Println("get thread id get thread err: ", err)
		return nil
	}

	return
}
