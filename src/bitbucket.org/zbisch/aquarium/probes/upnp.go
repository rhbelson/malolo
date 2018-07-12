package probes

import (
	"errors"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

type UPnPByteCountExperiment struct {
}

type UPnPByteCountResult struct {
	UPnPBytesDownloaded uint64
	UPnPBytesUploaded   uint64
}

func (ubce *UPnPByteCountExperiment) Run() (ExperimentResult, error) {
	result, err := ubce.tryGateway2()
	if err != nil {
		result, err := ubce.tryGateway1()
		return result, err
	} else {
		return result, err
	}
}

func (ubce *UPnPByteCountExperiment) GetCost() int {
	return 1
}

func (ubce *UPnPByteCountExperiment) GetTable() string {
	return "upnpcounters"
}

func (tre *UPnPByteCountExperiment) GetVersion() string {
	return "0.0.1"
}

func (ubce *UPnPByteCountExperiment) tryGateway1() (ExperimentResult, error) {
	if clients, _, err := internetgateway1.NewWANCommonInterfaceConfig1Clients(); err != nil {
		return nil, errors.New("Search for gateways failed.")
	} else {
		for _, client := range clients {
			bytesReceived, err := client.GetTotalBytesReceived()
			if err != nil {
				continue
			}
			bytesSent, err := client.GetTotalBytesSent()
			if err != nil {
				continue
			}
			result := UPnPByteCountResult{uint64(bytesReceived), uint64(bytesSent)}
			return result, nil
		}
		return nil, errors.New("Valid byte counters not found.")
	}
}

func (ubce *UPnPByteCountExperiment) tryGateway2() (ExperimentResult, error) {
	if clients, _, err := internetgateway2.NewWANCommonInterfaceConfig1Clients(); err != nil {
		return nil, errors.New("Search for gateways failed.")
	} else {
		for _, client := range clients {
			bytesReceived, err := client.GetTotalBytesReceived()
			if err != nil {
				continue
			}
			bytesSent, err := client.GetTotalBytesSent()
			if err != nil {
				continue
			}
			result := UPnPByteCountResult{uint64(bytesReceived), uint64(bytesSent)}
			return result, nil
		}
		return nil, errors.New("Valid byte counters not found.")
	}
}
