package protocol

import (
	"io"
)

type Encoder struct {
	w io.Writer
}

type Decoder struct {
	r io.Reader
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (e *Encoder) Encode(v any) error {
	return nil
}

func (d *Decoder) Decode(v any) error {
	return nil
}
