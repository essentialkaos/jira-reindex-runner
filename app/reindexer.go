package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"time"

	"github.com/essentialkaos/ek/v12/knf"
	"github.com/essentialkaos/ek/v12/log"
	"github.com/essentialkaos/ek/v12/req"
	"github.com/essentialkaos/ek/v12/timeutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const JIRA_DATE_TIME_FORMAT = "2006-01-02T15:04:05Z0700"

const (
	JIRA_ENDPOINT_CHECK    = "/rest/scriptrunner/latest/custom/reindexRequired"
	JIRA_ENDPOINT_REINDEX  = "/rest/api/2/reindex"
	JIRA_ENDPOINT_PROGRESS = "/rest/api/2/reindex/progress"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type ReindexRequestInfo struct {
	IsRequired bool   `json:"is_required"`
	User       string `json:"user"`
	Date       string `json:"date"`
}

type ReindexProgressInfo struct {
	CurrentProgress int    `json:"currentProgress"`
	CurrentSubTask  string `json:"currentSubTask"`
	IsFinished      bool   `json:"success"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (i *ReindexRequestInfo) GetDate() time.Time {
	d, _ := time.Parse(JIRA_DATE_TIME_FORMAT, i.Date)
	return d
}

// ////////////////////////////////////////////////////////////////////////////////// //

// runReindex starts re-index progress
func runReindex() int {
	isReindexRequired, err := checkIfReindexRequired()

	if err != nil {
		log.Crit(err.Error())
		return 1
	}

	if !isReindexRequired {
		log.Info("Re-index is not required. Exiting…")
		return 0
	}

	isReindexInProgress, err := checkReindexProgress()

	if err != nil {
		log.Crit(err.Error())
		return 1
	}

	if isReindexInProgress {
		log.Info("Re-index already in progress. Exiting…")
		return 0
	}

	err = startReindex()

	if err != nil {
		log.Crit(err.Error())
		return 1
	}

	return 0
}

// checkIfReindexRequired checks if re-index is required
func checkIfReindexRequired() (bool, error) {
	log.Info("Checking if re-index is required…")

	i := &ReindexRequestInfo{}

	statusCode, err := sendRequest(JIRA_ENDPOINT_CHECK, req.GET, nil, i)

	if err != nil {
		return false, fmt.Errorf("Can't get information from Jira: %v", err)
	}

	if statusCode != 200 {
		return false, fmt.Errorf("Can't get information from Jira: Jira returned status code %d", statusCode)
	}

	if i.IsRequired {
		log.Info(
			"Found reindex request (author: %v | created: %s)",
			i.User, timeutil.Format(i.GetDate(), "%Y/%m/%d %H:%M:%S"),
		)
	}

	return i.IsRequired, nil
}

// checkReindexProgress checks if re-index already in progress
func checkReindexProgress() (bool, error) {
	i := &ReindexProgressInfo{}
	statusCode, err := sendRequest(JIRA_ENDPOINT_PROGRESS, req.GET, nil, i)

	if statusCode != 200 {
		return false, err
	}

	return i.IsFinished == false, err
}

// startReindex starts and monitors re-index process
func startReindex() error {
	reindexType := knf.GetS(JIRA_REINDEX_TYPE, "BACKGROUND_PREFERRED")

	log.Info("Starting re-index (type: %s)…", reindexType)

	query := req.Query{"type": reindexType}
	statusCode, err := sendRequest(JIRA_ENDPOINT_REINDEX, req.POST, query, nil)

	if err != nil {
		return fmt.Errorf("Can't run re-index progress: %v", err)
	}

	if statusCode != 202 {
		return fmt.Errorf("Can't run re-index progress: Jira returned status code %d", statusCode)
	}

	log.Info("Re-index successfully started")

	lastSuccess := time.Now()

	time.Sleep(5 * time.Second)

	for range time.NewTicker(time.Minute).C {
		if time.Since(lastSuccess) >= 10*time.Minute {
			return fmt.Errorf("Can't get info about re-index progress more than 10 minutes")
		}

		i, err := getCurrentReindexProgress()

		if err != nil {
			log.Error("Can't check re-index progress: %v", err)
			continue
		}

		lastSuccess = time.Now()

		if i.IsFinished {
			break
		}

		log.Info("%s (%d%% done)", i.CurrentSubTask, i.CurrentProgress)
	}

	log.Info("Re-index successfully finished!")

	return nil
}

// getCurrentReindexProgress return info about re-index progress
func getCurrentReindexProgress() (*ReindexProgressInfo, error) {
	i := &ReindexProgressInfo{}
	statusCode, err := sendRequest(JIRA_ENDPOINT_PROGRESS, req.GET, nil, i)

	if err != nil {
		return nil, fmt.Errorf("Can't check re-index progress: %v", err)
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("Can't check re-index progress: Jira returned status code %d", statusCode)
	}

	return i, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// sendRequest sends request to JIRA
func sendRequest(endpoint, method string, query req.Query, result interface{}) (int, error) {
	r := req.Request{
		Method: method,
		URL:    knf.GetS(JIRA_URL) + endpoint,

		BasicAuthUsername: knf.GetS(JIRA_USERNAME),
		BasicAuthPassword: knf.GetS(JIRA_PASSWORD),

		AutoDiscard: true,
	}

	if query != nil {
		r.Query = query
	}

	resp, err := r.Do()

	if err != nil {
		return -1, err
	}

	if resp.StatusCode != 200 {
		return resp.StatusCode, nil
	}

	if result != nil {
		return resp.StatusCode, resp.JSON(result)
	}

	return resp.StatusCode, nil
}
