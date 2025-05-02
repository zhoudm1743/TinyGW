package io

import (
	"github.com/robfig/cron/v3"
	"zsxagw/core/config"
)

func InitNtp(config *config.Config) {
	command := "ntpdate " + config.Gateway.Ip
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

func ExecuteNtp(config *config.Config) {
	command := "ntpdate " + config.Gateway.Ip
	OsCommand(command)
	OsCommand("hwclock --systohc")
}
