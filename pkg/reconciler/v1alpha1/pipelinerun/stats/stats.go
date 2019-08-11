package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	pRunDurationInSeconds = stats.Float64(
		"pipelinerun_duration",
		"The pipelinerun execution time in seconds",
		stats.UnitDimensionless)

	pRunLatencyDistribution = view.Distribution(10, 30, 60, 300, 900, 1800, 3600, 5400, 10800, 21600, 43200, 86400)
)

type Reporter struct {
	pipeline    tag.Key
	pipelineRun tag.Key
	namespace   tag.Key
	status      tag.Key
}

func NewReporter() (*Reporter, error) {
	r := &Reporter{}

	pipeline, err := tag.NewKey("pipeline")
	if err != nil {
		return nil, err
	}
	r.pipeline = pipeline

	pipelineRun, err := tag.NewKey("pipelinerun")
	if err != nil {
		return nil, err
	}
	r.pipelineRun = pipelineRun

	namespace, err := tag.NewKey("namespace")
	if err != nil {
		return nil, err
	}
	r.namespace = namespace

	status, err := tag.NewKey("status")
	if err != nil {
		return nil, err
	}
	r.status = status

	err = view.Register(
		&view.View{
			Description: "The pipelinerun duration in seconds",
			Measure:     pRunDurationInSeconds,
			Aggregation: pRunLatencyDistribution,
			TagKeys:     []tag.Key{r.pipeline, r.pipelineRun, r.namespace, r.status},
		},
	)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reporter) Report(logger *zap.SugaredLogger, pr *v1alpha1.PipelineRun) {
	duration := time.Since(pr.Status.StartTime.Time)
	if pr.Status.CompletionTime != nil {
		duration = pr.Status.CompletionTime.Sub(pr.Status.StartTime.Time)
	}


	ctx, err := tag.New(
		context.Background(),
		tag.Insert(r.pipeline, pr.Spec.PipelineRef.Name),
		tag.Insert(r.pipelineRun, pr.Name),
		tag.Insert(r.namespace, pr.Namespace),
		tag.Insert(r.status, string(pr.Status.Conditions[0].Status)),
	)

	if err != nil {
		logger.Errorf("logging report %v \n", err)
		fmt.Printf("logging report %v \n", duration)
		return
	}

	stats.Record(ctx, pRunDurationInSeconds.M(float64(duration/time.Second)))
}
