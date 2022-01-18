package client

type Action string

type HttpResp struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}
