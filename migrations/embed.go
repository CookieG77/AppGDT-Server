// Small file only used to embed the migrations scripts in the compiled version.

package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
