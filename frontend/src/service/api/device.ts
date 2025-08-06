import { request } from '../http'

// 设备接口
export interface DeviceInfo {
  name: string
  typeID: string
  address: string
  collectorID: string
  online: boolean
  collectTime: number
  collectTotal: number
  collectSuccess: number
  reportTime: number
  reportTotal: number
  reportSuccess: number
  alarmStatus: boolean
  alarmReason: string
  alarmTime: number
  alarmTotal: number
  initialVal: number
  scale: number
  createdAt: number
  updatedAt: number
}

// 设备创建/更新请求
export interface DeviceRequest {
  name: string
  typeID: string
  address: string
  collectorID: string
  initialVal: number
  scale: number
}

// 创建设备
export function createDevice(data: DeviceRequest) {
  return request.Post<Service.ResponseResult<null>>('/device', data)
}

// 更新设备
export function updateDevice(data: DeviceRequest) {
  return request.Put<Service.ResponseResult<null>>('/device', data)
}

// 删除设备
export function deleteDevice(name: string) {
  return request.Delete<Service.ResponseResult<null>>(`/device/${name}`)
}

// 获取设备详情
export function getDevice(name: string) {
  return request.Get<Service.ResponseResult<DeviceInfo>>(`/device/${name}`)
}

// 获取设备列表
export function getDeviceList(params: { page: number; pageSize: number; typeID?: string; collectorID?: string }) {
  return request.Get<Service.ResponseResult<Service.PageData<DeviceInfo[]>>>('/devices', { params })
}

// 获取设备数量
export function getDeviceCount(params: { typeID?: string; collectorID?: string }) {
  return request.Get<Service.ResponseResult<{ count: number }>>('/device/count', { params })
} 