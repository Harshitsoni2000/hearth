package server

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestLANAddrsFormat(t *testing.T) {
	addrs := LANAddrs(8080)

	for _, a := range addrs {
		if !strings.HasPrefix(a, "http://") {
			t.Errorf("addr %q missing http:// prefix", a)
		}
		if !strings.HasSuffix(a, ":8080") {
			t.Errorf("addr %q missing :8080 suffix", a)
		}

		ipStr := strings.TrimSuffix(strings.TrimPrefix(a, "http://"), ":8080")
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Errorf("addr %q does not contain a valid IP", a)
		}
	}
}

func TestLANAddrsUsesGivenPort(t *testing.T) {
	addrs := LANAddrs(9999)

	for _, a := range addrs {
		if !strings.HasSuffix(a, ":9999") {
			t.Errorf("addr %q does not use requested port 9999", a)
		}
	}
}

func TestLANAddrsExcludesLoopback(t *testing.T) {
	addrs := LANAddrs(8080)

	for _, a := range addrs {
		if strings.Contains(a, "127.0.0.1") {
			t.Errorf("addr %q includes loopback address, want excluded", a)
		}
	}
}

func TestLANAddrsOnlyIPv4(t *testing.T) {
	addrs := LANAddrs(8080)

	for _, a := range addrs {
		ipStr := strings.TrimSuffix(strings.TrimPrefix(a, "http://"), ":8080")
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Fatalf("addr %q does not contain a valid IP", a)
		}
		if ip.To4() == nil {
			t.Errorf("addr %q is not IPv4", a)
		}
	}
}

// TestLANAddrsMatchesInterfaceAddrs cross-checks LANAddrs' output against a
// fresh, independent read of net.InterfaceAddrs, so the test fails if
// LANAddrs' filtering logic ever drifts from "IPv4, non-loopback".
func TestLANAddrsMatchesInterfaceAddrs(t *testing.T) {
	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("net.InterfaceAddrs: %v", err)
	}

	var want []string
	for _, a := range ifaceAddrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil || ipNet.IP.IsLoopback() {
			continue
		}
		want = append(want, fmt.Sprintf("http://%s:8080", ipNet.IP.String()))
	}

	got := LANAddrs(8080)

	if len(got) != len(want) {
		t.Fatalf("LANAddrs(8080) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("LANAddrs(8080)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
