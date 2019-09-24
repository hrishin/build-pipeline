package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	listers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1alpha1"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	"knative.dev/pkg/metrics"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	prDuration = stats.Float64(
		"pipelinerun_duration_seconds",
		"The pipelinerun execution time in seconds",
		stats.UnitDimensionless)
	pRunDistributions = view.Distribution(10, 30, 60, 300, 900, 1800, 3600, 5400, 10800, 21600, 43200, 86400)

	prCount = stats.Float64("pipelinerun_count",
		"number of pipelineruns",
		stats.UnitDimensionless)

	runningPrsCount = stats.Float64("running_pipelineruns_count",
		"Number of of pipelines running currently",
		stats.UnitDimensionless)
)

type Recorder struct {
	initialized bool
	lister      listers.PipelineRunLister
	logger      *zap.SugaredLogger

	pipeline    tag.Key
	pipelineRun tag.Key
	namespace   tag.Key
	status      tag.Key
}

func NewRecorder(logger *zap.SugaredLogger, lister listers.PipelineRunLister) (*Recorder, error) {
	r := &Recorder{
		initialized: true,
		logger: logger,
		lister: lister,
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
			Measure:     prCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{r.status},
		},
		&view.View{
			Description: runningPrsCount.Description(),
			Measure:     runningPrsCount,
			Aggregation: view.LastValue(),
		},
	)

	if err != nil {
		r.initialized = false
		return r, err
	}

	return r, nil
}

func (r *Recorder) Record(pr *v1alpha1.PipelineRun) {
	if !r.initialized {
		r.logger.Warnf("ignoring the metrics recording for %s , failed to initialize the metrics recorder", pr.Name)
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
		r.logger.Errorf("logging report %v \n", err)
		fmt.Printf("logging report %v \n", duration)
		return
	}

	metrics.Record(ctx, prDuration.M(float64(duration/time.Second)))
	metrics.Record(ctx, prCount.M(1))
}

func (r *Recorder) RunningPrsCount() {
	if !r.initialized {
		r.logger.Warnf("ignoring the metrics recording for %s , failed to initialize the metrics recorder")
		return
	}

	prs, err := r.lister.List(labels.Everything())
	if err != nil {
		r.logger.Errorf("failed to list pipelineruns while generating metrics : %v", err)
		return
	}

	var runningPrs int
	for _, pr := range prs {
		if !pr.IsDone() {
			runningPrs ++
		}
	}

	ctx, err := tag.New(
		context.Background(),
	)
	metrics.Record(ctx, runningPrsCount.M(float64(runningPrs)))
}
