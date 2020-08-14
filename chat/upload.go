package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	userId := req.FormValue("userid")
	file, header, err := req.FormFile("avatarFile") // アップロードされるバイト列を読み込むためにio.Reader型を取得。header(fileHeader)にはメタデータとかが入ってる
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	filename := filepath.Join("avatars", userId+filepath.Ext(header.Filename)) // Joinは /a/b/c みたいにファイルパス形式にしてくれるやつで、Extはstringから拡張子を抜き出す奴
	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "success!")
}
