package cloudwatch

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/pkg/errors"
)

type StreamOpts struct {
	Region       string
	Profile      string
	Group        string
	StreamPrefix string
	Forever      bool
}

type CloudWatchStreamer struct {
	opts       StreamOpts
	streamName *string
	done       chan struct{}
	once       *sync.Once
	eventsChan chan types.OutputLogEvent
	client     *cloudwatchlogs.Client
}

func NewStream(opts StreamOpts, done chan struct{}) *CloudWatchStreamer {
	c := &CloudWatchStreamer{
		opts:       opts,
		once:       &sync.Once{},
		done:       done,
		eventsChan: make(chan types.OutputLogEvent, 100),
	}

	if err := c.init(context.Background()); err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "failed_to_start"))
	}

	return c
}

func (c *CloudWatchStreamer) init(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(c.opts.Region),
		config.WithSharedConfigProfile(c.opts.Profile),
	)

	if err != nil {
		return errors.Wrap(err, "aws_config_load_failed")
	}

	since := time.Now().Add(-20 * time.Second)
	client := cloudwatchlogs.NewFromConfig(cfg)

	resp, err := client.FilterLogEvents(ctx, &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:        &c.opts.Group,
		LogStreamNamePrefix: &c.opts.StreamPrefix,
		Limit:               aws.Int32(1),
		StartTime:           aws.Int64(since.UnixMilli()),
	})
	if err != nil {
		return errors.Wrap(err, "failed_to_find_log_stream")
	}

	var streamName *string
	for _, logStream := range resp.Events {
		streamName = logStream.LogStreamName
	}

	if streamName == nil {
		return errors.New("log_stream_not_found")
	}

	c.client = client
	c.streamName = streamName

	return nil
}

func (c *CloudWatchStreamer) Run(ctx context.Context) chan types.OutputLogEvent {
	c.once.Do(func() {
		go c.loop(ctx)
	})

	return c.eventsChan
}

func (c *CloudWatchStreamer) getEventObject(ctx context.Context) (*cloudwatchlogs.GetLogEventsOutput, error) {
	if c.streamName == nil {
		return nil, errors.New("log_stream_not_found")
	}

	since := time.Now().Add(-20 * time.Second)

	out, err := c.client.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &c.opts.Group,
		LogStreamName: c.streamName,
		StartTime:     aws.Int64(since.UnixMilli()),
	})
	if err != nil {
		return nil, errors.Wrap(err, "get_log_events_failed")
	}

	return out, nil
}

func (c *CloudWatchStreamer) loop(ctx context.Context) {
	defer close(c.eventsChan)

	out, err := c.getEventObject(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
			for _, event := range out.Events {
				c.eventsChan <- event
			}

			if !c.opts.Forever {
				return
			}

			time.Sleep(2 * time.Second)
		}
	}

}
