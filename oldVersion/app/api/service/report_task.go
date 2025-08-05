package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type ReportTaskService interface {
	Add(addReq *req.ReportTaskReq) error
	Update(updateReq *req.ReportTaskReq) error
	Delete(name string) error
	Find(name string) (*resp.ReportTaskResp, error)
	FindAll() ([]resp.ReportTaskResp, error)
	List(pageReq *req.PageReq) (response.PageResp, error)
	Start(name string) error
	Stop(name string) error
}

type reportTaskService struct {
	repo repository.ReportTaskRepository
}

func (c reportTaskService) Add(addReq *req.ReportTaskReq) error {
	_, err := c.repo.Find(addReq.Name)
	if err == nil {
		return fmt.Errorf("采集任务名称 %s 已存在", addReq.Name)
	}
	var ct models.ReportTask
	response.Copy(&ct, addReq)
	err = c.repo.Save(&ct)
	if err != nil {
		return err
	}
	return nil
}

func (c reportTaskService) Update(updateReq *req.ReportTaskReq) error {
	var ct models.ReportTask
	response.Copy(&ct, updateReq)
	err := c.repo.Save(&ct)
	if err != nil {
		return fmt.Errorf("更新采集任务失败: %s", err.Error())
	}
	return nil
}

func (c reportTaskService) Delete(name string) error {
	err := c.repo.Delete(name)
	if err != nil {
		return fmt.Errorf("删除采集任务失败: %s", err.Error())
	}
	return nil
}

func (c reportTaskService) Find(name string) (*resp.ReportTaskResp, error) {
	find, err := c.repo.Find(name)
	if err != nil {
		return nil, fmt.Errorf("查询采集任务失败: %s", err.Error())
	}
	var resp resp.ReportTaskResp
	response.Copy(&resp, find)
	return &resp, nil
}

func (c reportTaskService) FindAll() ([]resp.ReportTaskResp, error) {
	findAll, err := c.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("查询所有采集任务失败: %s", err.Error())
	}
	var resps []resp.ReportTaskResp
	response.Copy(&resps, findAll)
	return resps, nil
}

func (c reportTaskService) List(page *req.PageReq) (response.PageResp, error) {
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

func (c reportTaskService) Start(name string) error {
	rt, err := c.repo.Find(name)
	if err != nil {
		return fmt.Errorf("查询上报任务失败: %s", err.Error())
	}
	rt.Status = 1
	err = c.repo.Save(&rt)
	if err != nil {
		return fmt.Errorf("启动上报任务失败: %s", err.Error())
	}
	return nil
}

func (c reportTaskService) Stop(name string) error {
	rt, err := c.repo.Find(name)
	if err != nil {
		return fmt.Errorf("查询上报任务失败: %s", err.Error())
	}
	rt.Status = 0
	err = c.repo.Save(&rt)
	if err != nil {
		return fmt.Errorf("停止上报任务失败: %s", err.Error())
	}
	return nil
}

func NewReportTaskService(repo repository.ReportTaskRepository) ReportTaskService {
	return &reportTaskService{
		repo: repo,
	}
}
