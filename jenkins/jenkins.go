// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package jenkins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const jenkinsURL = "http://192.168.1.10:8080"
const jenkinsJobName = "gui_kbfs_dokan"

func parseQueueNumber(locationString string) string {
	countIndex := strings.Index(locationString, "queue/item/")
	if countIndex == -1 {
		log.Printf("bad location format: %s", locationString)
		return ""
	}
	countIndex += len("queue/item/")
	return strings.TrimRight(locationString[countIndex:], "/")
}

// StartBuild starts the build and returns the queue number as a string from Jenkins
func StartBuild(clientRev string, kbfsRev string, jsonUpdateFilename string) (string, error) {

	u, err := url.Parse(jenkinsURL + "/job/" + jenkinsJobName + "/buildWithParameters")
	urlValues := url.Values{}
	if clientRev != "" {
		urlValues.Add("ClientRevision", clientRev)
	}
	if kbfsRev != "" {
		urlValues.Add("KFSRevision", kbfsRev)
	}
	if jsonUpdateFilename != "" {
		urlValues.Add("JSON_UPDATE_FILENAME", jsonUpdateFilename)
	}
	u.RawQuery = urlValues.Encode()
	buildurl := u.String()
	log.Printf("Posting: %s \n", buildurl)
	res, err := http.Post(buildurl, "", nil)
	defer func() { _ = res.Body.Close() }()

	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}
	if res.StatusCode != 201 {
		return "", fmt.Errorf("Build request returned %d", res.StatusCode)
	}
	loc, _ := res.Location()
	log.Print(robots)
	return fmt.Sprintf("Requested Jenkins build with queue ID %s", parseQueueNumber(loc.String())), nil
}

func stopBuildByID(buildID string) {
	log.Printf("Stopping build %s \n", buildID)
	// Of the form: http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	res, err := http.Post(jenkinsURL+"/job/"+jenkinsJobName+"/"+buildID+"/stop", "", nil)
	if err != nil {
		log.Print(err)
	}
	_, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		log.Print(err)
	}
	log.Printf("Status code: %v \n", res.StatusCode)
}

func getJSON(url string, target interface{}) error {
	fmt.Printf("posting to: %s\n", url)
	r, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = r.Body.Close() }()

	return json.NewDecoder(r.Body).Decode(target)
}

func cancelQueueItem(queueEntry string) {
	log.Printf("Looking to cancel queue item: %s\n", queueEntry)
	// Try removing from queue, even though it's probably not in there
	u, err := url.Parse(jenkinsURL + "/queue/cancelItem")
	urlValues := url.Values{}
	urlValues.Add("id", queueEntry)
	u.RawQuery = urlValues.Encode()
	cancelURL := u.String()

	res, err := http.Post(cancelURL, "", nil)

	if err != nil {
		log.Print(err)
		return
	}

	_, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("Status code: %v \n", res.StatusCode)
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

// StopBuild will stop/cancel the current build.
// Location is the return string from a build request, e.g.:
// http://192.168.1.10:8080/queue/item/34/
func StopBuild(queueEntry string) {

	// If the build has not started, you have the queueItem, then POST on:
	// http://<Jenkins_URL>/queue/cancelItem?id=<queueItem>
	// Otherwise, if it has started:
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	// ...but we have to fish out the build number

	cancelQueueItem(queueEntry)
	startedQ, err := strconv.Atoi(queueEntry)
	if err != nil {
		log.Printf("Bad queue item: %s", queueEntry)
		return
	}

	// Walk down the running jobs until we find the one we started
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	var currentBuild, currentQ int
	if currentBuild, currentQ, err = getLastBuildAndQueueNumbers(); err != nil {
		return
	}

	fmt.Printf("Got build number %d, currentQ = %d\n", currentBuild, currentQ)
	// Check the last 10 builds to see if we started it
	for i := 0; i < 10 && currentQ > startedQ; i++ {
		var f interface{}
		currentBuild--
		err = getJSON(jenkinsURL+"/job/"+jenkinsJobName+"/"+strconv.Itoa(currentBuild)+"/api/json", &f)
		if err != nil {
			log.Printf("No build %d", currentBuild)
			return
		}
		m := f.(map[string]interface{})
		if _, ok := m["queueId"]; ok {
			currentQ = int(m["queueId"].(float64))
		}
	}
	if currentQ == startedQ {
		stopBuildByID(strconv.Itoa(currentBuild))
	}
}
