package aquarium

import (
	"math/rand"
	"time"

	"github.com/beevik/ntp"
	log "github.com/sirupsen/logrus"
)

type TimeResponse struct {
	Valid bool
	Time  time.Time
}

type TimeLord struct {
	aqua             *Aquarium
	timeQueryChan    chan bool
	timeResponseChan chan TimeResponse
	shutdown         chan bool
	validClock       bool
	clockOffset      time.Duration
	lastUpdate       time.Time
	ttl              time.Duration
}

var ntpServers [4]string = [...]string{
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"3.pool.ntp.org",
	"4.pool.ntp.org"}

func NewTimeLord(aqua *Aquarium) *TimeLord {
	tl := TimeLord{
		aqua:             aqua,
		timeQueryChan:    make(chan bool),
		timeResponseChan: make(chan TimeResponse),
		shutdown:         make(chan bool),
		validClock:       false,
		clockOffset:      0 * time.Second,
		lastUpdate:       time.Unix(0, 0),
		ttl:              1024 * time.Second,
	}

	go tl.lordTime()
	return &tl
}

func pickRandomNtpServer() string {
	return ntpServers[rand.Intn(len(ntpServers))]
}

func (tl *TimeLord) checkOffset() {
	if tl.lastUpdate.Add(tl.ttl).Before(time.Now()) {
		log.WithFields(log.Fields{"lastupdate": tl.lastUpdate}).Info(
			"Timelord: Updating NTP clock.")
		ntpServer := pickRandomNtpServer()
		resp, err := ntp.Query(ntpServer)
		if err != nil {
			return
		}
		tl.clockOffset = resp.ClockOffset
		tl.validClock = true
		tl.lastUpdate = time.Now()
	}
}

func (tl *TimeLord) lordTime() {
	log.Info("Timelord: Starting.")
	quit := false
	tl.checkOffset()
	for !quit {
		select {
		case <-tl.timeQueryChan:
			tl.checkOffset()
			tl.timeResponseChan <- TimeResponse{tl.validClock, time.Now().UTC().Add(tl.clockOffset)}
		case <-tl.shutdown:
			quit = true
			log.Info("Timelord: Shutting down.")
			close(tl.shutdown)
			break
		}
	}
}

func (tl *TimeLord) GetTime() TimeResponse {
	tl.timeQueryChan <- true
	return <-tl.timeResponseChan
}
func (tl *TimeLord) AddTime(d map[string]interface{}, response TimeResponse) {
	d["utc_timestamp"] = response.Time
	d["ntp_sync"] = response.Valid
	tzname, tzoffset := time.Now().Zone()
	d["tz_name"] = tzname
	d["tz_offset"] = tzoffset
}
