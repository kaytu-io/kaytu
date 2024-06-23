package sdk

import (
	"context"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
)

type Processor interface {
	ReEvaluate(ctx context.Context, evaluate *golang.ReEvaluate)
	GetConfig(ctx context.Context) golang.RegisterConfig
	StartProcess(ctx context.Context, cmd string, flags map[string]string, kaytuAccessToken string, preferences []*golang.PreferenceItem, jobQueue *JobQueue) error
	SetStream(ctx context.Context, stream *StreamController)
}
