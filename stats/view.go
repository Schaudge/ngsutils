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
	"path/filepath"

	"github.com/Schaudge/ngsutils/db"
	"github.com/Schaudge/ngsutils/utils"
	"github.com/biogo/hts/bam"
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

// ExtractSvSamSet extract all break point context sam records
func ExtractSvSamSet(bamFile string, bpPair db.SvBpPair) error {
	// standard utils content seek for a special genome region
	bamReader = seekBamReader(bamFile)
	idx := createBaiReader(getBaiFromBamPath(bamFile))

	// output bam file settings
	outBam, err := os.Create(filepath.Dir(bamFile) + "/" + bpPair.Gene1 + "-" + bpPair.Gene2 + ".bam")
	panicError(err)
	bw, _ := bam.NewWriter(outBam, bamReader.Header().Clone(), 1)
	defer bw.Close()

	chr1, chr2 := utils.CtgName2Id(bpPair.Chr1), utils.CtgName2Id(bpPair.Chr2)
	orderedBpPair := [][]int{
		[]int{chr1, bpPair.Bp1, chr2, bpPair.Bp2},
		[]int{chr2, bpPair.Bp2, chr1, bpPair.Bp1},
	}
	if chr1 > chr2 || (chr1 == chr2 && bpPair.Bp1 > bpPair.Bp2) {
		orderedBpPair[0], orderedBpPair[1] = orderedBpPair[1], orderedBpPair[0]
	}

	for _, bp := range orderedBpPair {
		ref := bamReader.Header().Refs()[bp[0]]
		chunks, err := idx.Chunks(ref, bp[1]-500, bp[1]+500)
		panicError(err)
		i, err := bam.NewIterator(bamReader, chunks)
		panicError(err)
		for i.Next() {
			r := i.Record()
			if r.MateRef.ID() == bp[2] && bp[3]-500 < r.MatePos && r.MatePos < bp[3]+500 {
				//				sam, _ := i.Record().MarshalText()
				//				fmt.Printf("%s\n", sam)
				bw.Write(r)
			}
		}
	}

	return bamReader.Close()
}
