package main

import (
	"fmt"
	"log"

	"github.com/db-homework/controllers"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

func CreateForumControl(ctx *fasthttp.RequestCtx) {
	if ctx.UserValue("slug").(string) == "create" {
		controllers.CreateForum(ctx)
		return
	}
	controllers.CreateTread(ctx)
}

func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	router.POST("/api/user/:nickname/create", controllers.CreateUser)
	router.GET("/api/user/:nickname/profile", controllers.GetUser)
	router.POST("/api/user/:nickname/profile", controllers.UpdateUser)

	router.POST("/api/forum/:slug", CreateForumControl)
	router.GET("/api/forum/:slug/details", controllers.GetForum)
	router.POST("/api/forum/:slug/create", CreateForumControl)
	router.GET("/api/forum/:slug/threads", controllers.GetThreadsByForumSlug)

	router.POST("/api/thread/:slug_or_id/create", controllers.CreatePost)
	router.POST("/api/thread/:slug_or_id/vote", controllers.Vote)
	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
