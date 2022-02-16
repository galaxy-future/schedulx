package handler

type ServiceInfo struct {
	ServiceName  string `json:"service_name"`
	Description  string `json:"description"`
	Language     string `json:"language"`
	TmplExpandId string `json:"tmpl_expand_id"`
}

type Pager struct {
	PageNumber int `json:"page_num"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
}
