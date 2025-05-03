package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type CollectTaskService interface {
	Add(addReq *req.CollectTaskReq) error
	Update(updateReq *req.CollectTaskReq) error
	Delete(name string) error
	Find(name string) (*resp.CollectTaskResp, error)
	FindAll() ([]resp.CollectTaskResp, error)
	List(pageReq *req.PageReq) (response.PageResp, error)
}

type collectTaskService struct {
	repo repository.CollectTaskRepository
}

func (c collectTaskService) Add(addReq *req.CollectTaskReq) error {
	_, err := c.repo.Find(addReq.Name)
	if err == nil {
		return fmt.Errorf("采集任务名称 %s 已存在", addReq.Name)
	}
	var ct models.CollectTask
	response.Copy(&ct, addReq)
	err = c.repo.Save(&ct)
	if err != nil {
		return err
	}
	return nil
}

func (c collectTaskService) Update(updateReq *req.CollectTaskReq) error {
	var ct models.CollectTask
	response.Copy(&ct, updateReq)
	err := c.repo.Save(&ct)
	if err != nil {
		return fmt.Errorf("更新采集任务失败: %s", err.Error())
	}
	return nil
}

func (c collectTaskService) Delete(name string) error {
	err := c.repo.Delete(name)
	if err != nil {
		return fmt.Errorf("删除采集任务失败: %s", err.Error())
	}
	return nil
}

func (c collectTaskService) Find(name string) (*resp.CollectTaskResp, error) {
	find, err := c.repo.Find(name)
	if err != nil {
		return nil, fmt.Errorf("查询采集任务失败: %s", err.Error())
	}
	var resp resp.CollectTaskResp
	response.Copy(&resp, find)
	return &resp, nil
}

func (c collectTaskService) FindAll() ([]resp.CollectTaskResp, error) {
	findAll, err := c.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("查询所有采集任务失败: %s", err.Error())
	}
	var resps []resp.CollectTaskResp
	response.Copy(&resps, findAll)
	return resps, nil
}

func (c collectTaskService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	list, count, err := c.repo.List(limit, offset)
	if err != nil {
		return response.PageResp{}, fmt.Errorf("查询采集任务列表失败: %s", err.Error())
	}
	return response.PageResp{
		Count:    count,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Lists:    list,
	}, nil
}

func NewCollectTaskService(repo repository.CollectTaskRepository) CollectTaskService {
	return &collectTaskService{
		repo: repo,
	}
}
