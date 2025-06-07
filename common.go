package v2

type NormalRes[T any] struct {
	Code int    `json:"code"    dc:"code"`
	Data T      `json:"data"   dc:"data 可null"`
	Msg  string `json:"msg"     dc:"return msg"`
}

type ListRes[T any] struct {
	Rows  []T
	Total int `json:"total"`
}

func NewNormalRes[T any](data T, msg ...string) *NormalRes[T] {
	message := ""
	if len(msg) == 0 {
		message = "操作成功"
	} else {
		message = msg[0]
	}
	return &NormalRes[T]{
		Code: 1,
		Data: data,
		Msg:  message,
	}
}
func NewListRes[T any](data []T, total int) *ListRes[T] {
	return &ListRes[T]{
		Rows:  data,
		Total: total,
	}
}
