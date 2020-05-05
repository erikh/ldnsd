package dnsdb

import "testing"

func TestRecordValidation(t *testing.T) {
	table := map[string]struct {
		r       *Record
		success bool
	}{
		"basic": {
			r:       &Record{Host: "test", Address: "127.0.0.1"},
			success: true,
		},
		"empty host": {
			r:       &Record{Host: "", Address: "127.0.0.1"},
			success: false,
		},
		"empty ip": {
			r:       &Record{Host: "test", Address: ""},
			success: false,
		},
		"bad ip": {
			r:       &Record{Host: "test", Address: "abcdefgh"},
			success: false,
		},
		"ipv6 ip": {
			r:       &Record{Host: "test", Address: "fe80::1"},
			success: false,
		},
		"invalid ipv4 ip": {
			r:       &Record{Host: "test", Address: "256.1.1.1"},
			success: false,
		},
		"long string is looooooong": {
			r:       &Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "fe80::1"},
			success: false,
		},
		"long string is too looooooong": {
			r:       &Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "fe80::1"},
			success: false,
		},
		"long domain is looooooong": {
			r:       &Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "127.0.0.1"},
			success: true,
		},
		"long domain has a really long part": {
			r:       &Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "127.0.0.1"},
			success: false,
		},
	}

	for testName, result := range table {
		resultErr := result.r.Validate()
		if result.success && resultErr != nil {
			t.Fatalf("Result for %q should be success but was %v", testName, resultErr)
		}
		if !result.success && resultErr == nil {
			t.Fatalf("Result for %q should NOT be success but was.", testName)
		}
	}
}
