package aquarium

import (
	"math/rand"
	"sync"
	"time"

	"bitbucket.org/zbisch/aquarium/probes"

	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
)

type semaphore struct {
	s     chan bool
	mutex sync.Mutex
}

func (sema semaphore) P(n int) {
	sema.mutex.Lock()
	for i := 0; i < n; i++ {
		sema.s <- true
	}
	sema.mutex.Unlock()
}

func (sema semaphore) V(n int) {
	for i := 0; i < n; i++ {
		<-sema.s
	}
}

type ProbeMonitor struct {
	aqua             *Aquarium
	semaphore        semaphore
	SubmitExperiment chan probes.Experiment
	shutdown         chan bool
}

func newProbeMonitor(aqua *Aquarium) *ProbeMonitor {
	rand.Seed(time.Now().UTC().UnixNano())
	pm := ProbeMonitor{
		aqua,
		semaphore{make(chan bool, 10),
			sync.Mutex{}},
		make(chan probes.Experiment),
		make(chan bool)}
	go pm.monitor()
	return &pm
}

func (pm *ProbeMonitor) monitor() {
	for quit := false; !quit; {
		select {
		case <-pm.shutdown:
			log.Info("ProbeMon: Received shutdown signal.")
			quit = true
			close(pm.shutdown)
			break
		case exp := <-pm.SubmitExperiment:
			log.WithFields(log.Fields{"experiment": exp}).
				Info("ProbeMon: Running experiment")
			go pm.LaunchExperiment(exp)
			break
		}
	}
}

func (pm *ProbeMonitor) LaunchExperiment(exp probes.Experiment) {
	taskCost := exp.GetCost()
	table := exp.GetTable()
	pm.semaphore.P(taskCost)
	defer pm.semaphore.V(taskCost)
	timeLookup := pm.aqua.Time.GetTime()
	log.WithFields(log.Fields{"experiment": exp}).
		Info("ProbeMonitor: Launching measurment")
	results, err := exp.Run()
	if err != nil {
		return
	}
	if results != nil {
		m := structs.Map(results)
		pm.aqua.Time.AddTime(m, timeLookup)
		m["probe_version"] = exp.GetVersion()
		log.WithFields(log.Fields{"record": m, "table": table}).Info("Reporting record.")
		response := pm.aqua.Reporter.Report(table, m)
		log.WithFields(log.Fields{"response": response}).Info("Got response.")
	}
}
