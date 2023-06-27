package handler

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
)

var (
	SystemName = "system"
)

// MyUser description
type MyUser struct {
	User
	conn *ghttp.WebSocket
}

// MyGroup description
type MyGroup struct {
	Group
	Users []*MyUser
}

// Send description
//
// createTime: 2023-06-01 16:55:13
//
// author: hailaz
func (u *MyUser) Send(ctx context.Context, msg *MsgBody) error {
	glog.Debug(ctx, "Send msg", msg)
	// if u.Name == msg.Sender {
	// 	return nil
	// }
	glog.Debug(ctx, "SendTo", u.Name, u.Address)

	if u.conn != nil {
		// 本机
		if u.conn != nil {
			u.conn.WriteJSON(msg)
			// u.conn.WriteMessage(websocket.TextMessage, []byte(msg.Content))
		} else {
			return errors.New("User not found")
		}

	} else {
		// 跨机
		msg.UserName = u.Name
		gclient.New().PostContent(ctx, "http://"+u.Address+"/send", msg)

	}
	return nil
}

// MyHandler description
type MyHandler struct {
	userList      sync.Map
	groupList     sync.Map
	MasterAddress string
	LocalAddress  string
}

// NewHandler description
//
// createTime: 2023-06-01 16:58:54
//
// author: hailaz
func NewHandler(masterAddress string, localAddress string) *MyHandler {
	return &MyHandler{
		MasterAddress: masterAddress,
		LocalAddress:  localAddress,
	}
}

func (h *MyHandler) IsMaster() bool {
	return h.MasterAddress == h.LocalAddress
}

func (h *MyHandler) Login(ctx context.Context, msg *MsgBody) error {
	if msg.UserName == SystemName {
		return errors.New("system name is not allowed")
	}
	if user, ok := h.userList.Load(msg.UserName); ok {
		if _, ok := user.(MyUser); ok {
			return errors.New("user already login")
		}
	}
	user := MyUser{
		User: User{
			Name:    msg.UserName,
			Address: h.LocalAddress,
		},
		conn: msg.conn,
	}
	h.userList.Store(msg.UserName, user)
	msg.Sender = SystemName
	h.SendMsg(ctx, msg)
	return nil
}

func (h *MyHandler) Logout(ctx context.Context, msg *MsgBody) error {
	h.userList.Delete(msg.UserName)
	h.SendMsg(ctx, msg)
	return nil
}

func (h *MyHandler) LogoutWithCon(ctx context.Context, conn *ghttp.WebSocket) error {
	h.userList.Range(func(key, value interface{}) bool {
		if userObj, ok := value.(MyUser); ok {
			if userObj.conn != nil {
				glog.Debug(ctx, "LogoutWithCon", userObj.Name, userObj.conn.RemoteAddr().String(), conn.RemoteAddr().String())
				if userObj.conn.RemoteAddr().String() == conn.RemoteAddr().String() {
					// 不存在
					glog.Debug(ctx, "离线下线", userObj.Name)
					h.userList.Delete(userObj.Name)
				}
			}
		}
		return true
	})
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
	return userList, nil
}

func (h *MyHandler) SendMsg(ctx context.Context, msg *MsgBody) error {
	glog.Debugf(ctx, "SendMsg: %+v", *msg)
	switch msg.MsgType {
	case MsgTypeSendSingle:
		// 发送给单个用户
		if user, ok := h.userList.Load(msg.UserName); ok {
			glog.Debug(ctx, "SendMsg ", msg.UserName)
			if userObj, ok := user.(MyUser); ok {
				userObj.Send(ctx, msg)
			}
		} else {
			return errors.New("User not found")
		}
	case MsgTypeSendGroup:
		// 发送给群组
		if group, ok := h.groupList.Load(msg.GroupName); ok {
			if groupObj, ok := group.(MyGroup); ok {
				for _, user := range groupObj.Users {
					// 使用user.conn发送消息
					user.Send(ctx, msg)
				}
			}
		} else {
			return errors.New("Group not found")
		}
	case MsgTypeSendAll, MsgTypeLogin, MsgTypeLogout:
		// 发送给所有用户
		if msg.MsgType == MsgTypeLogin || msg.MsgType == MsgTypeLogout {
			userList, _ := h.UserList(ctx, msg)
			userListMsg := MsgBody{
				MsgType: MsgTypeUserList,
				Sender:  SystemName,
				Data:    userList,
			}
			h.userList.Range(func(key, value interface{}) bool {
				if userObj, ok := value.(MyUser); ok {
					// 使用userObj.conn发送消息
					userObj.Send(ctx, msg)
					userObj.Send(ctx, &userListMsg)
				}
				return true
			})
		} else {
			h.userList.Range(func(key, value interface{}) bool {
				if userObj, ok := value.(MyUser); ok {
					// 使用userObj.conn发送消息
					userObj.Send(ctx, msg)
				}
				return true
			})
		}

	default:
		return errors.New("invalid message type")
	}

	return nil

}

