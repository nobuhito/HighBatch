package highbatch

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
)

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

func startArranger() {
	ld("in startArranger")
	for {
		c := cron.New()

		specs := taskFileSerch()

		for i := range specs {
			spec := specs[i]
			li(fmt.Sprintf("%v", spec))
			if spec.Schedule != "" {
				c.AddFunc(spec.Schedule, func() {
					if spec, err := sendWorker(spec); err != nil {
						spec.ExitCode = 99
						spec.Output = err.Error()
						spec.Hostname = "unknown"
						spec.Completed = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

						go write(spec)
					}
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

func sendWorker(spec Spec) (Spec, error) {
	ld("in sendWorkder")
	spec.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))
	client := &http.Client{}
	for i := range spec.Machine {

		spec.Hostname = spec.Machine[i]

		workerPort := get("workers", spec.Hostname)
		if workerPort == "" {
			continue
		}

		spec.Id = spec.Started + "_" + spec.Hostname + "_" + spec.Key

		if err := writeDB(spec); err != nil {
			return spec, err
		}

		m, _ := json.Marshal(spec)

		req, err := http.NewRequest(
			"POST",
			"http://"+spec.Hostname+":"+workerPort+"/worker",
			bytes.NewBuffer(m),
		)
		if err != nil {
			return spec, err
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		// defer resp.Body.Close()
		if err != nil {
			return spec, err
		}

		if resp.StatusCode != 200 {
			err := errors.New("HTTP Status code error")
			return spec, err
		}
	}
	return spec, nil
}
