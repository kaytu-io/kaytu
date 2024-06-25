package controller

import (
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"sync/atomic"
)

type Optimizations[T golang.OptimizationItem | golang.ChartOptimizationItem] struct {
	itemsChan          chan *T
	inProcessItemCount atomic.Int32
	items              []*T

	summaryChan           chan string
	summaryTableChan      chan *golang.ResultSummaryTable
	inProcessSummaryCount atomic.Int32
	summary               string
	summaryTable          *golang.ResultSummaryTable

	selectedItem *T

	reEvaluateFunc func(id string, items []*golang.PreferenceItem)
	initializing   bool
}

func NewOptimizations[T golang.OptimizationItem | golang.ChartOptimizationItem]() *Optimizations[T] {
	o := Optimizations[T]{
		itemsChan:             make(chan *T, 1000),
		inProcessItemCount:    atomic.Int32{},
		initializing:          true,
		summaryChan:           make(chan string),
		summaryTableChan:      make(chan *golang.ResultSummaryTable),
		inProcessSummaryCount: atomic.Int32{},
	}
	go o.Process()
	go o.SummaryProcess()
	return &o
}

func (o *Optimizations[T]) Process() {
	defer func() {
		if r := recover(); r != nil {
			o.inProcessItemCount = atomic.Int32{}
			o.Process()
		}
	}()

	for newItem := range o.itemsChan {
		o.inProcessItemCount.Add(1)
		if o.initializing {
			o.initializing = false
		}
		updated := false
		for idx, i := range o.items {
			switch castedNewItem := any(newItem).(type) {
			case *golang.OptimizationItem:
				castedI := any(i).(*golang.OptimizationItem)
				if castedNewItem.Id == castedI.Id {
					o.items[idx] = newItem
					updated = true
					break
				}
			case *golang.ChartOptimizationItem:
				castedI := any(i).(*golang.ChartOptimizationItem)
				if castedNewItem.GetOverviewChartRow().GetRowId() == castedI.GetOverviewChartRow().GetRowId() {
					o.items[idx] = newItem
					updated = true
					break
				}
			}
		}
		if !updated {
			o.items = append(o.items, newItem)
		}
		o.inProcessItemCount.Add(-1)
	}
}

func (o *Optimizations[T]) SummaryProcess() {
	defer func() {
		if r := recover(); r != nil {
			o.inProcessSummaryCount = atomic.Int32{}
			o.SummaryProcess()
		}
	}()

	for {
		select {
		case msg := <-o.summaryChan:
			o.inProcessSummaryCount.Add(1)
			o.summary = msg
			o.inProcessSummaryCount.Add(-1)

		case msg := <-o.summaryTableChan:
			o.inProcessSummaryCount.Add(1)
			o.summaryTable = msg
			o.inProcessSummaryCount.Add(-1)
		}
	}
}

func (o *Optimizations[T]) SendItem(item *T) {
	o.itemsChan <- item
}

func (o *Optimizations[T]) Items() []*T {
	return o.items
}

func (o *Optimizations[T]) SetReEvaluateFunc(f func(id string, items []*golang.PreferenceItem)) {
	o.reEvaluateFunc = f
}

func (o *Optimizations[T]) SelectItem(i *T) {
	o.selectedItem = i
}

func (o *Optimizations[T]) SelectedItem() *T {
	return o.selectedItem
}

func (o *Optimizations[T]) ReEvaluate(id string, preferences []*golang.PreferenceItem) {
	o.reEvaluateFunc(id, preferences)
}

func (o *Optimizations[T]) GetInitialization() bool {
	return o.initializing
}

func (o *Optimizations[T]) SetInitialization(b bool) {
	o.initializing = b
}

func (o *Optimizations[T]) SetResultSummary(msg string) {
	o.summaryChan <- msg
}

func (o *Optimizations[T]) SetResultSummaryTable(msg *golang.ResultSummaryTable) {
	o.summaryTableChan <- msg
}

func (o *Optimizations[T]) GetResultSummary() string {
	return o.summary
}

func (o *Optimizations[T]) GetResultSummaryTable() *golang.ResultSummaryTable {
	return o.summaryTable
}

func (o *Optimizations[T]) IsProcessing() bool {
	if len(o.itemsChan) > 0 {
		return true
	}
	if len(o.summaryChan) > 0 {
		return true
	}
	if len(o.summaryTableChan) > 0 {
		return true
	}

	if o.inProcessItemCount.Load() > 0 {
		return true
	}
	if o.inProcessSummaryCount.Load() > 0 {
		return true
	}

	if o.initializing {
		return true
	}

	return false
}
