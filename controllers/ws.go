package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
)

type WsController struct {
	beego.Controller
}

func (this *WsController) Join() {
	username := this.GetString("username")
	room := this.GetString("room")
	if (len(username)) == 0 {
		this.Redirect("/", 302)
		return
	}
	if (len(room)) == 0 {
		this.Redirect("/", 302)
		return
	}
	this.Data["room"] = room
	this.Data["username"] = username
	this.Render()
}

func (this *WsController) WebSocket() {
	username := this.GetString("username")
	room := this.GetString("room")
	if (len(username)) == 0 {
		this.Redirect("/", 302)
		return
	}
	if (len(room)) == 0 {
		this.Redirect("/", 302)
		return
	}
	func() {
		//如果是同一个房间，名字一样的话，第二次进入会把第一次覆盖
		clients.Mu.Lock()
		defer clients.Mu.Unlock()
		for client := range clients.Client {
			if client.room == room && client.username == username {
				delete(clients.Client, client)
				client.conn.Close()
				clients.RoomNum[client.room] = clients.RoomNum[client.room] - 1
				msg := client.username + " 掉线了"
				client.BroadcastInfo(1, msg)
			}
		}
		//限制房间最大人数,得从新进房
		if clients.RoomNum[room] >= 2 {
			this.Redirect("/", 302)
			return
		}
	}()
	fmt.Println("clients.RoomNum[room]", clients.RoomNum[room])

	//创建新的连接
	conn, err := websocket.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil, 1024, 1024)
	if err != nil {
		log.Println(err)
		return
	}

	//这个是进入房间之后，前端和后端建立好连接，生成一个连接对象，我们这里叫订阅对象
	subTmp := &Subscriber{username: username, room: room, conn: conn, messages: make(chan []byte, 5)}
	subscribe <- subTmp

	go subTmp.writePump()
	subTmp.readPump(true)
}

var (
	subscribe chan *Subscriber
	broadcast chan []byte
	//客户端连接数
	clients WebSocketClient
)

//全网websocket连接端管理
type WebSocketClient struct {
	Client  map[*Subscriber]bool
	RoomNum map[string]int
	Mu      sync.RWMutex
}

//进房人员订阅信息
type Subscriber struct {
	username string
	room     string
	conn     *websocket.Conn
	messages chan []byte
}

//发送消息给前端模块
func (c *Subscriber) writePump() {
	for {
		select {
		case message := <-c.messages:
			fmt.Printf("send message is :%s\n", message)
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *Subscriber) BroadcastInfo(msgType int, msg string) {
	info := make(map[string]string)
	info["username"] = c.username
	info["room"] = c.room
	info["message"] = msg
	info["type"] = strconv.Itoa(msgType)
	tmpInfo, err := json.Marshal(info)
	if err != nil {
		return
	}
	broadcast <- tmpInfo
}

//读前端发过来的消息模块
func (c *Subscriber) readPump(first bool) {
	if first {
		msg := c.username + " 进房了"
		c.BroadcastInfo(1, msg)
	}
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.BroadcastInfo(0, string(message))
	}
}

func init() {
	go manager()
	go checkClient()
}

//检测client是否正常
func checkClient() {
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			func() {
				clients.Mu.Lock()
				defer clients.Mu.Unlock()
				for client := range clients.Client {
					err := client.conn.WriteMessage(websocket.TextMessage, []byte("heartbeat"))
					fmt.Println("len(clients)", len(clients.Client))
					if err != nil {
						client.conn.Close()
						delete(clients.Client, client)
						clients.RoomNum[client.room] = clients.RoomNum[client.room] - 1
						msg := client.username + " 掉线了"
						client.BroadcastInfo(1, msg)
					}
				}
			}()
			fmt.Println("len(clients)", len(clients.Client))
		}
	}
}

//主要是管理广播和进房的全局在线人员管理
func manager() {
	clients = WebSocketClient{Client: map[*Subscriber]bool{}, RoomNum: map[string]int{}}
	broadcast = make(chan []byte, 10)
	subscribe = make(chan *Subscriber)

	for {
		select {
		case tmpClient := <-broadcast:
			fmt.Println("broadcast length:", len(broadcast), clients)
			func() {
				clients.Mu.Lock()
				defer clients.Mu.Unlock()
				for client := range clients.Client {
					clientInfo := make(map[string]string)
					json.Unmarshal([]byte(tmpClient), &clientInfo)
					if clientInfo["room"] == client.room {
						select {
						case client.messages <- tmpClient:
						default:
							fmt.Println("clients", len(clients.Client))
							close(client.messages)
							delete(clients.Client, client)
							clients.RoomNum[client.room] = clients.RoomNum[client.room] - 1
							msg := client.username + " 掉线了"
							client.BroadcastInfo(1, msg)
						}
					}

				}
			}()
		case itemClient := <-subscribe:
			func() {
				clients.Mu.Lock()
				defer clients.Mu.Unlock()
				clients.Client[itemClient] = true
				clients.RoomNum[itemClient.room] = clients.RoomNum[itemClient.room] + 1
			}()
		}
	}
}
