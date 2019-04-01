package models

import (
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


