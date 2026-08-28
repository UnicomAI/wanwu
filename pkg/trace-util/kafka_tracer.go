// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace_util

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"

	"github.com/IBM/sarama"

	"go.opentelemetry.io/otel/codes"

	"github.com/dnwe/otelsarama"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type SyncProducer struct {
	sarama.SyncProducer
	cfg          config
	saramaConfig *sarama.Config
}

// SendMessage calls sarama.SyncProducer.SendMessage and traces the request.
func (p *SyncProducer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	span := startProducerSpan(ctx, p.cfg, p.saramaConfig.Version, msg)
	partition, offset, err = p.SyncProducer.SendMessage(msg)
	finishProducerSpan(span, partition, offset, err)
	return partition, offset, err
}

// SendMessages calls sarama.SyncProducer.SendMessages and traces the requests.
func (p *SyncProducer) SendMessages(ctx context.Context, msgs []*sarama.ProducerMessage) error {
	// Although there's only one call made to the SyncProducer, the messages are
	// treated individually, so we create a span for each one
	spans := make([]trace.Span, len(msgs))
	for i, msg := range msgs {
		spans[i] = startProducerSpan(ctx, p.cfg, p.saramaConfig.Version, msg)
	}
	err := p.SyncProducer.SendMessages(msgs)
	for i, span := range spans {
		finishProducerSpan(span, msgs[i].Partition, msgs[i].Offset, err)
	}
	return err
}

// WrapSyncProducer wraps a sarama.SyncProducer so that all produced messages
// are traced.
func WrapSyncProducer(saramaConfig *sarama.Config, producer sarama.SyncProducer, opts ...Option) *SyncProducer {
	cfg := newConfig(opts...)
	if saramaConfig == nil {
		saramaConfig = sarama.NewConfig()
	}

	return &SyncProducer{
		SyncProducer: producer,
		cfg:          cfg,
		saramaConfig: saramaConfig,
	}
}

type asyncProducer struct {
	sarama.AsyncProducer
	input         chan *sarama.ProducerMessage
	successes     chan *sarama.ProducerMessage
	errors        chan *sarama.ProducerError
	closeErr      chan error
	closeSig      chan struct{}
	closeAsyncSig chan struct{}
}

// Input returns the input channel.
func (p *asyncProducer) Input() chan<- *sarama.ProducerMessage {
	return p.input
}

// Successes returns the successes channel.
func (p *asyncProducer) Successes() <-chan *sarama.ProducerMessage {
	return p.successes
}

// Errors returns the errors channel.
func (p *asyncProducer) Errors() <-chan *sarama.ProducerError {
	return p.errors
}

// AsyncClose async close producer.
func (p *asyncProducer) AsyncClose() {
	close(p.input)
	close(p.closeAsyncSig)
}

// Close shuts down the producer and waits for any buffered messages to be
// flushed.
//
// Due to the implement of sarama, some messages may lose successes or errors status
// while closing.
func (p *asyncProducer) Close() error {
	close(p.input)
	close(p.closeSig)
	return <-p.closeErr
}

type producerMessageContext struct {
	span           trace.Span
	metadataBackup interface{}
}

