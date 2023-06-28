package commander

import (
	"context"
	"flag"
	"rift/display"
	"rift/internal/pkg/sqsreader"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type SQS struct {
	fs *flag.FlagSet

	region       string `validate:"required"`
	profile      string `validate:"required"`
	queueName    string `validate:"required"`
	formatJSON   bool   `validate:"required"`
	live         bool   `validate:"-"`
	pattern      string `validate:"-"`
	onlyMatching bool   `validate:"-"`
	protoFile    string `validate:"omitempty"`
}

func NewSQSWatch() *SQS {
	sw := &SQS{}

	fs := flag.NewFlagSet("sqs", flag.ExitOnError)
	fs.StringVar(&sw.region, "region", "us-west-2", "aws region")
	fs.StringVar(&sw.profile, "profile", "default", "aws profile. uses default")
	fs.StringVar(&sw.queueName, "queue", "", "queue name in aws sqs")
	fs.StringVar(&sw.protoFile, "proto", "", "proto file name to parse")

	fs.BoolVar(&sw.formatJSON, "json", false, "print in json format")
	fs.BoolVar(&sw.live, "live", false, "live tail logs")
	fs.StringVar(&sw.pattern, "grep", "", "grep pattern in logs")
	fs.BoolVar(&sw.onlyMatching, "only", false, "show only matching grep pattern")

	sw.fs = fs
	return sw
}

func (swc *SQS) Parse(ctx context.Context, cmdStr ...string) (err error) {
	// log.Println("running ", cmdStr)
	if swc.fs == nil {
		return ErrUnInitialized
	}

	err = swc.fs.Parse(cmdStr)
	if err != nil {
		err = errors.Wrap(err, "parse_failed")
		return
	}

	v := validator.New()
	err = v.Struct(swc)
	if err != nil {
		err = errors.Wrap(err, "validation_failed")
	}

	return err
}

func (sw *SQS) Run(ctx context.Context) (err error) {
	done := make(chan struct{}, 1)
	defer close(done)

	stream := sqsreader.NewStream(sqsreader.QueueOpts{
		Region:    sw.region,
		Profile:   sw.profile,
		QueueName: sw.queueName,
		Forever:   sw.live,
	}, done)

	formatter := display.SQSFormatter{TextFormat: !sw.formatJSON}

	eventChan := stream.Run(ctx)
	for event := range eventChan {
		if event.Body != nil {
			formatter.Display(sw.queueName, *event.Body)
		}
	}

	return nil
}
