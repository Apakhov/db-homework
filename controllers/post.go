package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/db-homework/models"
	"github.com/valyala/fasthttp"
)

func CreatePost(ctx *fasthttp.RequestCtx) {
	posts := make([]models.PostDescr, 0, 0)
	err := json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		fmt.Println("unmarshal err:", err)
	}
	fmt.Println("posts on query", posts)
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}
	var threadMiss bool
	var ps []models.Post
	if id == 0 {
		threadMiss, ps = models.CreatePost(&slug, nil, posts)
	} else {
		threadMiss, ps = models.CreatePost(nil, &id, posts)
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

	if ps != nil {
		resp, _ := json.Marshal(ps)

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
	resp, _ := json.Marshal(`{teapot:"teapot"}`)

	//fmt.Printf("hello, %s!\n%v\nerr: %s\nresp: %s\n", ctx.UserValue("nickname"), user, err, string(resp))
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
		fmt.Println("sort:asdasdsda: ", *sort, (*sort)[0] == 't', (*sort)[0])
	}

	if ps != nil {
		fmt.Println("on final contr", len(ps))
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
