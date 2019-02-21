# go-generror

Error generator for Go.

## Usage

```go
    //errcode {DetailCode}[,paramName paramType]...
```

with `go generate` command

```go
    //go:generate go-generror [Code]...
```

generated Error usage sample is following

```go
type NameSpec struct {
    lessThan int
    moreThan int
}

func (s NameSpec) Validate(name string) Error {
    //errcode NameIsInvalidLength,lessThan int,moreThan int
    if len(name) >= s.lessThan || len(name) <= s.moreThan {
        return ErrorBadRequest(errors.New("invalid name"), NameIsInvalidLengthError(s.lessThan, s.moreThan))
    }
    return nil
}
```

and

```go
func(w http.ResponseWriter, r *http.Request) {
    nameSpec := NameSpec{
        lessThan: 100,
        moreThan: 0,
    }
    err := nameSpec.Validate(r.Query().Get("name"))
    if err != nil {
        renderError(w, err)
        return
    }
    // ...
}

func renderError(w http.ResponseWriter, err Error) {
    switch {
        case err.IsUnknown():
            // ...
        case err.IsBadRequst():
            // ...
        // ...
    }
    // ...
}
```

### Example

def
[./_example/example.go](./_example/example.go)

generated
[./_example/error_gen.go](./_example/error_gen.go)

## Installation

```sh
$ go get github.com/hori-ryota/go-generror
```

## Extend output

We can define multi renderers. Like following

```go
templates := []template.Template{
    generror.GodefTmpl,
    customTmpl1,
    customTmpl2,
}

renderers := make([]func(generror.TemplParam), len(templates))
for i := range templates {
    tmpl := tempates[i]
    renderers[i] = func(param generror.TmplParam) error {

        param.ImportPackages = append(param.ImportPackages, "fmt")
        param.ImportPackages = append(param.ImportPackages, "strings")
        param.ImportPackages = append(param.ImportPackages, "go.uber.org/zap")
        param.ImportPackages = append(param.ImportPackages, "go.uber.org/zap/zapcore")
        param.ImportPackages = append(param.ImportPackages, "github.com/hori-ryota/zaperr")

        buf := new(bytes.Buffer)
        err := tmpl.Execute(buf, param)
        if err != nil {
            return err
        }

        out, err := format.Source(buf.Bytes())
        if err != nil {
            return err
        }
        return ioutil.WriteFile(dstFileName, out, 0644)
    }
}

return generror.Run(".", args[1:], renderers)
```

This mean we can create formatter interfaces. e.g.

template def

```go
type ErrorFormatter interface {
    {{- range .DetailErrorCodes }}
    {{ .Code }}Error(
        {{- range .Params -}}
        {{ .Name }} {{ .Type }},
        {{- end -}}
    ) string
    {{- end }}
}

type ErrorDetail struct {
    Code string
    Args []interface{}
}

func FormatError(formatter ErrorFormatter, err ErrorDetail) string {
    switch err.Code {
    {{- range .DetailErrorCodes }}
    case "{{ .Code }}":
        return formatter.{{ .Code }}Error(
            {{- range $i, $v := .Params -}}
            err.Args[{{ $i }}].({{ $v.Type }}),
            {{- end -}}
        )
    {{- end }}
    }
}
```

generated

```go
type ErrorFormatter interface {
    NameIsInvalidLengthError(lessThan int, moreThan int) string
    FooError(arg1 string, arg2 time.Time, arg3 int) string
    BarError() string
}

type ErrorDetail struct {
    Code string
    Args []interface{}
}

func FormatError(formatter ErrorFormatter, err ErrorDetail) string {
    switch err.Code {
    case "NameIsInvalidLength":
        return formatter.NameIsInvalidLengthError(err.Args[0].(int), err.Args[1].(int))
    case "Foo":
        return formatter.FooError(err.Args[0].(string), err.Args[1].(time.Time), err.Args[2].(int))
    case "Bar":
        return formatter.BarError()
    }
}
```

We needs only implements `ErrorFormatter` interface with benefits of Type-safe. Let's multilingual support!

Of course, this automatic generation is not limited to Go language. Since it can be generated also in Kotlin language etc. by devising it, it should be able to provide interface to frontend without tedious documentation.

And, of course, we can also generate documentation automatically.
