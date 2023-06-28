package display

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"rift/models/cloudwatchlog"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type CloudWatchFormatter struct {
	TextFormat   bool
	Pattern      string
	ShouldFilter bool
	Filter       Filter
	ProtoFile    string
	protoMessage *dynamic.Message
}

func (cf CloudWatchFormatter) Init(protoFile string) CloudWatchFormatter {
	ncf := cf

	parser := &protoparse.Parser{}
	descriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed_to_read_proto_file"))
	}

	if len(descriptors) < 1 {
		log.Fatal(errors.Wrap(err, "failed_to_parse_proto_file"))
	}

	ncf.ProtoFile = protoFile

	desc := descriptors[0].FindMessage("LogEvent")
	ncf.protoMessage = dynamic.NewMessage(desc)

	return ncf
}

func (cf CloudWatchFormatter) Display(streamName string, message string) {
	if cf.protoMessage == nil {
		log.Fatal("proto_file_missing")
	}

	vaiPlease := cf.protoMessage.(proto.Message)
	err := protojson.Unmarshal([]byte(message), vaiPlease)
	if err != nil {
		log.Println(errors.Wrap(err, message))
		return
	}

	formatter := cf.formatJSON
	if cf.TextFormat {
		formatter = cf.formatText
	}

	if !cf.ShouldFilter {
		formatter(event, streamName)
		return
	}

	if cf.Filter(message, cf.Pattern) {
		formatter(event, streamName)
	}

}

type CloudWatchJSONFormat struct {
	StreamName string `json:"stream"`
	*cloudwatchlog.LogEvent
}

func (cf CloudWatchFormatter) formatJSON(event *cloudwatchlog.LogEvent, streamName string) {
	f := &CloudWatchJSONFormat{StreamName: streamName, LogEvent: event}
	b, err := json.Marshal(f)
	if err != nil {
		log.Println("invalid json in log")
		return
	}

	fmt.Println(string(b))
}

func (cf CloudWatchFormatter) formatText(event *cloudwatchlog.LogEvent, streamName string) {
	space := " "
	displayText := bytes.Buffer{}
	displayText.WriteString(color.CyanString(streamName))
	displayText.WriteString(space)
	displayText.WriteString(color.CyanString(event.GetTime()))
	displayText.WriteString(space)
	displayText.WriteString(" request_id: ")
	displayText.WriteString(color.GreenString(event.GetRequestId()))
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
