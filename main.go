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

package main

import (
	"fmt"
	"github.com/Schaudge/ngsutils/db"
	"github.com/Schaudge/ngsutils/stats"
	"os"
)

func main() {

	accession, bam := os.Args[1], os.Args[2]

	fmt.Println("Begin to start some ngs-utils process:")
	svbps := db.GetSvRecordsFromDB(accession)

	for _, sv := range svbps {
		fmt.Printf("Gene1: %s with break point %s:%d, Gene2: %s with break point %s:%d\n",
			sv.Gene1, sv.Chr1, sv.Bp1, sv.Gene2, sv.Chr2, sv.Bp2)
		err := stats.ExtractSvSamSet(bam, sv)
		if err != nil {
			return
		}
	}

}
