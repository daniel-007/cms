package main

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-nsq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	writeWait  = 10 * time.Second
	pingPeriod = 60 * time.Second
	pongWait   = 90 * time.Second
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	config = nsq.NewConfig()
)

type notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
}

type message struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Duration int    `json:"duration,omitempty"`
}

func msgSendHandler(c *gin.Context) {
	var key ApiKey
	info := authInfo(c.Request.Header.Get("Authorization"))
	if err := key.Decode(info["key"]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Bad Api Key"})
		return
	}

	var form struct {
		To           string                 `json:"to"`
		Notification notification           `json:"notification"`
		Message      message                `json:"message"`
		Data         map[string]interface{} `json:"data"`
	}
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request"})
		return
	}
	if form.Notification.Title != "" {
		data, _ := json.Marshal(&form.Notification)
		var token Token
		token.Decode(form.To)
		if token.Pid != key.Pid {
			c.JSON(http.StatusNotAcceptable, gin.H{"status": "Bad Receiver Token"})
			return
		}
		topic := Topic(fmt.Sprintf("%d.uid.%s", key.Pid, token.Uid))
		if err := topic.Pub(data); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": err.Error()})
			return
		}
	}
	if form.Message.Content != "" {
		var token Token
		token.Decode(info["token"])
		if token.Pid != key.Pid {
			c.JSON(http.StatusNotAcceptable, gin.H{"status": "Bad Receiver Token"})
			return
		}
		var msg struct {
			From    string  `json:"from"`
			To      string  `json:"to"`
			Message message `json:"message"`
			Time    int64   `json:"time"`
		}
		msg.From = token.Uid
		msg.To = form.To
		msg.Message = form.Message
		msg.Time = time.Now().Unix()
		data, _ := json.Marshal(&msg)
		topic := Topic(fmt.Sprintf("%d.uid.%s", key.Pid, msg.To))
		if err := topic.Pub(data); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": err.Error()})
			return
		}

	}
	/*
		if len(form.Data) > 0 {
			var token Token
			token.Decode(form.To)
			if token.Pid != key.Pid {
				c.JSON(http.StatusNotAcceptable, gin.H{"status": "Bad Receiver Token"})
				return
			}
			data, _ := json.Marshal(v)
		}
	*/
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func msgRecvHandler(c *gin.Context) {
	var key ApiKey
	info := authInfo(c.Request.Header.Get("Authorization"))
	if err := key.Decode(info["key"]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Bad Api Key"})
		return
	}

	var token Token
	token.Decode(info["token"])
	if token.Pid != key.Pid {
		c.JSON(http.StatusNotAcceptable, gin.H{"status": "Bad Receiver Token"})
		return
	}

	topic := fmt.Sprintf("%d.uid.%s", key.Pid, token.Uid)
	channel := fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(info["token"])))
	fmt.Fprintf(gin.DefaultWriter, "[SUB] %v | %s/%s | %s |\n",
		time.Now().Format("2006/01/02 - 15:04:05"), topic, channel, c.Request.RemoteAddr)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	cons, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Println(err)
		return
	}
	cons.SetLogger(log.New(os.Stderr, "", log.Flags()), nsq.LogLevelError)
	cons.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		log.Println(string(m.Body))
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.TextMessage, m.Body)
	}))
	if err := cons.ConnectToNSQLookupds([]string{"127.0.0.1:4161"}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": err.Error()})
		return
	}

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			fmt.Println(message)
		}
		cons.Stop()
	}()

	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				fmt.Println(err)
				return
			}
		case <-cons.StopChan:
			ticker.Stop()
			return
		}
	}
}
