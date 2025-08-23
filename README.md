[![Go Reference](https://pkg.go.dev/badge/github.com/JJJJJJack/go-template-xlsx.svg)](https://pkg.go.dev/github.com/JJJJJJack/go-template-xlsx)

`go get github.com/JJJJJJack/go-template-xlsx`

# Notes

go-template-xlsx is based on the golang template standard library, thus it inherits its templating syntax to parse tokens inside the xlsx file.
The library doesn't change the original files and only reads it into memory to output a new file with the provided template values.

# Usage

First you need to create an instance of the object to load the xlsx file in and get the high-level APIs, you have 2 options to do so:

```go
XlsxTemplate, err := gotemplatexlsx.NewXlsxTemplateFromBytes(xlsxBytes)
if err != nil {
  // handle error
}
```

or

```go
XlsxTemplate, err := gotemplatexlsx.NewXlsxTemplateFromFilename(xlsxFilename)
if err != nil {
  // handle error
}
```

after obtaining the `XlsxTemplate` object it exposes the methods to create a new xlsx file based on the original templated one, let's walk through the usage for each one

> every function is provided with a Godoc comment, you can find all the exposed APIs in the `go_template_xlsx.go` file

## 1. Applying the template values

> here the `templateValues` variable could be any json marshallable value, the struct fields will be used as keys in the xlsx to search to access the value

```go
err := XlsxTemplate.Apply(templateValues)
if err != nil {
  // handle error
}
```

## 2. Saving the new xlsx as new file

```go
err := XlsxTemplate.Save(outputFilename)
if err != nil {
  // handle error
}
```

## 3. Read back bytes from new xlsx

```go
output := XlsxTemplate.Bytes()
```

Enjoy programmatically templating xlsx files from golang!

# Xlsx tempalte instructions examples

Let's say we have this json value as our `templateValues` variable:

```json
{
  "Text": "Hello, World!",
  "Text2": "Yaaaty",
  "Number": 42
}
```

and this is the templated xlsx file that we load into the `XlsxTemplate`:
![](https://github.com/JJJJJJack/jubilant-fortnight/blob/main/go-template-xlsx/input.xlsx.png)


now if we run this code
```go
template, _ := gotemplatexlsx.NewxlsxTemplateFromFilename(xlsxFilename)

template.Apply(templateValues)

template.Save("output.xlsx")
```

the `output.xlsx` file will be the result of the templating engine:

![](https://github.com/JJJJJJack/jubilant-fortnight/blob/main/go-template-xlsx/output.xlsx.png)

more details on the go template syntaxes on the [go-template-docx readme](https://github.com/JJJJJJack/go-template-docx/blob/main/README.md#1-fields-replacement)
> I only tested the text replacement and toNumberCell so far...

# Binary example

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