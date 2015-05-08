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

func startWatcher() {
	ld("in tartWatcher")
	zipfile = strings.Join([]string{"static", "file", "tasks.zip"}, string(os.PathSeparator))
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
				li(fmt.Sprint(event))
				ld("tasks change")
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
			li(fmt.Sprint(path))
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
