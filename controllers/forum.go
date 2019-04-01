package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/db-homework/models"
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
