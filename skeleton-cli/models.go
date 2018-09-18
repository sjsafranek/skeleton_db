package main

type ApiResponse struct {
	Status string          `json:"status"`
	Data   ApiResponseData `json:"data"`
	Error  string          `json:"error"`
}

type ApiResponseData struct {
	Value ApiResponseValue `json:"value"`
}

type ApiResponseValue struct {
	Value    string   `json:"value"`
	Diffs    []string `json:"diffs"`
	CreateAt string   `json:"create_at"`
	UpdateAt string   `json:"update_at"`
}
