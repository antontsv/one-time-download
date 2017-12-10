package main

import "testing"
import "os"

func TestGetAddress(t *testing.T) {
	type test struct {
		name  string
		input string
		env   string
		out   string
	}

	cases := []test{
		{"With default", "localhost:8765", "", "localhost:8765"},
		{"With set env var", "localhost:8765", "remote.site:4142", "remote.site:4142"},
		{"With set env var and empty input", "", "remote.site:4143", "remote.site:4143"},
		{"With invalid env var", "localhost:1234", "remote.site", "localhost:1234"},
		{"With invalid port in env var", "localhost:4761", "remote.site:171263", "localhost:4761"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := os.Setenv(bindAddressEnv, tc.env)
			if err != nil {
				t.Errorf("cannot set env var '%s' for test: %v", bindAddressEnv, err)
			}
			resAddr := getAddress(tc.input)
			if resAddr != tc.out {
				t.Errorf("expected %s, got %s", tc.out, resAddr)
			}
		})
	}
}
