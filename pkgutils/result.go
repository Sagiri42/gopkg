package pkgutils

import "go/types"

type Result[T any] struct {
	Success bool    `json:"success"`
	Message *string `json:"message,omitempty"`
	Data    *T      `json:"data,omitempty"`
}

func Success[T any](msg string, data T) Result[T] {
	return Result[T]{
		Success: true,
		Message: &msg,
		Data:    &data,
	}
}

func Failed[T any](msg string, data T) Result[T] {
	return Result[T]{
		Success: false,
		Message: &msg,
		Data:    &data,
	}
}

func SuccessMsg(msg string) Result[types.Nil] {
	return Result[types.Nil]{
		Success: true,
		Message: &msg,
		Data:    nil,
	}
}

func FailedMsg(msg string) Result[types.Nil] {
	return Result[types.Nil]{
		Success: false,
		Message: &msg,
		Data:    nil,
	}
}

func SuccessData[T any](data T) Result[T] {
	return Result[T]{
		Success: true,
		Message: nil,
		Data:    &data,
	}
}

func FailedData[T any](data T) Result[T] {
	return Result[T]{
		Success: false,
		Message: nil,
		Data:    &data,
	}
}
