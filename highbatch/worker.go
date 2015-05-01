package highbatch

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
	"math/rand"
)

func StartWorker() {
	Ld("in StartWorker")
	refleshTasks()
	sendKeepalive()
}

func refleshTasks() {
	Ld("in refleshTasks")
	if err := getTasks(); err != nil {
		Le(err)
	} else if Conf.Server.Name == "" {
		doUnzip("tasks.zip")
	}
}

func getTasks() error {
	Ld("in getTasks")
	url := "http://" + Conf.Client.Master.Hostname + ":8081/file/tasks.zip"
	resp, err := http.Get(url)
	if err != nil {
		Le(err)
		return err
	} else {

		defer resp.Body.Close()
		file, err := os.OpenFile("tasks.zip", os.O_CREATE|os.O_WRONLY, 0644)
		defer file.Close()
		if err != nil {
			Le(err)
		}

		io.Copy(file, resp.Body)
		return nil
	}
}

func sendKeepalive() {
	Ld("in sendKeepalive")
	for {
		re, err := getData("http://" + Conf.Client.Master.Hostname + ":8081/ka")
		if err != nil {
			Le(err)
		}

		info, _ := os.Stat("tasks.zip")
		downloadedDate := info.ModTime().Format("20060102150405")
		uploadedDate := strings.Replace(re, "\"", "", 2)
		Li(fmt.Sprint("%s : %s", uploadedDate, downloadedDate))
		if uploadedDate > downloadedDate {
			refleshTasks()
		}

		time.Sleep(60 * time.Second)
	}
}

func getData(url string) (string, error) {
	Ld("in getData")
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	b, _ := ioutil.ReadAll(response.Body)
	return string(b), nil
}

func worker(wo Spec) {
	Ld("in Worker")

	wo.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	cmd := strings.Split(wo.Cmd, " ")

	startTime := time.Now()

	if runtime.GOOS == "windows" {
		wo.ExitCode, wo.Output = executeWinCmd(wo.Name, cmd)
	}

	if wo.Error != "" {
		isMatch, err := regexp.MatchString(wo.Error, wo.Output)
		if err != nil {
			Le(err)
		}
		if isMatch {
			wo.ExitCode = 99
		}
	}

	duration := time.Now().Sub(startTime)
	wo.Duration = duration.String()
	wo.DurationInt = fmt.Sprint(duration.Nanoseconds())
	wo.Completed = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	Li(fmt.Sprint(wo))

	go write(wo)
}

func executeWinCmd(path string, cmdSlice []string) (exitCode int, output string) {
	Ld("in executeWinCmd")

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
