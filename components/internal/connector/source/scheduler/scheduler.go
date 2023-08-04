package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/tools"

	"github.com/FimGroup/logging"

	"github.com/reugn/go-quartz/quartz"
)

type GoQuartzSchedulerSourceConnectorGenerator struct {
	sched   quartz.Scheduler
	_logger logging.Logger
}

func NewGoQuartzSchedulerSourceConnectorGenerator() pluginapi.SourceConnectorGenerator {
	sched := quartz.NewStdScheduler()
	return &GoQuartzSchedulerSourceConnectorGenerator{
		sched:   sched,
		_logger: logging.GetLoggerManager().GetLogger("FimGroup.Component.SchedulerSourceConnector"),
	}
}

func (g *GoQuartzSchedulerSourceConnectorGenerator) OriginalGeneratorNames() []string {
	return []string{"job_scheduler"}
}

func (g *GoQuartzSchedulerSourceConnectorGenerator) GenerateSourceConnectorInstance(req pluginapi.SourceConnectorGenerateRequest) (pluginapi.SourceConnector, error) {
	instanceName := fmt.Sprint(tools.RandomString(), "@", req.InstanceName)
	cronTrigger, ok := req.Options["scheduler.cron"]
	if !ok {
		return nil, errors.New("scheduler.cron is not defined for instance:" + req.InstanceName)
	}
	location, ok := req.Options["scheduler.location"]
	if !ok {
		location = "Local"
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}

	instance := &GoQuartzSchedulerSourceConnector{
		name:        instanceName,
		cronTrigger: cronTrigger,
		loc:         loc,
		_logger:     g._logger,
		sched:       g.sched,
		container:   req.Container,
	}

	return instance, nil
}

func (g *GoQuartzSchedulerSourceConnectorGenerator) InitializeSubGeneratorInstance(req pluginapi.CommonSourceConnectorGenerateRequest) (pluginapi.SourceConnectorGenerator, error) {
	return nil, errors.New("sub job scheduler generator is not supported")
}

func (g *GoQuartzSchedulerSourceConnectorGenerator) Startup() error {
	g.sched.Start(context.Background())
	return nil
}

func (g *GoQuartzSchedulerSourceConnectorGenerator) Stop() error {
	g.sched.Stop()
	g.sched.Wait(context.Background())
	return nil
}

type GoQuartzSchedulerSourceConnector struct {
	name        string
	cronTrigger string
	loc         *time.Location
	_logger     logging.Logger

	sched     quartz.Scheduler
	container pluginapi.Container

	pipeline pluginapi.PipelineProcess

	job quartz.Job
}

func (g *GoQuartzSchedulerSourceConnector) Start() error {
	cron, err := quartz.NewCronTriggerWithLoc(g.cronTrigger, g.loc)
	if err != nil {
		return err
	}

	job := quartz.NewFunctionJob(func(ctx context.Context) (any, error) {
		// handling panic
		defer func() {
			if env := recover(); env != nil {
				g._logger.Error("handling panic in scheduler job:", env)
			}
		}()

		return nil, g.pipeline(g.container.NewModel())
	})

	return g.sched.ScheduleJob(context.Background(), job, cron)
}

func (g *GoQuartzSchedulerSourceConnector) Stop() error {
	return g.sched.DeleteJob(g.job.Key())
}

func (g *GoQuartzSchedulerSourceConnector) Reload() error {
	return nil
}

func (g *GoQuartzSchedulerSourceConnector) BindPipeline(process pluginapi.PipelineProcess) error {
	g.pipeline = process
	return nil
}
