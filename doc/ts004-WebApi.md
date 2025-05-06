# WebAPI 接口规范（TS004）

## 设备管理接口 `/api/device`

### 接口概览
| 方法   | 路径                | 描述         |
|--------|---------------------|--------------|
| POST   | /device            | 创建设备     |
| PUT    | /device            | 更新设备信息 |
| DELETE | /device/:name      | 删除设备     |
| GET    | /device/:name      | 查询单个设备 |
| GET    | /devices           | 查询所有设备 |
| POST   | /devices/import    | 批量导入设备 |
| GET    | /devices/export    | 导出设备列表 |

### 请求示例
```http
POST /api/device HTTP/1.1
Content-Type: application/json

{
  "name": "DDSU5886",
  "type": "电表",
  "protocol": "MODBUS-RTU",
  "address": "COM1:9600"
}
```

## 设备类型接口 `/api/device-type`

### 协议脚本管理
```http
POST /api/device-type/upload/:name HTTP/1.1
Content-Type: multipart/form-data

{
  "script": "设备解析脚本.lua",
  "config_schema": "设备配置JSON Schema"
}
```

### 设备属性管理
| 方法   | 路径                            | 描述         |
|--------|---------------------------------|--------------|
| POST   | /device-property/:name         | 添加设备属性 |
| PUT    | /device-property/:name/:propertyid | 更新属性配置 |
| DELETE | /device-property/:name/:propertyid | 删除设备属性 |

## 以太网配置接口 `/api/ethernet`

### 参数规范
```json
{
  "interface": "eth0",
  "ip": "192.168.1.100",
  "gateway": "192.168.1.1",
  "dns": ["8.8.8.8", "114.114.114.114"]
}
```

## 统一响应格式
```json
{
  "code": 200,
  "data": {},
  "message": "操作成功",
  "timestamp": 1718000000
}
```

### 错误代码表
| 代码 | 描述               |
|------|--------------------|
| 400  | 请求参数错误       |
| 401  | 未授权访问         |
| 404  | 资源不存在         |
| 500  | 服务器内部错误     |
| 503  | 设备通信异常       |