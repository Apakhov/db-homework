package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/Apakhov/db-homework/models"
	"github.com/valyala/fasthttp"
)

func CreateUser(ctx *fasthttp.RequestCtx) {
	user := &models.User{}
	json.Unmarshal(ctx.PostBody(), user)
	user.Nickname = ctx.UserValue("nickname").(string)
	fails := models.CreateUser(user)
	if fails != nil {
		resp, _ := json.Marshal(fails)

		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)

		return
	}
	resp, _ := json.Marshal(user)

	////fmt.Printf("hello, %s!\n%v\nerr: %s\nresp: %s\n", ctx.UserValue("nickname"), user, err, string(resp))
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(resp)
}

func GetUser(ctx *fasthttp.RequestCtx) {
	nick := ctx.UserValue("nickname").(string)
	user := models.GetUser(nick)
	if user != nil {
		resp, _ := json.Marshal(user)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "Can't find user by nickname: `, nick, `"}`)
	}
}

func UpdateUser(ctx *fasthttp.RequestCtx) {
	user := &models.User{}
	json.Unmarshal(ctx.PostBody(), user)
	user.Nickname = ctx.UserValue("nickname").(string)
	user, conflictNick := models.UpdateUser(user)
	if conflictNick != nil {
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		fmt.Fprint(ctx, `{"message": "This email is already registered by user: `, *conflictNick, `"}`)

		return
	} else if user == nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody([]byte(`{"message": "Can't find user with id #42\n"}`))
		return
	}
	resp, _ := json.Marshal(user)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBody(resp)
}
