package controller

import (
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
)

type Optimizations struct {
	itemsChan chan *golang.OptimizationItem
	items     []*golang.OptimizationItem

	selectedItem *golang.OptimizationItem

	reEvaluateFunc func(id string, items []*golang.PreferenceItem)
}

func NewOptimizations() *Optimizations {
	o := Optimizations{
		itemsChan: make(chan *golang.OptimizationItem, 1000),
	}
	go o.Process()
	return &o
}

func (o *Optimizations) Process() {
	for newItem := range o.itemsChan {
		updated := false
		for idx, i := range o.items {
			if newItem.Id == i.Id {
				o.items[idx] = newItem
				updated = true
				break
			}
		}
		if !updated {
			o.items = append(o.items, newItem)
		}
	}
}

func (o *Optimizations) SendItem(item *golang.OptimizationItem) {
	o.itemsChan <- item
}

func (o *Optimizations) Items() []*golang.OptimizationItem {
	return o.items
}

func (o *Optimizations) SetReEvaluateFunc(f func(id string, items []*golang.PreferenceItem)) {
	o.reEvaluateFunc = f
}

func (o *Optimizations) SelectItem(i *golang.OptimizationItem) {
	o.selectedItem = i
}

func (o *Optimizations) SelectedItem() *golang.OptimizationItem {
	return o.selectedItem
}

func (o *Optimizations) ReEvaluate(id string, preferences []*golang.PreferenceItem) {
	o.reEvaluateFunc(id, preferences)
}
