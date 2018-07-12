package updater

import (
	"os"
	"time"

	"bitbucket.org/zbisch/aquarium/app"

	"github.com/kardianos/osext"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	log "github.com/sirupsen/logrus"
)

var sigs chan os.Signal

type program struct{}

func Updater(app app.AppConfig) {
	var updater = &selfupdate.Updater{
		CurrentVersion: app.Version,
		ApiURL:         app.ApiURL,
		BinURL:         app.BinURL,
		DiffURL:        app.DiffURL,
		CmdName:        app.Name,
		ForceCheck:     true,
		Dir:            app.UpdateDir,
	}
	for {
		<-time.After(time.Minute * 10)
		filename, err := osext.Executable()
		if err != nil {
			continue
		}
		infoBefore, err := os.Stat(filename)
		if err != nil {
			continue
		}
		err = updater.BackgroundRun()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Error("Updater: Update failed")
			continue
		}
		infoAfter, err := os.Stat(filename)
		if err != nil {
			continue
		}
		log.WithFields(log.Fields{"lastmod": infoBefore.ModTime(),
			"currmod": infoAfter.ModTime()}).
			Info("Updater: Ran update.")
		if infoBefore.ModTime() != infoAfter.ModTime() {
			success := restart()
			if !success {
				log.Error("Updater: Restart failed.")
			}
		}
		<-time.After(time.Minute * 50)
	}
}
