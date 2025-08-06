import { request } from '../http'

// 设备属性接口
export interface DeviceProperty {
  name: string        // 名称，英文标识
  description: string // 描述，中文涵义
  type: string        // 类型:int,long,double,string
  length: number      // 长度
  decimal: number     // 小数位
  unit: string        // 计量单位
  value: any          // 数值
  reported: boolean   // 是否上报
  isAlarm: boolean    // 是否报警
  threshold: number   // 预警值 +-区间
  used: number
  autoCalc: boolean
  scale: number       // 倍率
}

// 设备类型接口
export interface DeviceTypeInfo {
  id: string
  name: string
  protocol: string
  description: string
  driver: string            // 驱动程序目录、文件名
  properties: DeviceProperty[] // 设备属性列表
  createdAt: number
  updatedAt: number
}

// 设备类型创建/更新请求
export interface DeviceTypeRequest {
  id: string
  name: string
  protocol: string
  description: string
  driver: string
  properties: DeviceProperty[]
}

// 创建设备类型
export function createDeviceType(data: DeviceTypeRequest) {
  return request.Post<Service.ResponseResult<null>>('/device-type', data)
}

// 更新设备类型
export function updateDeviceType(data: DeviceTypeRequest) {
  return request.Put<Service.ResponseResult<null>>('/device-type', data)
}

// 删除设备类型
export function deleteDeviceType(id: string) {
  return request.Delete<Service.ResponseResult<null>>(`/device-type/${id}`)
}

// 获取设备类型详情
export function getDeviceType(id: string) {
  return request.Get<Service.ResponseResult<DeviceTypeInfo>>(`/device-type/${id}`)
}

// 获取设备类型列表
export function getDeviceTypeList(params: { page: number; pageSize: number; name?: string }) {
  return request.Get<Service.ResponseResult<Service.PageData<DeviceTypeInfo[]>>>('/device-types', { params })
}

// 获取设备类型数量
export function getDeviceTypeCount(params: { name?: string }) {
  return request.Get<Service.ResponseResult<{ count: number }>>('/device-type/count', { params })
} 