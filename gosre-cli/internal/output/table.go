// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

// Table writes results to w as an aligned tab-separated table.
func Table(w io.Writer, results []domain.Result) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TIMESTAMP\tTARGET\tSTATUS\tDURATION\tERROR"); err != nil {
		return err
	}
	for _, r := range results {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%dms\t%s\n",
			r.Timestamp.Format("15:04:05"),
			r.TargetID,
			r.Status,
			r.Duration.Milliseconds(),
			r.Error,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
