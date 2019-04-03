package models

import (
	"time"
)

type User struct {
	About    string `json:"about"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
}

type ForumDescr struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}

type Forum struct {
	Posts   int64  `json:"posts"`
	Slug    string `json:"slug"`
	Threads int64  `json:"threads"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

type ThreadDescr struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Forum   string    `json:"forum"`
	Message string    `json:"message"`
	Slug    *string   `json:"slug"`
	Title   string    `json:"title"`
}

type Thread struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Forum   string    `json:"forum"`
	ID      int64     `json:"id"`
	Message string    `json:"message"`
	Slug    *string   `json:"slug"`
	Title   string    `json:"title"`
	Votes   int64     `json:"votes"`
}

type ThreadUPD struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

type PostDescr struct {
	Author  string `json:"author"`
	Message string `json:"message"`
	Parent  *int   `json:"parent"`
}

type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum"`
	ID       int       `json:"id"`
	IsEdited bool      `json:"isEdited"`
	Message  string    `json:"message"`
	Parent   *int      `json:"parent"`
	Thread   int       `json:"thread"`
}

type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int    `json:"voice"`
}

type PostInfo struct {
	U *User   `json:"author"`
	F *Forum  `json:"forum"`
	P *Post   `json:"post"`
	T *Thread `json:"thread"`
}

type DBInfo struct {
	Forums  int `json:"forum"`
	Posts   int `json:"post"`
	Threads int `json:"thread"`
	Users   int `json:"user"`
}
