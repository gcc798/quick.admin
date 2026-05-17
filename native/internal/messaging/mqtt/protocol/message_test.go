package protocol

import (
	"encoding/json"
	"testing"
)

func TestDownMsgToJSONDeviceFormat(t *testing.T) {
	msg := NewModeControlMsg("full_open")
	msg.MsgId = "42"

	payload, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(payload), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if body["fCode"] != float64(OptCodeModeControl) {
		t.Fatalf("unexpected fCode: %#v", body["fCode"])
	}
	if body["id"] != "42" {
		t.Fatalf("unexpected id: %#v", body["id"])
	}
	opt, ok := body["opt"].(map[string]any)
	if !ok {
		t.Fatalf("missing opt: %#v", body["opt"])
	}
	if opt["mode"] != float64(2) {
		t.Fatalf("unexpected mode: %#v", opt["mode"])
	}
	if _, ok := body["optCode"]; ok {
		t.Fatalf("old optCode field should not be emitted")
	}
	if _, ok := body["data"]; ok {
		t.Fatalf("old data field should not be emitted")
	}
	if _, ok := body["msgId"]; ok {
		t.Fatalf("old msgId field should not be emitted")
	}
}

func TestParseUpMsgTopLevelControlResp(t *testing.T) {
	payload := []byte(`{"type":0,"fCode":1003,"tm":1713600000,"id":"42","mac":"AA","mode":2,"state":1,"res":1}`)

	msg, err := ParseUpMsg(payload)
	if err != nil {
		t.Fatalf("ParseUpMsg failed: %v", err)
	}
	if msg.FCode != OptCodeModeControl {
		t.Fatalf("unexpected fCode: %d", msg.FCode)
	}
	if msg.MsgId != "42" {
		t.Fatalf("unexpected msgId: %q", msg.MsgId)
	}
	if msg.Timestamp != 1713600000 {
		t.Fatalf("unexpected timestamp: %d", msg.Timestamp)
	}

	resp, err := ParseControlRespData(msg.Data)
	if err != nil {
		t.Fatalf("ParseControlRespData failed: %v", err)
	}
	if resp.Mode != 2 || resp.State != 1 || resp.ResType != 1 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestParseUpMsgRetControlResp(t *testing.T) {
	payload := []byte(`{"type":0,"fCode":1003,"tm":1713600000,"id":"42","ret":{"mode":2,"state":1,"res":1}}`)

	msg, err := ParseUpMsg(payload)
	if err != nil {
		t.Fatalf("ParseUpMsg failed: %v", err)
	}
	resp, err := ParseControlRespData(msg.Data)
	if err != nil {
		t.Fatalf("ParseControlRespData failed: %v", err)
	}
	if resp.Mode != 2 || resp.State != 1 || resp.ResType != 1 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestParseHeartbeatDeviceFields(t *testing.T) {
	payload := []byte(`{"type":0,"fCode":1001,"tm":1713600000,"id":"h1","ret":{"mac":"AA","ver":"v1.2.3","door":1,"bind":1}}`)

	msg, err := ParseUpMsg(payload)
	if err != nil {
		t.Fatalf("ParseUpMsg failed: %v", err)
	}
	heartbeat, err := ParseHeartbeatData(msg.Data)
	if err != nil {
		t.Fatalf("ParseHeartbeatData failed: %v", err)
	}
	if heartbeat.FirmwareVer != "v1.2.3" {
		t.Fatalf("unexpected firmware version: %q", heartbeat.FirmwareVer)
	}
	if heartbeat.DoorConn == nil || *heartbeat.DoorConn != 1 {
		t.Fatalf("unexpected door state: %#v", heartbeat.DoorConn)
	}
	if heartbeat.Bind != 1 {
		t.Fatalf("unexpected bind state: %d", heartbeat.Bind)
	}
}

func TestParseLastWillData(t *testing.T) {
	payload := []byte(`{"type":1,"tm":1713600000,"ret":{"mac":"AA","sn":"SN001"}}`)

	msg, err := ParseUpMsg(payload)
	if err != nil {
		t.Fatalf("ParseUpMsg failed: %v", err)
	}
	lastWill, err := ParseLastWillData(msg.Data)
	if err != nil {
		t.Fatalf("ParseLastWillData failed: %v", err)
	}
	if lastWill.Mac != "AA" || lastWill.SnNum != "SN001" {
		t.Fatalf("unexpected last will: %#v", lastWill)
	}
}

func TestNewQueryDeviceInfoMsgUsesEmptyOpt(t *testing.T) {
	msg := NewQueryDeviceInfoMsg()
	msg.MsgId = "q1"

	payload, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(payload), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	opt, ok := body["opt"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected opt payload: %#v", body["opt"])
	}
	if len(opt) != 0 {
		t.Fatalf("query device info opt should be empty: %#v", opt)
	}
}

func TestParseDeviceInfoDataLinkFields(t *testing.T) {
	raw := json.RawMessage(`{"mac":"AA","sn":"SN001","ver":"v1.0.0","mode":2,"err":1,"bind":1,"ip":"192.168.1.2","rssi":-65,"link":{"door":-1,"D-sensor":1}}`)

	info, err := ParseDeviceInfoData(raw)
	if err != nil {
		t.Fatalf("ParseDeviceInfoData failed: %v", err)
	}
	if info.SnNum != "SN001" {
		t.Fatalf("unexpected snNum: %q", info.SnNum)
	}
	if info.Door == nil || *info.Door != -1 {
		t.Fatalf("unexpected door: %#v", info.Door)
	}
	if info.DSensor == nil || *info.DSensor != 1 {
		t.Fatalf("unexpected dSensor: %#v", info.DSensor)
	}
}

func TestParseUnbindRespData(t *testing.T) {
	resp, err := ParseUnbindRespData(json.RawMessage(`{"state":0}`))
	if err != nil {
		t.Fatalf("ParseUnbindRespData failed: %v", err)
	}
	if resp.State != 0 {
		t.Fatalf("unexpected state: %d", resp.State)
	}
}

func TestParseDoorConnectDataSensorOnly(t *testing.T) {
	resp, err := ParseDoorConnectData(json.RawMessage(`{"D-sensor":1}`))
	if err != nil {
		t.Fatalf("ParseDoorConnectData failed: %v", err)
	}
	if resp.DoorProvided {
		t.Fatalf("door should not be marked as provided: %#v", resp)
	}
	if resp.Door != nil {
		t.Fatalf("door should remain nil: %#v", resp.Door)
	}
	if resp.DSensor == nil || *resp.DSensor != 1 {
		t.Fatalf("unexpected dSensor: %#v", resp.DSensor)
	}
}

func TestParseDoorConnectDataConnectedField(t *testing.T) {
	resp, err := ParseDoorConnectData(json.RawMessage(`{"connected":1}`))
	if err != nil {
		t.Fatalf("ParseDoorConnectData failed: %v", err)
	}
	if !resp.DoorProvided {
		t.Fatalf("door should be marked as provided")
	}
	if resp.Door == nil || *resp.Door != 1 || resp.Connected != 1 {
		t.Fatalf("unexpected door state: %#v", resp)
	}
}
