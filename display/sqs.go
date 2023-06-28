package display

import (
	"fmt"
	"log"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
)

type SQSFormatter struct {
	TextFormat      bool
	Pattern         string
	ShouldFilter    bool
	Filter          Filter
	protoFile       string
	protoDescriptor *desc.MessageDescriptor
}

func NewSQSFormatter(sf SQSFormatter, protoFile string) *SQSFormatter {
	nsf := &sf
	if len(protoFile) == 0 {
		return nsf
	}

	parser := &protoparse.Parser{}
	descriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed_to_read_proto_file"))
	}

	if len(descriptors) < 1 {
		log.Fatal(errors.Wrap(err, "failed_to_parse_proto_file"))
	}

	nsf.protoFile = protoFile

	desc := descriptors[0].FindMessage("cloudwatchlog.LogEvent")
	if desc == nil {
		log.Fatal("couldnot find LogEvent messsage in cloudwatch package. please recheck proto file")
	}

	nsf.protoDescriptor = desc
	return nsf
}

func (sf *SQSFormatter) Display(queueName, message string) {
	event := DynamicCloudWatchLogEvent{
		"queue":   queueName,
		"message": message,
	}
	if !sf.ShouldFilter {
		fmt.Println(event)
		return
	}

	if sf.Filter(message, sf.Pattern) {
		fmt.Println(event)
	}
}

type DynamicSQSLogEvent map[string]string
