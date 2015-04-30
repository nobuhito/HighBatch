package highbatch

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web/middleware"
	"github.com/goji/glogrus"
	"github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"golang.org/x/text/transform"
	"bytes"
	"golang.org/x/net/html/charset"
	"os"
	"strings"
)

func route(m *web.Mux) {
	m.Get("/ka", kaHandler)
	m.Post("/worker", workerHandler)
	m.Post("/logger", loggerHandler)
	m.Get("/dump/:num", dumpHandler)
	m.Get("/dump", dumpHandler)
	m.Get("/exec/:key", execHandler)
	m.Get("/data/:machine/:task/:completed", dataHandler)
	m.Get("/data/:machine/:task", dataHandler)
	m.Get("/data/:machine", dataHandler)
	m.Get("/data", dataHandler)
	m.Get("/source/:name/:file", sourceHandler)
	m.Get("/", http.FileServer(http.Dir("public")))

	staticPattern := regexp.MustCompile("^/(css|js|img|file)")
	goji.Handle(staticPattern, http.FileServer(http.Dir("static")))
}

func mainHandler(c web.C, w http.ResponseWriter, r *http.Request) {

}

func StartWebserver() {

	if _, err := os.Stat("log"); err != nil {
		if err := os.Mkdir("log", 0666); err != nil {
			Le(err)
			Lw("can't create log directory")
		}
	}
	f, err := os.OpenFile("log/goji.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Le(err)
		Lw("can't create log file")
	}

	logger := logrus.Logger{
		Formatter: &logrus.TextFormatter{DisableColors: true},
		Out: f,
	}

	goji.Use(glogrus.NewGlogrus(&logger, "HighBatch"))
	goji.Abandon(middleware.Logger)

	flag.Set("bind", ":8081")
	route(goji.DefaultMux)
	goji.Serve()
}

func sourceHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	file := strings.Join([]string{"tasks", c.URLParams["name"], c.URLParams["file"]}, string(os.PathSeparator))
	body, err := ioutil.ReadFile(file)
	if err != nil {
		Le(err)
	}

	var f []byte
	encodings := []string{"sjis", "utf-8"}
	for _, enc := range encodings {
		if enc != "" {
			ee, _ := charset.Lookup(enc)
			if ee == nil {
				continue
			}
			var buf bytes.Buffer
			ic := transform.NewWriter(&buf, ee.NewDecoder())
			_, err := ic.Write(body)
			if err != nil {
				continue
			}
			err = ic.Close()
			if err != nil {
				continue
			}
			f = buf.Bytes()
			break
		}
	}

	fmt.Fprintf(w, string(f))
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
		f.completed = c.URLParams["completed"]
	}

	j, _ := json.Marshal(get(f))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(j))
}

func kaHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	j, _ := json.Marshal(changeDate)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(j))
}

func rootHandler(c web.C, w http.ResponseWriter, r *http.Request) {
}

func loggerHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	var wo Spec
	if err := json.NewDecoder(r.Body).Decode(&wo); err != nil {
		Le(err)
		return
	}

	err := writeDB(wo)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if wo.OnErrorStop == "" || wo.ExitCode == 0 {
		for i := range wo.Chain {
			file := strings.Join([]string{"tasks", wo.Chain[i], "spec.toml"}, string(os.PathSeparator))
			chainSpec := parseSpec(file)
			chainSpec.Route = wo.Route
			chainSpec.Route = append(chainSpec.Route, wo.Name)
			sendWorker(chainSpec)
		}
	}

	j, _ := json.Marshal(wo)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))

}

func execHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Println("execHandler")
	specs := taskFileSerch()
	for i := range specs {
		spec := specs[i]
		fmt.Println(spec.Key)
		fmt.Println(c.URLParams["key"])
		if spec.Key == c.URLParams["key"] {
			spec.Schedule = "manual"
			sendWorker(spec)
			break
		}
	}
	http.Redirect(w, r, "/", 301)
}

func workerHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	// var wi WorkerIn
	var s Spec
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		Le(err)
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
	items := dump(num)
	j, _ := json.Marshal(items)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(j))
}
