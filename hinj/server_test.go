package hinj

import "testing"

func TestNewHINJServerUnixAddr(t *testing.T) {
	url := "unix:///Users/myUser/file.sock"
	expectedNetwork := "unix"
	expectedString := "/Users/myUser/file.sock"
	server, err := NewHINJServer(url)
	if err != nil {
		t.Fatalf("NewHINJServer() returned an unexpected error: %s", err)
	}
	if server.Addr.Network() != expectedNetwork {
		t.Fatalf("error: expected network = %s, found %s", expectedNetwork, server.Addr.Network())
	}
	if server.Addr.String() != expectedString {
		t.Fatalf("error: expected String = %s, found %s", expectedString, server.Addr.String())
	}
}
