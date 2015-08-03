package OpenBattleGen

import (
	"fmt"
	"io"
)

func init() {
	RegisterExporter(new(NilExporter))
}

// NilExporter does nothing
type NilExporter struct {
}

// Languages returns all supported languages.
func (exp *NilExporter) Languages() []string {
	return []string{"nil"}
}

// Export exports the specified definitions using this exporter.
func (exp *NilExporter) Export(d *Definitions, w io.Writer) error {
	for _, msg := range d.Messages {
		fmt.Fprintln(w, msg)
	}
	return nil
}
