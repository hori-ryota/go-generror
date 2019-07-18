package generror_test

import (
	"bytes"
	"go/format"
	"log"
	"os"

	"github.com/hori-ryota/go-generror/generror"
)

func ExampleRun() {
	targetDir := "../_example"
	if err := generror.Run(
		targetDir,
		[]string{
			"Unknown", "BadRequest", "PermissionDenied", "NotFound",
		},
		[]func(param generror.TmplParam) error{
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
				_, err = os.Stdout.Write(out)
				return err
			},
		},
	); err != nil {
		log.Fatal(err)
	}
	// Output:
	// // Code generated ; DO NOT EDIT
	//
	// package example
	//
	// import (
	// 	"fmt"
	// 	"strconv"
	// 	"strings"
	//
	// 	"github.com/hori-ryota/zaperr"
	// 	"go.uber.org/zap"
	// 	"go.uber.org/zap/zapcore"
	// )
	//
	// type ErrorCode string
	//
	// const (
	// 	errorUnknown          ErrorCode = "Unknown"
	// 	errorBadRequest       ErrorCode = "BadRequest"
	// 	errorPermissionDenied ErrorCode = "PermissionDenied"
	// 	errorNotFound         ErrorCode = "NotFound"
	// )
	//
	// func (c ErrorCode) String() string {
	// 	return string(c)
	// }
	//
	// type Error interface {
	// 	Error() string
	// 	Details() []ErrorDetail
	//
	// 	IsUnknown() bool
	// 	IsBadRequest() bool
	// 	IsPermissionDenied() bool
	// 	IsNotFound() bool
	// }
	//
	// func newError(source error, code ErrorCode, details ...ErrorDetail) Error {
	// 	return errorImpl{
	// 		source:  source,
	// 		code:    code,
	// 		details: details,
	// 	}
	// }
	//
	// func ErrorUnknown(source error, details ...ErrorDetail) Error {
	// 	return newError(source, errorUnknown, details...)
	// }
	// func ErrorBadRequest(source error, details ...ErrorDetail) Error {
	// 	return newError(source, errorBadRequest, details...)
	// }
	// func ErrorPermissionDenied(source error, details ...ErrorDetail) Error {
	// 	return newError(source, errorPermissionDenied, details...)
	// }
	// func ErrorNotFound(source error, details ...ErrorDetail) Error {
	// 	return newError(source, errorNotFound, details...)
	// }
	//
	// type errorImpl struct {
	// 	source  error
	// 	code    ErrorCode
	// 	details []ErrorDetail
	// }
	//
	// func (e errorImpl) Error() string {
	// 	return fmt.Sprintf("%s:%s:%s", e.code, e.details, e.source)
	// }
	// func (e errorImpl) Details() []ErrorDetail {
	// 	return e.details
	// }
	//
	// func (e errorImpl) IsUnknown() bool {
	// 	return e.code == errorUnknown
	// }
	// func (e errorImpl) IsBadRequest() bool {
	// 	return e.code == errorBadRequest
	// }
	// func (e errorImpl) IsPermissionDenied() bool {
	// 	return e.code == errorPermissionDenied
	// }
	// func (e errorImpl) IsNotFound() bool {
	// 	return e.code == errorNotFound
	// }
	//
	// type ErrorDetail struct {
	// 	code ErrorDetailCode
	// 	args []string
	// }
	//
	// func newErrorDetail(code ErrorDetailCode, args ...string) ErrorDetail {
	// 	return ErrorDetail{
	// 		code: code,
	// 		args: args,
	// 	}
	// }
	//
	// func (e ErrorDetail) String() string {
	// 	return strings.Join(append([]string{e.code.String()}, e.args...), ",")
	// }
	//
	// func (c ErrorDetail) Code() ErrorDetailCode {
	// 	return c.code
	// }
	//
	// func (c ErrorDetail) Args() []string {
	// 	return c.args
	// }
	//
	// type ErrorDetailCode string
	//
	// func (c ErrorDetailCode) String() string {
	// 	return string(c)
	// }
	//
	// const ErrorDetailNameIsInvalidLength ErrorDetailCode = "NameIsInvalidLength"
	//
	// func NameIsInvalidLengthError(
	// 	lessThan int,
	// 	moreThan int,
	// ) ErrorDetail {
	// 	return newErrorDetail(
	// 		ErrorDetailNameIsInvalidLength,
	// 		strconv.FormatInt(int64(lessThan), 10),
	// 		strconv.FormatInt(int64(moreThan), 10),
	// 	)
	// }
	//
	// func (e errorImpl) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// 	zaperr.ToNamedField("sourceError", e.source).AddTo(enc)
	// 	zap.String("code", string(e.code)).AddTo(enc)
	// 	zap.Any("details", e.details)
	// 	return nil
	// }
}
