package models

import (
	"fmt"
	"sync"
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

var (
	timerRequests   = 0
	timerTimes      = float64(0)
	timerCrit       = 0.01
	timerElapsed    = 100
	timerMutex      = sync.Mutex{}
	timerCounterAll = make(map[string]int)
	timerCounter    = make(map[string]int)
	timerTimeAll    = make(map[string]float64)
	timerTime       = make(map[string]float64)
)

type Timer struct {
	message string
	time    time.Time
}

func newTimer(msg string) (t *Timer) {
	// t = &Timer{
	// 	message: msg,
	// 	time:    time.Now(),
	// }
	// return t
	return &Timer{}
}

func (t *Timer) stop() {
	// p := time.Now().Sub(t.time)
	// timerMutex.Lock()
	// timerRequests++
	// timerTimes += p.Seconds()
	// _, ok := timerCounter[t.message]
	// if !ok {
	// 	timerCounter[t.message] = 1
	// 	timerTime[t.message] = p.Seconds()
	// 	timerCounterAll[t.message] = 1
	// 	timerTimeAll[t.message] = p.Seconds()
	// } else {
	// 	timerCounter[t.message]++
	// 	timerTime[t.message] += p.Seconds()
	// 	timerCounterAll[t.message]++
	// 	timerTimeAll[t.message] += p.Seconds()
	// }
	// if timerCounter[t.message] == timerElapsed {
	// 	timerPrintStat(t.message)
	// 	timerCounter[t.message] = 0
	// 	timerTime[t.message] = 0
	// }
	// timerMutex.Unlock()
}

func timerPrintStat(message string) {
	importance := ""
	if timerTime[message]/float64(timerCounter[message]) > timerCrit {
		importance = "!!!!!!!!!!!"
	}
	if timerTime[message]/float64(timerCounter[message]) > timerCrit*10 {
		importance = "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
	}
	if timerTime[message]/float64(timerCounter[message]) > timerCrit*100 {
		importance = "RED ALERT!!!RED ALERT!!!RED ALERT!!!RED ALERT!!!RED ALERT!!!RED ALERT!!!RED ALERT!!!RED ALERT!!!"
	}
	fmt.Printf("MSG: %s\nAVGTloc: %f\nAVGTall: %f\nREQS: %f%s\nTIME: %f%s   %s\n\n",
		message,
		timerTime[message]/float64(timerCounter[message]),
		timerTimeAll[message]/float64(timerCounterAll[message]),
		float64(timerCounterAll[message])/float64(timerRequests),
		"%",
		timerTimeAll[message]/timerTimes,
		"%",
		importance,
	)
}
