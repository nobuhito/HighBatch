package highbatch

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/BurntSushi/toml"
	"github.com/go-fsnotify/fsnotify"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var zipfile string
var changeDate string

var Conf Config

type Config struct {
	Master MasterConfig
	Worker WorkerConfig
	Notify NotifyConfig
}

type ConfigClient struct {
	Master MasterConfig
	Worker WorkerConfig
}

type MasterConfig struct {
	Host string
	Port string
}

type WorkerConfig struct {
	Host     string
	Port     string
	LogLevel int
	IsMaster bool
}

type NotifyConfig struct {
	SmtpAuth SmtpAuth
	MailInfo MailInfo
}

type MailInfo struct {
	FromAddress string
	ToAddress   []string
	Host        string
	Port        string
}

type SmtpAuth struct {
	User string
	Pass string
}

type WorkerInfo struct {
	Host     string    `json:"host"`
	Port     string    `json:"port"`
	Datetime time.Time `json:"dt"`
	IsAlive  int       `json:"isAlive"`
}

type Spec struct {
	Id          string   `json:"id"`
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Cmd         string   `json:"cmd"`
	Schedule    string   `json:"schedule"`
	Route       []string `json:"route"`
	Chain       []string `json:"chain"`
	Error       string   `json:"error"`
	OnErrorStop string   `json:"onErrorStop"`
	Group       string   `json:"group"`
	Assets      []string `json:"assets"`

	Machine []string `json:"machine"`
	Tags    []string `json:"tags"`

	Hostname    string `json:"hostname"`
	Started     string `json:"started"`
	Completed   string `json:"completed"`
	Duration    string `json:"duration"`
	ExitCode    int    `json:"exitCode"`
	Output      string `json:"output"`
	DurationInt string `json:"durationInt"`
	Resolved    string `json:"resolved"`
}

type WorkersInfo []WorkerInfo

func loadConfig() (c Config) {
	if _, err := toml.DecodeFile("config.toml", &Conf); err != nil {
		le(err)
	}
	c = Conf

	if os.Getenv("HighBatchIsMaster") != "" {
		c.Worker.IsMaster = true
	}

	return
}

func startWatcher() {
	ld("in tartWatcher")
	zipfile = strings.Join([]string{"public", "static", "file", "tasks.zip"}, string(os.PathSeparator))
	checkTasks()
	watchTasks()
}

func checkTasks() {
	ld("in checkTasks")
	if _, err := os.Stat(zipfile); err != nil {
		doZip("tasks", zipfile)
	}

	info, _ := os.Stat(zipfile)
	changeDate = info.ModTime().Format("20060102150405")
}

func watchTasks() {
	ld("in watchTasks")
	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()
	if err != nil {
		le(err)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				ld("tasks change: " + event.Name)
				doZip("tasks", zipfile)
				info, _ := os.Stat(zipfile)
				changeDate = info.ModTime().Format("20060102150405")
			case err := <-watcher.Errors:
				le(err)
			}
		}
	}()

	if err := filepath.Walk("tasks",
		func(path string, info os.FileInfo, err error) error {

			if info.IsDir() {
				if err := watcher.Add(path); err != nil {
					le(err)
				}
			}

			return nil
		}); err != nil {
		le(err)
	}

	<-done
}

func doUnzip(path string) {
	ld("in doUnzip")
	if err := os.RemoveAll("tasks"); err != nil {
		le(err)
	}
	reader, err := zip.OpenReader(path)
	defer reader.Close()
	if err != nil {
		le(err)
	}

	var rc io.ReadCloser
	for _, f := range reader.File {
		rc, err = f.Open()
		defer rc.Close()
		if err != nil {
			le(err)
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, rc); err != nil {
			le(err)
		}

		s := f.Name
		d, _ := filepath.Split(s)
		if _, err := os.Stat(d); err != nil {
			os.MkdirAll(d, 0755)
		}

		if err := ioutil.WriteFile(s, buf.Bytes(), 0755); err != nil {
			le(err)
		}
	}
}

func doZip(archivePath string, zipPath string) {
	ld("in doZIP")
	file, err := os.Create(zipfile)
	if err != nil {
		le(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	if err := filepath.Walk(archivePath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			body, err := ioutil.ReadFile(path)
			if err != nil {
				le(err)
				return err
			}

			if body != nil {
				f, err := zw.Create(path)
				if err != nil {
					le(err)
					return err
				}

				if _, err := f.Write(body); err != nil {
					le(err)
					return err
				}
			}

			return nil
		}); err != nil {
		le(err)
	}
}

func copyAsset(path string, copyTo string) error {
	copyToParent := strings.Join([]string{"public", "tasks"}, string(os.PathSeparator))

	if _, err := os.Stat(copyToParent); err != nil {
		if err := os.Mkdir(copyToParent, 0777); err != nil {
			return err
		}
	}

	src, err := os.Open(path)
	defer src.Close()
	if err != nil {
		le(err)
		return err
	}

	dst, err := os.Create(copyToParent + copyTo)
	if err != nil {
		le(err)
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		le(err)
		return err
	}
	return nil
}

func taskFileSerch() (specs []Spec) {
	ld("in taskFileSerch")
	root := "tasks"

	if err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {

			isMatch, err := regexp.MatchString("\\.toml$", path)
			if err != nil {
				return err
			}

			if info.IsDir() || !isMatch {
				return nil
			}

			spec := parseSpec(path)
			specs = append(specs, spec)

			return nil

		}); err != nil {
		le(err)
	}
	return
}

func readAssets(task string) (asset string, err error) {
	body, err := ioutil.ReadFile(task)
	if err != nil {
		return "", err
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
	return string(f), nil
}

func findAssets(task string) (assets []string) {
	ld("in findAssets")
	if err := filepath.Walk(task,
		func(path string, info os.FileInfo, err error) error {
			isMatch, err := regexp.MatchString("\\.toml$", path)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !isMatch {
				file := strings.Split(path, string(os.PathSeparator))[2]
				assets = append(assets, file)
				return nil
			}

			return nil
		}); err != nil {
		le(err)
	}
	return assets
}

func parseSpec(path string) (s Spec) {
	ld("in parseSpec")

	if _, err := toml.DecodeFile(path, &s); err != nil {
		le(err)
	}

	name := strings.Split(path, string(os.PathSeparator))[1]
	key := md5.Sum([]byte(name))
	s.Key = hex.EncodeToString(key[:])
	s.Name = name
	s.Assets = findAssets(strings.Join([]string{"tasks", name}, string(os.PathSeparator)))

	return s
}
