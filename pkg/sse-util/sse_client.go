package sse_util

import (
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

const DONE_MSG = "data: [DONE]\n"

// SSEWriter designs sse writer to standardize and unify standard output methods (all sse returns can be used), and at the same time decouple it from the business as much as possible
type SSEWriter struct {
	client  *gin.Context
	label   string // Used for tags in SSE logs
	doneMsg string // When SSE ends, the end message sent to the front end will not be sent if it is empty; usually "data: [DONE]\n"
}

func NewSSEWriter(c *gin.Context, label, doneMsg string) *SSEWriter {
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	return &SSEWriter{
		client:  c,
		label:   label,
		doneMsg: doneMsg,
	}
}

// WriteStream streaming writing, identifying channels and writing to the front end in a loop
func (sw *SSEWriter) WriteStream(sseCh <-chan string, streamContextParams interface{},
	lineBuilder func(*gin.Context, string, interface{}) (string, bool, error),
	doneProcessor func(*gin.Context, interface{}) error) error {
	for s := range sseCh {
		var lineText = s
		if lineBuilder != nil {
			line, skip, err := lineBuilder(sw.client, s, streamContextParams)
			if err != nil {
				log.Errorf("[SSE]%v line %v build err: %v", sw.label, err)
				return err
			}
			if skip {
				continue
			}
			lineText = line
		}
		if err := sw.WriteLine(lineText, false, streamContextParams, doneProcessor); err != nil {
			return err
		}
	}
	return sw.WriteLine("", true, streamContextParams, doneProcessor)
}

// WriteLine writes a line to the client
func (sw *SSEWriter) WriteLine(lineText string, done bool, streamProcessParams interface{},
	doneProcessor func(*gin.Context, interface{}) error) error {

	var err error
	defer func() {
		if err != nil {
			log.Errorf("[SSE]%v err: %v", sw.label, err)
		} else if done {
			log.Debugf("[SSE]%v done", sw.label)
		} else {
			return
		}
		// err or done execute doneProcessor
		if doneProcessor != nil {
			if err := doneProcessor(sw.client, streamProcessParams); err != nil {
				log.Errorf("[SSE]%v doneProcessor err: %v", sw.label, err)
			}
		}
	}()

	if done {
		lineText = fmt.Sprintf("%v%v", lineText, sw.doneMsg)
	}
	// Write data
	log.Debugf("[SSE]%v write: %v", sw.label, lineText)
	_, err = sw.client.Writer.Write([]byte(lineText))
	if err != nil {
		err = fmt.Errorf("connection closed by web: %v", err)
		return err
	}
	sw.client.Writer.Flush()
	return nil
}
