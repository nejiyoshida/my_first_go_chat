package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/stretchr/objx"

	"github.com/stretchr/gomniauth"

	"github.com/nejiyoshida/go_chat/trace"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTPメソッドが実装してあれば、HTTPリクエストを処理できる
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// symc.Once型を利用すれば、複数のgoroutineに呼び出されても引数に渡した関数が一回だけ実行することが保証される
	// 今回はtemplateのparseが一回だけ。
	t.once.Do(func() {
		// ServeHTTPが実行されるまで	テンプレートの解析が行われないため、リソース消費を遅らせることができる
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	//t.templ.Execute(w, r)
	t.templ.Execute(w, data)
}

type Config struct {
	ID     map[string]string `json:"ID"`
	Secret map[string]string `json:"secret"`
}

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config Config
	json.Unmarshal(file, &config)

	var addr = flag.String("addr", ":8080", "アプリケーションのアドレス")
	flag.Parse() // addrに *stringが当てられる
	gomniauth.SetSecurityKey("セキュリティキー")
	gomniauth.WithProviders(
		google.New(config.ID["google"], config.Secret["google"], "http://localhost:8080/auth/callback/google"),
		github.New(config.ID["github"], config.Secret["github"], "http://localhost:8080/auth/callback/github"),
		facebook.New(config.ID["facebook"], config.Secret["facebook"], "http://localhost:8080/auth/callback/facebook"),
	)

	// avatar.goの中で宣言されたUseAuthAvatarは、メモリ上に生成されてないので余分にメモリを消費しない
	//r := newRoom(UseAuthAvatar)
	//r := newRoom(UseGravatar)
	r := newRoom(UseFileSystemAvatar)
	r.tracer = trace.New(os.Stdout)
	//http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.Handle("/room", r)                 // room構造体はserveHTTPを実装しているので、Handlerに登録できる
	http.HandleFunc("/auth/", loginHandler) // 構造体にServeHTTPを実装する以外に、シグネチャの一致する関数をHandleFuncで登録可能
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	http.HandleFunc("/uploader", uploadHandler)

	// StripPrefixで /avatars/ 以降を取り出すハンドラを作る
	// http.FileServeは静的ファイルの提供とか、一覧の作成とか
	// http.Dirは公開するフォルダを指定。
	http.Handle("/avatars/",
		http.StripPrefix("/avatars/",
			http.FileServer(http.Dir("./avatars"))))

	go r.run() // goroutineとしてチャットルーム処理の起動
	log.Println("サーバを起動します。ポート：", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe :", err)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "", // cookieが消えない場合のために空文字で上書き
		Path:   "/",
		MaxAge: -1, // MaxAgeが-1なのでcookieが消える(おおよそのブラウザで)
	})
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}
