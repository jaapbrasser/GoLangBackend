package dto

type CheckRepositoryRequest struct {
	Owner string `json:"owner" binding:"required"`
	Repo  string `json:"repo" binding:"required"`
}

type CheckRepositoryResponse struct {
	Exists  bool   `json:"exists"`
	HTMLURL string `json:"html_url,omitempty"`
}

type CreateIssueRequest struct {
	Owner string `json:"owner" binding:"required"`
	Repo  string `json:"repo" binding:"required"`
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type CreateIssueResponse struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
}

type GetIssueResponse struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
}
