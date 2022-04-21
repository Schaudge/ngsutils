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
	"github.com/biogo/hts/sam"
)

var bamReader *bam.Reader

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
		panic("Would not find a proper bai file for the input utils!")
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

func BamViewOnRegion(bamFile string, id, start, end int) error {
	// standard utils content seek for a special genome region
	bamReader = seekBamReader(bamFile)
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

// WriteSamSetOut writes all sam records sets into a bam file.
func WriteSamSetOut(outBamPath string, outSamSet []*sam.Record) error {
	fh, _ := os.Open(outBamPath)
	bw, _ := bam.NewWriter(fh, bamReader.Header().Clone(), 1)
	for _, record := range outSamSet {
		bw.Write(record)
	}
	return bw.Close()
}

// ExtractSvSamSet extract all break point context sam records
func ExtractSvSamSet(bamFile string, chr1, bp1, chr2, bp2 int) error {
	// standard utils content seek for a special genome region
	bamReader = seekBamReader(bamFile)
	idx := createBaiReader(getBaiFromBamPath(bamFile))
	if chr1 > chr2 {
		chr1, chr2 = chr2, chr1
	} else if chr1 == chr2 && bp1 > bp2 {
		bp1, bp2 = bp2, bp1
	}
	associatedBpPair := [][]int{
		[]int{chr1, bp1, chr2, bp2},
		[]int{chr2, bp2, chr1, bp1},
	}

	for _, obp := range associatedBpPair {
		ref := bamReader.Header().Refs()[obp[0]]
		chunks, err := idx.Chunks(ref, obp[1]-500, obp[1]+500)
		panicError(err)
		i, err := bam.NewIterator(bamReader, chunks)
		panicError(err)
		for i.Next() {
			r := i.Record()
			if r.Ref.ID() == r.MateRef.ID() && r.TempLen < 1000 {
				continue
			}
			sam, _ := i.Record().MarshalText()
			fmt.Printf("%s\n", sam)
		}
	}

	return bamReader.Close()
}
