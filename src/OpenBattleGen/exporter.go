package OpenBattleGen

import "io"

var (
	exporters = []Exporter{}
)

// Exporter is an interface which facilitates exporting of messages.
type Exporter interface {
	Languages() []string
	Export(*Definitions, io.Writer) error
}

// GetExporter returns the exporter for the specified language, or nil if no
// exporter could be found.
func GetExporter(lang string) Exporter {
	for _, exp := range exporters {
		for _, l := range exp.Languages() {
			if lang == l {
				return exp
			}
		}
	}
	return nil
}

// GetExporters returns all registered exporters.
func GetExporters() []Exporter {
	return exporters
}

// RegisterExporter registers an exporter.
func RegisterExporter(exp Exporter) {
	exporters = append(exporters, exp)
}
