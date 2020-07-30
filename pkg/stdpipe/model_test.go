package stdpipe

import (
    "bytes"
    "context"
    "strings"
    "testing"
    "time"
)

func TestEmptyPipeline(t *testing.T) {
    testdata := "testing123"

    in := bytes.NewBufferString(testdata)
    out := bytes.NewBuffer(make([]byte,0,32))

    pipe, err := NewPipeline(8)

    if (err != nil) {
        t.Errorf("Error creating Pipeline: %v", err)
    }

    drained := pipe.Process(context.Background(), in, out)

    d := <-drained

    if (!d) {
	t.Errorf("Error running Pipeline: pipeline not fully drained.\n")
    }

    if(out.String() != testdata) {
        t.Errorf("Error running Pipeline: in(%s) != out(%s)", testdata, out.String())
    }
}

func TestPipeSegment(t *testing.T) {
    testdata := "testing123"

    in := bytes.NewBufferString(testdata)
    out := bytes.NewBuffer(make([]byte,0,32))

    pipe, err := NewPipeline(8)

    if(err != nil) {
        t.Errorf("Error creating Pipeline: %v", err)
    }

    err = pipe.Add(NewUpperCaser())

    if(err != nil) {
        t.Errorf("Error adding PipeSegment: %v", err)
    }

    drained := pipe.Process(context.Background(), in, out)

    d := <-drained

    if (!d) {
	t.Errorf("Error running Pipeline: pipeline not fully drained.\n")
    }

    if(out.String() != strings.ToUpper(testdata)){
        t.Errorf("Error with UpperCaser PipeSegment: in(%s) !=> out(%s)", testdata, out.String())
    }
}

func NewUpperCaser() PipeSegment {
    return GenericPipeSegment{processor: makeUpper}
}

func makeUpper(in []byte) ([]byte, error) {
    return bytes.ToUpper(in), nil
}
func TestPipelineCancel(t *testing.T) {
    testdata := "testing123"

    in := bytes.NewBufferString(testdata)
    out := bytes.NewBuffer(make([]byte,0,32))

    pipe, err := NewPipeline(8)

    if(err != nil) {
        t.Errorf("Error creating Pipeline: %v", err)
    }

    err = pipe.Add(NewSleepyEcho())

    if(err != nil) {
        t.Errorf("Error adding PipeSegment: %v", err)
    }

    ctx, canceller := context.WithCancel(context.Background())

    drained := pipe.Process(ctx, in, out)

    canceller()

    d := <-drained

    if (d) {
	t.Errorf("Error cancelling Pipeline: pipeline fully drained.\n")
    }
}

func NewSleepyEcho() PipeSegment {
	return GenericPipeSegment{processor: sleepyEcho}
}

func sleepyEcho(in []byte) ([]byte, error) {
	t, _ := time.ParseDuration("5s")
	time.Sleep(t)
	return in, nil
}
