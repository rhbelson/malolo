/*
	Package aquarium provides a network measurement primitives and
	experiment coordination. Aquarium handles network resource
	allocation, timestamps, and application updates. As a result,
	Aquarium needs information about your application.

	Aquarium was developed by AquaLab at Northwestern University.
*/
package aquarium

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"bitbucket.org/zbisch/aquarium/app"
	"bitbucket.org/zbisch/aquarium/updater"

	"github.com/kardianos/osext"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
)

const AquariumVersion string = "0.3.0"

// An Aquarium is an instance of the Aquarium agent.
type Aquarium struct {
	Time            *TimeLord // Use to get current time
	ProbeMon        *ProbeMonitor
	Reporter        *Reporter
	Shutdown        chan bool
	AppVersion      string // Version of your application
	AquariumVersion string
	UUID            string
}

func (aqua *Aquarium) monitor() {
	<-aqua.Shutdown
	log.Info("Aquarium: Shutting down.")
	// Shut down measurement monitor
	log.Info("Aquarium: Telling Monitor to shut down.")
	aqua.ProbeMon.shutdown <- true
	<-aqua.ProbeMon.shutdown
	// Shut down Time Lord
	log.Info("Aquarium: Telling Timelord to shut down.")
	aqua.Time.shutdown <- true
	<-aqua.Time.shutdown
	close(aqua.Shutdown)
}

func (aqua *Aquarium) reportSysInfo() {
	data := make(map[string]interface{})
	data["os"] = runtime.GOOS
	data["arch"] = runtime.GOARCH
	data["client_version"] = aqua.AppVersion
	data["aquarium_version"] = aqua.AquariumVersion
	log.WithFields(log.Fields{"sysinfo": data}).
		Info("Aquarium: Reporting system info.")
	aqua.Reporter.Report("system-info", data)
}

func createNewConfig(fname string) map[string]interface{} {
	m := make(map[string]interface{})
	myUuid, err := uuid.NewV4()
	if err != nil {
		m["uuid"] = "ERROR MAKING UUID"
		return m
	}
	m["uuid"] = myUuid.String()
	saveConfig(m, fname)
	return m
}

func saveConfig(config map[string]interface{}, fname string) {
	data, err := json.Marshal(config)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(fname, data, 0644)
}

// StartAquarium creates, starts, and returns an Aquarium instance, when
// provided with a valid app.AppConfig.
func StartAquarium(program app.AppConfig) (aqua Aquarium) {
	var config map[string]interface{}
	filename := getExecRelativePath("config.json")
	log.WithFields(log.Fields{"file": filename}).
		Info("Aquarium: Selected config file.")
	configRawText, err := ioutil.ReadFile(filename)
	if err != nil {
		config = createNewConfig(filename)
		log.WithFields(log.Fields{"file": filename}).
			Debug("Aquarium: Config file not found. Creating one.")
	} else {
		config = make(map[string]interface{})
		err = json.Unmarshal(configRawText, &config)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "file": filename}).
				Error("Aquarium: Failed to parse config file. Creating one.")
			config = createNewConfig(filename)
		}
		if _, ok := config["uuid"]; !ok {
			config["uuid"] = "NO UUID IN FILE"
		}
	}
	myUuid, err := getString(config, "uuid")
	if err != nil {
		myUuid = "PROBLEM LOADING UUID"
	}
	aqua = Aquarium{nil,
		nil,
		nil,
		make(chan bool),
		program.Version,
		AquariumVersion,
		myUuid}
	aqua.Time = NewTimeLord(&aqua)
	aqua.Reporter = NewReporter(&aqua, program.Name)
	aqua.ProbeMon = newProbeMonitor(&aqua)
	go aqua.monitor()
	go aqua.reportSysInfo()
	go updater.Updater(program)
	return
}

func getString(m map[string]interface{}, key string) (string, error) {
	val, ok := m[key]
	if !ok {
		return "", errors.New("Key not found")
	}
	stdstring, ok := val.(string)
	if !ok {
		return "", errors.New("Could not convert to string")
	}
	return stdstring, nil
}

func getExecRelativePath(f string) string {
	filename, _ := osext.Executable()
	path := filepath.Join(filepath.Dir(filename), f)
	return path
}
