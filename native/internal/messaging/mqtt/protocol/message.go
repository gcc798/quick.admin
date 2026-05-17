package protocol

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DownMsg 下行消息结构（云平台 -> 设备）
type DownMsg struct {
	Type  int                    `json:"type"`         // 消息类型
	FCode int                    `json:"fCode"`        // 设备端功能码
	Opt   map[string]interface{} `json:"opt"`          // 设备端指令参数
	Tm    int64                  `json:"tm"`           // 设备端时间戳
	ID    string                 `json:"id,omitempty"` // 设备端消息ID

	OptCode   int         `json:"-"` // 旧 Go 代码使用的操作码
	Data      interface{} `json:"-"` // 旧 Go 代码使用的消息数据
	Timestamp int64       `json:"-"` // 旧 Go 代码使用的时间戳
	MsgId     string      `json:"-"` // 旧 Go 代码使用的消息ID
}

// UpMsg 上行消息结构（设备 -> 云平台）
type UpMsg struct {
	Type      int             `json:"type"`      // 消息类型
	FCode     int             `json:"fCode"`     // 功能码/操作码
	OptCode   int             `json:"optCode"`   // 兼容旧 Go 格式操作码
	Data      json.RawMessage `json:"data"`      // 消息数据
	Ret       json.RawMessage `json:"ret"`       // 设备端上行业务数据
	Opt       json.RawMessage `json:"opt"`       // 设备端参数结构
	Timestamp int64           `json:"timestamp"` // 时间戳
	Tm        int64           `json:"tm"`        // 设备端时间戳
	MsgId     string          `json:"msgId"`     // 消息ID
	ID        string          `json:"id"`        // 设备端消息ID
}

// NewDownMsg 创建组件实例。
func NewDownMsg(optCode int, data interface{}) *DownMsg {
	now := time.Now().Unix()
	msgID := generateMsgId()
	return &DownMsg{
		Type:      0,
		FCode:     optCode,
		Tm:        now,
		ID:        msgID,
		OptCode:   optCode,
		Data:      data,
		Timestamp: now,
		MsgId:     msgID,
	}
}

// NewModeControlMsg 创建组件实例。
func NewModeControlMsg(state string) *DownMsg {
	return NewDownMsg(OptCodeModeControl, ControlData{State: state})
}

// NewHeartbeatMsg 创建组件实例。
func NewHeartbeatMsg() *DownMsg {
	return NewDownMsg(OptCodeHeart, nil)
}

// NewUnbindMsg 创建组件实例。
func NewUnbindMsg() *DownMsg {
	return NewDownMsg(OptCodeUnbind, nil)
}

// NewReleaseAlarmMsg 创建组件实例。
func NewReleaseAlarmMsg() *DownMsg {
	return NewDownMsg(OptCodeAlarm, nil)
}

// NewQueryDeviceInfoMsg 创建组件实例。
func NewQueryDeviceInfoMsg() *DownMsg {
	return NewDownMsg(OptCodeQueryDeviceInfo, QueryInfoData{Type: "device"})
}

// NewDeviceCodeMatchMsg 创建组件实例。
func NewDeviceCodeMatchMsg(dev int) *DownMsg {
	return NewDownMsg(OptCodeDeviceCodeMatch, CodeMatchData{Dev: dev})
}

// NewDeviceCodeDeleteMsg 创建组件实例。
func NewDeviceCodeDeleteMsg(dev int) *DownMsg {
	return NewDownMsg(OptCodeDeviceCodeDelete, CodeDeleteData{Dev: dev})
}

