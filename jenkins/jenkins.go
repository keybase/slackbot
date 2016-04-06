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
const jenkinsJobName = "gui_kbfs_dokan_1"

func StartBuild(ClientRev string, KBFSRev string, JsonUpdateFilename string) (string, error) {

	buildurl := jenkinsURL + "/job/" + jenkinsJobName + "/buildWithParameters"
	var params string
	if len(ClientRev) > 0 {
		params = "ClientRevision=" + url.QueryEscape(ClientRev)
	}
	if len(KBFSRev) > 0 {
		if len(params) > 0 {
			params = params + "&"
		}

		params = params + "KFSRevision=" + url.QueryEscape(KBFSRev)
	}
	if len(JsonUpdateFilename) > 0 {
		if len(params) > 0 {
			params = params + "&"
		}
		params = params + "JSON_UPDATE_FILENAME=" + url.QueryEscape(JsonUpdateFilename)
	}
	if len(params) > 0 {
		buildurl = buildurl + "?" + params
	}
	// --data-urlencode json='{"parameter": [{"name":"id", "value":"123"}, {"name":"verbosity", "value":"high"}]}'
	log.Printf("Posting: %s \n", buildurl)
	res, err := http.Post(buildurl, "", nil)
	defer res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)

	if err != nil {
		//		log.Fatal(err)
		return "", err
	}
	if res.StatusCode != 201 {
		return "", fmt.Errorf("Build request returned %d", res.StatusCode)
	}
	loc, _ := res.Location()
	//    fmt.Printf("Status code: %v  Location: %v\n", res.StatusCode, loc)
	log.Print(robots)
	return loc.String(), nil
}

func stopBuildById(buildId string) {
	log.Printf("Stopping build %s \n", buildId)
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	res, err := http.Post(jenkinsURL+"/job/"+jenkinsJobName+"/"+buildId+"/stop", "", nil)
	if err != nil {
		log.Print(err)
	}
	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Print(err)
	}
	log.Printf("Status code: %v \n", res.StatusCode)
	//	fmt.Logf("%s", robots)
}

func getJson(url string, target interface{}) error {
	fmt.Printf("posting to: %s\n", url)
	r, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// stop/cancel the current build.
// Location is the return string from a build request, e.g.:
// http://192.168.1.10:8080/queue/item/34/
func StopBuild(Location string) {

	// If the build has not started, you have the queueItem, then POST on:
	// http://<Jenkins_URL>/queue/cancelItem?id=<queueItem>
	// Otherwise, if it has started:
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	// ...but we have to fish out the build number
	countIndex := strings.Index(Location, "queue/item/")
	if countIndex == -1 {
		log.Printf("bad location format: %s", Location)
		return
	}
	countIndex += len("queue/item/")

	// Parse out the queue number
	var startedQ int
	var err error
	if startedQ, err = strconv.Atoi(strings.TrimRight(Location[countIndex:], "/")); err != nil {
		log.Print(err)
		return
	}
	log.Printf("Looking to cancel queue item: %d\n", startedQ)
	// Try removing from queue, even though it's probably not in there
	res, err := http.Post(jenkinsURL+"/queue/cancelItem?id="+strconv.Itoa(startedQ), "", nil)
	if err != nil {
		log.Print(err)
		return
	}

	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("Status code: %v \n", res.StatusCode)

	// Walk down the running jobs until we find the one we started
	// http://<Jenkins_URL>/job/<Job_Name>/<Build_Number>/stop
	var f interface{}
	err = getJson(jenkinsURL+"/job/"+jenkinsJobName+"/lastBuild/api/json", &f)
	if err != nil {
		log.Printf("No lastBuild - %v\n", err)
		return
	}
	m := f.(map[string]interface{})
	// default currentQ to zero if not present
	if _, ok := m["queueId"]; !ok {
		log.Printf("No queueId in current build --- %v", m)
		return
	}

	if _, ok := m["number"]; !ok {
		log.Printf("No number in current build --- %v", m)
		return
	}

	currentBuild := int(m["number"].(float64))
	currentQ := int(m["queueId"].(float64))

	fmt.Printf("Got build number %d, currentQ = %d\n", currentBuild, currentQ)
	// Check the last 10 builds to see if we started it
	for i := 0; i < 10 && currentQ > startedQ; i++ {
		currentBuild--
		err = getJson(jenkinsURL+"/job/"+jenkinsJobName+"/"+strconv.Itoa(currentBuild)+"/api/json", &f)
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
		stopBuildById(strconv.Itoa(currentBuild))
	}
}
