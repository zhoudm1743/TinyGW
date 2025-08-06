package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userRepo *repo.UserRepo
}

func NewUserController(userRepo *repo.UserRepo) *UserController {
	return &UserController{
		userRepo: userRepo,
	}
}

// Create 创建用户
func (c *UserController) Create(ctx *gin.Context) {
	var userReq req.UserReq
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	user := &models.User{
		Name:     userReq.Name,
		Password: userReq.Password,
	}

	if err := c.userRepo.Create(user); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "用户创建成功", nil)
}

// Update 更新用户
func (c *UserController) Update(ctx *gin.Context) {
	var userReq req.UserReq
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	user, err := c.userRepo.Get(userReq.Name)
	if err != nil {
		FailWithNotFound(ctx, "用户不存在")
		return
	}

	user.Password = userReq.Password

	if err := c.userRepo.Update(user); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "用户更新成功", nil)
}

// Delete 删除用户
func (c *UserController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "用户名不能为空")
		return
	}

	user, err := c.userRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "用户不存在")
		return
	}

	if err := c.userRepo.Delete(user); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "用户删除成功", nil)
}

// Get 获取用户
func (c *UserController) Get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "用户名不能为空")
		return
	}

	user, err := c.userRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "用户不存在")
		return
	}

	userResp := resp.UserResp{
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	Success(ctx, userResp)
}

// List 获取用户列表
func (c *UserController) List(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	conds := make(map[string]interface{})
	total, users, err := c.userRepo.Find(page, pageSize, conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	var userResps []resp.UserResp
	for _, user := range users {
		userResps = append(userResps, resp.UserResp{
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	SuccessWithPage(ctx, total, page, pageSize, userResps)
}

// Count 获取用户数量
func (c *UserController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})
	count, err := c.userRepo.Count(conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	Success(ctx, gin.H{"count": count})
}

// Login 用户登录
func (c *UserController) Login(ctx *gin.Context) {
	var loginReq req.UserReq
	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	user, err := c.userRepo.Get(loginReq.Name)
	if err != nil {
		FailWithUnauthorized(ctx)
		return
	}

	if user.Password != loginReq.Password {
		FailWithUnauthorized(ctx)
		return
	}

	// 这里应该生成JWT token，但简化处理
	Success(ctx, gin.H{
		"user": resp.UserResp{
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		// 这里可以添加token
		"accessToken":  "mock-token",
		"refreshToken": "mock-refresh-token",
	})
}
