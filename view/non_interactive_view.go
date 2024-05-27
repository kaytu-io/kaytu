package view

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"os"
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
	return v
}

var bold = color.New(color.Bold)
var underline = color.New(color.Underline)

// OptimizationsString returns a string to show the optimization results and details
func (v *NonInteractiveView) OptimizationsString() (string, error) {
	var resultsString string

	for _, item := range v.items {
		resultsString += getItemString(item)
		resultsString += "\n──────────────────────────────────\n"
	}

	return resultsString, nil
}

func getItemString(item *golang.OptimizationItem) string {
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

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)
	var row table.Row
	var itemString string
	if item.Skipped {
		row = append(row, item.Id, item.ResourceType, item.Region, item.Platform, "Row Skipped")
		t.AppendRow(row)
		itemString += t.Render()
	} else {
		totalSaving := 0.0
		if !item.Loading && !item.Skipped && !item.LazyLoadingEnabled {
			for _, dev := range item.Devices {
				totalSaving += dev.CurrentCost - dev.RightSizedCost
			}
		}
		row = append(row, item.Id, item.ResourceType, item.Region, item.Platform, fmt.Sprintf("%s", utils.FormatPriceFloat(totalSaving)))
		t.AppendRow(row)
		itemString += t.Render()
		itemString += "\n    " + bold.Sprint("Devices") + ":"
		for _, dev := range item.Devices {
			itemString += "\n"
			itemString += getDeviceString(dev)
		}
	}

	return itemString
}

func getDeviceString(dev *golang.Device) string {
	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateHeader = false
	t.Style().Format.Header = text.FormatDefault

	var columns []table.ColumnConfig
	i := 1
	var headers table.Row
	headers = append(headers, underline.Sprint(""))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("ResourceType"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Runtime"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Current Cost"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("Right Sized Cost"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignRight,
		AlignHeader: text.AlignRight,
	})
	i++

	headers = append(headers, underline.Sprint("Savings"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignRight,
		AlignHeader: text.AlignRight,
	})
	i++

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)
	var row table.Row
	var itemString string
	row = append(row, "└─ "+dev.DeviceId, dev.ResourceType, dev.Runtime, dev.CurrentCost, dev.RightSizedCost, fmt.Sprintf("%s", utils.FormatPriceFloat(dev.CurrentCost-dev.RightSizedCost)))
	t.AppendRow(row)
	itemString += t.Render()
	itemString += "\n        " + bold.Sprint("Properties") + ":\n" + getPropertiesString(dev.Properties)
	return itemString
}

func getPropertiesString(properties []*golang.Property) string {
	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateHeader = false
	t.Style().Format.Header = text.FormatDefault

	var columns []table.ColumnConfig
	i := 1
	var headers table.Row
	headers = append(headers, underline.Sprint(""))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("Current"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Average Usage"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++
	headers = append(headers, underline.Sprint("Max Usage"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	headers = append(headers, underline.Sprint("Recommendation"))
	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignRight,
		AlignHeader: text.AlignRight,
	})
	i++

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)

	var itemString string
	for _, p := range properties {
		var row table.Row
		row = append(row, "└───── "+p.Key, p.Current, p.Average, p.Max, p.Recommended)
		t.AppendRow(row)
	}
	itemString += t.Render()
	return itemString
}

func (v *NonInteractiveView) PublishItem(item *golang.OptimizationItem) {
	v.itemsChan <- item
}

func (v *NonInteractiveView) PublishJobs(jobs *golang.JobResult) {
	v.jobChan <- jobs
}

func (v *NonInteractiveView) PublishError(err error) {
	v.errorChan <- err
}

func (v *NonInteractiveView) PublishResultsReady(ready *golang.ResultsReady) {
	v.resultsReady <- ready.Ready
}

