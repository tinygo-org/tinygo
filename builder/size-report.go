package builder

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
)

//go:embed size-report.html
var sizeReportBase string

func writeSizeReport(sizes *programSize, filename, pkgName string) error {
	tmpl, err := template.New("report").Parse(sizeReportBase)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not open report file: %w", err)
	}
	defer f.Close()

	// Prepare data for the report.
	type sizeLine struct {
		Name string
		Size *packageSize
	}
	programData := []sizeLine{}
	for _, name := range sizes.sortedPackageNames() {
		pkgSize := sizes.Packages[name]
		programData = append(programData, sizeLine{
			Name: name,
			Size: pkgSize,
		})
	}
	sizeTotal := map[string]uint64{
		"code":   sizes.Code,
		"rodata": sizes.ROData,
		"data":   sizes.Data,
		"bss":    sizes.BSS,
		"flash":  sizes.Flash(),
	}

	// Write the report.
	err = tmpl.Execute(f, map[string]any{
		"pkgName":   pkgName,
		"sizes":     programData,
		"sizeTotal": sizeTotal,
	})
	if err != nil {
		return fmt.Errorf("could not create report file: %w", err)
	}
	return nil
}
