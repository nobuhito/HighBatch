package highbatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type KeyValue struct {
	Key   string
	Value string
}
type KeyValues []KeyValue

type workerOuts []Spec

type filter struct {
	start   string
	machine string
	task    string
}

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
					return err
				}

				contents, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				var wo Spec
				if err := json.Unmarshal(contents, &wo); err != nil {
					return err
				}

				if err := sendMaster(wo); err != nil {
					return err
				}

				if err := deleteLocal(wo.Completed); err != nil {
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
		"http://"+Conf.Master.Host+":"+Conf.Master.Port+"/logger",
		bytes.NewBuffer(m),
	)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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

	if err := store("highbatch", key, value); err != nil {
		return err
	}

	return nil
}

func (w workerOuts) Len() int {
	return len(w)
}

func (w workerOuts) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func (w workerOuts) Less(i, j int) bool {
	return w[i].Started < w[j].Started
}

const dbname = "HighBatch.db"

func initdb() {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		le(err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("workers")) // Workerリスト
		return err
	}); err != nil {
		le(err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("highbatch")) // データ用
		return err
	}); err != nil {
		le(err)
	}

}

func store(bucketname, key string, value []byte) error {
	ld("in Store")
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		le(err)
		return err
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		if err := b.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	}); err != nil {
		le(err)
		return err
	}
	return nil
}

func get(bucketname, key string) (ret string) {
	ld("in get")
	db, err := bolt.Open(dbname, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		le(err)
	}
	defer db.Close()

	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		val := b.Get([]byte(key))
		if val != nil {
			ret = string(val)
		}
		return nil
	}); err != nil {
		le(err)
	}
	return ret
}

func getSpec(bucketname, key string) Spec {
	ld("in GetSpec")
	var spec Spec

	v := get(bucketname, key)
	if err := json.Unmarshal([]byte(v), &spec); err != nil {
		le(err)
	}

	return spec
}

func getWorkerList() (w WorkersInfo) {
	ld("in get worker list")
	db, err := bolt.Open(dbname, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		le(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("workers")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var workerinfo WorkerInfo
			_ = json.Unmarshal(v, &workerinfo)

			if time.Since(workerinfo.Datetime).Minutes() > 5 {
				workerinfo.IsAlive = 0
			} else {
				workerinfo.IsAlive = 1
			}
			w = append(w, workerinfo)
		}
		return nil
	})
	return w
}

func getSpecList(bucketname string, f filter) workerOuts {
	ld("in GetSpecLists")
	var wos workerOuts
	db, err := bolt.Open(dbname, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		le(err)
	}
	defer db.Close()

	reg := "^\\d{4}\\d{2}\\d{2}\\d{2}\\d{2}\\d{2}$"
	isMatch, err := regexp.MatchString(reg, f.start)
	if err != nil {
		le(err)
	}

	span := 14 // 初期値は30日分
	since := time.Now().AddDate(0, 0, span*-1).Format("20060102150405")

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketname)).Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var wo Spec
			if err := json.Unmarshal(v, &wo); err != nil {
				le(err)
			}

			if f.start == "" && wo.Started[0:14] < since {
				break
			}

			isMatch = filterCheck(f, wo)

			if isMatch {
				wos = append(wos, wo)
			}

		}
		return nil
	})

	sort.Sort(sort.Reverse(wos))
	return wos
}

func filterCheck(f filter, wo Spec) (isMatch bool) {
	isMatch = true

	if f.start != "" && f.start != wo.Started[0:14] {
		isMatch = false
	}

	if f.machine != "" && f.machine != wo.Hostname {
		isMatch = false
	}

	if f.task != "" && f.task != wo.Key {
		isMatch = false
	}

	return isMatch
}

func dump(bucketname string, num int) (kvs KeyValues) {
	ld("in Dump")
	db, err := bolt.Open(dbname, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		le(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		c := b.Cursor()

		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			kv := KeyValue{
				Key:   string(k),
				Value: string(v),
			}
			kvs = append(kvs, kv)
			i = i + 1
			if i >= num {
				return nil
			}
		}
		return nil
	})

	return kvs
}
