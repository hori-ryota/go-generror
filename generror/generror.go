package generror

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hori-ryota/go-genutil/strtype"
)

const (
	commentMarker = "//errcode "
)

func Run(targetDir string, errorCodes []string, renderers []func(TmplParam) error) error {
	files, err := ioutil.ReadDir(targetDir)
	if err != nil {
		return err
	}
	var pkgName string
	errorDetailComments := make([]string, 0, 50)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}
		filename := filepath.Join(targetDir, file.Name())
		comments, err := extractErrorDetailComments(filename)
		if err != nil {
			return err
		}
		errorDetailComments = append(errorDetailComments, comments...)

		if pkgName == "" {
			pkgName, err = extractPkgName(filename)
			if err != nil {
				return err
			}
		}
	}

	detailErrorCodes := make([]DetailErrorCodeInfo, len(errorDetailComments))
	for i, c := range errorDetailComments {
		c := strings.TrimPrefix(c, commentMarker)
		cs := strings.Split(c, ",")
		info := DetailErrorCodeInfo{
			Code:   strings.TrimSpace(cs[0]),
			Params: make([]ParamInfo, len(cs[1:])),
		}
		for i, param := range cs[1:] {
			ss := strings.Fields(param)
			info.Params[i] = ParamInfo{
				Name: ss[0],
				Type: ss[1],
			}
		}
		detailErrorCodes[i] = info
	}

	importPackages := make(map[string]string, 5)

	for _, c := range detailErrorCodes {
		for _, p := range c.Params {
			pkgs := strtype.ImportsForConverter(p.Type)
			for _, pkg := range pkgs {
				importPackages[path.Base(pkg)] = pkg
			}
		}
	}

	param := TmplParam{
		PackageName:      pkgName,
		ErrorCodes:       errorCodes,
		DetailErrorCodes: detailErrorCodes,
		ImportPackages:   importPackages,
	}

	for _, renderer := range renderers {
		if err := renderer(param); err != nil {
			return err
		}
	}
	return nil
}

type TmplParam struct {
	PackageName      string
	ErrorCodes       []string
	DetailErrorCodes []DetailErrorCodeInfo
	ImportPackages   map[string]string
}

type DetailErrorCodeInfo struct {
	Code   string
	Params []ParamInfo
}

type ParamInfo struct {
	Name string
	Type string
}

func ToStringMethod(param ParamInfo) string {
	return strtype.ToConverter(param.Type, param.Name)
}

func extractErrorDetailComments(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	results := make([]string, 0, 10)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(text, commentMarker) {
			results = append(results, text)
		}
	}
	return results, nil
}

func extractPkgName(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(text, "package ") {
			return strings.TrimPrefix(text, "package "), nil
		}
	}
	return "", nil
}
