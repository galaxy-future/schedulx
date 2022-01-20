package types

type ServiceInfo struct {
	ServiceName string `json:"service_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Domain      string `json:"domain"`
	Port        string `json:"port"`
}
