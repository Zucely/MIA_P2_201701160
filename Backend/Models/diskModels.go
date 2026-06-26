package Models

type MkdiskRequest struct {
	Size int    `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
	Path string `json:"path"`
}

type MkdiskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
