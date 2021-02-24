package http

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

func (receiver *Response) Ok(data interface{}) {
	receiver.Success = true
	receiver.Error = ""
	receiver.Data = data
}

func (receiver *Response) Err(err error) {
	receiver.Success = false
	receiver.Error = err.Error()
	receiver.Data = nil
}
