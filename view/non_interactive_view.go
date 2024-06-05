package view

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

type NonInteractiveView struct {
	runningJobsMap map[string]string
	failedJobsMap  map[string]string
	statusErr      string

	Optimizations *controller.Optimizations[golang.OptimizationItem]

	PluginCustomOptimizations *controller.Optimizations[golang.ChartOptimizationItem]
	OverviewChart             *golang.ChartDefinition
	DevicesChart              *golang.ChartDefinition

	errorChan chan error

	jobChan  chan *golang.JobResult
	jobs     []*golang.JobResult
	jobMutex sync.RWMutex

	resultsReady chan bool
}

func NewNonInteractiveView() *NonInteractiveView {
	v := &NonInteractiveView{
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

	for _, item := range v.Optimizations.Items() {
		resultsString += getItemString(item)
		resultsString += "\n──────────────────────────────────\n"
	}

	return resultsString, nil
}

func (v *NonInteractiveView) SetOptimizations(optimizations *controller.Optimizations[golang.OptimizationItem],
	pluginCustomOptimizations *controller.Optimizations[golang.ChartOptimizationItem],
	overviewChart *golang.ChartDefinition, devicesChart *golang.ChartDefinition) {
	v.Optimizations = optimizations
	v.PluginCustomOptimizations = pluginCustomOptimizations
	v.OverviewChart = overviewChart
	v.DevicesChart = devicesChart
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
		if p.Hidden {
			continue
		}
		var row table.Row
		row = append(row, "└───── "+p.Key, p.Current, p.Average, p.Max, p.Recommended)
		t.AppendRow(row)
	}
	itemString += t.Render()
	return itemString
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
					var csvHeaders []string
					var csvRows [][]string
					if v.Optimizations != nil {
						csvHeaders, csvRows = exportCsv(v.Optimizations.Items())
					} else {
						csvHeaders, csvRows = v.exportCustomCsv(v.PluginCustomOptimizations.Items())
					}
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
					var jsonData []byte
					var err error
					if v.Optimizations != nil {
						jsonValue := struct {
							Items []*golang.OptimizationItem
						}{
							Items: v.Optimizations.Items(),
						}
						jsonData, err = json.Marshal(jsonValue)
						if err != nil {
							return err
						}
					} else {
						jsonValue := struct {
							Items []map[string]any
						}{
							Items: convertOptimizeJson(v.PluginCustomOptimizations.Items()),
						}
						jsonData, err = json.Marshal(jsonValue)
						if err != nil {
							return err
						}
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
			os.Stderr.WriteString("\n" + err.Error())
			return nil
		}
	}
}

