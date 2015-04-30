package highbatch

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/robfig/cron"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"fmt"
)

type Spec struct {
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
}

func StartArranger() {
	Ld("in startArranger")
	for {
		c := cron.New()

		specs := taskFileSerch()

		for i := range specs {
			spec := specs[i]
			Li(fmt.Sprintf("%v", spec))
			if spec.Schedule != "" {
				c.AddFunc(spec.Schedule, func() {
					sendWorker(spec)
				})
			}
		}
		c.Start()
		time.Sleep(10 * 60 * time.Second)
		c.Stop()
		c = nil
	}
}

func taskFileSerch() (specs []Spec) {
	Ld("in taskFileSerch")
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
		Le(err)
	}
	return
}

func findAssets(task string) (assets []string) {
	Ld("in findAssets")
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
		Le(err)
	}
	return assets
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
		Le(err)
		return err
	}

	dst, err := os.Create(copyToParent + copyTo)
	if err != nil {
		Le(err)
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		Le(err)
		return err
	}
	return nil
}

func parseSpec(path string) (s Spec) {
	Ld("in parseSpec")

	if _, err := toml.DecodeFile(path, &s); err != nil {
		Le(err)
	}

	Li(fmt.Sprint(path))

	name := strings.Split(path, string(os.PathSeparator))[1]
	key := md5.Sum([]byte(name))
	s.Key = hex.EncodeToString(key[:])
	s.Name = name
	s.Assets = findAssets(strings.Join([]string{"tasks", name}, string(os.PathSeparator)))

	return s
}

func sendWorker(spec Spec) {
	Ld("in sendWorkder")
	spec.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))
	client := &http.Client{}
	for i := range spec.Machine {

		spec.Hostname = spec.Machine[i]

		if err := writeDB(spec); err != nil {
			Le(err)
		}

		m, _ := json.Marshal(spec)

		req, _ := http.NewRequest(
			"POST",
			"http://"+spec.Machine[i]+":8081/worker",
			bytes.NewBuffer(m),
		)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		if _, err := client.Do(req); err != nil {
			Le(err)
		}
	}
}
