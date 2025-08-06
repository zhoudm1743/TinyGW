import { request } from '../http'

// 设备日志接口
export interface DeviceLogInfo {
  id: number
  deviceID: string
  deviceName: string
  time: number
  originCmd: string
  originRes: string
  res: string
  property: string
  value: string
  status: number
  error: string
  createdAt: number
  updatedAt: number
}

// 设备日志创建请求
export interface DeviceLogRequest {
  deviceID: string
  time: number
  originCmd: string
  originRes: string
  res: string
  property: string
  value: string
  status: number
  error: string
}

// 创建设备日志
export function createDeviceLog(data: DeviceLogRequest) {
  return request.Post<Service.ResponseResult<null>>('/device-log', data)
}

// 删除设备日志
export function deleteDeviceLog(id: number) {
  return request.Delete<Service.ResponseResult<null>>(`/device-log/${id}`)
}

// 获取设备日志列表
export function getDeviceLogList(params: { 
  page: number
  pageSize: number
  deviceID?: string
  status?: number
  property?: string
  startTime?: number
  endTime?: number
}) {
  return request.Get<Service.ResponseResult<Service.PageData<DeviceLogInfo[]>>>('/device-logs', { params })
}

// 获取设备日志数量
export function getDeviceLogCount(params: { deviceID?: string; status?: number }) {
  return request.Get<Service.ResponseResult<{ count: number }>>('/device-log/count', { params })
}

// 清除历史日志
export function clearDeviceLogs(time?: number) {
  return request.Post<Service.ResponseResult<null>>('/device-logs/clear', {}, { params: { time } })
} 