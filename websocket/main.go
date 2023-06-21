package main

import (
	"encoding/json"
	"websocket/handler"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
)

// http://127.0.0.1:8199/websocket.html

var myHandler *handler.MyHandler

// init description
//
// createTime: 2023-06-01 17:24:12
//
// author: hailaz
func init() {
	myHandler = handler.NewHandler()
	glog.SetDefaultLogger(g.Log())
}

// HandlerWs description
//
// createTime: 2023-06-01 17:00:54
//
// author: hailaz
func HandlerWs(r *ghttp.Request) {
	var ctx = r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		glog.Error(ctx, err)
		r.Exit()
	}
	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}
		glog.Debug(ctx, "服务器收到消息：msgType:", msgType, " msg:", string(msg))
		switch msgType {
		case ghttp.WsMsgText:
			msgBody := handler.MsgBody{}
			err := json.Unmarshal(msg, &msgBody)
			if err != nil {
				glog.Error(ctx, err)
				continue
			}
			msgBody.TimeString = gtime.Now().Format("Y-m-d H:i:s")
			msgBody.SetConn(ws)
			switch msgBody.MsgType {
			case handler.MsgTypeLogin:
				myHandler.Login(ctx, &msgBody)
			case handler.MsgTypeLogout:
				myHandler.Logout(ctx, &msgBody)
			case handler.MsgTypeUserList:
				myHandler.UserList(ctx, &msgBody)
			case handler.MsgTypeSendSingle, handler.MsgTypeSendGroup, handler.MsgTypeSendAll:
				myHandler.SendMsg(ctx, &msgBody)
			case handler.MsgTypeAddGroup:
				myHandler.AddGroup(ctx, &msgBody)
			case handler.MsgTypeDelGroup:
				myHandler.DelGroup(ctx, &msgBody)
			case handler.MsgTypeGroupIn:
				myHandler.GroupIn(ctx, &msgBody)
			case handler.MsgTypeGroupOut:
				myHandler.GroupOut(ctx, &msgBody)
			case handler.MsgTypeGroupList:
				myHandler.GroupList(ctx, &msgBody)
			}
		}
	}
}

// HanderSend description
//
// createTime: 2023-06-21 16:49:12
//
// author: hailaz
func HanderSend(r *ghttp.Request) {
	var ctx = r.Context()
	msgBody := handler.MsgBody{}
	err := r.Parse(&msgBody)
	if err != nil {
		glog.Error(ctx, err)
		return
	}
	msgBody.UserName = "hailaz2"
	glog.Debugf(ctx, "服务器http收到消息：%+v", msgBody)
	myHandler.SendMsgFromHttp(ctx, &msgBody)

}

func main() {
	s := g.Server()
	s.BindHandler("/ws", HandlerWs)
	s.BindHandler("/send", HanderSend)
	s.SetServerRoot(gfile.MainPkgPath())
	// s.SetPort(8199)
	s.Run()
}
