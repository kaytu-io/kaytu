package sdk

import "github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"

type Processor interface {
	ReEvaluate(evaluate *golang.ReEvaluate)
	GetConfig() golang.RegisterConfig
	StartProcess(cmd string, flags map[string]string, kaytuAccessToken string, jobQueue *JobQueue) error
	SetStream(stream golang.Plugin_RegisterClient)
}
