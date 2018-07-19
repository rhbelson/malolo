package probes


import (
	"bytes"
	"os/exec"
	"time"

	"github.com/surol/speedtest-cli/speedtest"
	"fmt"
	"os"
	"flag"
	"log"
	"time"
)



type NdtExperiment struct {
	Target     string
	ResultChan chan *NdtResult
}

type NdttestResult struct {
	Command string
	Target  string
	Args    []string
	Stdout  string
	Stderr  string
}

func (se *NdtExperiment) Run() (ExperimentResult, error) {
	var cmd string
	var args []string
	var err error
	var result *NdtResult

	
	result=main()
	pe.sendResponse(result)
	return result, err
}

func (se *NdtExperiment) GetCost() int {
	return 1
}

func (pe *NdtExperiment) GetTable() string {
	return "ping"
}

func (se *NdtExperiment) sendResponse(result *NdtResult) {
	if pe.ResultChan != nil {
		select {
		case pe.ResultChan <- result:
			break
		case <-time.After(time.Minute):
			break
		}
	}
}

func (se *NdtExperiment) GetVersion() string {
	return "0.0.1"
}









func version() {
	fmt.Print(speedtest.Version)
}

func usage() {
	fmt.Fprint(os.Stderr, "Command line interface for testing internet bandwidth using speedtest.net.\n\n")
	flag.PrintDefaults()
}

func main() {
	opts := speedtest.ParseOpts()

	switch {
	case opts.Help:
		usage()
		return
	case opts.Version:
		version()
		return
	}

	client := speedtest.NewClient(opts)

	if opts.List {
		servers, err := client.AllServers()
		if err != nil {
			log.Fatalf("Failed to load server list: %v\n", err)
		}
		fmt.Println(servers)
		return
	}

	config, err := client.Config()
	if err != nil {
		log.Fatal(err)
	}

	client.Log("Testing from %s (%s)...\n", config.Client.ISP, config.Client.IP)

	server := selectServer(opts, client);

	downloadSpeed := server.DownloadSpeed()
	reportSpeed(opts, "Download", downloadSpeed)

	uploadSpeed := server.UploadSpeed()
	reportSpeed(opts, "Upload", uploadSpeed)
}

func reportSpeed(opts *speedtest.Opts, prefix string, speed int) {
	if opts.SpeedInBytes {
		fmt.Printf("%s: %.2f MiB/s\n", prefix, float64(speed) / (1 << 20))
	} else {
		fmt.Printf("%s: %.2f Mib/s\n", prefix, float64(speed) / (1 << 17))
	}
}

func selectServer(opts *speedtest.Opts, client *speedtest.Client) (selected *speedtest.Server) {
	if opts.Server != 0 {
		servers, err := client.AllServers()
		if err != nil {
			log.Fatal("Failed to load server list: %v\n", err)
			return nil
		}
		selected = servers.Find(opts.Server)
		if selected == nil {
			log.Fatalf("Server not found: %d\n", opts.Server)
			return nil
		}
		selected.MeasureLatency(speedtest.DefaultLatencyMeasureTimes, speedtest.DefaultErrorLatency)
	} else {
		servers, err := client.ClosestServers()
		if err != nil {
			log.Fatal("Failed to load server list: %v\n", err)
			return nil
		}
		selected = servers.MeasureLatencies(
			speedtest.DefaultLatencyMeasureTimes,
			speedtest.DefaultErrorLatency).First()
	}

	if opts.Quiet {
		log.Printf("Ping: %d ms\n", selected.Latency / time.Millisecond)
	} else {
		client.Log("Hosted by %s (%s) [%.2f km]: %d ms\n",
			selected.Sponsor,
			selected.Name,
			selected.Distance,
			selected.Latency / time.Millisecond)
	}

	return selected
}


