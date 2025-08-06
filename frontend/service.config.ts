/** 不同请求服务的环境配置 */
export const serviceConfig: Record<ServiceEnvType, Record<string, string>> = {
  dev: {
    url: 'http://localhost:8001/api',
  },
  test: {
    url: 'http://localhost:8001/api',
  },
  prod: {
    url: 'http://localhost:8001/api',
  },
}
