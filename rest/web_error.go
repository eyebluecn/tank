package rest

type WebError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (this *WebError) Error() string {
	return this.Msg
}
