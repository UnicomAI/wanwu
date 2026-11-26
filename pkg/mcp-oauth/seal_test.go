package mcpoauth

import (
	"bytes"
	"testing"
)

func TestCiphertextBindsOwnerAndResource(t *testing.T) {
	c, e := NewCodec(bytes.Repeat([]byte{7}, 32))
	if e != nil {
		t.Fatal(e)
	}
	o := Owner{"org", "alice"}
	g := &Grant{AccessToken: "private-access", RefreshToken: "private-refresh"}
	b, e := c.Seal(o, "https://resource.example/mcp", g)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(b, []byte(g.RefreshToken)) {
		t.Fatal("plaintext credential")
	}
	got, e := c.Open(o, "https://resource.example/mcp", b)
	if e != nil || got.RefreshToken != g.RefreshToken {
		t.Fatal(e)
	}
	for _, other := range []Owner{{"other", "alice"}, {"org", "bob"}} {
		if _, e = c.Open(other, "https://resource.example/mcp", b); e == nil {
			t.Fatal("identity substitution accepted")
		}
	}
	if _, e = c.Open(o, "https://different.example/mcp", b); e == nil {
		t.Fatal("resource substitution accepted")
	}
	b[len(b)-1] ^= 1
	if _, e = c.Open(o, "https://resource.example/mcp", b); e == nil {
		t.Fatal("tampering accepted")
	}
}
