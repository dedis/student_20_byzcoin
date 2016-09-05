package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnType(t *testing.T) {
	var tests = []struct {
		Value    string
		Expected ConnType
	}{
		{"tcp", PlainTCP},
		{"tls", TLS},
		{"purb", PURB},
		{"tcp4", InvalidConnType},
		{"_tls", InvalidConnType},
	}

	for _, str := range tests {
		if connType(str.Value) != str.Expected {
			t.Error("Wrong ConnType for " + str.Value)
		}
	}
}

func TestAddress(t *testing.T) {
	var tests = []struct {
		Value   string
		Valid   bool
		Type    ConnType
		Address string
		Host    string
		Port    string
	}{
		{"tls://10.0.0.4:2000", true, TLS, "10.0.0.4:2000", "10.0.0.4", "2000"},
		{"tcp://10.0.0.4:2000", true, PlainTCP, "10.0.0.4:2000", "10.0.0.4", "2000"},
		{"purb://10.0.0.4:2000", true, PURB, "10.0.0.4:2000", "10.0.0.4", "2000"},
		{"tls4://10.0.0.4:2000", false, InvalidConnType, "", "", ""},
		{"tls://1000.0.0.4:2000", false, InvalidConnType, "", "", ""},
		{"tls://10.0.0.4:20000000", false, InvalidConnType, "", "", ""},
		{"tls://10.0.0.4:-10", false, InvalidConnType, "", "", ""},
		{"tlsx10.0.0.4:2000", false, InvalidConnType, "", "", ""},
		{"tls:10.0.0.4x2000", false, InvalidConnType, "", "", ""},
		{"tlsx10.0.0.4x2000", false, InvalidConnType, "", "", ""},
		{"tlxblurdie", false, InvalidConnType, "", "", ""},
		{"tls://blublublu", false, InvalidConnType, "", "", ""},
	}

	for i, str := range tests {
		add := Address(str.Value)
		assert.Equal(t, str.Valid, add.Valid(), "Address (%d) %s", i, str.Value)
		assert.Equal(t, str.Type, add.ConnType(), "Address (%d) %s", i, str.Value)
		assert.Equal(t, str.Address, add.NetworkAddress())
		assert.Equal(t, str.Host, add.Host())
		assert.Equal(t, str.Port, add.Port())
	}
}
