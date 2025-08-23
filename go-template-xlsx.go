package gotemplatexlsx

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/JJJJJJack/go-template-xlsx/internal/xlsx"
	goziputils "github.com/JJJJJJack/go-zip-utils"
)

type XlsxTemplate struct {
	bytes      []byte
	xlsxZipMap goziputils.ZipMap
	output     bytes.Buffer
}

// NewXlsxTemplateFromFilename creates a new XlsxTemplate from the provided XLSX file.
func NewXlsxTemplateFromFilename(filename string) (*XlsxTemplate, error) {
	xlsxBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read XLSX file bytes: %w", err)
	}

	xlsxZipMap, err := goziputils.NewZipMapFromFilename(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create XLSX zip map: %w", err)
	}

	return &XlsxTemplate{
		bytes:      xlsxBytes,
		xlsxZipMap: xlsxZipMap,
		output:     bytes.Buffer{},
	}, nil
}

// NewXlsxTemplateFromBytes creates a new XlsxTemplate from the provided XLSX file bytes.
func NewXlsxTemplateFromBytes(xlsxBytes []byte) (*XlsxTemplate, error) {
	xlsxZipMap, err := goziputils.NewZipMapFromBytes(xlsxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create XLSX zip map: %w", err)
	}

	return &XlsxTemplate{
		bytes:      xlsxBytes,
		xlsxZipMap: xlsxZipMap,
		output:     bytes.Buffer{},
	}, nil
}

// Apply applies the template with the provided values to the XLSX file.
// The templateValues parameter can be any type that can be marshalled to JSON.
func (xt *XlsxTemplate) Apply(templateValues any) error {
	switch v := templateValues.(type) {
	case []byte:
		err := json.Unmarshal(v, &templateValues)
		if err != nil {
			return fmt.Errorf("error unmarshalling templateValues: %w", err)
		}
	}

	var sharedStringsNumbers map[int]string
	// key: old index, value: new index
	var sharedStringsNewIndexes map[int]int

	zipWriter := zip.NewWriter(&xt.output)

	// Copy all files except the ones that will be processed
	sheetNMatcher := regexp.MustCompile(`xl/worksheets/sheet\d*\.xml`)
	sharedStringsMatcher := regexp.MustCompile(`xl/(sharedStrings\d*)\.xml`)
	sharedStringsFilename := "xl/sharedStrings.xml"

	for filename, f := range xt.xlsxZipMap {
		switch {
		case
			sheetNMatcher.MatchString(filename),
			sharedStringsMatcher.MatchString(filename):
			continue
		}

		err := goziputils.CopyFile(zipWriter, f)
		if err != nil {
			return fmt.Errorf("unable to copy original embedding xlsx file '%s': %w", f.Name, err)
		}
	}

	// work on sharedStrings.xml
	sharedStringsFile := xt.xlsxZipMap[sharedStringsFilename]
	if sharedStringsFile == nil {
		return fmt.Errorf("shared strings file '%s' not found in embedded XLSX", sharedStringsFilename)
	}

	sharedStringsContent, err := goziputils.ReadZipFileContent(sharedStringsFile)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", sharedStringsFile.Name, err)
	}

	sharedStringsContent, err = xlsx.ApplyTemplateToCells(sharedStringsFile, templateValues, sharedStringsContent)
	if err != nil {
		return fmt.Errorf("error applying template to file '%s': %w", sharedStringsFile.Name, err)
	}

	sharedStringsContent, sharedStringsNumbers, sharedStringsNewIndexes, err = xlsx.GetReferencedSharedStringsByIndexAndCleanup(sharedStringsContent)
	if err != nil {
		return fmt.Errorf("error getting referenced shared strings from file '%s': %w", sharedStringsFile.Name, err)
	}

	sharedStringsCount := uint(0)
	for i := 1; ; i++ {
		sheetN := fmt.Sprintf("xl/worksheets/sheet%d.xml", i)

		f := xt.xlsxZipMap[sheetN]
		if f == nil {
			break
		}

		fileContent, err := goziputils.ReadZipFileContent(f)
		if err != nil {
			return fmt.Errorf("error reading zip file content '%s': %w", f.Name, err)
		}

		fileContent, err = xlsx.UpdateSheet(fileContent, sharedStringsNumbers, sharedStringsNewIndexes)
		if err != nil {
			return fmt.Errorf("error replacing shared strings indexes in file '%s': %w", f.Name, err)
		}

		sharedStringsRefs, err := xlsx.GetCountFromXml(fileContent)
		if err != nil {
			return fmt.Errorf("error getting shared strings refs count from file '%s': %w", f.Name, err)
		}

		sharedStringsCount += sharedStringsRefs

		err = goziputils.RewriteFileIntoZipWriter(zipWriter, f, fileContent)
		if err != nil {
			return fmt.Errorf("error writing file '%s': %w", f.Name, err)
		}
	}

	// need to be here, after all sheets have been processed we know the real count
	sharedStringsContent, err = xlsx.RecountSharedStringsCountAndUniqueCountAttributes(sharedStringsContent, sharedStringsCount)
	if err != nil {
		return fmt.Errorf("error recounting sharedStrings file '%s': %w", sharedStringsFile.Name, err)
	}

	err = goziputils.RewriteFileIntoZipWriter(zipWriter, sharedStringsFile, sharedStringsContent)
	if err != nil {
		return fmt.Errorf("error writing sharedStrings file '%s': %w", sharedStringsFile.Name, err)
	}

	err = zipWriter.Close()
	if err != nil {
		return fmt.Errorf("error closing zip writer: %w", err)
	}

	return nil
}

// Save saves the modified xlsx file to the specified filename.
func (xt *XlsxTemplate) Save(filename string) error {
	return os.WriteFile(filename, xt.output.Bytes(), 0644)
}

// Bytes returns the modified xlsx file as a byte slice.
// (empty if Apply was not used).
func (xt *XlsxTemplate) Bytes() []byte {
	return xt.output.Bytes()
}
