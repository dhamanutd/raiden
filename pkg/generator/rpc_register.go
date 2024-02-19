package generator

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

// ----- Define type, variable and constant -----
type (
	GenerateRegisterRpcData struct {
		Imports []string
		Package string
		Rpc     []string
	}
)

const (
	RpcRegisterFilename = "rpc.go"
	RpcRegisterDir      = "internal/bootstrap"
	RpcRegisterTemplate = `// Code generated by raiden-cli; DO NOT EDIT.
package {{ .Package }}
{{if gt (len .Imports) 0 }}
import (
{{- range .Imports}}
	{{.}}
{{- end}}
)
{{end }}
func RegisterRpc() {
	resource.RegisterRpc(
		{{- range .Rpc}}
		&rpc.{{.}}{},
		{{- end}}
	)
}
`
)

func GenerateRpcRegister(basePath string, projectName string, generateFn GenerateFn) error {
	rpcRegisterDir := filepath.Join(basePath, RpcRegisterDir)
	logger.Debugf("GenerateRpcRegister - create %s folder if not exist", rpcRegisterDir)
	if exist := utils.IsFolderExists(rpcRegisterDir); !exist {
		if err := utils.CreateFolder(rpcRegisterDir); err != nil {
			return err
		}
	}

	rpcDir := filepath.Join(basePath, RpcDir)
	logger.Debugf("GenerateRpcRegister - create %s folder if not exist", rpcDir)
	if exist := utils.IsFolderExists(rpcDir); !exist {
		if err := utils.CreateFolder(rpcDir); err != nil {
			return err
		}
	}

	// scan all controller
	rpcList, err := WalkScanRpc(rpcDir)
	if err != nil {
		return err
	}

	input, err := createRegisterRpcInput(projectName, rpcRegisterDir, rpcList)
	if err != nil {
		return err
	}

	logger.Debugf("GenerateRpcRegister - generate rpc to %s", input.OutputPath)
	return generateFn(input, nil)
}

func createRegisterRpcInput(projectName string, rpcRegisterDir string, rpcList []string) (input GenerateInput, err error) {
	// set file path
	filePath := filepath.Join(rpcRegisterDir, RpcRegisterFilename)

	// set imports path
	imports := []string{
		fmt.Sprintf("%q", "github.com/sev-2/raiden/pkg/resource"),
	}

	if len(rpcList) > 0 {
		rpcImportPath := fmt.Sprintf("%s/internal/rpc", utils.ToGoModuleName(projectName))
		imports = append(imports, fmt.Sprintf("%q", rpcImportPath))
	}

	// set passed parameter
	data := GenerateRegisterRpcData{
		Package: "bootstrap",
		Imports: imports,
		Rpc:     rpcList,
	}

	input = GenerateInput{
		BindData:     data,
		Template:     RpcRegisterTemplate,
		TemplateName: "rpcRegisterTemplate",
		OutputPath:   filePath,
	}

	return
}

func WalkScanRpc(rpcDir string) ([]string, error) {
	logger.Debugf("GenerateRpcRegister - scan %s for register all rpc", rpcDir)

	rpc := make([]string, 0)
	err := filepath.Walk(rpcDir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".go") {
			logger.Debugf("GenerateRpcRegister - collect rpc from %s", path)
			rs, e := getStructByBaseName(path, "RpcBase")
			if e != nil {
				return e
			}

			rpc = append(rpc, rs...)

		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return rpc, nil
}
