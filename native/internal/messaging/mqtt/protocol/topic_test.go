package protocol

import "testing"

func TestBuildTopicMatchesJavaDownlinkTopic(t *testing.T) {
	got := BuildTopic(NetTypeWiFi, "mac001", "sn001")
	want := "/edge-device/autodoorv1/mac001/sn001"
	if got != want {
		t.Fatalf("BuildTopic() = %q, want %q", got, want)
	}
}

func TestParseTopicMatchesJavaDownlinkTopic(t *testing.T) {
	netType, mac, sn, err := ParseTopic("/edge-device/autodoorv1/mac001/sn001")
	if err != nil {
		t.Fatalf("ParseTopic() error = %v", err)
	}
	if netType != "" || mac != "mac001" || sn != "sn001" {
		t.Fatalf("ParseTopic() = (%q, %q, %q), want (%q, %q, %q)", netType, mac, sn, "", "mac001", "sn001")
	}
}

func TestBuildSubscribeTopicMatchesJavaDownlinkTopic(t *testing.T) {
	got := BuildSubscribeTopic("", "", "")
	want := "/edge-device/autodoorv1/#"
	if got != want {
		t.Fatalf("BuildSubscribeTopic() = %q, want %q", got, want)
	}
}
