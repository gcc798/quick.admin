package data

import "testing"

func TestParseAttachmentExpireTime(t *testing.T) {
	t.Parallel()

	cases := []string{
		"2026-03-18T12:34:56Z",
		"2026-03-18 12:34:56",
		"2026-03-18T12:34:56",
	}
	for _, input := range cases {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			parsed, err := parseAttachmentExpireTime(input)
			if err != nil {
				t.Fatalf("parseAttachmentExpireTime(%q) error = %v", input, err)
			}
			if parsed == nil {
				t.Fatalf("parseAttachmentExpireTime(%q) = nil", input)
			}
		})
	}

	if _, err := parseAttachmentExpireTime("bad-time"); err == nil {
		t.Fatalf("parseAttachmentExpireTime(bad-time) error = nil, want non-nil")
	}
}
