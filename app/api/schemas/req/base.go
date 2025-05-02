package req

type PageReq struct {
	PageNo   int `form:"page,default=1" validate:"omitempty,gte=1"`            // 页码
	PageSize int `form:"pageSize,default=20" validate:"omitempty,gt=0,lte=60"` // 每页大小
}

type IdReq struct {
	ID uint `form:"id" validate:"required" json:"id" uri:"id"` // 主键ID
}

type NameReq struct {
	Name string `form:"name" validate:"required" json:"name" uri:"name"` // 名称
}
