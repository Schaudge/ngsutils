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

package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type SvBpPair struct {
	Chr1  string `db:"CHROM1"`
	Bp1   int    `db:"BP1"`
	Gene1 string `db:"GENE1"`
	Chr2  string `db:"CHROM2"`
	Bp2   int    `db:"BP2"`
	Gene2 string `db:"GENE2"`
}

var mysqlDB *sql.DB

func GetSvRecordsFromDB(accession string) []SvBpPair {
	defer mysqlDB.Close()
	rows, _ := mysqlDB.Query("SELECT `CHROM1`, `BP1`, `GENE1`, `CHROM2`, `BP2`, `GENE2` FROM sv_mutation"+
		" WHERE SAMPLE_ID = ?", accession)
	var svBpSet []SvBpPair
	var sv SvBpPair
	for rows.Next() {
		rows.Scan(&sv.Chr1, &sv.Bp1, &sv.Gene1, &sv.Chr2, &sv.Bp2, &sv.Gene2)
		svBpSet = append(svBpSet, sv)
	}
	return svBpSet
}

func init() {
	var confErr error
	mysqlDB, confErr = sql.Open("mysql", "guest:yY__kj20@tcp(10.0.0.1:3306)/variation_sites")
	if confErr != nil {
		panic("Mysql configuration error!")
	}
}
