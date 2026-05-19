package dto

type CheckRepositoryRequest struct {
	Owner string `json:"owner" binding:"required"`
	Repo  string `json:"repo" binding:"required"`
}

type CheckRepositoryResponse struct {
	Exists   bool   `json:"exists"`
	HTMLURL  string `json:"html_url,omitempty"`
}