package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runStop(os, dist string) error {
	logger.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if isContainerStopped {
		logger.Skip()
		return nil
	}

	err = dock.StopContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
