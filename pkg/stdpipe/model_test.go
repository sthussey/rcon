package stdpipe

import (
    "bytes"
    "context"
    "strings"
    "testing"
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

    <-drained

    out_string := out.String()

    if(out_string != testdata) {
        t.Errorf("Error running Pipeline: in(%s) != out(%s)", testdata, out_string)
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

    <-drained

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
