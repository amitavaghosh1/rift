package sqsreader

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
)

type QueueOpts struct {
	Region    string
	Profile   string
	QueueName string
	Forever   bool
}

type QueueStreamer struct {
	opts       QueueOpts
	queueURL   *string
	done       chan struct{}
	once       *sync.Once
	eventsChan chan types.Message
	client     *sqs.Client
}

func NewStream(opts QueueOpts, done chan struct{}) *QueueStreamer {
	q := &QueueStreamer{
		opts:       opts,
		once:       &sync.Once{},
		done:       done,
		eventsChan: make(chan types.Message, 10),
	}

	if err := q.init(context.Background(), 10); err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "failed_to_start"))
	}

	return q
}

func (s *QueueStreamer) init(ctx context.Context, retries int) (err error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(s.opts.Region),
		config.WithSharedConfigProfile(s.opts.Profile),
	)

	if err != nil {
		return errors.Wrap(err, "aws_config_load_failed")
	}

	client := sqs.NewFromConfig(cfg)

	input := &sqs.GetQueueUrlInput{
		QueueName: aws.String(s.opts.QueueName),
	}
	result, err := client.GetQueueUrl(ctx, input)
	if err != nil && retries <= 0 {
		log.Println(err)
		return
	}

	if err != nil {
		log.Println("retrying ", retries-1)
		s.init(ctx, retries-1)
	}

	s.client = client
	s.queueURL = result.QueueUrl

	return nil
}

func (s *QueueStreamer) Run(ctx context.Context) chan types.Message {
	s.once.Do(func() {
		go s.loop(ctx)
	})

	return s.eventsChan
}

func (s *QueueStreamer) loop(ctx context.Context) {
	defer close(s.eventsChan)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		default:
			gMInput := &sqs.ReceiveMessageInput{
				MessageAttributeNames: []string{
					string(types.QueueAttributeNameAll),
				},
				QueueUrl:            s.queueURL,
				MaxNumberOfMessages: 10,
				VisibilityTimeout:   int32(0),
			}

			result, err := s.client.ReceiveMessage(ctx, gMInput)
			if err != nil {
				log.Println(err)
				return
			}

			if len(result.Messages) == 0 {
				log.Println("empty messages")
				time.Sleep(10 * time.Second)

				if !s.opts.Forever {
					return
				}

				continue
			}

			for _, message := range result.Messages {
				s.eventsChan <- message
			}

			time.Sleep(10 * time.Second)
		}
	}
}
