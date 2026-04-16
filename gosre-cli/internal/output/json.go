// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package output

import (
	"encoding/json"
	"io"

	"github.com/gosre/gosre-sdk/domain"
)

// JSON encodes results as a JSON array to w.
func JSON(w io.Writer, results []domain.Result) error {
	return json.NewEncoder(w).Encode(results)
}
