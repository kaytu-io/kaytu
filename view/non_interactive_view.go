package view

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"sync"
	"time"
)

type NonInteractiveView struct {
	itemsChan chan *golang.OptimizationItem
	items     []*golang.OptimizationItem

	runningJobsMap map[string]string
	failedJobsMap  map[string]string
	statusErr      string

	errorChan chan error

	jobChan  chan *golang.JobResult
	jobs     []*golang.JobResult
	jobMutex sync.RWMutex

	resultsReady chan bool
}

func NewNonInteractiveView() *NonInteractiveView {
	v := &NonInteractiveView{
		itemsChan:      make(chan *golang.OptimizationItem, 1000),
		runningJobsMap: map[string]string{},
		failedJobsMap:  map[string]string{},
		jobMutex:       sync.RWMutex{},
		jobChan:        make(chan *golang.JobResult, 10000),
		errorChan:      make(chan error, 10000),
		resultsReady:   make(chan bool),
	}
	v.resultsReady <- false
	return v
}

var primary = color.New(color.FgHiCyan)

var bold = color.New(color.Bold)
var faint = color.New(color.Faint)
var underline = color.New(color.Underline)

var primaryLink = color.New(color.Underline).Add(color.Bold)

// OptimizationsString returns a string to show the optimization results and details
func (v *NonInteractiveView) OptimizationsString() (string, error) {
	var costString string

	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateHeader = false
	t.Style().Format.Header = text.FormatDefault

	var columns []table.ColumnConfig
	i := 1
	var headers table.Row
	headers = append(headers, underline.Sprint("ID"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("Resource Type"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Region"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Platform"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("Total Save"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignRight,
		AlignHeader: text.AlignRight,
	})
	i++

	t.AppendRow(table.Row{""})

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)

	for _, item := range v.items {
		var row table.Row
		totalSaving := 0.0
		if !item.Loading && !item.Skipped && !item.LazyLoadingEnabled {
			for _, dev := range item.Devices {
				totalSaving += dev.CurrentCost - dev.RightSizedCost
			}
		}
		row = append(row, item.Id, item.ResourceType, item.Region, item.Platform, fmt.Sprintf("$%.2f", totalSaving))

		t.AppendRow(row)
	}

	costString = t.Render()
	costString += "\n──────────────────────────────────\n"

	return costString, nil
}

func (v *NonInteractiveView) PublishItem(item *golang.OptimizationItem) {
	v.itemsChan <- item
}

func (v *NonInteractiveView) PublishJob(job *golang.JobResult) {
	v.jobChan <- job
}

func (v *NonInteractiveView) PublishError(err error) {
	v.errorChan <- err
}

func (v *NonInteractiveView) PublishResultsReady(ready *golang.ResultsReady) {
	v.resultsReady <- ready.Ready
}

func (v *NonInteractiveView) WaitAndShowResults() error {
	go v.WaitForJobs()
	go v.WaitForAllItems()
	for len(v.runningJobsMap) == 0 {
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(5 * time.Second)
	for v.itemLoadingExists() && len(v.runningJobsMap) > 0 {
		time.Sleep(5 * time.Second)
	}
	str, err := v.OptimizationsString()
	if err != nil {
		return err
	}
	fmt.Println(str)
	return nil
}

func (v *NonInteractiveView) itemLoadingExists() bool {
	for _, item := range v.items {
		if item.Loading {
			return true
		}
	}
	return false
}

func (v *NonInteractiveView) jobPendingExists() bool {
	for _, j := range v.jobs {
		if j.Done == false {
			return true
		}
	}
	return false
}

func (v *NonInteractiveView) WaitForJobs() {
	for {
		select {
		case job := <-v.jobChan:
			v.jobMutex.Lock()
			if !job.Done {
				v.runningJobsMap[job.Id] = job.Description
			} else {
				if _, ok := v.runningJobsMap[job.Id]; ok {
					delete(v.runningJobsMap, job.Id)
				}
			}
			if len(job.FailureMessage) > 0 {
				v.failedJobsMap[job.Id] = fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage)
			}
			v.jobMutex.Unlock()

		case err := <-v.errorChan:
			v.statusErr = fmt.Sprintf("Failed due to %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (v *NonInteractiveView) WaitForAllItems() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		select {
		case newItem := <-v.itemsChan:
			updated := false
			for idx, i := range v.items {
				if newItem.Id == i.Id {
					v.items[idx] = newItem
					updated = true
					break
				}
			}
			if !updated {
				v.items = append(v.items, newItem)
			}
		}
	}
}
