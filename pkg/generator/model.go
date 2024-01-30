package generator

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, van and constant -----
type Rls struct {
	CanWrite []string
	CanRead  []string
}

type GenerateModelData struct {
	Columns    []map[string]any
	Imports    []string
	StructName string
	Package    string
	RlsTag     string
	Schema     string
}

const (
	ModelDir      = "internal/models"
	ModelTemplate = `package {{ .Package }}
{{- if gt (len .Imports) 0 }}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)
{{- end }}

type {{ .StructName }} struct {
{{- range .Columns }}
	{{ .Name | ToGoIdentifier }} {{ .GoType }} ` + "`json:\"{{ .Name | ToSnakeCase }},omitempty\" column:\"{{ .Name | ToSnakeCase }}\"`" + `
{{- end }}

	Metadata string ` + "`schema:\"{{ .Schema}}\"`" + `
	Acl string ` + "`{{ .RlsTag }}`" + `
}
`
)

func GenerateModels(projectName string, tables []supabase.Table, rlsList supabase.Policies, generateFn GenerateFn) (err error) {
	internalFolderPath := filepath.Join(projectName, "internal")
	if exist := utils.IsFolderExists(internalFolderPath); !exist {
		if err := utils.CreateFolder(internalFolderPath); err != nil {
			return err
		}
	}

	folderPath := filepath.Join(projectName, ModelDir)
	if exist := utils.IsFolderExists(folderPath); !exist {
		if err := utils.CreateFolder(folderPath); err != nil {
			return err
		}
	}

	for i, v := range tables {
		searchTable := tables[i].Name
		GenerateModel(folderPath, v, rlsList.FilterByTable(searchTable), generateFn)
	}

	return nil
}

func GenerateModel(folderPath string, table supabase.Table, rlsList supabase.Policies, generateFn GenerateFn) error {
	// define binding func
	funcMaps := []template.FuncMap{
		{"ToGoIdentifier": utils.SnakeCaseToPascalCase},
		{"ToGoType": postgres.ToGoType},
		{"ToSnakeCase": utils.ToSnakeCase},
	}

	// map column data
	columns, importsPath := mapTableAttributes(table)
	rlsTag := generateRlsTag(rlsList)

	// define file path
	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.%s", table.Name, "go"))
	absolutePath, err := utils.GetAbsolutePath(filePath)
	if err != nil {
		return err
	}

	// set data
	data := GenerateModelData{
		Package:    "models",
		Imports:    importsPath,
		StructName: utils.SnakeCaseToPascalCase(table.Name),
		Columns:    columns,
		Schema:     table.Schema,
		RlsTag:     rlsTag,
	}

	// setup generate input param
	input := GenerateInput{
		BindData:     data,
		FuncMap:      funcMaps,
		Template:     ModelTemplate,
		TemplateName: "modelTemplate",
		OutputPath:   absolutePath,
	}

	return generateFn(input)
}
