import { request } from '../http'

// 采集任务接口
export interface CollectTaskInfo {
  name: string
  cron: string
  status: number
  deviceList: string[]
  createdAt: number
  updatedAt: number
}

// 采集任务创建/更新请求
export interface CollectTaskRequest {
  name: string
  cron: string
  status: number
  deviceList: string[]
}

// 创建采集任务
export function createCollectTask(data: CollectTaskRequest) {
  return request.Post<Service.ResponseResult<null>>('/collect-task', data)
}

// 更新采集任务
export function updateCollectTask(data: CollectTaskRequest) {
  return request.Put<Service.ResponseResult<null>>('/collect-task', data)
}

// 删除采集任务
export function deleteCollectTask(name: string) {
  return request.Delete<Service.ResponseResult<null>>(`/collect-task/${name}`)
}

// 获取采集任务详情
export function getCollectTask(name: string) {
  return request.Get<Service.ResponseResult<CollectTaskInfo>>(`/collect-task/${name}`)
}

// 获取采集任务列表
export function getCollectTaskList(params: { page: number; pageSize: number; status?: number }) {
  return request.Get<Service.ResponseResult<Service.PageData<CollectTaskInfo[]>>>('/collect-tasks', { params })
}

// 获取采集任务数量
export function getCollectTaskCount(params: { status?: number }) {
  return request.Get<Service.ResponseResult<{ count: number }>>('/collect-task/count', { params })
}

// 启动采集任务
export function startCollectTask(name: string) {
  return request.Post<Service.ResponseResult<null>>(`/collect-task/${name}/start`)
}

// 停止采集任务
export function stopCollectTask(name: string) {
  return request.Post<Service.ResponseResult<null>>(`/collect-task/${name}/stop`)
} 