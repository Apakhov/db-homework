package models

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx"
)

type Buffers struct {
	block chan struct{}
	bufs  map[int]*struct {
		buf  *bytes.Buffer
		busy bool
	}
	cur int
}

func (bs *Buffers) get() (int, *bytes.Buffer) {
	bs.block <- struct{}{}
	//fmt.Println("search")
	for i, b := range bs.bufs {
		//fmt.Println("next")
		if !b.busy {
			b.buf.Reset()
			b.busy = true
			<-bs.block
			return i, b.buf
		}
	}
	fmt.Println("NEW", bs.cur)
	bs.cur++
	fmt.Println(bs.cur)
	i := bs.cur
	n := &struct {
		buf  *bytes.Buffer
		busy bool
	}{bytes.NewBuffer(make([]byte, 100)), true}
	n.buf.Reset()
	bs.bufs[i] = n
	<-bs.block
	return i, n.buf
}

func (bs *Buffers) back(i int) {
	bs.block <- struct{}{}
	bs.bufs[i].busy = false
	<-bs.block
}

var bs *Buffers

func init() {
	bs = &Buffers{
		block: make(chan struct{}, 1),
		bufs: make(map[int]*struct {
			buf  *bytes.Buffer
			busy bool
		}),
	}
}

func CreatePost(threadSlug *string, threadID *int, pdescrs []Post) (conf bool, threadMiss bool, ps []Post) {
	tx, _ := conn.Begin()
	defer tx.Rollback()
	id := 0
	forum := ""
	if threadSlug != nil {
		//fmt.Println("finding by slug")
		row := tx.QueryRow("SELECT id, forum FROM threads WHERE slug = $1;", *threadSlug)
		if row.Scan(&id, &forum) != nil {
			return false, true, nil
		}
	} else {
		//fmt.Println("finding by id")
		row := tx.QueryRow("SELECT id, forum FROM threads WHERE id = $1;", *threadID)
		err := row.Scan(&id, &forum)
		if err != nil {
			//fmt.Println("id thread not found:", err)
			return false, true, nil
		}
	}
	if len(pdescrs) == 0 {
		//fmt.Println("zero posts got")
		return false, false, make([]Post, 0, 0)
	}

	ps = pdescrs
	for i, pdescr := range pdescrs {
		var parentID string
		if pdescr.Parent != nil {
			parentID = strconv.Itoa(*pdescr.Parent)
		}
		//fmt.Println("QUERY:::", queryBuffer.String())
		var row *pgx.Row
		if !pdescr.Created.IsZero() {
			pdescr.Created = time.Now()
		}
		// PostgreSQL считает с точностью до MS
		pdescr.Created = pdescr.Created.Round(time.Millisecond)
		if pdescr.Parent == nil || *pdescr.Parent <= 0 {
			row = tx.QueryRow(`INSERT INTO posts (author, forum, message, thread, parent, path, created) VALUES
			( $1, $2, $3, $4, NULL, '{}', $5) RETURNING  id;`,
				pdescr.Author, forum, pdescr.Message, strconv.Itoa(id), pdescr.Created)
		} else {
			row = tx.QueryRow(`INSERT INTO posts (author, forum, message, thread, parent, path, created) VALUES 
			( $1, $2, $3, $4, $5, CAST ((SELECT path FROM posts WHERE id = $6) AS BIGINT[]) || $6 , $7) RETURNING  id;`,
				pdescr.Author, forum, pdescr.Message, strconv.Itoa(id), parentID, pdescr.Parent, pdescr.Created)
		}

		p := &ps[i]
		err := row.Scan(&p.ID)
		if err != nil {
			//fmt.Println("post create err: ", err)
			//fmt.Println(queryBuffer.String())
			//fmt.Println("ERROR::::", err.Error())
			f := err.(pgx.PgError)
			//fmt.Printf("err: %+v\n%s\n", f, f.ConstraintName)
			if f.Code == "23503" && f.ConstraintName == "posts_author_fkey" {
				//fmt.Println("YA RETURNU!!!")
				return false, true, nil
			}
			return true, false, nil
		}
		//author, created, forum, id, isEdited, message, parent, thread
		//fmt.Println("scan err:", err)

		p.Forum = forum
		p.Thread = id
		//fmt.Println("post crated ok", p)
	}
	//fmt.Println("kek query:", queryBuffer.String())

	//fmt.Println("posts created: ", ps, queryBuffer.String())
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
		//fmt.Println(`getPostsFlat find thread err: `, err)
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
	defer rows.Close()
	if err != nil {
		//fmt.Println(`getPostsFlat find posts err: `, err)
		//fmt.Println(`query: `, queryBuffer.String())
		return
	}

	//fmt.Println(`!!!!!!!!!!!!!!!query: `, queryBuffer.String())
	ps = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		if p.Parent == nil {
			pa := 0
			p.Parent = &pa
		}
		if err != nil {
			//fmt.Println(`getPostsFlat scan posts err: `, err)
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
		//fmt.Println(`GetPostsTree find thread err: `, err)
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
	defer rows.Close()
	if err != nil {
		//fmt.Println(`GetPostsTree find posts err: `, err)
		//fmt.Println(`query: `, queryBuffer.String())
		return
	}
	//fmt.Println(`GetPostsTree query: `, queryBuffer.String())
	ps = make([]Post, 0, 0)
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		if p.Parent == nil {
			pa := 0
			p.Parent = &pa
		}
		if err != nil {
			//fmt.Println(`GetPostsTree scan posts err: `, err)
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
		//fmt.Println(`GetPostsParentTree find thread err: `, err)
		return
	}
	if since != nil {
		row := tx.QueryRow(`SELECT  parent FROM posts WHERE id = $1;`, since)
		var p *int
		err := row.Scan(&p)
		if err != nil {
			//fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAA")
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
	defer rows.Close()
	if err != nil {
		//fmt.Println(`GetPostsParentTree find main posts: `, err)
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
			//fmt.Println(`GetPostsParentTree scan main posts err: `, err)
			return nil
		}
		parentPosts = append(parentPosts, p)
	}

	ps := make([]Post, 0, 0)

	for _, parentPost := range parentPosts {
		ps = append(ps, parentPost)
		rows, err = tx.Query(fmt.Sprintf(getPostsParentTree, parentPost.ID))

		if err != nil {
			//fmt.Println("GetPostsParentTree query to childs err: ", err, "\nquery: ", fmt.Sprintf(getPostsParentTree, parentPost.ID))
		}
		//fmt.Println("GetPostsParentTree query success ")
		for rows.Next() {
			p := Post{}
			err := rows.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
			//fmt.Println(`GetPostsParentTree scan posts err: `, err, p)
			if err != nil {
				rows.Close()
				//fmt.Println(`GetPostsParentTree scan posts err: `, err)
				return nil
			}
			ps = append(ps, p)
		}
		//fmt.Println(len(ps))
		rows.Close()
	}
	//fmt.Println("on final", len(ps))
	return ps
}

const getPostInfoTPL = `
SELECT p.author, p.created, p.forum, p.id, p.isEdited, p.message, p.parent, p.thread,
       f.posts, f.slug, f.threads, f.title, f.author,
       u.about, u.email, u.fullname, u.nickname,
       t.author, t.created, t.forum, t.id, t.message, t.slug, t.title, t.votes
FROM ((posts p JOIN forums f on p.forum = f.slug AND p.id = $1) JOIN users u ON u.nickname = p.author) JOIN threads t ON p.thread = t.id`

// TODO: different queries
func GetPostInfo(id int, needAuthor, needForum, needThread bool) (pi *PostInfo) {
	tx, _ := conn.Begin()
	defer tx.Rollback()
	pi = &PostInfo{
		U: &User{},
		T: &Thread{},
		F: &Forum{},
		P: &Post{},
	}
	row := tx.QueryRow(getPostInfoTPL, id)
	err := row.Scan(
		&pi.P.Author, &pi.P.Created, &pi.P.Forum, &pi.P.ID, &pi.P.IsEdited, &pi.P.Message, &pi.P.Parent, &pi.P.Thread,
		&pi.F.Posts, &pi.F.Slug, &pi.F.Threads, &pi.F.Title, &pi.F.User,
		&pi.U.About, &pi.U.Email, &pi.U.Fullname, &pi.U.Nickname,
		&pi.T.Author, &pi.T.Created, &pi.T.Forum, &pi.T.ID, &pi.T.Message, &pi.T.Slug, &pi.T.Title, &pi.T.Votes,
	)
	if err == pgx.ErrNoRows {
		//fmt.Println("GetPostInfo post not found")
		return nil
	}
	if err != nil {
		//fmt.Println("GetPostInfo get post err", err)
		return nil
	}
	if !needAuthor {
		pi.U = nil
	}
	if !needForum {
		pi.F = nil
	}
	if !needThread {
		pi.T = nil
	}
	return
}

const UpdatePostTpl = `
UPDATE posts p
SET 
isEdited = isEdited or (CASE WHEN p.message != $1 THEN TRUE ELSE FALSE END),
message = $1
WHERE id = $2
RETURNING p.author, p.created, p.forum, p.id, p.isEdited, p.message, p.parent, p.thread`

func UpdatePost(id int, msg *string) (p *Post) {
	//fmt.Println("post query: ", p)
	tx, _ := conn.Begin()
	defer tx.Rollback()
	row := tx.QueryRow(UpdatePostTpl, msg, id)

	p = &Post{}
	err := row.Scan(&p.Author, &p.Created, &p.Forum, &p.ID, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
	if err == pgx.ErrNoRows {
		//fmt.Println("UpdatePost post not found")
		return nil
	}
	if err != nil {
		//fmt.Println("UpdatePost upd post err", err)
		return nil
	}
	tx.Commit()
	//fmt.Println("post upd: ", *p)
	return
}
