package transaction

import (
	m "github.com/elastic/apm-server/processor/model"
	"github.com/elastic/apm-server/utility"
	"github.com/elastic/beats/libbeat/common"
)

type Trace struct {
	Id               *int               `json:"id"`
	Name             string             `json:"name"`
	Type             string             `json:"type"`
	Start            float64            `json:"start"`
	Duration         float64            `json:"duration"`
	StacktraceFrames m.StacktraceFrames `json:"stacktrace"`
	Context          common.MapStr      `json:"context"`
	Parent           *int               `json:"parent"`

	TransformStacktrace m.TransformStacktrace
}

func (t *Trace) DocType() string {
	return "trace"
}

func (t *Trace) Transform(transactionId string) common.MapStr {
	enhancer := utility.NewMapStrEnhancer()
	tr := common.MapStr{}
	enhancer.Add(tr, "id", t.Id)
	enhancer.Add(tr, "transaction_id", transactionId)
	enhancer.Add(tr, "name", t.Name)
	enhancer.Add(tr, "type", t.Type)
	enhancer.Add(tr, "start", utility.MillisAsMicros(t.Start))
	enhancer.Add(tr, "duration", utility.MillisAsMicros(t.Duration))
	enhancer.Add(tr, "parent", t.Parent)
	st := t.transformStacktrace()
	if len(st) > 0 {
		enhancer.Add(tr, "stacktrace", st)
	}
	return tr
}

func (t *Trace) Mappings(pa *Payload, tx Event) (string, []m.SMapping, []m.FMapping) {
	return tx.Timestamp, []m.SMapping{
			{Key: "processor.name", Value: processorName},
			{Key: "processor.event", Value: t.DocType()},
		}, []m.FMapping{
			{Key: t.DocType(), Apply: func() common.MapStr { return t.Transform(tx.Id) }},
			{Key: "context", Apply: func() common.MapStr { return t.Context }},
			{Key: "context.app", Apply: pa.App.MinimalTransform},
		}
}

func (t *Trace) transformStacktrace() []common.MapStr {
	if t.TransformStacktrace == nil {
		t.TransformStacktrace = (*m.Stacktrace).Transform
	}
	st := m.Stacktrace{Frames: t.StacktraceFrames}
	return t.TransformStacktrace(&st)
}
