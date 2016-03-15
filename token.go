package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-web/tokenizer"
	"hash/crc32"
	"net/http"
	"time"
)

var (
	tokTokenizer *tokenizer.T
)

func init() {
	aesKey := []byte{
		0x86, 0x8e, 0xea, 0x52, 0xa1, 0xe4, 0x65, 0xb0,
		0xb8, 0x4a, 0xcd, 0xb7, 0x08, 0x64, 0xa1, 0x5c,
	}
	hmacKey := []byte{
		0xbf, 0x85, 0xe5, 0xa8, 0x4f, 0x2c, 0x24, 0xac,
		0xe2, 0xc3, 0xa0, 0xc1, 0x75, 0x7b, 0x14, 0xbf,
		0x9a, 0x5f, 0x0c, 0x15, 0x46, 0x94, 0x7f, 0xfa,
		0x89, 0x06, 0x72, 0x1a, 0x73, 0xc4, 0xc2, 0x41,
		0x02, 0x42, 0x5a, 0x00, 0x19, 0x58, 0xb3, 0x4d,
		0x2e, 0xe7, 0x83, 0x41, 0x21, 0x67, 0x79, 0xac,
		0x8e, 0x53, 0x2f, 0x39, 0xac, 0x73, 0x3a, 0xa1,
		0x6e, 0x49, 0xf0, 0x95, 0x67, 0x8f, 0x01, 0x77,
	}
	tokTokenizer, _ = tokenizer.New(aesKey, hmacKey, nil)
}

type Token struct {
	Pid     uint32    `json:"pid"`            // project id
	Uid     string    `json:"uid"`            // user/device id
	Name    string    `json:"name,omitempty"` // user/device name
	Profile string    `json:"profile,omitempty"`
	Client  string    `json:"client"` // type: android/ios/device
	Time    time.Time `json:"time"`   // created time
}

func (token *Token) Encode() (s string, err error) {
	data, err := json.Marshal(token)
	if err != nil {
		return
	}
	b, err := tokTokenizer.Encode(data)
	s = string(b)
	return
}

func (token *Token) Decode(s string) (err error) {
	data, _, err := tokTokenizer.Decode([]byte(s))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, token)
	return
}

func (token *Token) String() string {
	s, _ := token.Encode()
	return s
}

func tokenGenHandler(c *gin.Context) {
	var key ApiKey
	if err := key.Decode(authInfo(c.Request.Header.Get("Authorization"))["key"]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Bad Api Key"})
		return
	}

	var form struct {
		Uid     string `json:"uid" binding:"required"`
		Name    string `json:"name" binding:"required"`
		Profile string `json:"profile"`
		Client  string `json:"client" binding:"required"`
	}
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request"})
		return
	}

	token := Token{
		Pid:     key.Pid,
		Uid:     form.Uid,
		Name:    form.Name,
		Profile: form.Profile,
		Client:  form.Client,
		Time:    time.Now(),
	}
	ts := token.String()
	topic := Topic(fmt.Sprintf("%d.uid.%s", key.Pid, token.Uid))
	if err := topic.Create(); err == nil {
		channel := Channel(fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(ts))))
		channel.Create(topic)
	} else {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"token":  ts,
		"status": "ok",
	})
}
