package main

import (
	"strings"
	"websocket/handler"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
)

// http://127.0.0.1:8199/websocket.html

// init description
//
// createTime: 2023-06-01 17:24:12
//
// author: hailaz
func init() {

	glog.SetDefaultLogger(g.Log())
}

func main() {
	s := g.Server()
	masterAddress := "127.0.0.1:8080"
	localAddress := "127.0.0.1:" + strings.Split(s.GetListenedAddress(), ":")[1]

	myHandler := handler.NewHandler(masterAddress, localAddress)
	s.BindHandler("/ws", myHandler.HandlerWs)
	s.BindHandler("/send", myHandler.HanderSend)
	s.SetServerRoot(gfile.MainPkgPath())

	myHandler.UpdateTimer()

	// s.SetPort(8199)
	s.Run()
}
