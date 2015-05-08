package highbatch

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)


type WorkerInfo struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Datetime time.Time `json:"dt"`
	IsAlive int `json:"isAlive"`
}

type WorkersInfo []WorkerInfo

func startWorker() {
	ld("in StartWorker")
	refleshTasks()
	sendKeepalive()
}

func refleshTasks() {
	ld("in refleshTasks")
	if err := getTasks(); err != nil {
		le(err)
	} else if !Conf.Worker.IsMaster { // Master以外のとき
		doUnzip("tasks.zip")
	}
}

func getTasks() error {
	ld("in getTasks")
	url := "http://" + Conf.Master.Host + ":" + Conf.Master.Port + "/file/tasks.zip"
	resp, err := http.Get(url)
	if err != nil {
		le(err)
		return err
	} else {

		defer resp.Body.Close()
		file, err := os.OpenFile("tasks.zip", os.O_CREATE|os.O_WRONLY, 0644)
		defer file.Close()
		if err != nil {
			le(err)
		}

		io.Copy(file, resp.Body)
		return nil
	}
}

func sendKeepalive() {
	ld("in sendKeepalive")
	for {
		time.Sleep(60 * time.Second)
		m := Conf.Master
		w := Conf.Worker
		url := fmt.Sprintf("http://%s:%s/ka/%s/%s", m.Host, m.Port, w.Host, w.Port)
		re, err := getData(url)
		if err != nil {
			le(err)
		}

		info, _ := os.Stat("tasks.zip")
		downloadedDate := info.ModTime().Format("20060102150405")
		uploadedDate := strings.Replace(re, "\"", "", 2)
		if uploadedDate > downloadedDate {
			refleshTasks()
		}

	}
}

func getData(url string) (string, error) {
	ld("in getData")
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	b, _ := ioutil.ReadAll(response.Body)
	return string(b), nil
}

func worker(wo Spec) {
	ld("in Worker")

	wo.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	cmd := strings.Split(wo.Cmd, " ")

	startTime := time.Now()

	if runtime.GOOS == "windows" {
		wo.ExitCode, wo.Output = executeWinCmd(wo.Name, cmd)
	}

	if wo.Error != "" {
		isMatch, err := regexp.MatchString(wo.Error, wo.Output)
		if err != nil {
			le(err)
		}
		if isMatch {
			wo.ExitCode = 99
		}
	}

	duration := time.Now().Sub(startTime)
	wo.Duration = duration.String()
	wo.DurationInt = fmt.Sprint(duration.Nanoseconds())
	wo.Completed = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	go write(wo)
}

func executeWinCmd(path string, cmdSlice []string) (exitCode int, output string) {
	ld("in executeWinCmd")

	winbatchSlice := []string{"/c", "call"}
	jobSlice := append(winbatchSlice, cmdSlice...)

	cmd := exec.Command("cmd", jobSlice...)
	cmd.Dir = strings.Join([]string{"tasks", path}, string(os.PathSeparator))

	ret, err := cmd.CombinedOutput()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				err = nil
				exitCode = status.ExitStatus()
			}
		}
	}

	ret, _, err = transform.Bytes(japanese.ShiftJIS.NewDecoder(), ret)
	output = string(ret)

	return exitCode, output
}
