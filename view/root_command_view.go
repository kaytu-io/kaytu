package view

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"os"
	"time"
)

type RootCommandView struct {
	statusErr string
	errorChan chan error

	jobChan chan *golang.JobResult

	resultsReady chan bool
}

func NewRootCommandView() *RootCommandView {
	v := &RootCommandView{
		jobChan:      make(chan *golang.JobResult, 10000),
		errorChan:    make(chan error, 10000),
		resultsReady: make(chan bool),
	}
	return v
}

func (v *RootCommandView) PublishJobs(jobs *golang.JobResult) {
	v.jobChan <- jobs
}

func (v *RootCommandView) PublishError(err error) {
	v.errorChan <- err
}

func (v *RootCommandView) PublishResultsReady(ready *golang.ResultsReady) {
	v.resultsReady <- ready.Ready
}

func (v *RootCommandView) WaitAndShowResults() error {
	go v.WaitForJobs()
	for {
		select {
		case _ = <-v.resultsReady:
			return nil
		}
	}
}

func (v *RootCommandView) WaitForJobs() {
	for {
		select {
		case job := <-v.jobChan:
			if !job.Done {
				os.Stderr.WriteString(job.Description + " Running...\n")
			} else {
				os.Stderr.WriteString(job.Description + " Done.\n")
			}
			if len(job.FailureMessage) > 0 {
				if utils.MatchesLimitPattern(fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage)) {
					v.errorChan <- fmt.Errorf(utils.ContactUsMessage)
				}
			}
		case err := <-v.errorChan:
			os.Stderr.WriteString(err.Error() + "\n")
			v.statusErr = fmt.Sprintf("Failed due to %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
