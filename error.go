package main

import "net/http"

type MiError interface {
	// 自定义MiError接口内置error接口,表示要实现error接口的全部方法
	error
	Status() int
}

// 结构体 StatusError 实现了 MiError接口 同时也实现了 error接口
type StatusError struct {
	Code int
	// error接口类型的变量,表示只要实现error接口的值就可以赋值给 Err变量
	Err error
}

// 实现MiError接口内置error接口的 Error方法
func (se StatusError) Error() string {
	return se.Err.Error()
}

func (se StatusError) Status() int {
	return se.Code
}

func NewNotFindErr(err error) StatusError {
	return StatusError{
		Code: http.StatusNotFound,
		Err:  err,
	}
}

func NewBadReqErr(err error) StatusError {
	return StatusError{
		Code: http.StatusBadRequest,
		Err:  err,
	}
}
