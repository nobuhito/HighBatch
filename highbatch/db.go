package highbatch

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"gopkg.in/mgo.v2"
	"regexp"
	"sort"
	"time"
	// "gopkg.in/mgo.v2/bson"
)

type keyValue struct {
	key   string
	value string
}
type keyValues []keyValue

type workerOuts []Spec

type filter struct {
	completed string
	machine   string
	task      string
}

type MongoData struct {
	key   string
	value string
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

func store(key string, value []byte) error {
	if Conf.Server.MongoDbHost != "" {
		if err := storeMongoDb(key, value); err != nil {
			Le(err)
			return err
		}
	} else {
		if err := storeLevelDb(key, value); err != nil {
			Le(err)
			return err
		}
	}
	return nil
}

func storeMongoDb(key string, value []byte) error {
	session, err := mgo.Dial(Conf.Server.MongoDbHost)
	if err != nil {
		Le(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("Hibatch").C("log")
	if err := c.Insert(&MongoData{key: key, value: string(value)}); err != nil {
		Le(err)
	}
	return nil
}

func storeLevelDb(key string, value []byte) error {
	Ld("in Store")
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		Le(err)
	}
	defer db.Close()

	var writeOptions opt.WriteOptions
	writeOptions.Sync = true
	if err := db.Put([]byte(key), value, &writeOptions); err != nil {
		Le(err)
		return err
	}
	return nil
}

func get(f filter) workerOuts {
	Ld("in Get")
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		Le(err)
	}
	defer db.Close()

	var wos workerOuts

	reg := "^\\d{4}\\d{2}\\d{2}\\d{2}\\d{2}\\d{2}$"
	isMatch, err := regexp.MatchString(reg, f.completed)
	if err != nil {
		Le(err)
	}

	limit := []byte(time.Now().Format("20060102150405"))

	start := []byte(func() string {
		if isMatch {
			return f.completed
		} else {
			span := 90 // デフォルトは一ヶ月分
			return time.Now().AddDate(0, 0, span*-1).Format("20060102150405")
		}
	}())

	iter := db.NewIterator(&util.Range{Start: start, Limit: limit}, nil)
	for iter.Next() {

		var wo Spec
		if err := json.Unmarshal(iter.Value(), &wo); err != nil {
			Le(err)
		}

		isMatch = true

		if f.completed != "" && f.completed != wo.Completed[0:14] {
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

	sort.Sort(sort.Reverse(wos))
	return wos
}

func dump(num int) keyValues {
	Ld("in Dump")
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		Le(err)
	}
	defer db.Close()

	i := 0
	var keyValues keyValues
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		var keyValue keyValue
		keyValue.key = string(iter.Key())
		keyValue.value = string(iter.Value())
		keyValues = append(keyValues, keyValue)
		i = i + 1
		if i >= num {
			return keyValues
		}
	}
	return keyValues
}
