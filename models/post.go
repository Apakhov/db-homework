package models

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/jackc/pgx"
)

func CreatePost(threadSlug *string, threadID *int, pdescrs []PostDescr) (threadMiss bool, ps []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	id := 0
	forum := ""
	if threadSlug != nil {
		fmt.Println("finding by slug")
		row := tx.QueryRow("SELECT id, forum FROM threads WHERE slug = $1;", *threadSlug)
		if row.Scan(&id, &forum) != nil {
			return true, nil
		}
	} else {
		fmt.Println("finding by id")
		row := tx.QueryRow("SELECT id, forum FROM threads WHERE id = $1;", *threadID)
		err := row.Scan(&id, &forum)
		if err != nil {
			fmt.Println("id thread not found:", err)
			return true, nil
		}
	}
	if len(pdescrs) == 0 {
		fmt.Println("zero posts got")
		return false, make([]Post, 0, 0)
	}
	var queryBuffer bytes.Buffer
	queryBuffer.WriteString("INSERT INTO posts (author, forum, message, thread, parent, path) VALUES")
	l := len(pdescrs) - 1
	for i, pdescr := range pdescrs {
		var parentID string
		if pdescr.Parent != nil {
			parentID = strconv.Itoa(*pdescr.Parent)
		}
		queryBuffer.WriteString(`('`)
		queryBuffer.WriteString(pdescr.Author)
		queryBuffer.WriteString(`', '`)
		queryBuffer.WriteString(forum)
		queryBuffer.WriteString(`', '`)
		queryBuffer.WriteString(pdescr.Message)
		queryBuffer.WriteString(`', `)
		queryBuffer.WriteString(strconv.Itoa(id))
		queryBuffer.WriteString(`,`)
		if pdescr.Parent == nil || *pdescr.Parent <= 0 {
			queryBuffer.WriteString(` NULL `)
		} else {
			queryBuffer.WriteString(parentID)
		}
		queryBuffer.WriteString(`,`)
		if pdescr.Parent == nil || *pdescr.Parent <= 0 {
			queryBuffer.WriteString(`'{}'`)
		} else {
			queryBuffer.WriteString(`(SELECT path FROM posts WHERE id = `)
			queryBuffer.WriteString(parentID)
			queryBuffer.WriteString(`) || `)
			queryBuffer.WriteString(parentID)
		}
		queryBuffer.WriteByte(')')
		if i < l {
			queryBuffer.WriteByte(',')
		}
	}
	queryBuffer.WriteString(` RETURNING  author, created, forum, id, isEdited, message, parent, thread;`)

	rows, err := tx.Query(queryBuffer.String())
	if err != nil {
		fmt.Println("post create err: ", err)
		fmt.Println(queryBuffer.String())
		return
	}

	ps = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		var msg string
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &msg, &p.Parent, &p.Thread)
		fmt.Println("scan err:", err)
		p.Message = msg
		fmt.Println("post crated ok", p)
		ps = append(ps, p)
		fmt.Println("post crated ok", p)
	}
	fmt.Println("posts created: ", ps, queryBuffer.String())
	tx.Commit()
	return
}

func GetPostsFlat(threadSlug *string, threadID *int, limit *int, since *int, deck bool) (ps []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	var row *pgx.Row
	if threadID == nil {
		row = tx.QueryRow(`(SELECT id FROM threads WHERE slug = $1)`, threadSlug)
	} else {
		row = tx.QueryRow(`(SELECT id FROM threads WHERE id = $1)`, threadID)
	}
	threadIDNotPtr := 0
	threadID = &threadIDNotPtr
	err := row.Scan(threadID)
	if err == pgx.ErrNoRows {
		return
	}
	if err != nil {
		fmt.Println(`getPostsFlat find thread err: `, err)
		return
	}

	var queryBuffer bytes.Buffer

	queryBuffer.WriteString(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = `)
	queryBuffer.WriteString(strconv.Itoa(*threadID))
	if since != nil {
		if deck {
			queryBuffer.WriteString(` AND id < `)
			queryBuffer.WriteString(strconv.Itoa(*since))
		} else {
			queryBuffer.WriteString(` AND id > `)
			queryBuffer.WriteString(strconv.Itoa(*since))
		}
	}
	queryBuffer.WriteString(` ORDER BY id `)
	if deck {
		queryBuffer.WriteString(`DESC `)
	}
	if limit != nil {
		queryBuffer.WriteString(`LIMIT `)
		queryBuffer.WriteString(strconv.Itoa(*limit))
	}

	rows, err := tx.Query(queryBuffer.String())
	if err != nil {
		fmt.Println(`getPostsFlat find posts err: `, err)
		fmt.Println(`query: `, queryBuffer.String())
		return
	}

	fmt.Println(`!!!!!!!!!!!!!!!query: `, queryBuffer.String())
	ps = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		if p.Parent == nil {
			pa := 0
			p.Parent = &pa
		}
		if err != nil {
			fmt.Println(`getPostsFlat scan posts err: `, err)
			return nil
		}
		ps = append(ps, p)
	}
	return
}

func GetPostsTree(threadSlug *string, threadID *int, limit *int, since *int, deck bool) (ps []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	var row *pgx.Row
	if threadID == nil {
		row = tx.QueryRow(`(SELECT id FROM threads WHERE slug = $1)`, threadSlug)
	} else {
		row = tx.QueryRow(`(SELECT id FROM threads WHERE id = $1)`, threadID)
	}
	threadIDNotPtr := 0
	threadID = &threadIDNotPtr
	err := row.Scan(threadID)
	if err == pgx.ErrNoRows {
		return
	}
	if err != nil {
		fmt.Println(`GetPostsTree find thread err: `, err)
		return
	}

	var queryBuffer bytes.Buffer

	queryBuffer.WriteString(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts
	WHERE thread = `)
	queryBuffer.WriteString(strconv.Itoa(*threadID))
	if since != nil {
		if deck {
			queryBuffer.WriteString(` AND (path || id::INTEGER) < (SELECT path || id::INTEGER from posts WHERE id = `)
			queryBuffer.WriteString(strconv.Itoa(*since))
			queryBuffer.WriteString(`)`)
		} else {
			queryBuffer.WriteString(` AND (path || id::INTEGER) > (SELECT path || id::INTEGER from posts WHERE id = `)
			queryBuffer.WriteString(strconv.Itoa(*since))
			queryBuffer.WriteString(`)`)
		}
	}
	queryBuffer.WriteString(` ORDER BY path || id::INTEGER `)
	if deck {
		queryBuffer.WriteString(`DESC `)
	}
	if limit != nil {
		queryBuffer.WriteString(`LIMIT `)
		queryBuffer.WriteString(strconv.Itoa(*limit))
	}

	rows, err := tx.Query(queryBuffer.String())
	if err != nil {
		fmt.Println(`GetPostsTree find posts err: `, err)
		fmt.Println(`query: `, queryBuffer.String())
		return
	}
	fmt.Println(`GetPostsTree query: `, queryBuffer.String())
	ps = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		if p.Parent == nil {
			pa := 0
			p.Parent = &pa
		}
		if err != nil {
			fmt.Println(`GetPostsTree scan posts err: `, err)
			return nil
		}
		ps = append(ps, p)
	}
	return
}

