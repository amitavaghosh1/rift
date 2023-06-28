package display

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

type CloudWatchFormatter struct {
	TextFormat      bool
	Pattern         string
	ShouldFilter    bool
	Filter          Filter
	protoFile       string
	protoDescriptor *desc.MessageDescriptor
}

func NewCloudWatchFormatter(cf CloudWatchFormatter, protoFile string) *CloudWatchFormatter {
	ncf := &cf

	parser := &protoparse.Parser{}
	descriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed_to_read_proto_file"))
	}

	if len(descriptors) < 1 {
		log.Fatal(errors.Wrap(err, "failed_to_parse_proto_file"))
	}

	ncf.protoFile = protoFile

	desc := descriptors[0].FindMessage("cloudwatchlog.LogEvent")
	if desc == nil {
		log.Fatal("couldnot find LogEvent messsage in cloudwatch package. please recheck proto file")
	}

	ncf.protoDescriptor = desc

	return ncf
}

func (cf *CloudWatchFormatter) Display(streamName string, message string) {
	md := dynamic.NewMessage(cf.protoDescriptor)

	err := md.UnmarshalJSON([]byte(message))
	if err != nil {
		log.Println(errors.Wrap(err, message))
		return
	}

	formatter := cf.formatJSON
	if cf.TextFormat {
		formatter = cf.formatText
	}

	event := DynamicCloudWatchLogEvent{}
	for _, field := range md.GetKnownFields() {
		fieldName := field.GetName()
		event[fieldName] = fmt.Sprintf("%v", md.GetFieldByName(fieldName))
	}

	if !cf.ShouldFilter {
		formatter(event, streamName)
		return
	}

	if cf.Filter(message, cf.Pattern) {
		formatter(event, streamName)
	}

}

type DynamicCloudWatchLogEvent map[string]string

// Some default methods
func (de DynamicCloudWatchLogEvent) GetTime() string {
	timeStr, ok := de["time"]
	if ok {
		return color.CyanString(timeStr)
	}
	return timeStr
}

func (de DynamicCloudWatchLogEvent) GetRequestId() string {
	requestID, ok := de["request_id"]
	if ok {
		return color.GreenString(requestID)
	}

	return requestID
}

func (de DynamicCloudWatchLogEvent) GetEnvironment() string {
	env, ok := de["environment"]
	if ok {
		return env
	}

	return "unparsed"
}

func (de DynamicCloudWatchLogEvent) GetLevel() string {
	level, ok := de["level"]
	if ok {
		return level
	}

	return "unknown"
}

func (de DynamicCloudWatchLogEvent) GetMsg() string {
	msg, ok := de["msg"]
	if ok {
		return msg
	}

	return ""
}

func (cf *CloudWatchFormatter) formatJSON(event DynamicCloudWatchLogEvent, streamName string) {
	eventWithStream := event
	eventWithStream["stream"] = streamName

	b, err := json.Marshal(eventWithStream)
	if err != nil {
		log.Println("invalid json in log")
		return
	}

	fmt.Println(string(b))
}

func (cf *CloudWatchFormatter) formatText(event DynamicCloudWatchLogEvent, streamName string) {
	space := " "
	displayText := bytes.Buffer{}
	displayText.WriteString(color.CyanString(streamName))
	displayText.WriteString(space)
	displayText.WriteString(event.GetTime())
	displayText.WriteString(space)
	displayText.WriteString(" request_id: ")
	displayText.WriteString(event.GetRequestId())
	displayText.WriteString(space)
	displayText.WriteString(" environment: ")
	displayText.WriteString(event.GetEnvironment())
	displayText.WriteString(space)

	displayText.WriteString(" level: ")
	level := event.GetLevel()

	levelColor := color.WhiteString
	switch strings.ToLower(level) {
	case "error":
		levelColor = color.RedString
	case "info":
		levelColor = color.GreenString
	case "warn":
		levelColor = color.YellowString
	default:
	}
	displayText.WriteString(levelColor(level))
	displayText.WriteString(space)
	displayText.WriteString(" msg:")
	displayText.WriteString(event.GetMsg())

	fmt.Println(displayText.String())
}