func (v *NonInteractiveView) WaitAndShowResults(nonInteractiveFlag string) error {
	go v.WaitForAllItems()
	go v.WaitForJobs()
	for {
		select {
		case ready := <-v.resultsReady:
			if ready == true {
				if nonInteractiveFlag == "table" {
					str, err := v.OptimizationsString()
					if err != nil {
						return err
					}
					os.Stdout.WriteString(str)
				} else if nonInteractiveFlag == "csv" {
					csvHeaders, csvRows := exportCsv(v.items)
					out := os.Stdout
					writer := csv.NewWriter(out)

					err := writer.Write(csvHeaders)
					if err != nil {
						return err
					}

					for _, row := range csvRows {
						err := writer.Write(row)
						if err != nil {
							return err
						}
					}
					writer.Flush()
					err = out.Close()
					if err != nil {
						return err
					}
				} else if nonInteractiveFlag == "json" {
					jsonValue := struct {
						Items []*golang.OptimizationItem
					}{
						Items: v.items,
					}
					jsonData, err := json.Marshal(jsonValue)
					if err != nil {
						return err
					}

					out := os.Stdout
					if err != nil {
						return err
					}

					_, err = out.Write(jsonData)
					if err != nil {
						return err
					}
					err = out.Close()
					if err != nil {
						return err
					}
				} else {
					os.Stderr.WriteString("output mode not recognized!")
				}
				return nil
			}
		case err := <-v.errorChan:
			os.Stderr.WriteString(err.Error())
			return nil
		}
	}
}

func (v *NonInteractiveView) WaitAndReturnResults(nonInteractiveFlag string) (string, error) {
	go v.WaitForAllItems()
	go v.WaitForJobs()
	for {
		select {
		case ready := <-v.resultsReady:
			if ready == true {
				if nonInteractiveFlag == "table" {
					str, err := v.OptimizationsString()
					if err != nil {
						return "", err
					}
					return str, nil
				} else if nonInteractiveFlag == "csv" {
					csvHeaders, csvRows := exportCsv(v.items)
					s := &bytes.Buffer{}
					writer := csv.NewWriter(s)

					err := writer.Write(csvHeaders)
					if err != nil {
						return "", err
					}

					for _, row := range csvRows {
						err := writer.Write(row)
						if err != nil {
							return "", err
						}
					}
					writer.Flush()
					return s.String(), nil
				} else if nonInteractiveFlag == "json" {
					jsonValue := struct {
						Items []*golang.OptimizationItem
					}{
						Items: v.items,
					}
					jsonData, err := json.Marshal(jsonValue)
					if err != nil {
						return "", err
					}

					return string(jsonData), nil
				} else {
					os.Stderr.WriteString("output mode not recognized!")
				}
				return "", nil
			}
		case err := <-v.errorChan:
			os.Stderr.WriteString(err.Error())
			return "", nil
		}
	}
}

func (v *NonInteractiveView) itemLoadingExists() bool {
	for _, item := range v.items {
		if item.Loading {
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
				os.Stderr.WriteString(job.Description + " Done.\n")
				if _, ok := v.runningJobsMap[job.Id]; ok {
					delete(v.runningJobsMap, job.Id)
				}
			}
			if len(job.FailureMessage) > 0 {
				v.failedJobsMap[job.Id] = fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage)
			}
			v.jobMutex.Unlock()

		case err := <-v.errorChan:
			os.Stderr.WriteString(err.Error() + "\n")
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

func exportCsv(items []*golang.OptimizationItem) ([]string, [][]string) {
	headers := []string{
		"Item-ID", "Item-ResourceType", "Item-Region", "Item-Platform", "Item-TotalSave",
		"Device-ID", "Device-ResourceType", "Device-Runtime", "Device-CurrentCost", "Device-RightSizedCost", "Device-Savings",
		"Property-Name", "Property-Current", "Property-Average", "Property-Max", "Property-Recommendation",
	}
	var rows [][]string
	for _, i := range items {
		totalSaving := float64(0)
		for _, d := range i.Devices {
			totalSaving = totalSaving + (d.CurrentCost - d.RightSizedCost)
		}
		for _, d := range i.Devices {
			for _, p := range d.Properties {
				rows = append(rows, []string{
					i.Id, i.ResourceType, i.Region, i.Platform, fmt.Sprintf("%s", utils.FormatPriceFloat(totalSaving)),
					d.DeviceId, d.ResourceType, d.Runtime, fmt.Sprintf("%s", utils.FormatPriceFloat(d.CurrentCost)), fmt.Sprintf("%s", utils.FormatPriceFloat(d.RightSizedCost)), fmt.Sprintf("%s", utils.FormatPriceFloat(d.CurrentCost-d.RightSizedCost)),
					p.Key, p.Current, p.Average, p.Max, p.Recommended,
				})
			}
		}
	}
	return headers, rows
}
