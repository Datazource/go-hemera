package hemera

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
)

const (
	// RequestType represent the request with default request / reply semantic
	RequestType = "request"
	// PubsubType represent the request with publish / subscribe semantic
	PubsubType = "pubsub"
	// RequestTimeout is the maxiumum act timeout in miliseconds
	RequestTimeout = 2000
)

var (
	ErrAddTopicRequired = errors.New("Topic is required")
	ErrActTopicRequired = errors.New("Topic is required")
)

func GetDefaultOptions() Options {
	opts := Options{
		Timeout: RequestTimeout,
	}
	return opts
}

// Option is a function on the options for hemera
type Option func(*Options) error

type Options struct {
	Timeout time.Duration
}

// Reply is function type to represent the callback handler
type Reply func(interface{})
type addHandler func(Pattern, Reply)
type actHandler func(ClientResult)

//Pattern the default struct to represent a pattern
type Pattern map[string]interface{}

// Hemera is the main struct
type Hemera struct {
	Conn *nats.Conn
	Opts Options
}

type request struct {
	ID          string `json:"id"`
	RequestType string `json:"type"`
}

type ClientResult interface{}

// Error is the default error struct
type Error struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Code    int16  `json:"code"`
}

type trace struct {
	TraceID      string `json:"traceId"`
	ParentSpanID string `json:"parentSpanId"`
	SpanID       string `json:"spanId"`
	Timestamp    int64  `json:"timestamp"`
	Service      string `json:"service"`
	Method       string `json:"method"`
	Duration     int64  `json:"duration"`
}

type packet struct {
	Pattern  Pattern                `json:"pattern"`
	Meta     map[string]interface{} `json:"meta"`
	Delegate map[string]interface{} `json:"delegate"`
	Result   interface{}            `json:"result"`
	Trace    trace                  `json:"trace"`
	Request  request                `json:"request"`
	Error    *Error                 `json:"error"`
}

// New create a new Hemera struct
func NewHemera(conn *nats.Conn, options ...Option) (Hemera, error) {
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return Hemera{Opts: opts}, err
		}
	}
	return Hemera{Conn: conn, Opts: opts}, nil
}

// Timeout is an Option to set the timeout for a act request
func Timeout(t time.Duration) Option {
	return func(o *Options) error {
		o.Timeout = t
		return nil
	}
}

// Add is a method to subscribe on a specific topic
func (h *Hemera) Add(p Pattern, handler addHandler) (bool, error) {
	topic, ok := p["topic"].(string)

	if !ok {
		log.Fatal("Topic is required in Add definition")
		return false, ErrAddTopicRequired
	}

	h.Conn.QueueSubscribe(topic, topic, func(m *nats.Msg) {
		pack := packet{}
		json.Unmarshal(m.Data, &pack)

		handler(pack.Pattern, func(payload interface{}) {
			response := packet{
				Pattern: p,
				Request: request{
					ID:          nuid.Next(),
					RequestType: RequestType,
				},
			}

			// Check if error or message was passed
			he, ok := payload.(Error)
			if ok {
				response.Error = &he
			} else {
				response.Result = payload
			}
			// Encode to JSON
			data, _ := json.Marshal(&response)
			// Send
			h.Conn.Publish(m.Reply, data)
		})
	})

	return true, nil
}

// Act is a method to send a message to a NATS subscriber which the specific topic
func (h *Hemera) Act(p Pattern, handler actHandler) (bool, error) {

	topic, ok := p["topic"].(string)

	if !ok {
		log.Fatal("Topic is required in Act call")
		return false, ErrActTopicRequired
	}

	request := packet{
		Pattern: p,
		Request: request{
			ID:          nuid.Next(),
			RequestType: RequestType,
		},
	}

	data, _ := json.Marshal(&request)
	m, err := h.Conn.Request(topic, data, h.Opts.Timeout*time.Millisecond)

	if err != nil {
		log.Fatal("Act could not be executed")
		return false, err
	}

	pack := packet{}
	mErr := json.Unmarshal(m.Data, &pack)

	if mErr != nil {
		log.Fatal("Act response could not be unmarshalled")
		return false, err
	}

	if pack.Error != nil {
		handler(pack.Error)
	} else {
		handler(pack.Result)
	}

	return true, nil
}
