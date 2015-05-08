package main

import (
	"./highbatch"
	"github.com/BurntSushi/toml"
	"os"
	"bytes"
	"io/ioutil"
	"fmt"
	"runtime"
	"path/filepath"
	"github.com/kardianos/osext"
)

func main() {

	fullexecpath, err := osext.Executable()
	if err != nil {
		fmt.Println(err)
	}

	dir, _ := filepath.Split(fullexecpath)
	fmt.Println(dir)

	if err := os.Chdir(dir); err != nil {
		fmt.Println(err)
	}

	if bootCheck() && cmdCheck(fullexecpath) {
		highbatch.ServiceInit()
	}

}

func cmdCheck(path string) bool {

	if runtime.GOOS == "windows" {

		if _, err := os.Stat("command"); err != nil {
			if err := os.Mkdir("command", 0600); err != nil {
				fmt.Println(err)
				return false
			}
		}

		installCmdFile := "command/win_service_install.bat"
		if _, err := os.Stat(installCmdFile); err != nil {
			cmd := path + " install"
			if err := ioutil.WriteFile(installCmdFile, []byte(cmd), 0644); err != nil {
				fmt.Println(err)
				return false
			}
		}

		uninstallCmdFile := "command/win_service_uninstall.bat"
		if _, err := os.Stat(uninstallCmdFile); err != nil {
			cmd := path + " uninstall"
			if err := ioutil.WriteFile(uninstallCmdFile, []byte(cmd), 0644); err != nil {
				fmt.Println(err)
				return false
			}
		}
	}
	return true
}

func bootCheck() bool {
	configfile := "config.toml"

	if _, err := os.Stat(configfile); err != nil {
		var config highbatch.ClientConfigFile

		tag := []string{"test"}
		config.Client.Tag = tag
		config.Client.Master.Hostname = "localhost"
		config.Client.Master.Port = "8081"

		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		if err := encoder.Encode(config); err != nil {
			fmt.Println(err);
		}
		if err := ioutil.WriteFile(configfile, []byte(buffer.String()), 0644); err != nil {
			fmt.Println(err)
		}
		fmt.Println("create config file")
		return false
	} else {
		return true
	}
}
