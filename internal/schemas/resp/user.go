package resp

type UserResp struct {
	Name      string `json:"name"`
	Password  string `json:"password,omitempty"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}
