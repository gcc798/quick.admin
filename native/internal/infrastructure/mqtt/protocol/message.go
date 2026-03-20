package protocol

import (
	"encoding/json"
	"time"
)

type DownMsg struct {
	OptCode   int
	Data      interface{}
	Timestamp int64
	MsgId     string
}
type UpMsg struct {
	OptCode   int
	Data      json.RawMessage
	Timestamp int64
	MsgId     string
}

func NewDownMsg(optCode int, data interface{}) *DownMsg {
	return &DownMsg{OptCode: optCode, Data: data, Timestamp: time.Now().Unix()}
}
func (m *DownMsg) ToJSON() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type ControlData struct{ State string }
type QueryInfoData struct{ Type string }
type HeartbeatData struct {
	Mac, SnNum, NetType, IpAddress, FirmwareVer string
	SignalQuality                               int
	DoorConn                                    *int
}
type DoorStateData struct {
	Mac, State, Reason string
	Timestamp          int64
}
type DeviceInfoData struct{ Mac, SnNum, Model, FirmwareVer, HardwareVer, Manufacturer string }
type DoorConnectData struct {
	Mac       string
	Connected int
	Timestamp int64
}
type AlarmData struct {
	Mac, AlarmType, AlarmMsg string
	Level                    int
	Timestamp                int64
}

type ControlRespData struct {
	Mac     string `json:"mac"`
	Mode    int    `json:"mode"`
	State   int    `json:"state"`
	ResType int    `json:"res"`
}

type CodeOperationData struct {
	Dev int `json:"dev"`
}

func ParseUpMsg(payload []byte) (*UpMsg, error) {
	var msg UpMsg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
func ParseHeartbeatData(data json.RawMessage) (*HeartbeatData, error) {
	var hb HeartbeatData
	if err := json.Unmarshal(data, &hb); err != nil {
		return nil, err
	}
	return &hb, nil
}
func ParseDoorStateData(data json.RawMessage) (*DoorStateData, error) {
	var ds DoorStateData
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}
func ParseDeviceInfoData(data json.RawMessage) (*DeviceInfoData, error) {
	var di DeviceInfoData
	if err := json.Unmarshal(data, &di); err != nil {
		return nil, err
	}
	return &di, nil
}
func ParseDoorConnectData(data json.RawMessage) (*DoorConnectData, error) {
	var dc DoorConnectData
	if err := json.Unmarshal(data, &dc); err != nil {
		return nil, err
	}
	return &dc, nil
}
func ParseAlarmData(data json.RawMessage) (*AlarmData, error) {
	var a AlarmData
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
func ParseControlRespData(data json.RawMessage) (*ControlRespData, error) {
	var resp ControlRespData
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
