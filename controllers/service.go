package controllers

import (
	"encoding/json"

	"github.com/Apakhov/db-homework/models"
	"github.com/valyala/fasthttp"
)

func GetInfo(ctx *fasthttp.RequestCtx) {
	info := models.GetInfo()
	if info == nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	} else {
		resp, _ := json.Marshal(info)

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBody(resp)
	}

}

func Clear(ctx *fasthttp.RequestCtx) {
	models.Clear()

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
}
