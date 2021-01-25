package main

import (
	"github.com/sirupsen/logrus"
	"github.com/vulcanize/gap-filler/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("exit")
	}
}
