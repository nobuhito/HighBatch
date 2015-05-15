package highbatch

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/goji/glogrus"
	"github.com/kardianos/service"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type program struct{}

var logger service.Logger

var loglevel int // 1:info 2:error 3:warn

func route(m *web.Mux) {
	m.Get("/ka/:host/:port", kaHandler)
	m.Post("/worker", workerHandler)
	m.Post("/logger", loggerHandler)
	m.Get("/dump/:num", dumpHandler)
	m.Get("/dump", dumpHandler)
	m.Get("/worker/dump", workerDumpHandler)
	m.Get("/worker/list", workerListHandler)
	m.Get("/exec/:key", execHandler)
	m.Get("/resolve/:id", resolveHandler)
	m.Get("/data/:machine/:task/:completed", dataHandler)
	m.Get("/data/:machine/:task", dataHandler)
	m.Get("/data/:machine", dataHandler)
	m.Get("/data", dataHandler)
	m.Get("/conf/data", confDataHandler)
	m.Get("/conf", confHandler)
	m.Get("/source/:name/:file", sourceHandler)
	m.Post("/webhook", webhookHnadler)
	m.Get("/", mainHandler)

	staticPattern := regexp.MustCompile("^/(css|js|img|file)")
	goji.Handle(staticPattern, http.FileServer(http.Dir("public/static")))

	docPattern := regexp.MustCompile("^/(doc)")
	goji.Handle(docPattern, http.FileServer(http.Dir("public")))
}

func mainHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getHtml("index()"))
}

func confHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getHtml("conf()"))
}

func startWebserver() {

	if _, err := os.Stat("log"); err != nil {
		if err := os.Mkdir("log", 0666); err != nil {
			le(err)
			lw("can't create log directory")
		}
	}
	f, err := os.OpenFile("log/goji.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		le(err)
		lw("can't create log file")
	}

	logger := logrus.Logger{
		Formatter: &logrus.TextFormatter{DisableColors: true},
		Out:       f,
	}

	goji.Use(glogrus.NewGlogrus(&logger, "HighBatch"))
	goji.Abandon(middleware.Logger)

	flag.Set("bind", ":" + Conf.Worker.Port)
	route(goji.DefaultMux)
	goji.Serve()
}

func sourceHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	file := strings.Join([]string{"tasks", c.URLParams["name"], c.URLParams["file"]}, string(os.PathSeparator))
	source, err := readAssets(file)
	if err != nil {
		le(err)
	}

	fmt.Fprintf(w, source)
}

func confDataHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	j, _ := json.Marshal(Conf)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(j))
}

func dataHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	var f filter
	if c.URLParams["machine"] != "" {
		f.machine = c.URLParams["machine"]
	}
	if c.URLParams["task"] != "" {
		f.task = c.URLParams["task"]
	}
	if c.URLParams["conpletedatetime"] != "" {
		f.start = c.URLParams["completed"]
	}

	j, _ := json.Marshal(getSpecList("highbatch", f))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(j))
}

func kaHandler(c web.C, w http.ResponseWriter, r *http.Request) {

	v := WorkerInfo{c.URLParams["host"], c.URLParams["port"], time.Now(), 0}
	values, err := json.Marshal(&v)
	if err != nil {
		le(err)
	}
	_ = store("workers", c.URLParams["host"], values)

	j, _ := json.Marshal(changeDate)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(j))
}

func rootHandler(c web.C, w http.ResponseWriter, r *http.Request) {
}

func webhookHnadler(c web.C, w http.ResponseWriter, r *http.Request) {

	var spec Spec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		le(err)
		return
	}

	if spec.Name == "" {
		http.Error(w, "Job name required.", http.StatusInternalServerError)
		return
	}

	spec.Schedule = "WebHook"
	spec.Assets = nil
	spec.Route = []string{}
	if spec.Completed == "" {
		spec.Completed = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))
	}
	if spec.Started == "" {
		spec.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	}
	if spec.Hostname == "" {
		spec.Hostname = strings.Split(r.Host, ":")[0]
	}

	key := md5.Sum([]byte(spec.Name))
	spec.Key = hex.EncodeToString(key[:])
	spec.Id = spec.Started + "_" + spec.Hostname + "_" + spec.Key

	if err := writeDB(spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if spec.ExitCode != 0 {
		notify(spec)
	}

	taskChain(spec)

	j, _ := json.Marshal(spec)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

func loggerHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	var wo Spec
	if err := json.NewDecoder(r.Body).Decode(&wo); err != nil {
		le(err)
		return
	}

	err := writeDB(wo)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if wo.ExitCode != 0 {
		notify(wo)
	}

	taskChain(wo)

	j, _ := json.Marshal(wo)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))

}

func execHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	specs := taskFileSerch()
	for i := range specs {
		spec := specs[i]
		if spec.Key == c.URLParams["key"] {
			spec.Schedule = "Manual"
			sendWorker(spec)
			break
		}
	}
	http.Redirect(w, r, "/", 301)
}

func resolveHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	spec := getSpec("highbatch", c.URLParams["id"])
	spec.Resolved = time.Now().Format("20060102150405")
	if err := writeDB(spec); err != nil {
		le(err)
	}
	http.Redirect(w, r, "/", 301)

}

func workerHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	// var wi WorkerIn
	var s Spec
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		le(err)
		return
	}

	go worker(s)
	j, _ := json.Marshal(s)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

func dumpHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(c.URLParams["num"])
	if err != nil {
		num = 50 // 初期値は50
	}
	items := dump("highbatch", num)
	j, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

func workerDumpHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	items := dump("workers", 10)
	j, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

func workerListHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	items := getWorkerList()
	j, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {

	logger.Info("start HighBatch")

	// 設定ファイルの読み込み
	loadConfig()
	logInit(Conf.Worker.LogLevel)

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
	svcConfig := &service.Config{
		Name:        "HighBatch",
		DisplayName: "HighBatch client",
		Description: "Client for HighBatch. http://highbatch:8081",
	}

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

	if len(os.Args) > 1 {
		serviceRegist(s)
		return
	}

	err = s.Run()
	if err != nil {
		fmt.Println(err)
		logger.Error(err)
	}

}

func serviceRegist(s service.Service) {
	verb := os.Args[1]
	switch verb {
	case "install":
		err := s.Install()
		if err != nil {
			fmt.Printf("Failed to install %s\n", err)
			return
		}
	case "uninstall":
		err := s.Uninstall()
		if err != nil {
			fmt.Printf("Faild to uninstall %s\n", err)
			return
		}
	}
	return
}

func logInit(l int) {
	loglevel = l
	if l == 0 {
		loglevel = 3
	}
	if os.Getenv("HighBatchLogLevel") != "" {
		envLogLevel, err := strconv.Atoi(os.Getenv("HighBatchLogLevel"))
		if err != nil {
			loglevel = 3
		} else {
			loglevel = envLogLevel
		}
	}
}

func ld(msg string) {
	if loglevel < 2 {
		logger.Info(msg)
	}
}

func li(msg string) {
	if loglevel < 2 {
		logger.Info(msg)
	}
}

func lw(msg string) {
	if loglevel < 3 {
		logger.Warning(msg)
	}
}

func le(err error) {
	logger.Error(err)
}

func getData(url string) (string, error) {
	ld("in getData")
	timeout := time.Duration(3 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	response, err := client.Get(url)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		err := errors.New("HTTP Status code error")
		return "", err
	}

	b, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return string(b), nil
}
