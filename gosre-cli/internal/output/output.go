// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package output

import (
	"fmt"
	"io"

	"github.com/gosre/gosre-sdk/domain"
)

// Format selects the output representation for check results.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Write renders results to w in the requested format.
func Write(w io.Writer, f Format, results []domain.Result) error {
	switch f {
	case FormatTable:
		return Table(w, results)
	case FormatJSON:
		return JSON(w, results)
	default:
		return fmt.Errorf("unknown output format: %s", f)
	}
}
