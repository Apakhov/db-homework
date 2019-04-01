package models

import (
	"bytes"
	"fmt"
	"strconv"
)

func CreatePost(threadSlug *string, threadID *int, pdescrs []PostDescr) (threadMiss bool, ps []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()

	id := 0
	forum := ""
	if threadSlug != nil {
		fmt.Println("finding by slug")
		row := conn.QueryRow("SELECT id, forum FROM threads WHERE slug = $1;", *threadSlug)
		if row.Scan(&id, &forum) != nil {
			return true, nil
		}
	} else {
		fmt.Println("finding by id")
		row := conn.QueryRow("SELECT id, forum FROM threads WHERE id = $1;", *threadID)
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
	queryBuffer.WriteString("INSERT INTO posts (author, forum, message, thread, parent) VALUES")
	for _, pdescr := range pdescrs {
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
			queryBuffer.WriteString(strconv.Itoa(*pdescr.Parent))
		}
		queryBuffer.WriteByte(')')
	}
	queryBuffer.WriteString(` RETURNING  author, created, forum, id, isEdited, message, parent, thread;`)

	rows, err := conn.Query(queryBuffer.String())
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
