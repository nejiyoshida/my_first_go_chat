package main

import (
	"log"
	"net/http"

	"github.com/stretchr/objx"

	"github.com/gorilla/websocket"
	"github.com/nejiyoshida/go_chat/trace"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type room struct {
	forward chan *message    // 他のクライアントに送るメッセージを保存するチャネル
	join    chan *client     // チャットルームに参加しようとしているユーザのためのチャネル
	leave   chan *client     // チャットルームから退室しようとしているユーザのためのチャネル
	clients map[*client]bool // ユーザXがこのルームに参加しているかどうかを示す
	tracer  trace.Tracer     // roomの操作ログを受ける
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// socket通信。 *Conn, err が返る。
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("cookieの取得に失敗しました", err)
		return
	}

	// ユーザの作成
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value), // MustFromBase64でmapに変換
	}

	r.join <- client                     // ユーザがこの部屋に参加
	defer func() { r.leave <- client }() // このメソッド終了時にユーザが退室
	go client.write()
	// read()の中でclient.socket.ReadMessage()が呼ばれて、エラーが出るまでメッセージを受け続ける
	client.read()
}

/*
宣言方法
chan 型
chan <- 型：送信用チャネル
<- chan 型：受信用チャネル

チャネル <- 変数とか　で、チャネルに値を送信
例
msg := make(chan string)
msg <- "msgチャネルへ送信するよ"

変数 := <-チャネル　で変数に値を受信とか
<-チャネル　でclose検知待ちとか
val, ok := <-チャネル　で値が入ったかクローズされたかとか

*/

func (r *room) run() {
	for {
		select {
		// joinチャネルにメッセージを受ける（誰かがチャットルームに入ってくる）場合
		// r.joinの値をclientに取り込む
		case client := <-r.join:
			r.clients[client] = true // 入室者をtrueで記録
			r.tracer.Trace("新しいクライアントが入室しました")
		case client := <-r.leave:
			delete(r.clients, client) // falseではなくて、キー自体を消す
			// TODO closeについては後で書く
			close(client.send)
			r.tracer.Trace("クライアントが退室しました")
		case msg := <-r.forward: // メッセージを受ける場合
			r.tracer.Trace("メッセージを受信しました：", msg.Message)
			for client := range r.clients {
				select {
				// 各clientのメッセージを入れるチャネルにmsgが送られるとき
				case client.send <- msg:
					r.tracer.Trace("    - メッセージがクライアントに送信されました")
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("    - メッセージ送信に失敗しました。クライアントとの接続を切ります")
				}
			}
		}
	}
}
