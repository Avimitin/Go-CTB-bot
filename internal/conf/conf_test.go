package conf

import "testing"

func TestConfPath(t *testing.T) {
	t.Log(confPath())
}

func TestReadToken(t *testing.T) {
	if ReadToken() != "abcdefg" {
		t.Error("Unexpected token")
		t.FailNow()
	}
}
