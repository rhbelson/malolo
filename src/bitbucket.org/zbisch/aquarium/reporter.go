package aquarium

import (
	"bitbucket.org/zbisch/aquarium/cattle"

	log "github.com/sirupsen/logrus"
)

type Reporter struct {
	aqua *Aquarium
}

func NewReporter(aqua *Aquarium, dbname string) *Reporter {
	log.Info("Reporter: Starting up.")
	cattle.Init("reports.aqualab.cs.northwestern.edu", 8088, true, dbname)
	return &Reporter{aqua}
}

func (r *Reporter) Report(table string, data map[string]interface{}) (response cattle.ReportResult) {
	//marshalled, _ := json.Marshal(results)
	if data == nil {
		return cattle.ReportResult{Success: true}
	} else {
		data["uuid"] = r.aqua.UUID
		response = cattle.ReportSync(table, data)
		return response
	}
}
