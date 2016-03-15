package main

import (
	"testing"
	"time"
)

func TestApiKey(t *testing.T) {
	key := ApiKey{
		Pid:  1000,
		Name: "LTE Watch",
		Time: time.Now(),
	}
	s := key.String()
	t.Log(s)

	var key2 ApiKey
	if err := key2.Decode(s); err != nil {
		t.Error(err)
		return
	}
	t.Log(key2)
}
