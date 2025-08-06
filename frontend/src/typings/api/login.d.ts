/// <reference path="../global.d.ts"/>

namespace Api {
  namespace Login {
    /** 用户信息 */
    interface UserInfo {
      /** 用户ID */
      id: number
      /** 用户名 */
      name: string
      /** 头像 */
      avatar: string
      /** 角色 */
      role: Entity.RoleType[]
      /** 邮箱 */
      email: string
      /** 手机号 */
      phone: string
      /** 创建时间 */
      createdAt: number
      /** 更新时间 */
      updatedAt: number
    }

    /** 登录响应 */
    interface Info {
      /** 用户信息 */
      userInfo: UserInfo
      /** 访问token */
      accessToken: string
      /** 刷新token */
      refreshToken: string
    }
  }
}
