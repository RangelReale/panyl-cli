package panylcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RangelReale/panyl/v2"
)

type DefaultOutput struct {
}

var _ panyl.Output = (*DefaultOutput)(nil)

func NewDefaultOutput() *DefaultOutput {
	return &DefaultOutput{}
}

func (o *DefaultOutput) OnItem(ctx context.Context, item *panyl.Item) (cont bool) {
	var out bytes.Buffer

	// level
	level := item.Metadata.StringValue(panyl.MetadataLevel)
	if level == "" {
		level = "unknown"
	}

	// timestamp
	if ts, ok := item.Metadata[panyl.MetadataTimestamp]; ok {
		out.WriteString(fmt.Sprintf("%s ", ts.(time.Time).Local().Format("2006-01-02 15:04:05.000")))
	}

	// application
	if application := item.Metadata.StringValue(panyl.MetadataApplication); application != "" {
		out.WriteString(fmt.Sprintf("| %s | ", application))
	}

	// level
	out.WriteString(fmt.Sprintf("[%s] ", level))

	// format
	if format := item.Metadata.StringValue(panyl.MetadataFormat); format != "" {
		out.WriteString(fmt.Sprintf("(%s) ", format))
	}

	// category
	if category := item.Metadata.StringValue(panyl.MetadataCategory); category != "" {
		out.WriteString(fmt.Sprintf("{{%s}} ", category))
	}

	// message
	if msg := item.Metadata.StringValue(panyl.MetadataMessage); msg != "" {
		out.WriteString(msg)
	} else if item.Line != "" {
		out.WriteString(item.Line)
	} else if len(item.Data) > 0 {
		dt, err := json.Marshal(item.Data)
		if err != nil {
			fmt.Printf("Error marshaling data to json: %s\n", err.Error())
			return
		}
		out.WriteString(fmt.Sprintf("| %s", string(dt)))
	}

	fmt.Println(out.String())
	return true
}

func (o *DefaultOutput) OnFlush(ctx context.Context) {}

func (o *DefaultOutput) OnClose(ctx context.Context) {}
