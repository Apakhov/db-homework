package models

import (
	"fmt"

	"github.com/jackc/pgx"
)

const createForumTpl = `
INSERT INTO  forums (slug, title, author) VALUES
($1, $2, (SELECT nickname FROM users WHERE nickname = $3))
RETURNING author;`

const findForumBySlug = `
SELECT posts, slug, threads, title, author FROM forums
WHERE slug = $1;`

func CreateForum(fdescr *ForumDescr) (*Forum, *string) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	row := tx.QueryRow(findForumBySlug, fdescr.Slug)
	forumConf := Forum{}
	err := row.Scan(&forumConf.Posts, &forumConf.Slug, &forumConf.Threads, &forumConf.Title, &forumConf.User)
	if err != pgx.ErrNoRows {
		return &forumConf, nil
	}

	row = tx.QueryRow(createForumTpl, fdescr.Slug, fdescr.Title, fdescr.User)

	nick := ""
	err = row.Scan(&nick)
	if err != nil {
		return nil, &fdescr.User
	}
	tx.Commit()
	fdescr.User = nick
	return nil, nil
}

const getForumTpl = `
SELECT posts, slug, threads, title, author FROM forums
WHERE slug = $1`

func GetForum(slug string) *Forum {
	rows, _ := conn.Query(getForumTpl, slug)
	defer rows.Close()
	if rows.Next() {
		f := Forum{Slug: slug}
		rows.Scan(&f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
		return &f
	}
	return nil
}

func GetUsers(slug *string, limit *int, since *string, desc *bool) (us []User, ok bool) {
	tx, _ := conn.Begin()
	row := tx.QueryRow(`SELECT id FROM forums WHERE slug = $1`, *slug)
	var threadID int
	err := row.Scan(&threadID)
	if err == pgx.ErrNoRows {
		return
	}
	if err != nil {
		fmt.Println("GetUsers get forum id err:", err)
		return
	}

	var rows *pgx.Rows
	if limit == nil {
		if since == nil {
			if desc == nil {
				rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
				 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX";`, slug)
			} else {
				if *desc {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX" DESC;`, slug)
				} else {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX";`, slug)
				}
			}
		} else {
			if desc == nil {
				rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
				 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX";`, slug)
			} else {
				if *desc {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX" DESC;`, slug)
				} else {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX";`, slug)
				}
			}
		}
	} else {
		if since == nil {
			if desc == nil {
				rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
				 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX" LIMIT $2;`, slug, limit)
			} else {
				if *desc {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX" DESC LIMIT $2;`, slug, limit)
				} else {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 ORDER BY users.nickname COLLATE "POSIX" LIMIT $2;`, slug, limit)
				}
			}
		} else {
			if desc == nil {
				rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
				 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 WHERE users.nickname > $2 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX" LIMIT $3;`, slug, since, limit)
			} else {
				if *desc {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 WHERE users.nickname < $2 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX" DESC LIMIT $3;`, slug, since, limit)
				} else {
					rows, err = tx.Query(`SELECT about, email, fullname, users.nickname FROM
					 users JOIN users_forums ON users.nickname = users_forums.nickname AND forum = $1 WHERE users.nickname > $2 COLLATE "POSIX" ORDER BY users.nickname COLLATE "POSIX" LIMIT $3;`, slug, since, limit)
				}
			}
		}
	}
	if err != nil {
		fmt.Println("GetUsers get users err:", err)
		return
	}
	defer rows.Close()
	us = make([]User, 0, 0)
	for rows.Next() {
		var u User
		err = rows.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
		if err != nil {
			fmt.Println("GetUsers scan users err:", err)
			return
		}
		us = append(us, u)
	}
	ok = true
	return
}
