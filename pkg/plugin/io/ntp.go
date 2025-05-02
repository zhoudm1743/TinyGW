package io

import (
	"github.com/robfig/cron/v3"
	"tinyGW/pkg/service/conf"
)

func InitNtp(config *conf.Config) {
	command := "ntpdate " + config.Cloud.Host
	//OsCommand(command)
	//OsCommand("hwclock --systohc")
	OsCommand(command + " && hwclock --systohc")
	c := cron.New()
	c.AddFunc("0 0 0 * * ?", func() {
		//OsCommand(command)
		//OsCommand("hwclock --systohc")
		OsCommand(command + " && hwclock --systohc")
	})
	c.Start()
}

func ExecuteNtp(config *conf.Config) {
	command := "ntpdate " + config.Cloud.Host
	OsCommand(command)
	OsCommand("hwclock --systohc")
}
