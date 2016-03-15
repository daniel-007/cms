package main

import (
	"testing"
	"time"
)

func TestToken(t *testing.T) {
	token := Token{
		Pid:    1000,
		Uid:    "1000",
		Name:   "ginuerzh",
		Client: "android",
		Time:   time.Now(),
	}
	s := token.String()
	t.Log(s)

	var token2 Token
	if err := token2.Decode(s); err != nil {
		t.Error(err)
		return
	}
	t.Log(token2)
}