func (h *MyHandler) SendMsgFromHttp(ctx context.Context, msg *MsgBody) error {
	glog.Debugf(ctx, "SendMsg: %+v", *msg)

	if user, ok := h.userList.Load(msg.UserName); ok {
		glog.Debug(ctx, "SendMsg ", msg.UserName)
		if userObj, ok := user.(MyUser); ok {
			userObj.Send(ctx, msg)
		}
	} else {
		return errors.New("User not found")
	}

	return nil

}

func (h *MyHandler) UserListFromHttp(ctx context.Context, msg *MsgBody) ([]User, error) {
	var userList []User
	h.userList.Range(func(key, value interface{}) bool {
		if user, ok := value.(MyUser); ok {
			glog.Debug(ctx, "当前用户列表 ", user.Name, user.Address)
			userList = append(userList, user.User)
		}
		return true
	})
	return userList, nil
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
			groupObj.Users = append(groupObj.Users, &user)
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
			Users: []*MyUser{&user},
		}
		h.groupList.Store(msg.GroupName, group)
	}

	msg.Send(*msg)
	return nil
}

func (h *MyHandler) DelGroup(ctx context.Context, msg *MsgBody) error {
	if group, ok := h.groupList.Load(msg.GroupName); ok {
		if groupObj, ok := group.(MyGroup); ok {
			for i, user := range groupObj.Users {
				if user.Name == msg.UserName {
					// 从群组中移除用户
					groupObj.Users = append(groupObj.Users[:i], groupObj.Users[i+1:]...)
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
			groupObj.Users = append(groupObj.Users, &user)
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
			for i, user := range groupObj.Users {
				if user.Name == msg.UserName {
					// 从群组中移除用户
					groupObj.Users = append(groupObj.Users[:i], groupObj.Users[i+1:]...)
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
	msg.Data = groupList
	msg.Send(*msg)
	return groupList, nil
}

// HandlerWs description
//
// createTime: 2023-06-01 17:00:54
//
// author: hailaz
func (h *MyHandler) HandlerWs(r *ghttp.Request) {
	var ctx = r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		glog.Error(ctx, err)
		r.Exit()
	}
	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			// glog.Error(ctx, err, ws.Conn.RemoteAddr())
			h.LogoutWithCon(ctx, ws)
			return
		}
		glog.Debug(ctx, "服务器收到消息：msgType:", msgType, " msg:", string(msg))
		switch msgType {
		case ghttp.WsMsgText:
			msgBody := MsgBody{}
			err := json.Unmarshal(msg, &msgBody)
			if err != nil {
				glog.Error(ctx, err)
				continue
			}
			msgBody.TimeString = gtime.Now().Format("Y-m-d H:i:s")
			msgBody.SetConn(ws)
			switch msgBody.MsgType {
			case MsgTypeLogin:
				h.Login(ctx, &msgBody)
			case MsgTypeLogout:
				h.Logout(ctx, &msgBody)
			case MsgTypeUserList:
				h.UserList(ctx, &msgBody)
			case MsgTypeSendSingle, MsgTypeSendGroup, MsgTypeSendAll:
				h.SendMsg(ctx, &msgBody)
			case MsgTypeAddGroup:
				h.AddGroup(ctx, &msgBody)
			case MsgTypeDelGroup:
				h.DelGroup(ctx, &msgBody)
			case MsgTypeGroupIn:
				h.GroupIn(ctx, &msgBody)
			case MsgTypeGroupOut:
				h.GroupOut(ctx, &msgBody)
			case MsgTypeGroupList:
				h.GroupList(ctx, &msgBody)
			}
		}
	}
}

func (h *MyHandler) CheckAndAddUser(ctx context.Context, msg *MsgBody) error {
	// 判断msg.Data类型
	// g.DumpWithType(msg)

	if userListStr, ok := msg.Data.(string); ok {
		userList := make([]User, 0)
		err := json.Unmarshal([]byte(userListStr), &userList)
		if err != nil {
			return err
		}
		// 移除不在线用户
		h.userList.Range(func(key, value interface{}) bool {
			if userObj, ok := value.(MyUser); ok {
				// 判断当前上报机器的用户是否在线
				if h.IsMaster() {
					if userObj.Address == msg.Sender {
						isExist := false
						for _, user := range userList {
							if userObj.Name == user.Name {
								isExist = true
								break
							}
						}
						if !isExist {
							// 不存在
							glog.Debug(ctx, "delete when User not exist", userObj.Name)
							h.userList.Delete(userObj.Name)
							msg.MsgType = MsgTypeLogout
							h.SendMsg(ctx, msg)
						}
					}
				} else {
					isExist := false
					for _, user := range userList {
						if userObj.Name == user.Name {
							isExist = true
							break
						}
					}
					if !isExist {
						// 不存在
						glog.Debug(ctx, "delete when User not exist", userObj.Name)
						h.userList.Delete(userObj.Name)
						msg.MsgType = MsgTypeLogout
						h.SendMsg(ctx, msg)
					}
				}

			}
			return true
		})

		// 添加新用户
		for _, user := range userList {
			userName := user.Name
			if _, ok := h.userList.Load(userName); ok {
				// 已经存在
				// glog.Debug(ctx, "User already exist", userName)

			} else {
				// 不存在
				glog.Debug(ctx, "User not exist", userName)
				if h.IsMaster() {
					// 如果是主节点，需要判断是否是当前上报机器的用户
					if user.Address == msg.Sender {
						user := MyUser{
							User: User{
								Name:    userName,
								Address: user.Address,
							},
						}
						h.userList.Store(userName, user)
						msg.MsgType = MsgTypeLogin
						h.SendMsg(ctx, msg)
					}

				} else {

					user := MyUser{
						User: User{
							Name:    userName,
							Address: user.Address,
						},
					}
					h.userList.Store(userName, user)
					msg.MsgType = MsgTypeLogin
					h.SendMsg(ctx, msg)
				}

			}
		}
	} else {
		// 移除该节点所有用户
		h.userList.Range(func(key, value interface{}) bool {
			if userObj, ok := value.(MyUser); ok {
				if userObj.Address == msg.Sender {
					// 不存在
					glog.Debug(ctx, "移除该节点所有用户", userObj.Name)
					h.userList.Delete(userObj.Name)
					msg.MsgType = MsgTypeLogout
					h.SendMsg(ctx, msg)
				}
			}
			return true
		})
	}

	return nil
}

// HanderSend description
//
// createTime: 2023-06-21 16:49:12
//
// author: hailaz
func (h *MyHandler) HanderSend(r *ghttp.Request) {
	var ctx = r.Context()
	msgBody := MsgBody{}
	err := r.Parse(&msgBody)
	if err != nil {
		glog.Error(ctx, err)
		return
	}
	// msgBody.UserName = "hailaz2"
	// glog.Debugf(ctx, "服务器http收到消息：%+v", msgBody)
	switch msgBody.MsgType {
	case MsgTypeSendSingle, MsgTypeSendGroup, MsgTypeSendAll:

		h.SendMsgFromHttp(ctx, &msgBody)
	case MsgTypeUserList:
		if msgBody.Sender == SystemName {
			return
		}
		glog.Debugf(ctx, "收到[%s]上报: %v", msgBody.Sender, msgBody.Data)
		h.CheckAndAddUser(ctx, &msgBody)

		// 返回当前在线用户列表
		userList, _ := h.UserListFromHttp(ctx, nil)
		glog.Debugf(ctx, "UserListFromHttp：userList:%+v", userList)
		r.Response.WriteJsonExit(userList)
	}

}

// HanderSend description
//
// createTime: 2023-06-21 16:49:12
//
// author: hailaz
func (h *MyHandler) UpdateTimer() {
	ctx := gctx.New()
	// 发送当前在线用户列表
	gtimer.AddSingleton(ctx, time.Second*3, func(ctx context.Context) {
		// glog.Debug(ctx, "定时器：", time.Now().Format("2006-01-02 15:04:05"))

		if h.IsMaster() {
			// 主机不上报

		} else {
			userList, _ := h.UserListFromHttp(ctx, nil)
			// glog.Debug(ctx, "userList:", userList)
			msgBody := MsgBody{
				MsgType: MsgTypeUserList,
				Data:    userList,
				Sender:  h.LocalAddress,
			}
			// 上报到主服务器
			respUserList := gclient.New().PostContent(ctx, "http://"+h.MasterAddress+"/send", msgBody)

			// 处理主服务器返回的用户列表
			msgBody.Data = respUserList
			msgBody.Sender = h.MasterAddress
			glog.Debugf(ctx, "收到[%s]返回: %v", msgBody.Sender, msgBody.Data)
			h.CheckAndAddUser(ctx, &msgBody)
		}

	})
	glog.Debugf(ctx, "启动上报: [%s]%s", h.MasterAddress, h.LocalAddress)
}
