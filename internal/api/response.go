package api

type GenericResponse struct {
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Changed int    `json:"changed"`
}
