package highbatch

import (
	"github.com/boltdb/bolt"
	"encoding/json"
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
	start     string
	machine   string
	task      string
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
const bucketname = "hibhbatch"

func initdb() {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		le(err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketname))
		return err
	}); err != nil {
		le(err)
	}
}

func store(key string, value []byte) error {
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

func getOne(id string) Spec {
	ld("in GetOne")
	var spec Spec
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		le(err)
		}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		v := b.Get([]byte(id))
		if err := json.Unmarshal(v, &spec); err != nil {
			le(err)
		}
		return nil
	})

	return spec
}

func getLists(f filter) workerOuts {
	ld("in GetLists")
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

	span := 30 										// 初期値は30日分
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

func dump(num int) (kvs KeyValues) {
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
				Key: string(k),
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
