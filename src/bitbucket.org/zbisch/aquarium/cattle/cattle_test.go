package cattle

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MakeDummyData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["foo"] = "bar"
	data["somenum"] = 5
	submap := make(map[string]string)
	submap["second-level"] = "yeah!"
	data["mymap"] = submap

	return data
}

func TestWhenCallingInit_GlobalVariablesAreSetCorrectly(t *testing.T) {
	host := "testHost.com"
	port := 1234
	useSSL := true
	project := "dummy"

	Init(host, port, useSSL, project)

	assert.Equal(t, host, config.host, "config.host should be set to: %v", host)
	assert.Equal(t, port, config.port, "config.port should be set to: %v", port)
	assert.Equal(t, useSSL, config.useSSL, "config.useSSL should be set to: %v", useSSL)
	assert.Equal(t, project, config.project, "config.project should be set to: %v", project)
	assert.Equal(t, false, client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify,
		"InsecureSkipVerify should be: %v", false)

	//fmt.Println(client.Transport)
}

// TODO Init with invalid port numbers should return error

func TestWhenCallingReportSync_ReturnsReportResultSuccessfully(t *testing.T) {
	data := MakeDummyData()

	// initialize global variables
	Init("reports.aqualab.cs.northwestern.edu", 8088, true, "dummy")
	result := ReportSync("dummy", data)

	assert.NotNil(t, result, "ReportResult should not be nil")
	assert.True(t, result.Success, "ReportResult.Success should be true")
	assert.Nil(t, result.Err, "ReportResult.Err should be nil")
	assert.Equal(t, 200, result.StatusCode, "ReportResult.StatusCode should be 200")
	assert.Equal(t, make(map[string]string), result.Reply, "ReportResult.Reply should be empty")
}

func TestWhenCallingReportSyncFails_DataIsWrittenToTempFile(t *testing.T) {
	data := MakeDummyData()

	// Init with invalid project to cause ReportSync to fail
	Init("reports.aqualab.cs.northwestern.edu", 8088, true, "not-dummy")
	result := ReportSync("dummy", data)

	assert.NotNil(t, result, "ReportResult should not be nil")
	assert.False(t, result.Success, "ReportResult.Success should be false")
	assert.NotNil(t, result.Err, "ReportResult.Err should not be nil")
	assert.NotEqual(t, 200, result.StatusCode, "ReportResult.StatusCode should not be 200")
	assert.Equal(t, "womp", result.Reply["cattle"],
		"ReportResult.Reply should contain {'cattle': 'womp'}")

	// TODO add check that data was written to temp file

	Init("reports.aqualab.cs.northwestern.edu", 8088, true, "dummy")
	result = ReportSync("not-dummy", data)

	assert.NotNil(t, result, "ReportResult should not be nil")
	assert.False(t, result.Success, "ReportResult.Success should be false")
	assert.NotNil(t, result.Err, "ReportResult.Err should not be nil")
	assert.NotEqual(t, 200, result.StatusCode, "ReportResult.StatusCode should not be 200")
	assert.Equal(t, "womp", result.Reply["cattle"],
		"ReportResult.Reply should contain {'cattle': 'womp'}")

	// TODO add check that data was written to temp file
}

func TestWhenCallingReportSyncToInvalidHost_DataIsWrittenToTempFile(t *testing.T) {
	data := MakeDummyData()

	// Init with invalid project to cause ReportSync to fail
	Init("report1s.aqualab.cs.northwestern.edu", 8088, true, "dummy")
	result := ReportSync("dummy", data)

	fmt.Println(result)
	assert.NotNil(t, result, "ReportResult should not be nil")
	assert.False(t, result.Success, "ReportResult.Success should be false")
	assert.NotNil(t, result.Err, "ReportResult.Err should not be nil")
	assert.NotEqual(t, 200, result.StatusCode, "ReportResult.StatusCode should not be 200")

	// TODO add check that data was written to temp file
}
