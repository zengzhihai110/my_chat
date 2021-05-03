package routers

import (
	"mychat/controllers"

	"github.com/astaxie/beego"
)

func init() {
	//登录界面
	beego.Router("/", &controllers.MainController{})
	//websocket连接
	beego.Router("/ws", &controllers.WsController{}, "get:WebSocket")
	//进房的方法
	beego.Router("/join", &controllers.WsController{}, "get:Join")
}
