package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/pkg/util"
)

type AccountService interface {
	Login(loginReq *req.AccountLoginReq) (res *resp.AccountLoginResp, err error)
}

type accountService struct {
	repo repository.UserRepository
}

func (a accountService) Login(loginReq *req.AccountLoginReq) (res *resp.AccountLoginResp, err error) {
	account, err := a.repo.Find(loginReq.Username)
	if err != nil {
		return nil, fmt.Errorf("用户名不存在")
	}
	if account.Password != loginReq.Password {
		return nil, fmt.Errorf("密码错误")
	}
	token, err := util.JwtUtil.GenerateToken(account.Username)
	if err != nil {
		return nil, fmt.Errorf("生成token失败")
	}
	return &resp.AccountLoginResp{
		Username: account.Username,
		Token:    token,
	}, nil
}

func NewAccountService(repository repository.UserRepository) AccountService {
	return &accountService{repo: repository}
}
