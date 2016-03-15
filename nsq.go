package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	nsqdAddr = "127.0.0.1:4151"
)

type Topic string

func (t Topic) Create() (err error) {
	return httpPost(fmt.Sprintf("http://%s/topic/create?topic=%s", nsqdAddr, t), "", nil)
}

func (t Topic) Delete() (err error) {
	return httpPost(fmt.Sprintf("http://%s/topic/delete?topic=%s", nsqdAddr, t), "", nil)
}

func (t Topic) Pub(data []byte) (err error) {
	return httpPost(fmt.Sprintf("http://%s/pub?topic=%s", nsqdAddr, t), "", bytes.NewReader(data))
}

type Channel string

func (c Channel) Create(t Topic) (err error) {
	return httpPost(fmt.Sprintf("http://%s/channel/create?topic=%s&channel=%s", nsqdAddr, t, c), "", nil)
}

func (c Channel) Delete(t Topic) (err error) {
	return httpPost(fmt.Sprintf("http://%s/channel/delete?topic=%s&channel=%s", nsqdAddr, t, c), "", nil)
}

func httpPost(url string, bodyType string, body io.Reader) (err error) {
	resp, err := http.Post(url, bodyType, body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
	}
	resp.Body.Close()
	return
}
