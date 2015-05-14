# HighBatch

[![Join the chat at https://gitter.im/nobuhito/HighBatch](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/nobuhito/HighBatch?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

HighBatch は WebUI を持ったバッチ処理等を管理するシンプルなジョブスケジュールシステムです。

HighBatch is simple job scheduler sysytem with a web console (commonly called batch processing). 

![main](https://raw.githubusercontent.com/nobuhito/HighBatch/master/public/doc/intro/img/highbatch_main_1.png)

## Consepts

- 簡単なセットアップ Easy setup
- シンプルな作り Simple design
- 商用のシステムに比べて **ゆるふわ** な運用 Loose & Fluffy system as compared to Commercial

## Features

- サーバー毎、タスク毎に実行履歴を表示 Show execution history
- 日時指定起動や間隔指定起動、順番を指定した起動をサポート Support multiple execution
- 管理画面からのタスク再実行 Instruct the Job of the re-run
- MasterとWorkerの通信はJSONでのHTTP通信 Master and Worker communication in HTTP
- 正常終了以外の場合にメールで通知 Notification by e-mail
- 記録のみ用にWebhookでの登録も可能（未実装）Starting in WebHook (Unimplemented)

## Installation

### Master

```
git clone https://github.com/nobuhito/HighBatch.git
cd HighBatch
cp config.toml.sample config.toml
nano config.toml
go get ...
go build main.go
./main &
```

*See [more infomation (in Japanese)](http://www.slideshare.net/NobuhitoSato/highbatch/22)*

### Worker

1. Windows用にビルドしたExeファイルを任意のフォルダにコピー
1. ダブルクリックして一回起動
1. config.tomlができるのでUTF8を解釈できるエディタで編集
1. ダブルクリックして再度起動

*See [more infomation (in Japanese)](http://www.slideshare.net/NobuhitoSato/highbatch/26)*

## Configuration for job

*See [configuration page (in Japanese)](http://www.slideshare.net/NobuhitoSato/highbatch/35)*

## Screenshots

![list](https://github.com/nobuhito/HighBatch/raw/master/public/doc/intro/img/highbatch_list_min.png)

![error](https://github.com/nobuhito/HighBatch/raw/master/public/doc/intro/img/highbatch_error.png)

![assets](https://github.com/nobuhito/HighBatch/raw/master/public/doc/intro/img/highbatch_assets.png)

*See [more screenshots (in Japanese)](http://www.slideshare.net/NobuhitoSato/highbatch/40)*

## Documentation

[HighBatchとは？ (in Japanese)](https://github.com/nobuhito/HighBatch/raw/master/public/doc/intro/highbatch.pdf)

## Requirements

- [Golang](https://golang.org/)
- [bolt](https://github.com/boltdb/bolt) (A low-level key/value database for Go.)
- [cron](https://github.com/robfig/cron) (a cron library for go)
- [fsnotify](https://github.com/go-fsnotify/fsnotify) (File system notifications for Go.)
- [Goji](https://github.com/zenazn/goji) (Goji is a minimalistic web framework for Golang that's high in antioxidants.)
- [service](https://github.com/kardianos/service) (Run go programs as a service on major platforms.)
- [toml](https://github.com/BurntSushi/toml) (TOML parser for Golang with reflection.)
- [jQuery](https://jquery.com/)
- [Bootstrap](http://getbootstrap.com/)
- [TreeView](https://github.com/jonmiles/bootstrap-treeview) (Tree View for Twitter Bootstrap)
- [highlight.js](https://highlightjs.org/) (Syntax highlighting for the Web)

## Support

Github の [Issues](https://github.com/nobuhito/highbatch/issues)  か [Gitter](https://gitter.im/nobuhito/HighBatch) を利用してください。

You can use [Github Issues](https://github.com/nobuhito/highbatch/issues) or [Gitter](https://gitter.im/nobuhito/HighBatch).

## License

MIT


