package conf

import (
	"io/ioutil"
	"log"
	"os/user"

	"gopkg.in/ini.v1"
)

func confPath() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	files, err := ioutil.ReadDir(u.HomeDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.Name() == ".config" && file.IsDir() {
			configDir, err := ioutil.ReadDir(u.HomeDir + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}
			for _, config := range configDir {
				if config.Name() == "ctbot.ini" {
					return u.HomeDir + "/" + file.Name() + "/" + config.Name()
				}
			}
		}
	}
	log.Fatalf("Can't found config %s/.config/ctbot.ini", u.HomeDir)
	return ""
}

func ReadToken() string {
	cfg, err := ini.Load(confPath())
	if err != nil {
		log.Fatal(err)
	}
	return cfg.Section("bot").Key("token").String()
}

func ReadUser() *ini.Section {
	cfg, err := ini.Load(confPath())
	if err != nil {
		log.Fatal(err)
	}
	return cfg.Section("user")
}

func ReadUsrInfo() (int, int64) {
	cfg := ReadUser()
	channel, err := cfg.Key("channel").Int64()
	if err != nil {
		log.Fatal(err)
	}
	admin, err := cfg.Key("admin").Int()
	if err != nil {
		log.Fatal(err)
	}
	return admin, channel
}
