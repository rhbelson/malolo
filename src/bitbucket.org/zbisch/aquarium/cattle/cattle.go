/*
   The Go client for wrangler. Ted's first nontrivial Go project.
*/
package cattle

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Holds package-level state.
type cattleConfig struct {
	host                         string
	port                         int
	useSSL                       bool
	project                      string
	asynchronousReportingChannel chan Report
}

var (
	config    cattleConfig
	reportFmt string
	client    *http.Client
)

// A report for Cattle to send to Wrangler
type Report struct {
	destinationDataset string
	data               map[string]interface{}
}

// Represents the outcome of a report attempt.
type ReportResult struct {
	Success    bool
	Err        error
	StatusCode int
	Reply      map[string]string
}

// String returns a user-readable string for printing or logging report summaries
func (rr ReportResult) String() string {
	return fmt.Sprintf("ReportResult\n"+
		"Success: %t\n"+
		"Err: %v\n"+
		"StatusCode: %v\n"+
		"Reply: %v\n", rr.Success, rr.Err, rr.StatusCode, rr.Reply)
}

// Initialize the library. These values serve as defaults for future reports.
func Init(host string, port int, useSSL bool, project string) {
	config = cattleConfig{
		host, port, useSSL, project, make(chan Report),
	}

	reportFmt = "POST %v HTTP/1.0\r\n"
	reportFmt += "Content-Length: %d\r\n"
	reportFmt += "\r\n"
	reportFmt += "%s\r\n"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

// ReportSync sends data synchronously to the server. Defining data as a
// map[string]interface{} allows us to put all data types into the same map and
// plays nicely with the underlying JSON lib.
func ReportSync(dataset string, data map[string]interface{}) (result ReportResult) {
	result.Success = false
	result.Reply = make(map[string]string)
	// This will ensure that any error ends up in the ReportResult.
	var err error = nil
	defer func() { result.Err = err }()

	// Set up the HTTP destination for this report.
	protocol := "http"
	if config.useSSL {
		protocol = "https"
	}
	resource := fmt.Sprintf("%s/%s", config.project, dataset)
	connStr := fmt.Sprintf("%s://%v:%d/%s", protocol, config.host, config.port,
		resource)

	// Prepare the report data.
	marshalled, err := json.Marshal(data)
	if err != nil {
		return
	}
	msgReader := bytes.NewReader(marshalled)
	log.WithFields(log.Fields{"msg": string(marshalled)}).Debug("Cattle: Reporting results.")

	// Perform the actual network communication.
	resp, err := client.Post(connStr, "text/plain", msgReader)
	if err != nil {
		return
	}
	result.StatusCode = resp.StatusCode
	defer resp.Body.Close()
	if result.StatusCode != 200 {
		//fmt.Println("error:", result.StatusCode, result, result.Reply)
		result.Reply["cattle"] = "womp"
	}

	// Unpack the response.
	replyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println("response:", err)
		return
	}
	err = json.Unmarshal(replyBytes, &result.Reply)
	if err != nil {
		//fmt.Println("Response:", err)
		return
	}

	// Determine end-to-end success.
	if result.StatusCode == 200 {
		if _, hasErr := result.Reply["error"]; !hasErr {
			result.Success = true
		}
	}

	return
}

// dataset: the name of the destination dataset in Wrangler
// data: the data to send
// result: the result of sending the data (in asynchronous case, trivially returns success)
func ReportAsynchronous(dataset string, data map[string]interface{}) (result ReportResult) {
	var report Report
	report = Report{dataset, data}

	config.asynchronousReportingChannel <- report

	result.Success = true
	result.Err = nil
	result.StatusCode = 0
	result.Reply = nil

	return
}

func main() {
	data := make(map[string]interface{})
	data["foo"] = "bar"
	data["somenum"] = 5
	submap := make(map[string]string)
	submap["second-level"] = "yeah!"
	data["mymap"] = submap

	Init("reports.aqualab.cs.northwestern.edu", 8088, true, "dummy")
	result := ReportSync("dummy", data)
	fmt.Println("result: ", result)
}
