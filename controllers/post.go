package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Apakhov/db-homework/models"
	"github.com/valyala/fasthttp"
)

func CreatePost(ctx *fasthttp.RequestCtx) {
	posts := make([]models.Post, 0, 0)
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		//fmt.Println("unmarshal err:", err)
	}
	// if len(posts) == 0 {
	// 	resp, _ := json.Marshal(posts)

	// 	ctx.SetStatusCode(fasthttp.StatusCreated)
	// 	ctx.SetContentType("application/json")
	// 	ctx.SetBody(resp)
	// 	return
	// }
	//fmt.Println("posts on query", posts)
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}
	var threadMiss bool
	var ps []models.Post
	var conf bool
	if id == 0 {
		conf, threadMiss, ps = models.CreatePost(&slug, nil, posts)
	} else {
		conf, threadMiss, ps = models.CreatePost(nil, &id, posts)
	}
	if threadMiss {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		if id == 0 {
			fmt.Fprint(ctx, `{"message": "Can't find thread by slug: `, slug, `"}`)
		} else {
			fmt.Fprint(ctx, `{"message": "Can't find thread by id: `, id, `"}`)
		}
		return
	}

	if conf {
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Parent post was created in another thread"}`)

		return
	}

	if ps != nil {
		resp, _ := json.Marshal(ps)

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
	resp, _ := json.Marshal(`{teapot:"teapot"}`)

	////fmt.Printf("hello, %s!\n%v\nerr: %s\nresp: %s\n", ctx.UserValue("nickname"), user, err, string(resp))
	ctx.SetStatusCode(fasthttp.StatusTeapot)
	ctx.SetContentType("plain/text")
	ctx.SetBody(resp)
}

func GetPosts(ctx *fasthttp.RequestCtx) {
	args := ctx.URI().QueryArgs()
	var limit *int
	if args.Has("limit") {
		temp, _ := strconv.Atoi(string(args.Peek("limit")))
		limit = &temp
	}
	var since *int
	if args.Has("since") {
		temp, _ := strconv.Atoi(string(args.Peek("since")))
		since = &temp
	}
	var desc bool
	if args.Has("desc") {
		desc, _ = strconv.ParseBool(string(args.Peek("desc")))
	}
	var sort *string
	if args.Has("sort") {
		temp := string(args.Peek("sort"))
		sort = &temp
	}
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}
	var ps []models.Post
	if sort == nil || (*sort)[0] == 'f' {
		if id == 0 {
			ps = models.GetPostsFlat(&slug, nil, limit, since, desc)
		} else {
			ps = models.GetPostsFlat(nil, &id, limit, since, desc)
		}
	} else if (*sort)[0] == 't' {
		if id == 0 {
			ps = models.GetPostsTree(&slug, nil, limit, since, desc)
		} else {
			ps = models.GetPostsTree(nil, &id, limit, since, desc)
		}
	} else if (*sort)[0] == 'p' {
		if id == 0 {
			ps = models.GetPostsParentTree(&slug, nil, limit, since, desc)
		} else {
			ps = models.GetPostsParentTree(nil, &id, limit, since, desc)
		}
	}
	if sort != nil {
		//fmt.Println("sort:asdasdsda: ", *sort, (*sort)[0] == 't', (*sort)[0])
	}

	if ps != nil {
		//fmt.Println("on final contr", len(ps))
		resp, _ := json.Marshal(ps)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		if id == 0 {
			fmt.Fprint(ctx, `{"message": "Can't find thread by slug: `, slug, `"}`)
		} else {
			fmt.Fprint(ctx, `{"message": "Can't find thread by id: `, id, `"}`)
		}
		return
	}
}

func GetPostInfo(ctx *fasthttp.RequestCtx) {
	args := ctx.URI().QueryArgs()
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))
	needAuthor := false
	needForum := false
	needThread := false
	if args.Has("related") {
		temp := string(args.Peek("related"))
		related := strings.Split(temp, ",")
		//fmt.Println(temp, related)
		for _, r := range related {
			if !needAuthor && r[0] == 'u' {
				needAuthor = true
				continue
			}
			if !needForum && r[0] == 'f' {
				needForum = true
				continue
			}
			if !needThread && r[0] == 't' {
				needThread = true
				continue
			}
		}

	}

	pi := models.GetPostInfo(id, needAuthor, needForum, needThread)
	if pi == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find post by id: `, id, `"}`)
	} else {
		resp, _ := json.Marshal(pi)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	}

}

type Message struct {
	Msg string `json:"message"`
}

func UpdatePost(ctx *fasthttp.RequestCtx) {
	message := Message{}
	json.Unmarshal(ctx.PostBody(), &message)
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))
	var p *models.Post
	if message.Msg != "" {
		p = models.UpdatePost(id, &message.Msg)
	} else {
		pi := models.GetPostInfo(id, false, false, false)
		if pi != nil {
			p = pi.P
		}
	}
	if p == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find post by id: `, id, `"}`)
	} else {
		resp, _ := json.Marshal(p)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	}
}
