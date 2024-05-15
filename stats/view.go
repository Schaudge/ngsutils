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
	"strings"

	"github.com/Schaudge/ngsutils/db"
	"github.com/Schaudge/ngsutils/utils"
	"github.com/biogo/hts/bam"
)

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

// seekBamReader creates an io.ReadSeeker BAM reader from file.
func seekBamReader(bh *os.File) *bam.Reader {
	reader, err := bam.NewReader(io.ReadSeeker(bh), 1)
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
	fh, err := ioutil.ReadFile(baiFile) // auto open/close file
	panicError(err)
	reader, err := bam.ReadIndex(bytes.NewReader(fh))
	panicError(err)
	return reader
}

func BamViewOnRegion(bamFile string, id, start, end int) error {
	// standard utils for records seek on a special genome region
	bh, err := os.Open(bamFile) // close if open success
	panicError(err)
	defer func(bh *os.File) {
		err := bh.Close()
		if err != nil {

		}
	}(bh)
	bamReader := seekBamReader(bh)
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
	accession := strings.Split(filepath.Base(bamFile), "_")
	outBamFile := filepath.Dir(bamFile) + "/" + accession[0] + "_" + bpPair.Gene1 + "-" + bpPair.Gene2 + ".bam"
	defer func(bamFile string) {
		err := utils.CreateBamIndex(bamFile)
		if err != nil {

		}
	}(outBamFile)

	// standard utils for records seek on a special genome region
	bh, err := os.Open(bamFile) // close if open success
	panicError(err)
	defer func(bh *os.File) {
		err := bh.Close()
		if err != nil {

		}
	}(bh)
	bamReader := seekBamReader(bh)
	idx := createBaiReader(getBaiFromBamPath(bamFile))

	// output bam file settings
	ob, err := os.Create(outBamFile)
	panicError(err)
	bw, _ := bam.NewWriter(ob, bamReader.Header().Clone(), 1)
	defer func(ob *os.File) {
		err := ob.Close()
		if err != nil {

		}
	}(ob)
	defer func(bw *bam.Writer) {
		err := bw.Close()
		if err != nil {

		}
	}(bw)

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
				err := bw.Write(r)
				if err != nil {
					return err
				}
			}
		}
	}

	return bamReader.Close()
}
