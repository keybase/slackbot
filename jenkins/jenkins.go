// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const jenkinsURL = "https://ci.keyba.se"
const jenkinsJobName = "windows-installer"

type CrumbResult struct {
	Crumb string
	Err   error
	Time  time.Time
}

var lastCrumb CrumbResult
var crumbGetChan chan bool
var crumbResultChan chan CrumbResult

// Make a worker gofunc to cache the Jenkins crumb for
// a certain amount of time
func init() {
	crumbGetChan = make(chan bool)
	crumbResultChan = make(chan CrumbResult)
	go func() {
		for {
			<-crumbGetChan
			if time.Since(lastCrumb.Time) > time.Hour || lastCrumb.Err != nil || lastCrumb.Crumb == "" {
				lastCrumb.Crumb, lastCrumb.Err = getJenkinsCrumb()
				if lastCrumb.Err == nil {
					lastCrumb.Time = time.Now()
				}
			}
			crumbResultChan <- lastCrumb
		}
	}()
}

// SetLastCrumb is for testing
func SetLastCrumb(crumb CrumbResult) {
	lastCrumb = crumb
}

// GetLastCrumb is for testing
func GetLastCrumb() CrumbResult {
	return lastCrumb
}

func parseQueueNumber(locationString string) string {
	countIndex := strings.Index(locationString, "queue/item/")
	if countIndex == -1 {
		log.Printf("bad location format: %s", locationString)
		return ""
	}
	countIndex += len("queue/item/")
	return strings.TrimRight(locationString[countIndex:], "/")
}

func getJenkinsCrumb() (string, error) {
	name := os.Getenv("JENKINS_WINDOWS_USERNAME")
	password := os.Getenv("JENKINS_WINDOWS_PASSWORD")
	if name == "" || password == "" {
		err := errors.New("Jenkins Windows username and password required")
		return err.Error(), err
	}
	u, err := url.Parse(jenkinsURL + "/crumbIssuer/api/json")
	if err != nil {
		return err.Error(), err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err.Error(), err
	}
	req.SetBasicAuth(name, password)
	res, err := client.Do(req)
	defer func() { _ = res.Body.Close() }()
	if err != nil {
		return err.Error(), err
	}
	response, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err.Error(), err
	}
	if res.StatusCode < 200 || res.StatusCode > 201 {
		err = fmt.Errorf("Build request returned %d", res.StatusCode)
		return err.Error(), err
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return err.Error(), err
	}
	crumb := responseMap["crumb"]
	if crumb == nil || crumb.(string) == "" {
		err = errors.New("no crumb in response")
		return err.Error(), err
	}
	return crumb.(string), nil
}

func doJenkinsPost(buildurl string) (*http.Response, error) {
	log.Printf("Posting: %s \n", buildurl)

	crumbGetChan <- true
	crumb := <-crumbResultChan
	if crumb.Err != nil {
		return nil, crumb.Err
	}
	name := os.Getenv("JENKINS_WINDOWS_USERNAME")
	password := os.Getenv("JENKINS_WINDOWS_PASSWORD")
	if name == "" || password == "" {
		return nil, errors.New("Jenkins Windows username and password required")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", buildurl, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(name, password)
	req.Header.Add("Jenkins-Crumb", crumb.Crumb)
	return client.Do(req)
}

// StartBuild starts the build and returns the queue number as a string from Jenkins.
// Also enables/resumes auto building.
func StartBuild(clientRev string, kbfsRev string, updateChannel string) (string, error) {
	u, err := url.Parse(jenkinsURL + "/job/" + jenkinsJobName + "/buildWithParameters")
	if err != nil {
		return err.Error(), err
	}
	urlValues := url.Values{}
	urlValues.Add("SlackBot", "true")
	if clientRev != "" {
		urlValues.Add("ClientRevision", clientRev)
	}
	if kbfsRev != "" {
		urlValues.Add("KBFSRevision", kbfsRev)
	}
	if updateChannel != "" {
		urlValues.Add("UpdateChannel", updateChannel)
	}
	u.RawQuery = urlValues.Encode()
	buildurl := u.String()

	res, err := doJenkinsPost(buildurl)
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}

	if err != nil {
		return err.Error(), err
	}
	robots, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err.Error(), err
	}
	if res.StatusCode != 201 {
		err = fmt.Errorf("Build request returned %d", res.StatusCode)
		return err.Error(), err
	}
	loc, _ := res.Location()
	log.Printf("StartBuild robots: %v", robots)

	err = controlJob(true)

	return fmt.Sprintf("Requested Jenkins build with queue ID %s", parseQueueNumber(loc.String())), err
}

