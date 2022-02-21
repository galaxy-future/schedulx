package types

import "time"

type Integration struct {
	Host     string     `json:"host"`
	Account  string     `json:"account"`
	Password string     `json:"-"`
	Type     string     `json:"type"`
	CreateAt *time.Time `json:"create_at"`
	CreateBy string     `json:"create_by"`
}
