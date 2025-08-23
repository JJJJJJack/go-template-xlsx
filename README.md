# Usage

> This is the actual source code of the released binaries
```go
func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <file.xlsx> <values.json>\n", os.Args[0])
		return
	}

	xlsxFilename := os.Args[1]
	jsonFilename := os.Args[2]

	jsonBytes, err := os.ReadFile(jsonFilename)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	templateValues := any(nil)
	err = json.Unmarshal(jsonBytes, &templateValues)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	xlsxTemplate, err := gotemplatexlsx.NewXlsxTemplateFromFilename(xlsxFilename)
	if err != nil {
		panic(err)
	}

	err = xlsxTemplate.Apply(jsonBytes)
	if err != nil {
		panic(err)
	}

	name := xlsxFilename[:strings.LastIndex(xlsxFilename, ".")] + "_output.xlsx"
	err = xlsxTemplate.Save(name)
	if err != nil {
		panic(err)
	}
}
```