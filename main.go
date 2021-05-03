package main

import (
	"github.com/astaxie/beego"
	_ "mychat/routers"
)

func main() {
	beego.Run()
}
