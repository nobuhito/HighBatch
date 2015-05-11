package highbatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func startArranger() {
	ld("in startArranger")
	for {
		c := cron.New()

		specs := taskFileSerch()

		for i := range specs {
			spec := specs[i]
			if spec.Schedule != "" {
				c.AddFunc(spec.Schedule, func() {
					taskKick(spec)
				})
			}
		}
		c.Start()
		time.Sleep(10 * 60 * time.Second)
		c.Stop()
		c = nil
	}
}

func taskChain(spec Spec) {
	if spec.OnErrorStop == "" || spec.ExitCode == 0 {
		for i := range spec.Chain {
			file := strings.Join([]string{"tasks", spec.Chain[i], "spec.toml"}, string(os.PathSeparator))
			chainSpec := parseSpec(file)
			chainSpec.Route = spec.Route
			chainSpec.Route = append(chainSpec.Route, spec.Name)
			taskKick(chainSpec)
		}
	}
}

func taskKick(spec Spec) {
	if spec, err := sendWorker(spec); err != nil {
		spec.ExitCode = 99
		spec.Output = err.Error()
		spec.Hostname = "unknown"
		spec.Completed = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))

		go write(spec)
	}
}

func sendCheck(spec Spec) (WorkerInfo, error) {
	var w WorkerInfo
	err := json.Unmarshal([]byte(get("workers", spec.Hostname)), &w)

	if err != nil {
		err = errors.New("invalid port number\n" + err.Error())
		return w, err
	}

	if time.Since(w.Datetime).Minutes() > 5 {
		duration := time.Since(w.Datetime).String()
		err := errors.New("No keep-alive sent from client in over " + duration)
		return w, err
	}

	if w.Port == "" {
		err := errors.New("invalid port number.\n port: " + w.Port)
		return w, err
	}

	return w, nil
}

func sendWorker(spec Spec) (Spec, error) {
	ld("in sendWorkder")
	spec.Started = fmt.Sprint(time.Now().Format("20060102150405"), rand.Intn(9))
	client := &http.Client{}
	for i := range spec.Machine {

		spec.Hostname = spec.Machine[i]
		spec.Id = spec.Started + "_" + spec.Hostname + "_" + spec.Key

		w, err := sendCheck(spec)
		if err != nil {
			return spec, err
		}

		err = writeDB(spec)
		if err != nil {
			return spec, err
		}

		m, _ := json.Marshal(spec)

		req, err := http.NewRequest(
			"POST",
			"http://"+spec.Hostname+":"+w.Port+"/worker",
			bytes.NewBuffer(m),
		)
		if err != nil {
			return spec, err
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			return spec, err
		}

		if resp.StatusCode != 200 {
			err := errors.New("HTTP Status code error")
			return spec, err
		}
	}
	return spec, nil
}
