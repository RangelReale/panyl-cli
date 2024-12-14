package panylcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RangelReale/panyl"
)

type Output struct {
}

func NewOutput() *Output {
	return &Output{}
}

func (o *Output) OnResult(p *panyl.Process) (cont bool) {
	var out bytes.Buffer

	// level
	level := p.Metadata.StringValue(panyl.MetadataLevel)
	if level == "" {
		level = "unknown"
	}

	// timestamp
	if ts, ok := p.Metadata[panyl.MetadataTimestamp]; ok {
		out.WriteString(fmt.Sprintf("%s ", ts.(time.Time).Local().Format("2006-01-02 15:04:05.000")))
	}

	// application
	if application := p.Metadata.StringValue(panyl.MetadataApplication); application != "" {
		out.WriteString(fmt.Sprintf("| %s | ", application))
	}

	// level
	if level != "" {
		out.WriteString(fmt.Sprintf("[%s] ", level))
	}

	// format
	if format := p.Metadata.StringValue(panyl.MetadataFormat); format != "" {
		out.WriteString(fmt.Sprintf("(%s) ", format))
	}

	// category
	if category := p.Metadata.StringValue(panyl.MetadataCategory); category != "" {
		out.WriteString(fmt.Sprintf("{{%s}} ", category))
	}

	// message
	if msg := p.Metadata.StringValue(panyl.MetadataMessage); msg != "" {
		out.WriteString(msg)
	} else if p.Line != "" {
		out.WriteString(p.Line)
	} else if len(p.Data) > 0 {
		dt, err := json.Marshal(p.Data)
		if err != nil {
			fmt.Printf("Error marshaling data to json: %s\n", err.Error())
			return
		}
		out.WriteString(fmt.Sprintf("| %s", string(dt)))
	}

	fmt.Println(out.String())
	return true
}

func (o *Output) OnFlush() {}

func (o *Output) OnClose() {}
