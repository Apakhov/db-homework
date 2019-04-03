package main

import (
	"fmt"
	"log"

	"github.com/Apakhov/db-homework/controllers"

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
	router.GET("/api/forum/:slug/users", controllers.GetUsers)

	router.POST("/api/thread/:slug_or_id/create", controllers.CreatePost)
	router.POST("/api/thread/:slug_or_id/vote", controllers.Vote)
	router.GET("/api/thread/:slug_or_id/details", controllers.GetThread)
	router.POST("/api/thread/:slug_or_id/details", controllers.UpdateThread)
	router.GET("/api/thread/:slug_or_id/posts", controllers.GetPosts)

	router.GET("/api/post/:id/details", controllers.GetPostInfo)
	router.POST("/api/post/:id/details", controllers.UpdatePost)

	router.GET("/api/service/status", controllers.GetInfo)
	router.POST("/api/service/clear", controllers.Clear)

	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
