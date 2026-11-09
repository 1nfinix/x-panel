package main

import "testing"

func TestListenAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{name: "empty", want: ":3002"},
		{name: "whitespace", address: "  ", want: ":3002"},
		{name: "IPv4", address: "127.0.0.1", want: "127.0.0.1:3002"},
		{name: "IPv6", address: "::1", want: "[::1]:3002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listenAddress(tt.address, "3002"); got != tt.want {
				t.Fatalf("listen address = %q, want %q", got, tt.want)
			}
		})
	}
}
