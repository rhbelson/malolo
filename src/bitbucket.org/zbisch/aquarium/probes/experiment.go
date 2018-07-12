package probes

type ExperimentResult interface {
}

type Experiment interface {
	Run() (ExperimentResult, error)
	GetCost() int
	GetTable() string
	GetVersion() string
}

type ReportableError interface {
	GetErrorTable() string
}

type errorReport struct {
	table    string
	errorMsg string
}

func (e *errorReport) Error() string {
	return e.errorMsg
}

func (e *errorReport) GetErrorTable() string {
	return e.table
}
