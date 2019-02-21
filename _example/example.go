package example

import "errors"

//go:generate go-generror Unknown BadRequest PermissionDenied NotFound

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
