package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Apakhov/db-homework/models"
	"github.com/valyala/fasthttp"
)

func CreateForum(ctx *fasthttp.RequestCtx) {

	forum := &models.ForumDescr{}
	json.Unmarshal(ctx.PostBody(), forum)
	forumConf, nameConf := models.CreateForum(forum)

	if forumConf != nil {
		resp, _ := json.Marshal(forumConf)

		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	} else if nameConf != nil {

		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find user by nickname: `, *nameConf, `"}`)

		return
	}
	resp, _ := json.Marshal(forum)

	//fmt.Printf("hello, %s!\n%v\nerr: %s\nresp: %s\n", ctx.UserValue("nickname"), user, err, string(resp))
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(resp)
}

func GetForum(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	forum := models.GetForum(slug)
	if forum != nil {
		resp, _ := json.Marshal(forum)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find forum by slug: `, slug, `"}`)
	}
}

func GetUsers(ctx *fasthttp.RequestCtx) {
	args := ctx.URI().QueryArgs()
	slug := ctx.UserValue("slug").(string)

	var since *string
	var limit *int
	var desc *bool

	if args.Has("limit") {
		l, _ := strconv.Atoi(string(args.Peek("limit")))
		limit = &l
	}
	if args.Has("desc") {
		d, _ := strconv.ParseBool(string(args.Peek("desc")))
		desc = &d
	}
	if args.Has("since") {
		s := string(args.Peek("since"))
		since = &s
	}
	fmt.Println("sluuuuuuuuuuug:", slug)
	us, ok := models.GetUsers(&slug, limit, since, desc)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find forum by slug: `, slug, `"}`)
	} else {
		resp, _ := json.Marshal(us)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	}
}
