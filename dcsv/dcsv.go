/*
 * A Go package to handles access to a CSV file as Dataset
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dcsv handles access to a CSV file as Dataset
package dcsv

import (
	"encoding/csv"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"io"
	"os"
)

type DCSV struct {
	filename   string
	fieldNames []string
	hasHeader  bool
	separator  rune
	numFields  int
}

type DCSVConn struct {
	dataset       *DCSV
	file          *os.File
	reader        *csv.Reader
	currentRecord ddataset.Record
	err           error
}

func New(
	filename string,
	hasHeader bool,
	separator rune,
	fieldNames []string,
) ddataset.Dataset {
	return &DCSV{
		filename:   filename,
		fieldNames: fieldNames,
		hasHeader:  hasHeader,
		separator:  separator,
		numFields:  len(fieldNames),
	}
}

func (c *DCSV) Open() (ddataset.Conn, error) {
	f, r, err := makeCsvReader(c.filename, c.separator, c.hasHeader)
	if err != nil {
		return nil, err
	}
	r.Comma = c.separator

	return &DCSVConn{
		dataset:       c,
		file:          f,
		reader:        r,
		currentRecord: make(ddataset.Record, c.numFields),
		err:           nil,
	}, nil
}

func (c *DCSV) GetFieldNames() []string {
	return c.fieldNames
}

func (cc *DCSVConn) Next() bool {
	if cc.err != nil {
		return false
	}
	if cc.reader == nil {
		cc.err = ddataset.ErrConnClosed
		return false
	}
	row, err := cc.reader.Read()
	if err == io.EOF {
		return false
	} else if err != nil {
		cc.Close()
		cc.err = err
		return false
	}
	if err := cc.makeRowCurrentRecord(row); err != nil {
		cc.Close()
		cc.err = err
		return false
	}
	return true
}

func (cc *DCSVConn) Err() error {
	return cc.err
}

func (cc *DCSVConn) Read() ddataset.Record {
	return cc.currentRecord
}

func (cc *DCSVConn) Close() error {
	err := cc.file.Close()
	cc.file = nil
	cc.reader = nil
	return err
}

func (cc *DCSVConn) getNumFields() int {
	return cc.dataset.numFields
}

func (cc *DCSVConn) getFieldNames() []string {
	return cc.dataset.fieldNames
}

func (cc *DCSVConn) makeRowCurrentRecord(row []string) error {
	fieldNames := cc.dataset.GetFieldNames()
	if len(row) != cc.getNumFields() {
		cc.err = ddataset.ErrWrongNumFields
		cc.Close()
		return cc.err
	}
	for i, field := range row {
		l, err := dlit.New(field)
		if err != nil {
			cc.Close()
			cc.err = err
			return err
		}
		cc.currentRecord[fieldNames[i]] = l
	}
	return nil
}

func makeCsvReader(
	filename string,
	separator rune,
	hasHeader bool,
) (*os.File, *csv.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(f)
	r.Comma = separator
	if hasHeader {
		_, err := r.Read()
		if err != nil {
			return nil, nil, err
		}
	}
	return f, r, err
}
