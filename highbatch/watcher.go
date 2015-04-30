package highbatch

import (
	"archive/zip"
	"bytes"
	"github.com/go-fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

var zipfile string
var changeDate string

func StartWatcher() {
	Ld("in tartWatcher")
	zipfile = strings.Join([]string{"static", "file", "tasks.zip"}, string(os.PathSeparator))
	checkTasks()
	watchTasks()
}

func checkTasks() {
	Ld("in checkTasks")
	if _, err := os.Stat(zipfile); err != nil {
		doZip("tasks", zipfile)
	}

	info, _ := os.Stat(zipfile)
	changeDate = info.ModTime().Format("20060102150405")
}

func watchTasks() {
	Ld("in watchTasks")
	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()
	if err != nil {
		Le(err)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				Li(fmt.Sprint(event))
				Ld("tasks change")
				doZip("tasks", zipfile)
				info, _ := os.Stat(zipfile)
				changeDate = info.ModTime().Format("20060102150405")
			case err := <-watcher.Errors:
				Le(err)
			}
		}
	}()

	if err := filepath.Walk("tasks",
		func(path string, info os.FileInfo, err error) error {

			if info.IsDir() {
				if err := watcher.Add(path); err != nil {
					Le(err)
				}
			}

			return nil
		}); err != nil {
		Le(err)
	}

	<-done
}

func doUnzip(path string) {
	Ld("in doUnzip")
	if err := os.RemoveAll("tasks"); err != nil {
		Le(err)
	}
	reader, err := zip.OpenReader(path)
	defer reader.Close()
	if err != nil {
		Le(err)
	}

	var rc io.ReadCloser
	for _, f := range reader.File {
		rc, err = f.Open()
		defer rc.Close()
		if err != nil {
			Le(err)
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, rc); err != nil {
			Le(err)
		}

		s := f.Name
		d, _ := filepath.Split(s)
		if _, err := os.Stat(d); err != nil {
			os.MkdirAll(d, 0755)
		}

		if err := ioutil.WriteFile(s, buf.Bytes(), 0755); err != nil {
			Le(err)
		}
	}
}

func doZip(archivePath string, zipPath string) {
	Ld("in doZIP")
	file, err := os.Create(zipfile)
	if err != nil {
		Le(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	if err := filepath.Walk(archivePath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			Li(fmt.Sprint(path))
			body, err := ioutil.ReadFile(path)
			if err != nil {
				Le(err)
				return err
			}

			if body != nil {
				f, err := zw.Create(path)
				if err != nil {
					Le(err)
					return err
				}

				if _, err := f.Write(body); err != nil {
					Le(err)
					return err
				}
			}

			return nil
		}); err != nil {
		Le(err)
	}
}
