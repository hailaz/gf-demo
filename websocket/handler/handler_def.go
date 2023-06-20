package handler

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/net/ghttp"
)

type Handler interface {
	Login(ctx context.Context, msg *MsgBody) error
	Logout(ctx context.Context, msg *MsgBody) error
	UserList(ctx context.Context, msg *MsgBody) ([]User, error)
	SendMsg(ctx context.Context, msg *MsgBody) error
	AddGroup(ctx context.Context, msg *MsgBody) error
	DelGroup(ctx context.Context, msg *MsgBody) error
	GroupIn(ctx context.Context, msg *MsgBody) error
	GroupOut(ctx context.Context, msg *MsgBody) error
	GroupList(ctx context.Context, msg *MsgBody) ([]Group, error)
}

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
	MsgType    MsgType
	Content    string
	Sender     string
	UserName   string
	GroupName  string
	TimeString string
	conn       *ghttp.WebSocket
}

// SetConn description
//
// createTime: 2023-06-01 17:47:20
//
// author: hailaz
func (msg *MsgBody) SetConn(conn *ghttp.WebSocket) {
	msg.conn = conn
}

// Resp description
type Resp struct {
	Code int
	Msg  string
	Data interface{}
}

// Send description
//
// createTime: 2023-06-01 17:40:09
//
// author: hailaz
func (msg *MsgBody) Send(obj interface{}) error {
	if msg.conn == nil {
		return errors.New("conn is nil")
	}
	resp := Resp{
		Code: 0,
		Msg:  "success",
		Data: obj,
	}
	return msg.conn.WriteJSON(resp)
}

// User description
type User struct {
	Name string
}

// Group description
type Group struct {
	Name  string
	Users []*User
}
