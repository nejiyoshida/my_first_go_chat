package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// チャットを行っている一人のユーザ
type client struct {
	socket   *websocket.Conn
	send     chan *message // メッセージが送られるチャネル
	room     *room         // クライアントが参加しているルーム
	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg // 他のクライアントに投げる
		} else {
			break
		}
		/*
			if _, msg, err := c.socket.ReadMessage(); err == nil {
				c.room.forward <- msg
			} else {
				break
			}
		*/
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		// WriteMessageの内部でconn.NextWriterが呼ばれていて、メッセージを書き終わったらcloseされる感じか
		//if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
}
