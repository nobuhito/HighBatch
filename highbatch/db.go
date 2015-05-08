package highbatch

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"regexp"
	"sort"
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
	db, err := bolt.Open(dbname, 0600, nil)
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
	db, err := bolt.Open(dbname, 0600, nil)
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
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		le(err)
	}
	defer db.Close()

	reg := "^\\d{4}\\d{2}\\d{2}\\d{2}\\d{2}\\d{2}$"
	isMatch, err := regexp.MatchString(reg, f.start)
	if err != nil {
		le(err)
	}

	span := 30 // 初期値は30日分
	since := time.Now().AddDate(0, 0, span*-1).Format("20060102150405")

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketname)).Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var wo Spec
			if err := json.Unmarshal(v, &wo); err != nil {
				le(err)
			}

			isMatch = true

			if f.start == "" && wo.Started[0:14] < since {
				break
			}

			if f.start != "" && f.start != wo.Started[0:14] {
				isMatch = false
			}

			if f.machine != "" && f.machine != wo.Hostname {
				isMatch = false
			}

			if f.task != "" && f.task != wo.Key {
				isMatch = false
			}

			if isMatch {
				wos = append(wos, wo)
			}

		}
		return nil
	})

	sort.Sort(sort.Reverse(wos))
	return wos
}

func dump(bucketname string, num int) (kvs KeyValues) {
	ld("in Dump")
	db, err := bolt.Open(dbname, 0600, nil)
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
