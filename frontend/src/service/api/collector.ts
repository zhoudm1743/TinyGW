import { request } from '../http'

// 串口配置
export interface Serial {
  name: string
  deviceName: string
  baudRate: number
  dataBit: number
  stopBit: string
  check: string
}

// TCP客户端配置
export interface TcpClient {
  name: string
  ip: string
  port: number
}

// TCP服务端配置
export interface TcpServer {
  name: string
  port: number
}

// MQTT配置
export interface Mqtt {
  name: string
}

// 通道配置
export interface Channel {
  name: string
}

// 4G GPRS配置
export interface FourGPRS {
  name: string
}

// 采集器接口
export interface CollectorInfo {
  name: string
  type: string
  address: string
  serial: Serial
  tcpClient: TcpClient
  tcpServer: TcpServer
  mqtt: Mqtt
  channel: Channel
  fourGPRS: FourGPRS
  timeout: number
  interval: number
  enable: boolean
  createdAt: number
  updatedAt: number
}

// 采集器创建/更新请求
export interface CollectorRequest {
  name: string
  type: string
  address: string
  serial: Serial
  tcpClient: TcpClient
  tcpServer: TcpServer
  mqtt: Mqtt
  channel: Channel
  fourGPRS: FourGPRS
  timeout: number
  interval: number
  enable: boolean
}

// 创建采集器
export function createCollector(data: CollectorRequest) {
  return request.Post<Service.ResponseResult<null>>('/collector', data)
}

// 更新采集器
export function updateCollector(data: CollectorRequest) {
  return request.Put<Service.ResponseResult<null>>('/collector', data)
}

// 删除采集器
export function deleteCollector(name: string) {
  return request.Delete<Service.ResponseResult<null>>(`/collector/${name}`)
}

// 获取采集器详情
export function getCollector(name: string) {
  return request.Get<Service.ResponseResult<CollectorInfo>>(`/collector/${name}`)
}

// 获取采集器列表
export function getCollectorList(params: { page: number; pageSize: number; type?: string; enable?: boolean }) {
  return request.Get<Service.ResponseResult<Service.PageData<CollectorInfo[]>>>('/collectors', { params })
}

// 获取采集器数量
export function getCollectorCount(params: { type?: string; enable?: boolean }) {
  return request.Get<Service.ResponseResult<{ count: number }>>('/collector/count', { params })
} 