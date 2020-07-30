package stdpipe

import (
	"context"
	"fmt"
	"io"
)

const BUF_MIN = 0
const BUF_MAX = 1024 * 1024

type PipeSegment interface {
	Start(ctx context.Context, in <-chan []byte, out chan<- []byte)
}

type GenericPipeSegment struct {
	processor func(in []byte) ([]byte, error)
}

type Pipeline interface {
	Add(s PipeSegment) error
	SetBufferSize(s int) error
	Process(ctx context.Context, source io.Reader, sink io.Writer) chan bool
	Feed(ctx context.Context, out chan<- []byte)
	Drain(ctx context.Context, in <-chan []byte, drained chan bool)
}

type PipelineWrapper struct {
    *GenericPipeline
}

type GenericPipeline struct {
	source   io.Reader
	sink     io.Writer
	segments []PipeSegment
	bufSize  int
}

func NewPipeline(bufSize int) (Pipeline, error) {
	if bufSize == 0 {
		bufSize = 1024
	}

	if bufSize < BUF_MIN || bufSize > BUF_MAX {
		return nil, fmt.Errorf("Invalid buffer size %d.", bufSize)
	}

	gp := GenericPipeline{bufSize: bufSize, segments: make([]PipeSegment, 0)}
	return PipelineWrapper{GenericPipeline: &gp}, nil
}

func (p *GenericPipeline) Add(s PipeSegment) error {
	p.segments = append(p.segments, s)
	return nil
}

func (p *GenericPipeline) SetBufferSize(s int) error {
	p.bufSize = s
	return nil
}

func (p *GenericPipeline) Process(ctx context.Context, source io.Reader, sink io.Writer) chan bool {
	p.source = source
	p.sink = sink
	drained := make(chan bool)

	if len(p.segments) == 0 {
		/* Channel buffer size should be configurable */
		c := make(chan []byte, 1)
		go p.Feed(ctx, c)
		go p.Drain(ctx, c, drained)
	} else {
		var in chan []byte = nil
		var out chan []byte = nil

		for i, s := range p.segments {
			if i == 0 {
				in = make(chan []byte, 1)
				go p.Feed(ctx, in)
			}
			out = make(chan []byte, 1)
			go s.Start(ctx, in, out)
			in = out
		}
		go p.Drain(ctx, out, drained)
	}

	return drained
}

func (p GenericPipeline) Feed(ctx context.Context, c chan<- []byte) {
	for {
		buf := make([]byte, p.bufSize)
		l, err := p.source.Read(buf)
		if err == io.EOF {
			fmt.Printf("Source EOF, starting pipeline shutdown.\n")
			break
		} else if err != nil {
			fmt.Printf("Error reading source - %v\n", err)
		} else {
			select {
				case <-ctx.Done():
					fmt.Printf("Context cancelled (%v), closing pipeline.\n", ctx.Err())
					break
				case c <- buf[0:l]:
			}
		}
	}
	close(c)
}

func (p GenericPipeline) Drain(ctx context.Context, c <-chan []byte, drained chan bool) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Context cancelled (%v), closing pipeline.\n", ctx.Err())
			drained <- false
			break
		case buf := <-c:
			if buf == nil {
				fmt.Printf("Ingress channel closed, pipeline fully drained. Exiting.\n")
				drained <- true
				break
			}
			_, err := p.sink.Write(buf)
			if err != nil {
				fmt.Printf("Error writing to sink - %v\n", err)
			}
		}
	}
	close(drained)
}

func (ps GenericPipeSegment) Start(ctx context.Context, in <-chan []byte, out chan<- []byte) {
	for {
		select {
			case <-ctx.Done():
				fmt.Printf("Context cancelled(%v), closing pipeline segment.\n", ctx.Err())
				close(out)
				return
			case buf, ok := <-in:
				if !ok {
					fmt.Printf("Upstream pipe closed, propagating shutdown.\n")
					close(out)
					return
				} else {
					var err error
					buf, err = ps.processor(buf)
					if err != nil {
						fmt.Printf("Error running processor: %v\n", err)
						continue
					}
					out <- buf
				}
		}

	}
	close(out)
}
