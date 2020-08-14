package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	client := new(client)
	url, err := authAvatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("値が存在しない場合、ErrNoAvatarURLを返します。")
	}

	testURL := "http://url-to-avatar/"
	client.userData = map[string]interface{}{"avatar_url": testURL}
	url, err = authAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("値があるならエラーを返すべきではない")
	} else {
		if url != testURL {
			t.Error("返すURLが正しくない")
		}
	}

}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	client := new(client)
	client.userData = map[string]interface{}{"userid": "nankan0h@shvalue"}
	url, err := gravatarAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("gravatarAvatar.GetAvatarURLがエラーを返すのはおかしいで")
	}
	if url != "//www.gravatar.com/avatar/abc" {
		t.Errorf("GravatarAvitar.GetAvatarURLが%sという誤った値を返しました", url)
	}
}

func TestFileSystemAvatar(t *testing.T) {
	filename := filepath.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename, []byte{}, 0777)
	defer func() { os.Remove(filename) }()

	var fileSystemAvatar FileSystemAvatar
	client := new(client)
	client.userData = map[string]interface{}{"userid": "abc"}
	url, err := fileSystemAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("fileSystemAvatar.GetAvatarURLがエラーを返すのはおかしいで")
	}
	if url != "avatars/abc.jpg" {
		t.Errorf("fileSystemAvatar.GetAvatarURLが%sという誤った値を返しました", url)
	}
}