func stopBuildByID(buildID string) error {
	log.Printf("Stopping build %s \n", buildID)
	// Of the form: http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	res, err := doJenkinsPost(jenkinsURL + "/job/" + jenkinsJobName + "/" + buildID + "/stop")
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err == nil {
		log.Printf("Status code: %v \n", res.StatusCode)
	}
	return err
}

func getJSON(url string, target interface{}) error {
	fmt.Printf("posting to: %s\n", url)
	r, err := doJenkinsPost(url)
	if err != nil {
		return err
	}
	defer func() { _ = r.Body.Close() }()

	return json.NewDecoder(r.Body).Decode(target)
}

func cancelQueueItem(queueEntry string) error {
	log.Printf("Looking to cancel queue item: %s\n", queueEntry)
	// Try removing from queue, even though it's probably not in there
	u, err := url.Parse(jenkinsURL + "/queue/cancelItem")
	if err != nil {
		return err
	}
	urlValues := url.Values{}
	urlValues.Add("id", queueEntry)
	u.RawQuery = urlValues.Encode()
	cancelURL := u.String()

	res, err := doJenkinsPost(cancelURL)

	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err == nil {
		fmt.Printf("Status code: %v \n", res.StatusCode)
	}
	return err
}

func getLastBuildAndQueueNumbers() (currentBuild int, currentQ int, err error) {
	var f interface{}
	err = getJSON(jenkinsURL+"/job/"+jenkinsJobName+"/lastBuild/api/json", &f)
	if err != nil {
		log.Printf("No lastBuild - %v\n", err)
		return
	}
	m := f.(map[string]interface{})
	// default currentQ to zero if not present
	if _, ok := m["queueId"]; !ok {
		err = fmt.Errorf("No queueId in current build --- %v", m)
		return
	}

	if _, ok := m["number"]; !ok {
		err = fmt.Errorf("No number in current build --- %v", m)
		return
	}

	currentBuild = int(m["number"].(float64))
	currentQ = int(m["queueId"].(float64))
	return
}

func controlJob(enable bool) error {
	command := "disable"
	if enable {
		command = "enable"
	}
	log.Printf("%s job\n", command)
	// Of the form: http://<Jenkins_URL>/job/<Job_Name>/disable
	_, err := doJenkinsPost(jenkinsURL + "/job/" + jenkinsJobName + "/" + command)
	return err
}

// StopBuild will disable auto builds and
// stop/cancel the current build, if specified.
// Location is the return string from a build request, e.g.:
// http://192.168.1.10:8080/queue/item/34/
func StopBuild(queueEntry string) error {

	err := controlJob(false)

	if queueEntry == "" {
		return err
	}

	// If the build has not started, you have the queueItem, then POST on:
	// http://<Jenkins_URL>/queue/cancelItem?id=<queueItem>
	// Otherwise, if it has started:
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	// ...but we have to fish out the build number

	err = cancelQueueItem(queueEntry)
	if err != nil {
		return err
	}
	startedQ, err2 := strconv.Atoi(queueEntry)
	if err2 != nil {
		log.Printf("Bad queue item: %s", queueEntry)
		return err2
	}

	// Walk down the running jobs until we find the one we started
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	var currentBuild, currentQ int
	if currentBuild, currentQ, err = getLastBuildAndQueueNumbers(); err != nil {
		return err
	}

	fmt.Printf("Got build number %d, currentQ = %d\n", currentBuild, currentQ)
	// Check the last 10 builds to see if we started it
	for i := 0; i < 10 && currentQ > startedQ; i++ {
		var f interface{}
		currentBuild--
		err = getJSON(jenkinsURL+"/job/"+jenkinsJobName+"/"+strconv.Itoa(currentBuild)+"/api/json", &f)
		if err != nil {
			log.Printf("No build %d", currentBuild)
			return err
		}
		m := f.(map[string]interface{})
		if _, ok := m["queueId"]; ok {
			currentQ = int(m["queueId"].(float64))
		}
	}
	if currentQ == startedQ {
		err = stopBuildByID(strconv.Itoa(currentBuild))
	}
	return err
}