const getPostsParentTree = `SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts
WHERE path && '{%d}'
ORDER BY path || id::INTEGER;`

const getPostsParentTreeSince = `SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts
WHERE path && '{%d}' AND id > %d
ORDER BY path || id::INTEGER;`

const getPostsParentTreeSinceDesc = `SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts
WHERE path && '{%d}' AND id < %d
ORDER BY path || id::INTEGER;`

func GetPostsParentTree(threadSlug *string, threadID *int, limit *int, since *int, deck bool) (parentPosts []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	var row *pgx.Row
	if threadID == nil {
		row = tx.QueryRow(`SELECT id FROM threads WHERE slug = $1;`, threadSlug)
	} else {
		row = tx.QueryRow(`SELECT id FROM threads WHERE id = $1;`, threadID)
	}

	threadIDNotPtr := 0
	threadID = &threadIDNotPtr
	err := row.Scan(threadID)
	if err == pgx.ErrNoRows {
		return
	}
	if err != nil {
		fmt.Println(`GetPostsParentTree find thread err: `, err)
		return
	}
	if since != nil {
		row := tx.QueryRow(`SELECT  parent FROM posts WHERE id = $1;`, since)
		var p *int
		err := row.Scan(&p)
		if err != nil {
			fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAA")
		} else {
			if p != nil {
				*since = *p
			}
		}
	}
	var rows *pgx.Rows
	if limit != nil {
		if deck {
			if since == nil {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL ORDER BY id DESC LIMIT $2;`, threadID, limit)
			} else {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL AND id < $3 ORDER BY id DESC LIMIT $2;`, threadID, limit, since)
			}
		} else {
			if since == nil {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL ORDER BY id LIMIT $2;`, threadID, limit)
			} else {
				fmt.Println("LOL HERE")
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL AND id > $3 ORDER BY id LIMIT $2;`, threadID, limit, since)
			}
		}
	} else {
		if deck {
			if since == nil {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL ORDER BY id DESC;`, threadID)
			} else {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL AND id < $2 ORDER BY id DESC;`, threadID, since)
			}
		} else {
			if since == nil {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL ORDER BY id;`, threadID)
			} else {
				rows, err = tx.Query(`SELECT author, created, forum, id, isEdited, message, parent, thread FROM posts WHERE thread = $1 AND parent ISNULL AND id > $2 ORDER BY id;`, threadID, since)
			}
		}
	}
	if err != nil {
		fmt.Println(`GetPostsParentTree find main posts: `, err)
		return
	}

	parentPosts = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		pa := 0
		p.Parent = &pa

		if err != nil {
			rows.Close()
			fmt.Println(`GetPostsParentTree scan main posts err: `, err)
			return nil
		}
		parentPosts = append(parentPosts, p)
	}
	rows.Close()
	ps := make([]Post, 0, 0)

	for _, parentPost := range parentPosts {
		ps = append(ps, parentPost)
		rows, err = tx.Query(fmt.Sprintf(getPostsParentTree, parentPost.ID))

		if err != nil {
			fmt.Println("GetPostsParentTree query to childs err: ", err, "\nquery: ", fmt.Sprintf(getPostsParentTree, parentPost.ID))
		}
		fmt.Println("GetPostsParentTree query success ")
		for rows.Next() {
			p := Post{}
			err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
			fmt.Println(`GetPostsParentTree scan posts err: `, err, p)
			if err != nil {
				rows.Close()
				fmt.Println(`GetPostsParentTree scan posts err: `, err)
				return nil
			}
			ps = append(ps, p)
		}
		fmt.Println(len(ps))
		rows.Close()
	}
	fmt.Println("on final", len(ps))
	return ps
}
