// Copyright © 2022 Schaudge King <yuanshenran@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package stats

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/biogo/hts/bam"
)

func panicError(err error) {
	if err != nil {
		panic(err)
		os.Exit(-1)
	}
}

// seekBamReader creates an io.ReadSeeker BAM reader from file.
func seekBamReader(bamFile string) *bam.Reader {
	fh, err := os.Open(bamFile)
	panicError(err)
	reader, err := bam.NewReader(io.ReadSeeker(fh), 1)
	panicError(err)
	return reader
}

// getProperBai given the proper bai path
func getBaiFromBamPath(bamFile string) string {
	baiFile := bamFile + ".bai"
	if _, err := os.Stat(baiFile); err == nil {
		return baiFile
	}
	baiFile = bamFile[:len(bamFile)-4] + ".bai"
	if _, err := os.Stat(baiFile); err == nil {
		return baiFile
	} else {
		panic("Would not find a proper bai file for the input bam!")
	}
}

// createBaiReader creates a BAI reader from file path.
func createBaiReader(baiFile string) *bam.Index {
	fh, err := ioutil.ReadFile(baiFile)
	panicError(err)
	reader, err := bam.ReadIndex(bytes.NewReader(fh))
	panicError(err)
	return reader
}

func ViewOnRegion(bamFile string, id, start, end int) error {
	// standard bam content seek for a special genome region
	bamReader := seekBamReader(bamFile)
	idx := createBaiReader(getBaiFromBamPath(bamFile))

	ref := bamReader.Header().Refs()[id]
	chunks, err := idx.Chunks(ref, start, end)
	panicError(err)
	i, err := bam.NewIterator(bamReader, chunks)
	panicError(err)
	for i.Next() {
		sam, _ := i.Record().MarshalText()
		fmt.Printf("%s\n", sam)
	}
	return i.Close()
}
