/*
```go
    //errcode {DetailCode}[,paramName paramType]...
```

with `go generate` command

```go
    //go:generate go-generror [Code]...
```
*/
package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"log"
	"os"

	"github.com/hori-ryota/go-generror/generror"
)

const (
	dstFileName = "error_gen.go"
)

func main() {
	if err := Main(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Main(args []string) error {
	return generror.Run(".", args[1:], []func(param generror.TmplParam) error{
		func(param generror.TmplParam) error {

			importPackages := map[string]string{
				"fmt":     "fmt",
				"strings": "strings",
				"zap":     "go.uber.org/zap",
				"zapcore": "go.uber.org/zap/zapcore",
				"zaperr":  "github.com/hori-ryota/zaperr",
			}
			for k, v := range param.ImportPackages {
				importPackages[k] = v
			}

			param.ImportPackages = importPackages

			buf := new(bytes.Buffer)
			err := generror.GodefTmpl.Execute(buf, param)
			if err != nil {
				return err
			}

			out, err := format.Source(buf.Bytes())
			if err != nil {
				return err
			}
			return ioutil.WriteFile(dstFileName, out, 0644)
		},
	})
}
