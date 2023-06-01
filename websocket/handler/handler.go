package handler

import (
	"context"
	"errors"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

var (
	ServerList = []string{
		"127.0.0.1:8199",
	}
)

// MyUser description
type MyUser struct {
	User
	conn    *ghttp.WebSocket
	address string // ip:port 跨机的时候用
}

// MyGroup description
type MyGroup struct {
	Group
	users []*MyUser
}

// SendMsg description
//
// createTime: 2023-06-01 16:55:13
//
// author: hailaz
func (u *MyUser) SendMsg(ctx context.Context, msg *MsgBody) error {
	glog.Debug(ctx, "SendTo", u.Name, u.address)
	glog.Debug(ctx, "SendMsg", msg)
	if u.address == "localhost" {
		// 本机
		if u.conn != nil {
			u.conn.WriteJSON(MsgBody{
				MsgType: msg.MsgType,
				Content: msg.Content,
				Sender:  msg.Sender,
			})
			// u.conn.WriteMessage(websocket.TextMessage, []byte(msg.Content))
		} else {
			return errors.New("User not found")
		}

	} else {
		// 跨机
		// ...
	}
	return nil
}

// MyHandler description
type MyHandler struct {
	userList  sync.Map
	groupList sync.Map
}

// NewHandler description
//
// createTime: 2023-06-01 16:58:54
//
// author: hailaz
func NewHandler() *MyHandler {
	return &MyHandler{}
}

func (h *MyHandler) Login(ctx context.Context, msg *MsgBody) error {
	user := MyUser{
		User: User{
			Name: msg.UserName,
		},
		conn:    msg.conn,
		address: "localhost",
	}
	h.userList.Store(msg.UserName, user)
	msg.Send(nil)
	return nil
}

func (h *MyHandler) Logout(ctx context.Context, msg *MsgBody) error {
	h.userList.Delete(msg.UserName)
	return nil
}

func (h *MyHandler) UserList(ctx context.Context, msg *MsgBody) ([]User, error) {
	var userList []User
	h.userList.Range(func(key, value interface{}) bool {
		if user, ok := value.(MyUser); ok {
			userList = append(userList, user.User)
		}
		return true
	})
	msg.Send(userList)
	return userList, nil
}

func (h *MyHandler) SendMsg(ctx context.Context, msg *MsgBody) error {
	if msg.MsgType == MsgTypeSendSingle {
		// 发送给单个用户
		if user, ok := h.userList.Load(msg.UserName); ok {
			if userObj, ok := user.(MyUser); ok {
				userObj.SendMsg(ctx, msg)
			}
		} else {
			return errors.New("User not found")
		}
	} else if msg.MsgType == MsgTypeSendGroup {
		// 发送给群组
		if group, ok := h.groupList.Load(msg.GroupName); ok {
			if groupObj, ok := group.(MyGroup); ok {
				for _, user := range groupObj.users {
					// 使用user.conn发送消息
					user.SendMsg(ctx, msg)
				}
			}
		} else {
			return errors.New("Group not found")
		}
	} else if msg.MsgType == MsgTypeSendAll {
		// 发送给所有用户
		h.userList.Range(func(key, value interface{}) bool {
			if userObj, ok := value.(MyUser); ok {
				// 使用userObj.conn发送消息
				userObj.SendMsg(ctx, msg)
			}
			return true
		})
	} else {
		return errors.New("Invalid message type")
	}
	return nil
}

func (h *MyHandler) AddGroup(ctx context.Context, msg *MsgBody) error {
	if group, ok := h.groupList.Load(msg.GroupName); ok {
		if groupObj, ok := group.(MyGroup); ok {
			user := MyUser{
				User: User{
					Name: msg.UserName,
				},
				conn: msg.conn,
			}
			groupObj.users = append(groupObj.users, &user)
			h.groupList.Store(msg.GroupName, groupObj)
		}
	} else {
		user := MyUser{
			User: User{
				Name: msg.UserName,
			},
			conn: msg.conn,
		}
		group := MyGroup{
			Group: Group{
				Name: msg.GroupName,
			},
			users: []*MyUser{&user},
		}
		h.groupList.Store(msg.GroupName, group)
	}
	return nil
}

func (h *MyHandler) DelGroup(ctx context.Context, msg *MsgBody) error {
	if group, ok := h.groupList.Load(msg.GroupName); ok {
		if groupObj, ok := group.(MyGroup); ok {
			for i, user := range groupObj.users {
				if user.Name == msg.UserName {
					// 从群组中移除用户
					groupObj.users = append(groupObj.users[:i], groupObj.users[i+1:]...)
					break
				}
			}
			h.groupList.Store(msg.GroupName, groupObj)
		}
	}
	return nil
}

func (h *MyHandler) GroupIn(ctx context.Context, msg *MsgBody) error {
	if group, ok := h.groupList.Load(msg.GroupName); ok {
		if groupObj, ok := group.(MyGroup); ok {
			user := MyUser{
				User: User{
					Name: msg.UserName,
				},
				conn: msg.conn,
			}
			groupObj.users = append(groupObj.users, &user)
			h.groupList.Store(msg.GroupName, groupObj)
		}
	} else {
		return errors.New("Group not found")
	}
	return nil
}

func (h *MyHandler) GroupOut(ctx context.Context, msg *MsgBody) error {
	if group, ok := h.groupList.Load(msg.GroupName); ok {
		if groupObj, ok := group.(MyGroup); ok {
			for i, user := range groupObj.users {
				if user.Name == msg.UserName {
					// 从群组中移除用户
					groupObj.users = append(groupObj.users[:i], groupObj.users[i+1:]...)
					break
				}
			}
			h.groupList.Store(msg.GroupName, groupObj)
		}
	} else {
		return errors.New("Group not found")
	}
	return nil
}

func (h *MyHandler) GroupList(ctx context.Context, msg *MsgBody) ([]Group, error) {
	var groupList []Group
	h.groupList.Range(func(key, value interface{}) bool {
		if group, ok := value.(Group); ok {
			groupList = append(groupList, group)
		}
		return true
	})
	msg.Send(groupList)
	return groupList, nil
}
