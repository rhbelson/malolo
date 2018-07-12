package probes

type HeartbeatExperiment struct {
}

type HeartbeatResult struct {
}

func (pe *HeartbeatExperiment) Run() (ExperimentResult, error) {
	var result *HeartbeatResult

	result = &HeartbeatResult{}
	return result, nil
}

func (pe *HeartbeatExperiment) GetCost() int {
	return 1
}

func (pe *HeartbeatExperiment) GetTable() string {
	return "heartbeat"
}

func (pe *HeartbeatExperiment) GetVersion() string {
	return "0.0.1"
}
