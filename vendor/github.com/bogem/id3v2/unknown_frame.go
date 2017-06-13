// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"io"

	"github.com/bogem/id3v2/util"
)

// UnknownFrame is used for frames, which id3v2 so far doesn't know how to
// parse and write it. It just contains an unparsed byte body of the frame.
type UnknownFrame struct {
	body []byte
}

func (uk UnknownFrame) Size() int {
	return len(uk.body)
}

func (uk UnknownFrame) WriteTo(w io.Writer) (n int64, err error) {
	var i int
	i, err = w.Write(uk.body)
	return int64(i), err
}

func parseUnknownFrame(rd io.Reader) (Framer, error) {
	body, err := util.ReadAll(rd)
	return UnknownFrame{body: body}, err
}
