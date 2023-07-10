package handler

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// MsgType description
type MsgType string

const (
	MsgTypeLogin      MsgType = "login"
	MsgTypeLogout     MsgType = "logout"
	MsgTypeUserList   MsgType = "user_list"
	MsgTypeSendSingle MsgType = "send_single"
	MsgTypeSendGroup  MsgType = "send_group"
	MsgTypeSendAll    MsgType = "send_all"
	MsgTypeAddGroup   MsgType = "add_group"
	MsgTypeDelGroup   MsgType = "del_group"
	MsgTypeGroupIn    MsgType = "group_in"
	MsgTypeGroupOut   MsgType = "group_out"
	MsgTypeGroupList  MsgType = "group_list"
)

// MsgBody description
type MsgBody struct {
	MsgType    MsgType          `json:"msgType"`
	Content    string           `json:"content"`
	Sender     string           `json:"sender"`
	UserName   string           `json:"userName"`
	GroupName  string           `json:"groupName"`
	TimeString string           `json:"timeString"`
	Code       int              `json:"code"`
	Msg        string           `json:"msg"`
	Data       interface{}      `json:"data"`
	conn       *ghttp.WebSocket `json:"-"`
}

// SetConn description
//
// createTime: 2023-06-01 17:47:20
//
// author: hailaz
func (msg *MsgBody) SetConn(conn *ghttp.WebSocket) {
	msg.conn = conn
}

// Send description
//
// createTime: 2023-06-01 17:40:09
//
// author: hailaz
func (msg *MsgBody) Send(resp MsgBody) error {
	if msg.conn == nil {
		return errors.New("conn is nil")
	}
	glog.Debug(context.Background(), "服务器发送消息：msgType:", resp.MsgType, " msg:", resp)
	return msg.conn.WriteJSON(resp)
}

// User description
type User struct {
	Name    string `json:"name"`
	Address string `json:"address"` // ip:port 跨机的时候用
}

// Group description
type Group struct {
	Name  string  `json:"name"`
	Users []*User `json:"users"`
}