// WrapAsyncProducer wraps a sarama.AsyncProducer so that all produced messages
// are traced. It requires the underlying sarama Config, so we can know whether
// or not successes will be returned.
//
// If `Return.Successes` is false, there is no way to know partition and offset of
// the message.
func WrapAsyncProducer(saramaConfig *sarama.Config, p sarama.AsyncProducer, opts ...Option) sarama.AsyncProducer {
	cfg := newConfig(opts...)
	if saramaConfig == nil {
		saramaConfig = sarama.NewConfig()
	}

	wrapped := &asyncProducer{
		AsyncProducer: p,
		input:         make(chan *sarama.ProducerMessage),
		successes:     make(chan *sarama.ProducerMessage),
		errors:        make(chan *sarama.ProducerError),
		closeErr:      make(chan error),
		closeSig:      make(chan struct{}),
		closeAsyncSig: make(chan struct{}),
	}

	var (
		mtx                     sync.Mutex
		producerMessageContexts = make(map[interface{}]producerMessageContext)
	)

	// Spawn Input producer goroutine.
	go func() {
		for {
			select {
			case <-wrapped.closeSig:
				wrapped.closeErr <- p.Close()
				return
			case <-wrapped.closeAsyncSig:
				p.AsyncClose()
				return
			case msg, ok := <-wrapped.input:
				if !ok {
					continue // wait for closeAsyncSig
				}
				span := startProducerSpan(context.Background(), cfg, saramaConfig.Version, msg)

				// Create message context, backend message metadata
				mc := producerMessageContext{
					metadataBackup: msg.Metadata,
					span:           span,
				}

				// Remember metadata using span ID as a cache key
				msg.Metadata = span.SpanContext().SpanID()
				if saramaConfig.Producer.Return.Successes {
					mtx.Lock()
					producerMessageContexts[msg.Metadata] = mc
					mtx.Unlock()
				} else {
					// If returning successes isn't enabled, we just finish the
					// span right away because there's no way to know when it will
					// be done.
					mc.span.End()
				}

				p.Input() <- msg
			}
		}
	}()

	// Sarama will consume all the successes and errors by itself while closing,
	// so we need to end these spans ourselves.
	var cleanupWg sync.WaitGroup

	// Spawn Successes consumer goroutine.
	cleanupWg.Add(1)
	go func() {
		defer func() {
			close(wrapped.successes)
			cleanupWg.Done()
		}()
		for msg := range p.Successes() {
			key := msg.Metadata
			mtx.Lock()
			if mc, ok := producerMessageContexts[key]; ok {
				delete(producerMessageContexts, key)
				finishProducerSpan(mc.span, msg.Partition, msg.Offset, nil)
				msg.Metadata = mc.metadataBackup // Restore message metadata
			}
			mtx.Unlock()
			wrapped.successes <- msg
		}
	}()

	// Spawn Errors consumer goroutine.
	cleanupWg.Add(1)
	go func() {
		defer func() {
			close(wrapped.errors)
			cleanupWg.Done()
		}()
		for errMsg := range p.Errors() {
			key := errMsg.Msg.Metadata
			mtx.Lock()
			if mc, ok := producerMessageContexts[key]; ok {
				delete(producerMessageContexts, key)
				finishProducerSpan(mc.span, errMsg.Msg.Partition, errMsg.Msg.Offset, errMsg.Err)
				errMsg.Msg.Metadata = mc.metadataBackup // Restore message metadata
			}
			mtx.Unlock()
			wrapped.errors <- errMsg
		}
	}()

	// Spawn spans cleanup goroutine.
	go func() {
		// wait until both consumer goroutines are closed
		cleanupWg.Wait()
		// end all remaining spans
		mtx.Lock()
		for _, mc := range producerMessageContexts {
			mc.span.End()
		}
		mtx.Unlock()
	}()

	return wrapped
}

// msgPayloadSize returns the approximate estimate of message size in bytes.
//
// For kafka version <= 0.10, the message size is
// ~ 4(crc32) + 8(timestamp) + 4(key len) + 4(value len) + 4(message len) + 1(attrs) + 1(magic).
//
// For kafka version >= 0.11, the message size with varint encoding is
// ~ 5 * (crc32, key len, value len, message len, attrs) + timestamp + 1 byte (magic).
// + header key + header value + header key len + header value len.
func msgPayloadSize(msg *sarama.ProducerMessage, kafkaVersion sarama.KafkaVersion) int {
	maximumRecordOverhead := 5*binary.MaxVarintLen32 + binary.MaxVarintLen64 + 1
	producerMessageOverhead := 26
	version := 1
	if kafkaVersion.IsAtLeast(sarama.V0_11_0_0) {
		version = 2
	}
	size := producerMessageOverhead
	if version >= 2 {
		size = maximumRecordOverhead
		for _, h := range msg.Headers {
			size += len(h.Key) + len(h.Value) + 2*binary.MaxVarintLen32
		}
	}
	if msg.Key != nil {
		size += msg.Key.Length()
	}
	if msg.Value != nil {
		size += msg.Value.Length()
	}
	return size
}

func startProducerSpan(ctx context.Context, cfg config, version sarama.KafkaVersion, msg *sarama.ProducerMessage) trace.Span {
	// If there's a span context in the message, use that as the parent context.
	carrier := otelsarama.NewProducerMessageCarrier(msg)
	ctx = cfg.Propagators.Extract(ctx, carrier)

	// Create a span.
	attrs := []attribute.KeyValue{
		semconv.MessagingSystem("kafka"),
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingMessagePayloadSizeBytes(msgPayloadSize(msg, version)),
		semconv.MessagingOperationPublish,
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}
	ctx, span := cfg.Tracer.Start(ctx, fmt.Sprintf("%s publish", msg.Topic), opts...)

	if version.IsAtLeast(sarama.V0_11_0_0) {
		// Inject current span context, so consumers can use it to propagate span.
		cfg.Propagators.Inject(ctx, carrier)
	}

	return span
}

func finishProducerSpan(span trace.Span, partition int32, offset int64, err error) {
	span.SetAttributes(
		semconv.MessagingMessageID(strconv.FormatInt(offset, 10)),
		semconv.MessagingKafkaDestinationPartition(int(partition)),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
