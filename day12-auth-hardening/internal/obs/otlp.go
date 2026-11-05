package obs

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OTLPExporter ships spans as OTLP/HTTP JSON (POST {endpoint}/v1/traces).
// Jaeger (all-in-one) and the OpenTelemetry Collector accept this on :4318.
// Export never blocks the request path: a full queue drops the span.
type OTLPExporter struct {
	url     string
	service string
	client  *http.Client
	q       chan SpanData
	done    chan struct{}
	wg      sync.WaitGroup
	m       *Metrics
	every   time.Duration
	batch   int
}

// NewOTLPExporter: endpoint like "http://jaeger:4318". Plain http is only
// meant for the compose network; use https endpoints anywhere else.
func NewOTLPExporter(endpoint, service string, m *Metrics) *OTLPExporter {
	return newOTLPExporter(endpoint, service, m, 2*time.Second, 64, 2048)
}

func newOTLPExporter(endpoint, service string, m *Metrics, every time.Duration, batch, queue int) *OTLPExporter {
	e := &OTLPExporter{
		url: strings.TrimRight(endpoint, "/") + "/v1/traces", service: service,
		client: &http.Client{Timeout: 3 * time.Second},
		q:      make(chan SpanData, queue), done: make(chan struct{}), m: m,
		every: every, batch: batch,
	}
	e.wg.Add(1)
	go e.loop()
	return e
}

func (e *OTLPExporter) Export(d SpanData) {
	select {
	case e.q <- d:
	default:
		if e.m != nil {
			e.m.SpansDropped.Inc()
		}
	}
}

func (e *OTLPExporter) loop() {
	defer e.wg.Done()
	t := time.NewTicker(e.every)
	defer t.Stop()
	var buf []SpanData
	flush := func() {
		if len(buf) > 0 {
			e.send(buf)
			buf = nil
		}
	}
	for {
		select {
		case d := <-e.q:
			buf = append(buf, d)
			if len(buf) >= e.batch {
				flush()
			}
		case <-t.C:
			flush()
		case <-e.done:
			for {
				select {
				case d := <-e.q:
					buf = append(buf, d)
				default:
					flush()
					return
				}
			}
		}
	}
}

// Shutdown flushes what is queued and stops the loop.
func (e *OTLPExporter) Shutdown(ctx context.Context) error {
	close(e.done)
	ch := make(chan struct{})
	go func() { e.wg.Wait(); close(ch) }()
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *OTLPExporter) send(spans []SpanData) {
	body := EncodeOTLP(e.service, spans)
	req, err := http.NewRequest(http.MethodPost, e.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := e.client.Do(req)
	ok := err == nil && res.StatusCode/100 == 2
	if res != nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<16))
		res.Body.Close()
	}
	if e.m != nil {
		if ok {
			e.m.OTLPExport.Inc("ok")
		} else {
			e.m.OTLPExport.Inc("error")
		}
	}
}

type otlpAttr struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

func toAttr(a Attr) otlpAttr {
	v := map[string]any{}
	switch x := a.Val.(type) {
	case string:
		v["stringValue"] = x
	case bool:
		v["boolValue"] = x
	case int:
		v["intValue"] = strconv.Itoa(x) // OTLP JSON: int64 as string
	case int64:
		v["intValue"] = strconv.FormatInt(x, 10)
	default:
		v["stringValue"] = fmt.Sprint(x)
	}
	return otlpAttr{Key: a.Key, Value: v}
}

// EncodeOTLP builds an ExportTraceServiceRequest in OTLP JSON encoding
// (ids are lowercase hex, 64-bit numbers are strings).
func EncodeOTLP(service string, spans []SpanData) []byte {
	type status struct {
		Code    int    `json:"code"`
		Message string `json:"message,omitempty"`
	}
	type span struct {
		TraceID string     `json:"traceId"`
		SpanID  string     `json:"spanId"`
		Parent  string     `json:"parentSpanId,omitempty"`
		Name    string     `json:"name"`
		Kind    int        `json:"kind"`
		Start   string     `json:"startTimeUnixNano"`
		End     string     `json:"endTimeUnixNano"`
		Attrs   []otlpAttr `json:"attributes,omitempty"`
		Status  status     `json:"status"`
	}
	out := make([]span, 0, len(spans))
	for _, d := range spans {
		s := span{
			TraceID: hex.EncodeToString(d.Context.TraceID[:]),
			SpanID:  hex.EncodeToString(d.Context.SpanID[:]),
			Name:    d.Name, Kind: int(d.Kind),
			Start:  strconv.FormatInt(d.Start.UnixNano(), 10),
			End:    strconv.FormatInt(d.End.UnixNano(), 10),
			Status: status{Code: 1},
		}
		if d.Parent != [8]byte{} {
			s.Parent = hex.EncodeToString(d.Parent[:])
		}
		for _, a := range d.Attrs {
			s.Attrs = append(s.Attrs, toAttr(a))
		}
		if d.ErrMsg != "" {
			s.Status = status{Code: 2, Message: d.ErrMsg}
		}
		out = append(out, s)
	}
	doc := map[string]any{"resourceSpans": []any{map[string]any{
		"resource": map[string]any{"attributes": []otlpAttr{toAttr(Attr{"service.name", service})}},
		"scopeSpans": []any{map[string]any{
			"scope": map[string]any{"name": "claim-tracker-lite/obs"},
			"spans": out,
		}},
	}}}
	b, _ := json.Marshal(doc)
	return b
}
