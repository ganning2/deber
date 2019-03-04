package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runPackage(os, dist string) error {
	logger.Info("Packaging software")

	err := dock.ExecContainer(names.Container(), "sudo", "apt-get", "update")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apt-get -y")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "dpkg-buildpackage", "-tc")
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
