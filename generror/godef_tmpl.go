package generror

import (
	"text/template"

	"github.com/hori-ryota/go-genutil/genutil"
)

var GodefTmpl = template.Must(template.New("errTmpl").Funcs(map[string]interface{}{
	"ToStringMethod": ToStringMethod,
	"FmtImports":     genutil.GoFmtImports,
}).Parse(`
// Code generated ; DO NOT EDIT

package {{ .PackageName }}

{{ FmtImports .ImportPackages }}

type ErrorCode string

const (
	{{ range .ErrorCodes }}
	error{{ . }} ErrorCode = "{{ . }}"
	{{- end }}
)

func (c ErrorCode) String() string {
	return string(c)
}

type Error interface {
	Error() string
	Details() []ErrorDetail
	{{ range .ErrorCodes }}
	Is{{ . }}() bool
	{{- end }}
}

func newError(source error, code ErrorCode, details ...ErrorDetail) Error {
	return errorImpl{
		source:  source,
		code:    code,
		details: details,
	}
}

{{ range .ErrorCodes }}
func Error{{ . }}(source error, details ...ErrorDetail) Error {
	return newError(source, error{{ . }}, details...)
}
{{- end }}

type errorImpl struct {
	source  error
	code    ErrorCode
	details []ErrorDetail
}

func (e errorImpl) Error() string {
	return fmt.Sprintf("%s:%s:%s", e.code, e.details, e.source)
}
func (e errorImpl) Details() []ErrorDetail {
	return e.details
}
{{ range .ErrorCodes }}
func (e errorImpl) Is{{ . }}() bool {
	return e.code == error{{ . }}
}
{{- end }}

type ErrorDetail struct {
	code ErrorDetailCode
	args []string
}

func newErrorDetail(code ErrorDetailCode, args ...string) ErrorDetail {
	return ErrorDetail{
		code: code,
		args: args,
	}
}

func (e ErrorDetail) String() string {
	return strings.Join(append([]string{e.code.String()}, e.args...), ",")
}

func (c ErrorDetail) Code() ErrorDetailCode {
	return c.code
}

func (c ErrorDetail) Args() []string {
	return c.args
}

type ErrorDetailCode string

func (c ErrorDetailCode) String() string {
	return string(c)
}

{{ range .DetailErrorCodes }}
const ErrorDetail{{ .Code }} ErrorDetailCode = "{{ .Code }}"
func {{ .Code }}Error(
	{{- range .Params }}
	{{ .Name }} {{ .Type }},
	{{- end }}
) ErrorDetail {
	return newErrorDetail(
		ErrorDetail{{ .Code }},
		{{- range .Params }}
		{{ ToStringMethod . }},
		{{- end }}
	)
}
{{- end }}

func (e errorImpl) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	zaperr.ToNamedField("sourceError", e.source).AddTo(enc)
	zap.String("code", string(e.code)).AddTo(enc)
	zap.Any("details", e.details)
	return nil
}
`))
