package metrics

import (
	"context"
	"fmt"
	"time"

	"knative.dev/pkg/metrics"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	prDuration = stats.Float64(
		"pipelinerun_duration_seconds",
		"The pipelinerun execution time in seconds",
		stats.UnitDimensionless)

	prCount = stats.Float64("" +
		"pipelinerun_count",
		"number of pipelineruns",
		stats.UnitDimensionless)

	pRunDistributions = view.Distribution(10, 30, 60, 300, 900, 1800, 3600, 5400, 10800, 21600, 43200, 86400)
)

type Recorder struct {
	pipeline    tag.Key
	pipelineRun tag.Key
	namespace   tag.Key
	status      tag.Key
	initialized bool
}

func NewRecorder() (*Recorder, error) {
	r := &Recorder{
		initialized: true,
	}

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
			Description: prDuration.Description(),
			Measure:     prDuration,
			Aggregation: pRunDistributions,
			TagKeys:     []tag.Key{r.pipeline, r.pipelineRun, r.namespace, r.status},
		},
		&view.View{
			Description: prCount.Description(),
			Measure: prCount,
			Aggregation: view.Count(),
			TagKeys: []tag.Key{r.status},
		},
	)

	if err != nil {
		r.initialized = false
		return r, err
	}

	return r, nil
}

func (r *Recorder) Record(logger *zap.SugaredLogger, pr *v1alpha1.PipelineRun) {
	if !r.initialized {
		logger.Warnf("ignoring the metrics recording for %s , failed to initialize the metrics recorder", pr.Name)
		return
	}

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

	metrics.Record(ctx, prDuration.M(float64(duration/time.Second)))
	metrics.Record(ctx, prCount.M(1))
}
