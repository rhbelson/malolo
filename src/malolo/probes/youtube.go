package probes

import (
	"time"

	"github.com/pkg/browser"
	"github.com/toqueteos/webbrowser"
)

type YoutubeExperiment struct {
}

type YoutubeResult struct {
}

func (yte *YoutubeExperiment) Run() (ExperimentResult, error) {
	url := "http://127.0.0.1:42000/auto/youtube/index.html"
	err := browser.OpenURL(url)
	if err != nil {
		webbrowser.Open(url)
	}
	<-time.After(time.Minute * 10)
	return nil, nil
}

func (yte *YoutubeExperiment) GetCost() int {
	return 9
}

func (yte *YoutubeExperiment) GetTable() string {
	return "youtube"
}

func (ubce *YoutubeExperiment) GetVersion() string {
	return "0.0.1"
}
