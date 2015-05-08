package highbatch

import (
	"fmt"
	"github.com/kardianos/service"
	"os"
)

type program struct{}

var logger service.Logger

func (p *program) Start(s service.Service) error {
	fmt.Println("HighBatch service start.")
	go p.run()
	return nil
}

func (p *program) run() {

	logInit("debug")
	ld("init logging")

	ld("start HighBatch")

	// 設定ファイルの読み込み
	loadConfig()
	ld("load config")

	if Conf.Worker.IsMaster { // マスターの場合

		// DBの作成
		initdb()
		ld("init boltdb")

		// バッチ処理指示エンジンの起動
		go startArranger()
		ld("start arranger")

		// バッチ処理設定関係の変更監視とダウンロード用ZIP作成
		go startWatcher()
		ld("start watcher()")
	}

	// 処理結果の未送信情報再送信
	go reSend()
	ld("start resend")

	// バッチ処理エンジンの起動
	go startWorker()
	ld("start worker")

	// 通信用HTTPサーバー起動
	startWebserver()

	return
}

func (p *program) Stop(s service.Service) error {
	logger.Info("HighBatch service stop.")
	return nil
}

func ServiceInit() {
	fmt.Println("in service init")
	svcConfig := &service.Config{
		Name:        "HighBatch",
		DisplayName: "HighBatch client",
		Description: "Client for HighBatch. http://highbatch:8081",
	}

	fmt.Println(svcConfig)

	prog := &program{}
	s, err := service.New(prog, svcConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	logger, err = s.Logger(nil)
	if err != nil {
		fmt.Println(err)
	}

	logger.Info(os.Getwd())

	if len(os.Args) > 1 {
		fmt.Println(len(os.Args))
		verb := os.Args[1]
		switch verb {
		case "install":
			err = s.Install()
			if err != nil {
				fmt.Println("Failed to install %s\n", err)
				return
			}
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				fmt.Println("Faild to uninstall %s\n", err)
				return
			}
		}
		return
	}

	err = s.Run()
	if err != nil {
		fmt.Println(err)
		logger.Error(err)
	}

}

func logInit(loglevel string) {

}

func ld(msg string) {
	logger.Info(msg)
}

func li(msg string) {
	logger.Info(msg)
}

func lw(msg string) {
	logger.Warning(msg)
}

func le(err error) {
	logger.Error(err)
}
