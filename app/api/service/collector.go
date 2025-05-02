package service

import (
	"errors"
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/api/schemas/resp"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
	"tinyGW/pkg/service/event"
)

type CollectorService interface {
	Add(addReq *req.CollectorSaveReq) error
	Update(updateReq *req.CollectorSaveReq) error
	Delete(name string) error
	Find(name string) (*resp.CollectorResp, error)
	FindAll() ([]resp.CollectorResp, error)
	List(pageReq *req.PageReq) (response.PageResp, error)
}

type collectorService struct {
	collectorRepo   repository.CollectorRepository
	deviceRepo      repository.DeviceRepository
	collectTaskRepo repository.CollectTaskRepository
	event           *event.EventService
}

func (c collectorService) Add(addReq *req.CollectorSaveReq) error {
	_, err := c.collectorRepo.Find(addReq.Name)
	if err == nil {
		return fmt.Errorf("采集接口已存在")
	}
	var coll models.Collector
	response.Copy(&coll, addReq)
	err = c.collectorRepo.Save(&coll)
	if err != nil {
		return fmt.Errorf("保存采集接口失败:%v", err)
	}
	c.event.Publish(event.Event{
		Name: "collector_add",
		Data: coll,
	})
	return nil
}

func (c collectorService) Update(updateReq *req.CollectorSaveReq) error {
	var coll models.Collector
	response.Copy(&coll, updateReq)
	err := c.collectorRepo.Save(&coll)
	if err != nil {
		return fmt.Errorf("更新采集接口失败:%v", err)
	}
	c.deviceRepo.CollectorChanged(coll)
	c.event.Publish(event.Event{
		Name: "collector_update",
		Data: coll,
	})
	return nil
}

func (c collectorService) Delete(name string) error {
	find, err := c.collectorRepo.Find(name)
	if err != nil {
		return fmt.Errorf("删除采集接口失败:%v", err)
	}
	if c.deviceRepo.CollectorIsUsed(&find) {
		return errors.New("已经有设备引用了采集接口，不能删除！")
	}
	err = c.collectorRepo.Delete(name)
	if err != nil {
		return fmt.Errorf("删除采集接口失败:%v", err)
	}
	c.event.Publish(event.Event{
		Name: "collector_delete",
		Data: name,
	})
	return nil
}

func (c collectorService) Find(name string) (*resp.CollectorResp, error) {
	find, err := c.collectorRepo.Find(name)
	if err != nil {
		return nil, fmt.Errorf("查询采集接口失败:%v", err)
	}
	res := resp.CollectorResp{}
	response.Copy(&res, find)
	return &res, nil
}

func (c collectorService) FindAll() ([]resp.CollectorResp, error) {
	list, err := c.collectorRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("查询采集接口列表失败:%v", err)
	}
	res := make([]resp.CollectorResp, 0)
	response.Copy(&res, list)
	return res, nil
}

func (c collectorService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	list, total, err := c.collectorRepo.List(limit, offset)
	if err != nil {
		return response.PageResp{}, fmt.Errorf("查询采集接口列表失败:%v", err)
	}
	var res []resp.CollectorResp
	response.Copy(&res, list)
	return response.PageResp{
		Count:    total,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Lists:    res,
	}, nil
}

func NewCollectorService(repo repository.CollectorRepository, deviceRepo repository.DeviceRepository, collectTaskRepo repository.CollectTaskRepository, event *event.EventService) CollectorService {
	return &collectorService{
		collectorRepo:   repo,
		deviceRepo:      deviceRepo,
		collectTaskRepo: collectTaskRepo,
		event:           event,
	}
}
