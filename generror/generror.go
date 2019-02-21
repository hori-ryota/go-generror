package generror

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

	importPackages := make([]string, 0, 5)

	for _, c := range detailErrorCodes {
		for _, p := range c.Params {
			_, pkgs := JudgeToStringMethod(p)
			importPackages = append(importPackages, pkgs...)
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
	ImportPackages   []string
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
	m, _ := JudgeToStringMethod(param)
	return m
}

func JudgeToStringMethod(param ParamInfo) (toStringMethod string, needsPkgs []string) {
	switch param.Type {
	case "string":
		return param.Name, nil
	case "bool":
		return fmt.Sprintf("strconv.FormatBool(%s)", param.Name), []string{"strconv"}
	case "uint64":
		return fmt.Sprintf("strconv.FormatUint(%s, 10)", param.Name), []string{"strconv"}
	case "uint", "uint8", "uint16", "uint32":
		return fmt.Sprintf("strconv.FormatUint(uint64(%s), 10)", param.Name), []string{"strconv"}
	case "int64":
		return fmt.Sprintf("strconv.FormatInt(%s, 10)", param.Name), []string{"strconv"}
	case "int", "int8", "int16", "int32":
		return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", param.Name), []string{"strconv"}
	default:
		return fmt.Sprintf("fmt.Sprint(%s)", param.Name), []string{"fmt"}
	}
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

func FmtImports(pkgs []string) string {
	if len(pkgs) == 0 {
		return ""
	}

	groups := make([][]string, 2)

	for _, pkg := range pkgs {
		if len(strings.Split(pkg, "/")) < 3 && !strings.Contains(pkg, ".") {
			groups[0] = append(groups[0], pkg)
			continue
		}
		groups[1] = append(groups[1], pkg)
	}

	b := new(bytes.Buffer)
	for _, group := range groups {
		group := group
		sort.Slice(group, func(i, j int) bool {
			return group[i] < group[j]
		})
		for _, pkg := range group {
			_, err := b.WriteString(strconv.Quote(pkg))
			if err != nil {
				panic(err)
			}
			_, err = b.WriteRune('\n')
			if err != nil {
				panic(err)
			}
		}
		_, err := b.WriteRune('\n')
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf(`import (
%s
		)`,
		b.String(),
	)
}
