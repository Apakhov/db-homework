package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Apakhov/db-homework/models"
	"github.com/valyala/fasthttp"
)

func CreateTread(ctx *fasthttp.RequestCtx) {
	thread := &models.ThreadDescr{}
	json.Unmarshal(ctx.PostBody(), thread)
	thread.Forum = ctx.UserValue("slug").(string)
	nameMiss, slugMiss, th, ok := models.CreateThread(thread)
	//fmt.Printf("got from CreateThread\n name: %s\n slug: %s\n thread: %+v\n ok: %v", nameMiss, slugMiss, th, ok)
	if ok {
		resp, _ := json.Marshal(th)

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	} else if nameMiss {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find user by nickname: `, thread.Author, `"}`)

		return
	} else if slugMiss {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find forum by slug: `, thread.Forum, `"}`)

		return
	} else {
		resp, _ := json.Marshal(th)

		ctx.SetStatusCode(fasthttp.StatusConflict)
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

func GetThreadsByForumSlug(ctx *fasthttp.RequestCtx) {
	args := ctx.URI().QueryArgs()
	slug := ctx.UserValue("slug").(string)

	since := time.Time{}
	limit := -1
	desc := false

	if args.Has("limit") {
		limit, _ = strconv.Atoi(string(args.Peek("limit")))
	}
	if args.Has("desc") {
		desc, _ = strconv.ParseBool(string(args.Peek("desc")))
	}
	if args.Has("since") {
		layout := "2006-01-02T15:04:05.000Z"
		var err error
		since, err = time.Parse(layout, string(args.Peek("since")))
		if err != nil {
			//fmt.Println("time parse err:", err)
		}
		//fmt.Println("time: ", since)
	}
	//fmt.Printf("!!!!!!!!!!!! %v, %v, %v\n", since, limit, desc)
	ths, forumConf, ok := models.GetThreadsByForumSlug(&slug, &limit, &since, desc)
	//fmt.Printf("got from GetThreadsBySlug\n threads: %+v\n ok: %v\n am: %v\n slug: %v\n", ths, ok, len(ths), slug)
	if forumConf {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find forum by slug: `, slug, `"}`)

		return
	}
	if !ok {
		resp, _ := json.Marshal(`{teapot:"not ok"}`)

		ctx.SetStatusCode(fasthttp.StatusTeapot)
		ctx.SetContentType("plain/text")
		ctx.SetBody(resp)

		return
	}

	resp, _ := json.Marshal(ths)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBody(resp)
}

func Vote(ctx *fasthttp.RequestCtx) {
	vote := models.Vote{}
	err := json.Unmarshal(ctx.PostBody(), &vote)
	if err != nil {
		//fmt.Println("Vote unmarshal err:", err)
	}
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}

	var th *models.Thread
	if id == 0 {
		th = models.VoteSlug(&slug, &vote)
	} else {
		th = models.VoteID(id, &vote)
	}
	if th == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody([]byte(`{"message": "Can't find thread by slug or user by nickname"}`))
		return
	} else {
		resp, _ := json.Marshal(th)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
}

func UpdateThread(ctx *fasthttp.RequestCtx) {
	thUPD := models.ThreadUPD{}
	err := json.Unmarshal(ctx.PostBody(), &thUPD)
	if err != nil {
		//fmt.Println("UpdateThread unmarshal err:", err)
	}
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}

	var th *models.Thread
	if id == 0 {
		th = models.UpdateThreadSlug(&slug, &thUPD)
	} else {
		th = models.UpdateThreadID(&id, &thUPD)
	}
	if th == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody([]byte(`{"message": "Can't find thread by slug or id"}`))
		return
	} else {
		resp, _ := json.Marshal(th)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
}

func GetThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		id = 0
	}

	var th *models.Thread
	if id == 0 {
		th = models.GetThreadSlug(&slug)
	} else {
		th = models.GetThreadID(&id)
	}
	if th == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody([]byte(`{"message": "Can't find thread by slug or id"}`))
		return
	} else {
		resp, _ := json.Marshal(th)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
}
