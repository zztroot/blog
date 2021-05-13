package main

import (
	"zzt_blog/routers"
)

type MainStruct struct {
	Router routers.RouterStruct
}

func main() {
	m := MainStruct{}
	m.Router.Init()
}
