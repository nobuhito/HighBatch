package highbatch

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
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

func startWorker() {
	ld("in StartWorker")
	refleshTasks()
	sendKeepalive()
}

func sendKeepalive() {
	ld("in sendKeepalive")
	errorCount := 0
	for {
		time.Sleep(60 * time.Second)
		m := Conf.Master
		w := Conf.Worker
		url := fmt.Sprintf("http://%s:%s/ka/%s/%s/%s", m.Host, m.Port, w.Host, w.Port, Version)
		re, err := getData(url)
		if err != nil {
			if errorCount%10 == 0 { // エラーは10分に一回程度
				lw(err.Error())
			}
			errorCount += 1
		} else {
			info, err := os.Stat("tasks.zip")
			if err != nil {
				refleshTasks()
			} else {
				downloadedDate := info.ModTime().Format("20060102150405")
				uploadedDate := strings.Replace(re, "\"", "", 2)
				if uploadedDate > downloadedDate {
					refleshTasks()
				}
			}
			errorCount = 0
		}
	}
}

func refleshTasks() {
	ld("in refleshTasks")

	err := getTasks()
	if err != nil {
		le(err)
	}

	if !Conf.Worker.IsMaster { // Master以外のとき
		doUnzip("tasks.zip")
	}
}

func getTasks() error {
	ld("in getTasks")
	url := "http://" + Conf.Master.Host + ":" + Conf.Master.Port + "/file/tasks.zip"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.OpenFile("tasks.zip", os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func worker(wo Spec) {
	ld("in Worker")

	wo.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

	cmd := strings.Split(wo.Cmd, " ")

	startTime := time.Now()

	wo.ExitCode, wo.Output = executeCmd(wo.Name, cmd)

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

func executeCmd(path string, cmdSlice []string) (exitCode int, output string) {
	ld("in executeCmd")

	shell := ""
	if runtime.GOOS == "windows" {
		shell = "cmd"
		winbatchSlice := []string{"/c", "call"}
		cmdSlice = append(winbatchSlice, cmdSlice...)
	} else {
		shell = "/bin/sh"
		cmdSlice = append([]string{"-xec"}, cmdSlice...)
	}

	cmd := exec.Command(shell, cmdSlice...)
	if runtime.GOOS == "linux" {
		cmd.Env = os.Environ();
	}
	cmd.Dir = strings.Join([]string{"tasks", path}, string(os.PathSeparator))

	ret, err := cmd.CombinedOutput()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				err = nil
				exitCode = status.ExitStatus()
			}
		} else {
			ret = []byte(err.Error())
			exitCode = 99
		}
	}

	if runtime.GOOS == "windows" {
		ret, _, err = transform.Bytes(japanese.ShiftJIS.NewDecoder(), ret)
	}
	output = string(ret)

	return exitCode, output
}
