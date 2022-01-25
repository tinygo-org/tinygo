// The following is copied from Go 1.17 official implementation and
// modified to accommodate TinyGo.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestBuffers_read(t *testing.T) {
	const story = "once upon a time in Gopherland ... "
	buffers := Buffers{
		[]byte("once "),
		[]byte("upon "),
		[]byte("a "),
		[]byte("time "),
		[]byte("in "),
		[]byte("Gopherland ... "),
	}
	got, err := io.ReadAll(&buffers)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != story {
		t.Errorf("read %q; want %q", got, story)
	}
	if len(buffers) != 0 {
		t.Errorf("len(buffers) = %d; want 0", len(buffers))
	}
}

func TestBuffers_consume(t *testing.T) {
	tests := []struct {
		in      Buffers
		consume int64
		want    Buffers
	}{
		{
			in:      Buffers{[]byte("foo"), []byte("bar")},
			consume: 0,
			want:    Buffers{[]byte("foo"), []byte("bar")},
		},
		{
			in:      Buffers{[]byte("foo"), []byte("bar")},
			consume: 2,
			want:    Buffers{[]byte("o"), []byte("bar")},
		},
		{
			in:      Buffers{[]byte("foo"), []byte("bar")},
			consume: 3,
			want:    Buffers{[]byte("bar")},
		},
		{
			in:      Buffers{[]byte("foo"), []byte("bar")},
			consume: 4,
			want:    Buffers{[]byte("ar")},
		},
		{
			in:      Buffers{nil, nil, nil, []byte("bar")},
			consume: 1,
			want:    Buffers{[]byte("ar")},
		},
		{
			in:      Buffers{nil, nil, nil, []byte("foo")},
			consume: 0,
			want:    Buffers{[]byte("foo")},
		},
		{
			in:      Buffers{nil, nil, nil},
			consume: 0,
			want:    Buffers{},
		},
	}
	for i, tt := range tests {
		in := tt.in
		in.consume(tt.consume)
		if !reflect.DeepEqual(in, tt.want) {
			t.Errorf("%d. after consume(%d) = %+v, want %+v", i, tt.consume, in, tt.want)
		}
	}
}

func TestBuffers_WriteTo(t *testing.T) {
	for _, name := range []string{"WriteTo", "Copy"} {
		for _, size := range []int{0, 10, 1023, 1024, 1025} {
			t.Run(fmt.Sprintf("%s/%d", name, size), func(t *testing.T) {
				testBuffer_writeTo(t, size, name == "Copy")
			})
		}
	}
}

func testBuffer_writeTo(t *testing.T, chunks int, useCopy bool) {
	var want bytes.Buffer
	for i := 0; i < chunks; i++ {
		want.WriteByte(byte(i))
	}

	var b bytes.Buffer
	buffers := make(Buffers, chunks)
	for i := range buffers {
		buffers[i] = want.Bytes()[i : i+1]
	}
	var n int64
	var err error
	if useCopy {
		n, err = io.Copy(&b, &buffers)
	} else {
		n, err = buffers.WriteTo(&b)
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(buffers) != 0 {
		t.Fatal(fmt.Errorf("len(buffers) = %d; want 0", len(buffers)))
	}
	if n != int64(want.Len()) {
		t.Fatal(fmt.Errorf("Buffers.WriteTo returned %d; want %d", n, want.Len()))
	}
	all, err := io.ReadAll(&b)
	if !bytes.Equal(all, want.Bytes()) || err != nil {
		t.Fatal(fmt.Errorf("read %q, %v; want %q, nil", all, err, want.Bytes()))
	}
}
