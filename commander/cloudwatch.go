package commander

import (
	"context"
	"flag"
	"log"
	"rift/display"
	"rift/internal/pkg/cloudwatch"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type CloudWatch struct {
	fs *flag.FlagSet `validate:"-"`

	region       string `validate:"required"`
	profile      string `validate:"required"`
	group        string `validate:"required"`
	streamName   string `validate:"required"`
	formatJSON   bool   `validate:"required"`
	protoFile    string `validate:"required"`
	live         bool   `validate:"-"`
	pattern      string `validate:"-"`
	onlyMatching bool   `validate:"-"`
}

func NewCloudWatch() *CloudWatch {
	cw := &CloudWatch{}

	fs := flag.NewFlagSet("cloudwatch", flag.ExitOnError)

	fs.StringVar(&cw.region, "region", "us-west-2", "aws region")
	fs.StringVar(&cw.profile, "profile", "default", "aws profile. uses default")

	fs.StringVar(&cw.group, "group", "", "cloudwatch log group name")
	fs.StringVar(&cw.streamName, "stream", "", "cloudwatch log stream name")
	fs.StringVar(&cw.protoFile, "proto", "", "proto file name to parse")

	fs.BoolVar(&cw.formatJSON, "json", false, "print in json format")
	fs.BoolVar(&cw.live, "live", false, "live tail logs")
	fs.StringVar(&cw.pattern, "grep", "", "grep pattern in logs")
	fs.BoolVar(&cw.onlyMatching, "only", false, "show only matching grep pattern")

	cw.fs = fs
	return cw
}

func (cwc *CloudWatch) Parse(cx context.Context, cmdStr ...string) (err error) {
	log.Println("running ", cmdStr)

	if cwc.fs == nil {
		return ErrUnInitialized
	}

	err = cwc.fs.Parse(cmdStr)
	if err != nil {
		err = errors.Wrap(err, "parse_failed")
		return
	}

	v := validator.New()
	err = v.Struct(cwc)
	if err != nil {
		err = errors.Wrap(err, "validation_failed")
	}
	return
}

func (cwc *CloudWatch) Run(ctx context.Context) (err error) {
	done := make(chan struct{}, 1)
	defer close(done)

	stream := cloudwatch.NewStream(cloudwatch.StreamOpts{
		Region:       cwc.region,
		Group:        cwc.group,
		StreamPrefix: cwc.streamName,
		Profile:      cwc.profile,
		Forever:      cwc.live,
	}, done)

	formatter := display.CloudWatchFormatter{
		TextFormat:   !cwc.formatJSON,
		Pattern:      cwc.pattern,
		ShouldFilter: len(cwc.pattern) > 0,
		Filter:       display.SimpleFilter,
		Protofile:    cwc.protoFile,
	}
	eventChan := stream.Run(ctx)
	for event := range eventChan {
		if event.Message != nil {
			formatter.Display(cwc.streamName, *event.Message)
		}
	}

	return nil
}