func (v *NonInteractiveView) WaitAndReturnResults(nonInteractiveFlag string) (string, error) {
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
					var csvHeaders []string
					var csvRows [][]string
					if v.Optimizations != nil {
						csvHeaders, csvRows = exportCsv(v.Optimizations.Items())
					} else {
						csvHeaders, csvRows = v.exportCustomCsv(v.PluginCustomOptimizations.Items())
					}
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
					var jsonData []byte
					var err error
					if v.Optimizations != nil {
						jsonValue := struct {
							Items []*golang.OptimizationItem
						}{
							Items: v.Optimizations.Items(),
						}
						jsonData, err = json.Marshal(jsonValue)
						if err != nil {
							return "", err
						}
					} else {
						jsonValue := struct {
							Items []map[string]any
						}{
							Items: convertOptimizeJson(v.PluginCustomOptimizations.Items()),
						}
						jsonData, err = json.Marshal(jsonValue)
						if err != nil {
							return "", err
						}
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
				if utils.MatchesLimitPattern(fmt.Sprintf("%s failed due to %s", job.Description, job.FailureMessage)) {
					v.errorChan <- fmt.Errorf(utils.ContactUsMessage)
				}
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
				if p.Hidden {
					continue
				}
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

func (v *NonInteractiveView) exportCustomCsv(items []*golang.ChartOptimizationItem) ([]string, [][]string) {
	var rowsMap []map[string]string
	for _, i := range items {
		row := make(map[string]string)
		for key, value := range i.OverviewChartRow.Values {
			if strings.HasPrefix(key, "x_kaytu") {
				continue
			}
			row[fmt.Sprintf("Item-%s", key)] = removeANSI(value.Value)
		}
		for _, d := range i.DevicesChartRows {
			for key, value := range d.Values {
				if strings.HasPrefix(key, "x_kaytu") {
					continue
				}
				row[fmt.Sprintf("Device-%s", key)] = removeANSI(value.Value)
			}
			//for key, value := range i.DevicesProperties {
			//	if key != d.RowId {
			//		continue
			//	}
			//	for _, p := range value.Properties {
			//		row[fmt.Sprintf("Device-%s-current", p.Key)] = removeANSI(p.Current)
			//		row[fmt.Sprintf("Device-%s-recommended", p.Key)] = removeANSI(p.Recommended)
			//		row[fmt.Sprintf("Device-%s-average", p.Key)] = removeANSI(p.Average)
			//		row[fmt.Sprintf("Device-%s-max", p.Key)] = removeANSI(p.Max)
			//	}
			//} // TODO
		}
		rowsMap = append(rowsMap, row)
	}
	var itemHeaders []string
	var deviceHeaders []string
	for _, value := range v.OverviewChart.Columns {
		itemHeaders = append(itemHeaders, fmt.Sprintf("Item-%s", value.Id))
	}
	for _, value := range v.DevicesChart.Columns {
		deviceHeaders = append(deviceHeaders, fmt.Sprintf("Device-%s", value.Id))
	}
	var rows [][]string
	for _, row := range rowsMap {
		var itemRow []string
		var deviceRow []string
		for _, header := range itemHeaders {
			if _, ok := row[header]; ok {
				itemRow = append(itemRow, row[header])
			} else {
				itemRow = append(itemRow, "")
			}
		}
		for _, header := range deviceHeaders {
			if _, ok := row[header]; ok {
				deviceRow = append(deviceRow, row[header])
			} else {
				deviceRow = append(deviceRow, "")
			}
		}
		rows = append(rows, append(itemRow, deviceRow...))
	}
	return append(itemHeaders, deviceHeaders...), rows
}

func convertOptimizeJson(items []*golang.ChartOptimizationItem) []map[string]any {
	var mappedItems []map[string]any
	for _, i := range items {
		item := make(map[string]any)
		for key, value := range i.OverviewChartRow.Values {
			if strings.HasPrefix(key, "x_kaytu") {
				continue
			}
			item[key] = removeANSI(value.Value)
		}
		devices := make(map[string]map[string]any)
		for _, d := range i.DevicesChartRows {
			device := make(map[string]any)
			for key, value := range d.Values {
				if strings.HasPrefix(key, "x_kaytu") {
					continue
				}
				device[key] = removeANSI(value.Value)
			}
			devices[d.RowId] = device
		}
		for k, d := range i.DevicesProperties {
			for _, p := range d.Properties {
				devices[k][toSnakeCase(p.Key)] = struct {
					Current     string `json:"current"`
					Average     string `json:"average"`
					Max         string `json:"max"`
					Recommended string `json:"recommended"`
				}{
					Current:     p.Current,
					Average:     p.Average,
					Max:         p.Max,
					Recommended: p.Recommended,
				}
			}
		}
		var devicesArray []map[string]any
		for _, d := range devices {
			devicesArray = append(devicesArray, d)
		}
		item["devices"] = devicesArray
		mappedItems = append(mappedItems, item)
	}
	return mappedItems
}

func removeANSI(text string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(text, "")
}

func toSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")

	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && s[i-1] != '_' {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	s = result.String()

	re := regexp.MustCompile(`_+`)
	s = re.ReplaceAllString(s, "_")

	return s
}