// ToJSON 执行业务逻辑。
func (m *DownMsg) ToJSON() (string, error) {
	if m.FCode == 0 {
		m.FCode = m.OptCode
	}
	if m.Tm == 0 {
		m.Tm = time.Now().Unix()
	}
	m.Timestamp = m.Tm
	if m.MsgId != "" {
		m.ID = m.MsgId
	} else if m.ID != "" {
		m.MsgId = m.ID
	} else {
		m.MsgId = generateMsgId()
		m.ID = m.MsgId
	}
	if m.Opt == nil {
		m.Opt = buildDownOpt(m.OptCode, m.Data)
		if m.Opt == nil {
			m.Opt = map[string]interface{}{}
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ============ 下行消息 Data 结构 ============

// ControlData 模式控制数据
type ControlData struct {
	State string `json:"state"` // full_open/half_open/close/auto
}

// QueryInfoData 查询设备信息数据
type QueryInfoData struct {
	Type string `json:"type"` // 查询类型
}

// CodeMatchData 设备对码数据
type CodeMatchData struct {
	Dev int `json:"dev"` // 对码设备类型：0=主机，1=安防门磁
}

// CodeDeleteData 删除对码数据
type CodeDeleteData struct {
	Dev int `json:"dev"` // 对码设备类型：0=主机，1=安防门磁
}

// ============ 上行消息 Data 结构 ============

// HeartbeatData 心跳数据
type HeartbeatData struct {
	Mac           string `json:"mac"`           // MAC地址
	SnNum         string `json:"snNum"`         // 序列号
	NetType       string `json:"netType"`       // 网络类型 wifi/4g
	IpAddress     string `json:"ipAddress"`     // IP地址
	FirmwareVer   string `json:"firmwareVer"`   // 固件版本
	Ver           string `json:"ver"`           // 设备端固件版本
	Mode          int    `json:"mode"`          // 设备端当前模式
	Err           int    `json:"err"`           // 设备端异常码
	SignalQuality int    `json:"signalQuality"` // 信号质量
	DoorConn      *int   `json:"doorConn"`      // 门连接状态：0=断开，1=连接
	Door          *int   `json:"door"`          // 设备端门连接状态
	DSensor       *int   `json:"D-sensor"`      // 设备端门磁状态
	Bind          int    `json:"bind"`          // 绑定状态：0=未绑定，1=已绑定
}

// UnbindRespData 解绑响应数据
type UnbindRespData struct {
	State int `json:"state"` // 状态：1=成功，0=失败
}

// ModeControlRespData 模式控制响应数据
type ModeControlRespData struct {
	Mac     string `json:"mac"`     // MAC地址
	Mode    int    `json:"mode"`    // 模式
	State   int    `json:"state"`   // 状态
	ResType int    `json:"resType"` // 结果类型
}

// AlarmTriggerData 告警触发数据
type AlarmTriggerData struct {
	Mac       string `json:"mac"`       // MAC地址
	AlarmType string `json:"alarmType"` // 告警类型
	AlarmMsg  string `json:"alarmMsg"`  // 告警消息
	Level     int    `json:"level"`     // 告警级别
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// DeviceInfoRespData 设备信息响应数据
type DeviceInfoRespData struct {
	Mac     string `json:"mac"`   // MAC地址
	SnNum   string `json:"snNum"` // 序列号
	Mode    int    `json:"mode"`  // 模式
	Err     int    `json:"err"`   // 报警状态
	Ver     string `json:"ver"`   // 固件版本
	Bind    int    `json:"bind"`  // 绑定状态
	IP      string `json:"ip"`    // IP地址
	RSSI    int    `json:"rssi"`  // 信号强度
	Door    *int   `json:"door"`
	DSensor *int   `json:"D-sensor"`
}

// DoorConnectData 门连接状态数据
type DoorConnectData struct {
	Mac          string `json:"mac"`       // MAC地址
	Connected    int    `json:"connected"` // 连接状态：0=断开，1=连接
	Door         *int   `json:"door"`      // 设备端门连接状态
	DSensor      *int   `json:"D-sensor"`  // 设备端门磁状态
	Timestamp    int64  `json:"timestamp"` // 时间戳
	DoorProvided bool   `json:"-"`
}

// DoorExceptionData 门异常事件数据
type DoorExceptionData struct {
	Mac       string `json:"mac"`       // MAC地址
	Exception string `json:"exception"` // 异常类型
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// LastWillData 遗嘱消息数据
type LastWillData struct {
	Mac   string `json:"mac"`   // MAC地址
	Sn    string `json:"sn"`    // 设备端序列号
	SnNum string `json:"snNum"` // 序列号
}

// CodeMatchRespData 对码响应数据
type CodeMatchRespData struct {
	Mac   string `json:"mac"`   // MAC地址
	Dev   int    `json:"dev"`   // 对码设备类型
	State int    `json:"state"` // 状态：1=成功，0=失败
	Sn    string `json:"sn"`    // 设备SN
}

// DoorStateData 门状态数据
type DoorStateData struct {
	Mac   string `json:"mac"`   // MAC地址
	State string `json:"state"` // 门状态：open/close
}

// DeviceInfoData 设备信息数据
type DeviceInfoData struct {
	Mac     string `json:"mac"`   // MAC地址
	SnNum   string `json:"snNum"` // 序列号
	Mode    int    `json:"mode"`  // 模式
	Err     int    `json:"err"`   // 报警状态
	Ver     string `json:"ver"`   // 固件版本
	Bind    int    `json:"bind"`  // 绑定状态
	IP      string `json:"ip"`    // IP地址
	RSSI    int    `json:"rssi"`  // 信号强度
	Door    *int   `json:"door"`
	DSensor *int   `json:"D-sensor"`
}

// AlarmData 告警数据
type AlarmData struct {
	Mac       string `json:"mac"`       // MAC地址
	Err       int    `json:"err"`       // 设备端告警码
	AlarmType string `json:"alarmType"` // 告警类型
	AlarmMsg  string `json:"alarmMsg"`  // 告警消息
	Level     int    `json:"level"`     // 告警级别
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// ControlRespData 控制响应数据
type ControlRespData struct {
	Mac     string `json:"mac"`     // MAC地址
	Mode    int    `json:"mode"`    // 模式
	State   int    `json:"state"`   // 状态
	Res     int    `json:"res"`     // 设备端结果类型
	ResType int    `json:"resType"` // 结果类型
}

// ============ 解析函数 ============

// ParseUpMsg 解析上行消息
func ParseUpMsg(payload []byte) (*UpMsg, error) {
	var msg UpMsg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	if msg.FCode == 0 {
		msg.FCode = msg.OptCode
	}
	if msg.MsgId == "" {
		msg.MsgId = msg.ID
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = msg.Tm
	}
	if len(msg.Data) == 0 || string(msg.Data) == "null" {
		if len(msg.Ret) > 0 && string(msg.Ret) != "null" {
			msg.Data = msg.Ret
		} else if len(msg.Opt) > 0 && string(msg.Opt) != "null" {
			msg.Data = msg.Opt
		} else {
			msg.Data = json.RawMessage(payload)
		}
	}
	return &msg, nil
}

// ParseHeartbeatData 解析心跳数据
func ParseHeartbeatData(data json.RawMessage) (*HeartbeatData, error) {
	var hb HeartbeatData
	if err := json.Unmarshal(data, &hb); err != nil {
		return nil, err
	}
	if hb.FirmwareVer == "" {
		hb.FirmwareVer = hb.Ver
	}
	if hb.DoorConn == nil && hb.Door != nil {
		hb.DoorConn = hb.Door
	}
	return &hb, nil
}

// ParseDoorStateData 解析门状态数据
func ParseDoorStateData(data json.RawMessage) (*DoorStateData, error) {
	var ds DoorStateData
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}

// ParseUnbindRespData 解析解绑响应数据
func ParseUnbindRespData(data json.RawMessage) (*UnbindRespData, error) {
	var resp UnbindRespData
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseDeviceInfoData 解析设备信息数据
func ParseDeviceInfoData(data json.RawMessage) (*DeviceInfoData, error) {
	var di DeviceInfoData
	if err := json.Unmarshal(data, &di); err != nil {
		return nil, err
	}
	if di.SnNum == "" {
		var extra struct {
			Sn string `json:"sn"`
		}
		_ = json.Unmarshal(data, &extra)
		di.SnNum = extra.Sn
	}
	var link struct {
		Link struct {
			Door    *int `json:"door"`
			DSensor *int `json:"D-sensor"`
		} `json:"link"`
	}
	if err := json.Unmarshal(data, &link); err == nil {
		if di.Door == nil {
			di.Door = link.Link.Door
		}
		if di.DSensor == nil {
			di.DSensor = link.Link.DSensor
		}
	}
	return &di, nil
}

// ParseLastWillData 解析遗嘱消息数据
func ParseLastWillData(data json.RawMessage) (*LastWillData, error) {
	var lw LastWillData
	if err := json.Unmarshal(data, &lw); err != nil {
		return nil, err
	}
	if lw.SnNum == "" {
		lw.SnNum = lw.Sn
	}
	return &lw, nil
}

// ParseDoorConnectData 解析门连接数据
func ParseDoorConnectData(data json.RawMessage) (*DoorConnectData, error) {
	var dc DoorConnectData
	if err := json.Unmarshal(data, &dc); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, ok := raw["door"]; ok {
			dc.DoorProvided = true
		}
		if _, ok := raw["connected"]; ok {
			dc.DoorProvided = true
			if dc.Door == nil {
				door := dc.Connected
				dc.Door = &door
			}
		}
	}
	if dc.Door != nil {
		dc.Connected = *dc.Door
	}
	return &dc, nil
}

// ParseAlarmData 解析告警数据
func ParseAlarmData(data json.RawMessage) (*AlarmData, error) {
	var a AlarmData
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	if a.AlarmType == "" && a.Err != 0 {
		a.AlarmType = strconv.Itoa(a.Err)
	}
	return &a, nil
}

// ParseControlRespData 解析控制响应数据
func ParseControlRespData(data json.RawMessage) (*ControlRespData, error) {
	var resp ControlRespData
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.ResType == 0 && resp.Res != 0 {
		resp.ResType = resp.Res
	}
	return &resp, nil
}

// ParseCodeMatchRespData 执行业务逻辑。
func ParseCodeMatchRespData(data json.RawMessage) (*CodeMatchRespData, error) {
	var resp CodeMatchRespData
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ============ 辅助函数 ============

func buildDownOpt(optCode int, data interface{}) map[string]interface{} {
	switch v := data.(type) {
	case nil:
		if optCode == OptCodeAlarm {
			return map[string]interface{}{"state": 0}
		}
		return nil
	case ControlData:
		return map[string]interface{}{"mode": stateToMode(v.State)}
	case *ControlData:
		return map[string]interface{}{"mode": stateToMode(v.State)}
	case QueryInfoData:
		return map[string]interface{}{}
	case *QueryInfoData:
		return map[string]interface{}{}
	case CodeMatchData:
		return map[string]interface{}{"dev": v.Dev}
	case *CodeMatchData:
		return map[string]interface{}{"dev": v.Dev}
	case CodeDeleteData:
		return map[string]interface{}{"dev": v.Dev}
	case *CodeDeleteData:
		return map[string]interface{}{"dev": v.Dev}
	case map[string]interface{}:
		return v
	}
	var opt map[string]interface{}
	b, err := json.Marshal(data)
	if err == nil {
		_ = json.Unmarshal(b, &opt)
	}
	return opt
}

func stateToMode(state string) int {
	switch strings.ToLower(state) {
	case "close":
		return 0
	case "auto":
		return 1
	case "full_open", "open":
		return 2
	case "half_open":
		return 3
	default:
		return 0
	}
}

// generateMsgId 生成消息ID
func generateMsgId() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), randomString(8))
}

// randomString 生成随机字符串
func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(b)
}
