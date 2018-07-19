/*
	Malolo provides a simple Aquarium app that pings google every 30 seconds.
*/

package main

//go:generate go get -t malolo
//go:generate go get -u github.com/GeertJohan/go.rice
//go:generate go get -u github.com/GeertJohan/go.rice/rice
//go:generate $GOPATH/bin/rice embed-go

import (
	"flag"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"bitbucket.org/zbisch/aquarium"
	"bitbucket.org/zbisch/aquarium/app"
	"bitbucket.org/zbisch/aquarium/probes"

	// "github.com/surol/speedtest-cli"
	"github.com/GeertJohan/go.rice"
	"github.com/kardianos/osext"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// App settings used by Aquarium.
// Settings used for reporting data and automatic checking for updates.
//
// NOTE: Data reporting will not work until you create a valid project on the
// data reporting server.
var appConfig app.AppConfig = app.AppConfig{
	Name:        "malolo",
	DisplayName: "Malolo",
	Version:     "0.0.1",
	ApiURL:      "https://aquarium.aqualab.cs.northwestern.edu/",
	BinURL:      "https://aquarium.aqualab.cs.northwestern.edu/",
	DiffURL:     "https://aquarium.aqualab.cs.northwestern.edu/",
	UpdateDir:   "update/",
	Description: "Inflight network monitor.",
}

// Settings for registering application with OS service monitor.
var config service.Config = service.Config{
	Name:        appConfig.Name,
	DisplayName: appConfig.DisplayName,
	Description: appConfig.Description}

// systemLogger is used for logging service info to OS.
// Use logrus for non-service-specific messages.
var systemLogger service.Logger

// shutdown_signal_chan is the channel for incoming control signals.
var shutdown_signal_chan chan os.Signal

var dataMutex = &sync.Mutex{}
var data map[string]string

func init() {
	// init is called before main when starting the program.
	//
	// This template configures the log settings in this function.
	//
	// Optional example: Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// You can also use lumberjack to set up rollling logs. Can either output
	// logs to stdout or lumberjack. Pick one, comment out the other.
	exe, err := osext.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := path.Dir(exe)
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(exeDir, appConfig.Name+".log"),
		MaxSize:    1, // In megabytes.
		MaxBackups: 3,
		MaxAge:     3, // In days (0 means keep forever?)
	}
	log.SetOutput(lumberjackLogger)
	// Output to stdout instead of the default stderr.
	// Note: can be any io.Writer.
	log.SetOutput(os.Stdout)

	// Only log the Info level or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	// Configure and parse the command line flag options
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	// Set up the application to be a system service
	// May not be required by all applications
	prog := &Program{}
	s, err := service.New(prog, &config)
	if err != nil {
		log.Fatal(err)
	}
	systemLogger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// If service flag set, configure service
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	// If no flags set, run application.
	// Run() will not return unless there is an error.
	err = s.Run()
	if err != nil {
		systemLogger.Error(err)
	}
}

// program struct and methods match interface for registering application as
// a system service.
type Program struct{}

// Start is called by system's service monitor.
// Calls to Start should not block (i.e., must return immediately).
// Do any work asynchronously.
func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}
func handler(w http.ResponseWriter, r *http.Request) {
	dataMutex.Lock()
	w.Write([]byte(data["ping"]))
	dataMutex.Unlock()
}
// run is where you start your actual program. This function should block while
// the application is running and return upon termination or completion.
func (p *Program) run() {
	log.WithFields(log.Fields{"app": appConfig.Name}).Info("Starting app.")
	rand.Seed(time.Now().UTC().UnixNano())

	aqua := aquarium.StartAquarium(appConfig)

	go pingLauncher(aqua)
	go tracerouteLauncher(aqua)
	go speedtestLauncher(aqua)
	data = make(map[string]string)
	http.Handle("/", http.FileServer(rice.MustFindBox("www").HTTPBox()))
	http.HandleFunc("/dynamic", handler)
	//http.ListenAndServe(":8080", nil)
	log.Println("UI Server: Starting up!")
	go http.ListenAndServe(":42000", nil)

	shutdown_signal_chan = make(chan os.Signal)
	signal.Notify(shutdown_signal_chan, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown_signal_chan
	aqua.Shutdown <- true
	<-aqua.Shutdown
}

// Stop is called when the OS service monitor wants to kill your application.
// Stop should not block. Return within a few seconds.
func (p *Program) Stop(s service.Service) error {
	go func() {
		for shutdown_signal_chan == nil {
			// Make sure that shutdown channel has been initilized.
		}
		shutdown_signal_chan <- syscall.SIGINT
	}()
	return nil
}

// pingLauncher is a simple example of a goroutine that launches a ping
// measurement ever 10 + rand * 5 seconds.

func pingLauncher(aqua aquarium.Aquarium) {
	for {
		select {
		case <-time.After(10*time.Second + time.Second*time.Duration(rand.Int63n(5))):
			result := make(chan *probes.PingResult)
			aqua.ProbeMon.SubmitExperiment <- &probes.PingExperiment{"google.com", result}
			temp := (<-result).Stdout
			dataMutex.Lock()
			data["ping"] = temp
			dataMutex.Unlock()
		}
	}
}


func tracerouteLauncher(aqua aquarium.Aquarium) {
	for {
		select {
		case <-time.After(10*time.Second + time.Second*time.Duration(rand.Int63n(5))):
			result := make(chan *probes.TracerouteResult)
			aqua.ProbeMon.SubmitExperiment <- &probes.TracerouteExperiment{"google.com", result}
			temp := (<-result).Stdout
			dataMutex.Lock()
			data["ping"] = temp
			dataMutex.Unlock()
		}
	}
}

func speedtestLauncher(aqua aquarium.Aquarium) {
	for {
		select {
		case <-time.After(20*time.Second + time.Second*time.Duration(rand.Int63n(15))):
			result := make(chan *probes.SpeedtestResult)
			aqua.ProbeMon.SubmitExperiment <- &probes.SpeedtestExperiment{"google.com", result}
			temp := (<-result).Stdout
			dataMutex.Lock()
			data["ping"] = temp
			dataMutex.Unlock()
		}
	}
}


