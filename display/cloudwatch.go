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

	"google.golang.org/protobuf/encoding/protojson"
)

type CloudWatchFormatter struct {
	TextFormat bool
}

func (cf CloudWatchFormatter) Display(streamName string, message string) {
	event := &cloudwatchlog.LogEvent{}
	err := protojson.Unmarshal([]byte(message), event)
	if err != nil {
		log.Println(errors.Wrap(err, message))
		return
	}

	if cf.TextFormat {
		cf.formatText(event, streamName)
		return
	}

	cf.formatJSON(event, streamName)
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
