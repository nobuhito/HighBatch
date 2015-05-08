package highbatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func write(wo Spec) {
	ld("in write")
	rand.Seed(time.Now().UnixNano())

	if err := writeLocal(wo); err != nil {
		le(err)
		return
	}

	if err := sendMaster(wo); err != nil {
		le(err)
		return
	}

	if err := deleteLocal(wo.Completed); err != nil {
		le(err)
		return
	}
}

func deleteLocal(filename string) error {
	ld("in deleteLocal")
	f := strings.Join([]string{"temp", filename}, string(os.PathSeparator))
	if _, err := os.Stat(f); err == nil {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	ld("out deleteLocal")
	return nil
}

func writeLocal(wo Spec) error {
	ld("in writeLocal")
	if _, err := os.Stat("temp"); err != nil {
		if err := os.Mkdir("temp", 0777); err != nil {
			return err
		}
	}

	m, err := json.Marshal(wo)
	if err != nil {
		return err
	}

	file := strings.Join([]string{"temp", wo.Completed}, string(os.PathSeparator))
	if err := ioutil.WriteFile(file, m, 0777); err != nil {
		return err
	}

	return nil
}

func reSend() {
	ld("in ReSend")
	if _, err := os.Stat("temp"); err != nil {
		if err := os.Mkdir("temp", 0777); err != nil {
			le(err)
		}
	}

	for {
		time.Sleep(15 * 60 * time.Second)
		root := "temp"
		if err := filepath.Walk(root,
			func(path string, info os.FileInfo, err error) error {

				if info.IsDir() {
					return nil
				}

				file := strings.Split(path, string(os.PathSeparator))[1]
				isMatch, err := regexp.MatchString("^\\d{15}$", file)
				if err != nil || !isMatch {
					le(err)
					return err
				}

				contents, err := ioutil.ReadFile(path)
				if err != nil {
					le(err)
					return err
				}

				var wo Spec
				if err := json.Unmarshal(contents, &wo); err != nil {
					le(err)
					return err
				}

				if err := sendMaster(wo); err != nil {
					le(err)
					return err
				}

				if err := deleteLocal(wo.Completed); err != nil {
					le(err)
					return err
				}

				time.Sleep(10 * time.Second)
				return nil

			}); err != nil {
			le(err)
		}
	}
}

func sendMaster(wo Spec) error {
	ld("in sendMaster")
	m, err := json.Marshal(wo)
	if err != nil {
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest(
		"POST",
		"http://"+Conf.Client.Master.Hostname+":8081/logger",
		bytes.NewBuffer(m),
	)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		err := errors.New("Status code error")
		return err
	}
	return nil
}

func writeDB(wo Spec) error {
	ld("in writeDB")

	key := wo.Id
	value, err := json.Marshal(wo)

	if err != nil {
		return err
	}

	if err := store(key, value); err != nil {
		return err
	}

	return nil
}
