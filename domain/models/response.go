package models

type HttpResponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status,omitempty"`
}

func NewHttpResponse[T any](data T, message string, status int) HttpResponse[T] {
	return HttpResponse[T]{
		Data:    data,
		Message: message,
		Status:  status,
	}
}
