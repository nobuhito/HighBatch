package main

import (
	"./highbatch"
	"github.com/BurntSushi/toml"
	"os"
	"bytes"
	"io/ioutil"
	"fmt"
	"path/filepath"
)

func main() {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir)
	if err != nil {
		fmt.Println(err)
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Println(err)
	}

	highbatch.LogInit("debug")

	if bootCheck() {							// 設定ファイルの有無
		highbatch.Ld("start HighBatch")

		// 設定ファイルの読み込み
		highbatch.LoadConfig()
		highbatch.Ld("load config")

		if highbatch.Conf.Server.Name != "" { // マスターの場合

			// DBの作成
			highbatch.Initdb()
			highbatch.Ld("init boltdb")

			// バッチ処理指示エンジンの起動
			go highbatch.StartArranger()
			highbatch.Ld("start arranger")

			// バッチ処理設定関係の変更監視とダウンロード用ZIP作成
			go highbatch.StartWatcher()
			highbatch.Ld("start watcher()")
		}

		// 処理結果の未送信情報再送信
		go highbatch.ReSend()
		highbatch.Ld("start resend")

		// バッチ処理エンジンの起動
		go highbatch.StartWorker()
		highbatch.Ld("start worker")

		// 通信用HTTPサーバー起動
		highbatch.StartWebserver()
	}
}

func bootCheck() bool {
	configfile := "config.toml"

	if _, err := os.Stat(configfile); err != nil {
		var config highbatch.ClientConfigFile

		tag := []string{"test"}
		config.Client.Tag = tag
		config.Client.Master.Hostname = "localhost"
		config.Client.Master.Port = "8081"

		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		if err := encoder.Encode(config); err != nil {
			highbatch.Le(err);
		}
		if err := ioutil.WriteFile(configfile, []byte(buffer.String()), 0644); err != nil {
			highbatch.Le(err)
		}
		highbatch.Lw("create config file")
		return false
	} else {
		return true
	}
}
