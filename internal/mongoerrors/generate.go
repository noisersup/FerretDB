// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build ignore

package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"text/template"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// line represents a single mapping.
type line struct {
	ErrorName  string
	ErrorCode  string
	MongoError int
}

// extraMongoErrors contains MongoDB error codes FerretDB uses and error_mappings.csv does not include
var extraMongoErrors = map[string]int{
	"Unset":                         0,
	"UserNotFound":                  11,
	"UnsupportedFormat":             12,
	"Unauthorized":                  13,
	"ProtocolError":                 17,
	"AuthenticationFailed":          18,
	"MaxTimeMSExpired":              50,
	"CommandNotFound":               59,
	"OperationFailed":               96,
	"ClientMetadataCannotBeMutated": 186,
	"InvalidUUID":                   207,
	"NotImplemented":                238,
	"MechanismUnavailable":          334,
	"UnsupportedOpQueryCommand":     352,
	"Location16979":                 16979,
	"Location40621":                 40621,
	"Location50687":                 50687,
	"Location50692":                 50692,
	"Location50840":                 50840,
	"Location5739101":               5739101,
}

func main() {
	opts := &logging.NewHandlerOpts{
		Base:  "console",
		Level: slog.LevelDebug,
	}
	logging.SetupDefault(opts, "")

	ctx := context.Background()
	l := slog.Default()

	// temporary workaround while DocumentDB repo/submodule is private
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		l.WarnContext(ctx, "Skipping code generation on GitHub Actions")
		os.Exit(0)
	}

	flag.Parse()

	mongoErrors := parseErrorMappings(ctx, l)
	data := parseDocumentDBCodes(ctx, l, mongoErrors)

	slices.SortFunc(data, func(a, b line) int {
		return cmp.Or(
			cmp.Compare(a.MongoError, b.MongoError),
			cmp.Compare(a.ErrorName, b.ErrorName),
		)
	})

	// "// Code generated" is intentionally not on the next line
	// to prevent generate.go itself being marked as generated on GitHub.
	t := template.Must(template.New("").Parse(`// Code generated by "generate.go"; DO NOT EDIT.

package mongoerrors

// MongoDB error names and codes.
const (
	{{- range .}}
	Err{{.ErrorName}} = Code({{.MongoError}}) // {{.ErrorName}}
	{{- end}}
)

// Mapping of PostgreSQL/DocumentDB error codes to MongoDB error codes.
var pgCodes = map[string]Code{
	{{- range .}}
	{{- if .ErrorCode}}
	"{{.ErrorCode}}": Err{{.ErrorName}}, // {{.MongoError}}
	{{- end}}
	{{- end}}
}
`)).Option("missingkey=error")

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		l.Log(ctx, logging.LevelFatal, "Can't render template", logging.Error(err))
	}

	if err := os.WriteFile("codes.go", buf.Bytes(), 0o666); err != nil {
		l.Log(ctx, logging.LevelFatal, "Can't write template file", logging.Error(err))
	}
}

// parseErrorMappings parses error_mappings.csv and adds a few codes not defined there.
func parseErrorMappings(ctx context.Context, l *slog.Logger) map[string]int {
	f, err := os.Open(filepath.FromSlash("../../build/postgres-documentdb/documentdb/error_mappings.csv"))
	if err != nil {
		l.Log(ctx, logging.LevelFatal, "Can't open file", logging.Error(err))
	}
	defer f.Close() //nolint:errcheck // we only read it

	r := csv.NewReader(f)
	r.FieldsPerRecord = 2

	res := make(map[string]int)
	codes := make(map[int]struct{})

	for {
		records, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			l.Log(ctx, logging.LevelFatal, "Can't read record", logging.Error(err))
		}

		mongoError, errorName := records[0], records[1]
		if mongoError == "ErrorMapping" {
			continue
		}

		mongoErrorCode, err := strconv.Atoi(mongoError)
		if err != nil {
			l.Log(ctx, logging.LevelFatal, "Can't parse ErrorMapping", logging.Error(err))
		}

		if _, ok := codes[mongoErrorCode]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate Mongo error code", slog.Int("code", mongoErrorCode))
		}

		if _, ok := res[errorName]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate Mongo error name", slog.String("name", errorName))
		}

		codes[mongoErrorCode] = struct{}{}
		res[errorName] = mongoErrorCode
	}

	for errorName, mongoErrorCode := range extraMongoErrors {
		if _, ok := codes[mongoErrorCode]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate Mongo error code", slog.Int("code", mongoErrorCode))
		}

		if _, ok := res[errorName]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate Mongo error name", slog.String("name", errorName))
		}

		codes[mongoErrorCode] = struct{}{}
		res[errorName] = mongoErrorCode
	}

	return res
}

// parseDocumentDBCodes parses documentdb_codes.txt.
func parseDocumentDBCodes(ctx context.Context, l *slog.Logger, mongoErrors map[string]int) []line {
	f, err := os.Open(filepath.FromSlash("../../build/postgres-documentdb/documentdb/pg_documentdb_core/include/utils/documentdb_codes.txt"))
	if err != nil {
		l.Log(ctx, logging.LevelFatal, "Can't open file", logging.Error(err))
	}
	defer f.Close() //nolint:errcheck // we only read it

	r := csv.NewReader(f)
	r.FieldsPerRecord = 3

	var res []line

	names := make(map[string]struct{})
	pCodes := make(map[string]struct{})

	for {
		records, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			l.Log(ctx, logging.LevelFatal, "Can't read record", logging.Error(err))
		}

		errorName, errorCode, _ := records[0], records[1], records[2]
		if errorName == "ErrorName" {
			continue
		}

		if _, ok := names[errorName]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate Mongo error name", slog.String("name", errorName))
		}

		if _, ok := pCodes[errorCode]; ok {
			l.Log(ctx, logging.LevelFatal, "Duplicate PostgreSQL error code", slog.String("code", errorCode))
		}

		names[errorName] = struct{}{}
		pCodes[errorCode] = struct{}{}

		mongoError := mongoErrors[errorName]
		if mongoError == 0 {
			l.Log(ctx, logging.LevelFatal, "No Mongo error code for name", slog.String("name", errorName))
		}

		delete(mongoErrors, errorName)

		res = append(res, line{
			ErrorName:  errorName,
			ErrorCode:  errorCode,
			MongoError: mongoError,
		})
	}

	if !reflect.DeepEqual(mongoErrors, extraMongoErrors) {
		l.Log(ctx, logging.LevelFatal, "Extra error codes", slog.Any("codes", mongoErrors))
	}

	for errorName, mongoError := range mongoErrors {
		res = append(res, line{
			ErrorName:  errorName,
			MongoError: mongoError,
		})
	}

	return res
}
