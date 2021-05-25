package agent

import (
	"time"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/errors"
	hc "github.com/Axway/agent-sdk/pkg/util/healthcheck"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

type periodicStatusUpdate struct {
	jobs.Job
	previousActivityTime time.Time
	currentActivityTime  time.Time
	prevStatus           string
}

var statusUpdate *periodicStatusUpdate

func (su *periodicStatusUpdate) Ready() bool {
	if runStatusUpdateCheck() != nil {
		return false
	}
	// Do not start until status will be running// get the status from the health check and jobs
	status := su.getCombinedStatus()
	if status != AgentRunning {
		return false
	}

	log.Debug("Periodic status update is ready")
	su.currentActivityTime = time.Now()
	su.previousActivityTime = su.currentActivityTime
	return true
}

func (su *periodicStatusUpdate) Status() error {
	// error out if the agent name does not exist
	err := runStatusUpdateCheck()
	if err != nil {
		return err
	}
	return nil
}

func (su *periodicStatusUpdate) Execute() error {
	// error out if the agent name does not exist
	err := runStatusUpdateCheck()
	if err != nil {
		log.Error(errors.ErrPeriodicCheck.FormatError("periodic status updater"))
		return err
	}

	// get the status from the health check and jobs
	status := su.getCombinedStatus()

	if su.prevStatus != status {
		UpdateLocalActivityTime()
	}

	// if the last timestamp for an event has changed, update the resource
	if time.Time(su.currentActivityTime).After(time.Time(su.previousActivityTime)) {
		log.Tracef("Activity change detected at %s, from previous activity at %s, updating status", su.currentActivityTime, su.previousActivityTime)
		log.Tracef("*********PREV STATUS %v*************", su.prevStatus)
		log.Tracef("*********NEW  STATUS %v*************", status)
		UpdateStatus(status, "")
		su.prevStatus = status
		su.previousActivityTime = su.currentActivityTime
	}
	return nil
}

// StartPeriodicStatusUpdate - starts a job that runs the periodic status updates
func StartPeriodicStatusUpdate() {
	interval := agent.cfg.GetReportActivityFrequency()
	statusUpdate = &periodicStatusUpdate{}
	_, err := jobs.RegisterDetachedIntervalJob(statusUpdate, interval)

	if err != nil {
		log.Error(errors.Wrap(errors.ErrStartingPeriodicStatusUpdate, err.Error()))
	}
}

func (su *periodicStatusUpdate) getCombinedStatus() string {
	status := su.getJobPoolStatus()
	hcStatus := su.getHealthcheckStatus()
	if hcStatus != AgentRunning {
		status = hcStatus
	}
	return status
}

// getJobPoolStatus
func (su *periodicStatusUpdate) getJobPoolStatus() string {
	status := jobs.GetStatus()

	// update the status only if not running
	if status == jobs.PoolStatusStopped.String() {
		return AgentUnhealthy
	}
	return AgentRunning
}

// getHealthcheckStatus
func (su *periodicStatusUpdate) getHealthcheckStatus() string {
	hcStatus := hc.GetGlobalStatus()

	// update the status only if not running
	if hcStatus == string(hc.FAIL) {
		return AgentUnhealthy
	}
	return AgentRunning
}

// runStatusUpdateCheck - returns an error if agent name is blank
func runStatusUpdateCheck() error {
	if agent.cfg.GetAgentName() == "" {
		return errors.ErrStartingPeriodicStatusUpdate
	}
	return nil
}

// UpdateLocalActivityTime - updates the local activity timestamp for the event to compare against
func UpdateLocalActivityTime() {
	statusUpdate.currentActivityTime = time.Now()
}

func getLocalActivityTime() time.Time {
	return statusUpdate.currentActivityTime
}
