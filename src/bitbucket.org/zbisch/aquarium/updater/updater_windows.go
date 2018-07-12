package updater

import (
	"os/exec"

	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
)

func restart() bool {
	filename, err := osext.Executable()
	if err != nil {
		return false
	}
	log.WithFields(log.Fields{"exe": filename}).
		Info("Updater: Attempting restart.")
	args := []string{"--service", "restart"}
	c := exec.Command(filename, args...)
	err = c.Run()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).
			Error("Updater: Problem restarting.")
		return false
	}
	// Should not reach this return when running as a service.
	return true
}
