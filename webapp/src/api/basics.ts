import { http } from "@/utils/http";
import Cookies from "js-cookie";
import {TokenKey} from "@/utils/auth";

const token = JSON.parse(Cookies.get(TokenKey)).accessToken;

//采集接口
export const collectorApi = {
    // 获取采集接口列表
    list: (params?: object) => {
        return http.request("get", "/api/collectors", { params }, {headers: {"Token": token} });
    },
    // 更新采集接口
    edit: (data?: object) => {
        return http.request("put", "/api/collector", { data }, {headers: {"Token": token} });
    },
    // 删除采集接口
    delete: (name?: string) => {
        return http.request("delete", `/api/collector/${name}`,{}, {headers: {"Token": token} });
    },
    // 增加采集接口
    add: (data?: object) => {
        return http.request("post", "/api/collector", { data }, {headers: {"Token": token} });
    },
    page: (params?: object) => {
        return http.request("get", "/api/collector/list", { params }, {headers: {"Token": token} });
    }
};

//设备类型
export const deviceApi = {
    page: (params?: object) => {
        return http.request("get", "/api/device-type/list", { params }, {headers: {"Token": token} });
    },
    // 获取类型列表
    list: (data?:object) => {
        return http.request("get", "/api/device-types", { data }, {headers: {"Token": token}})
    },
    // 更新类型
    edit: (data?:object) => {
        return http.request("put", "/api/device-type", { data }, {headers: {"Token": token}})
    },
    // 删除类型
    delete: (name?:string) => {
        return http.request("delete", `/api/device-type/${name}`,{}, {headers: {"Token": token}})
    },
    // 添加类型
    add: (data?:object) => {
        return http.request("post", "/api/device-type", { data }, {headers: {"Token": token}})
    },
    // 上传驱动(.zip)文件接口
    upload: (name?:string, data?:object) => {
        return http.request("post", `/api/device-type/upload/${name}`, { FileName: name, ...data }, {
            headers: { "Content-Type": "multipart/form-data", "Token": token }
        });
    },
    // 更新设备类型规则
    editDetail: (name?:string, data?:object) => {
        console.log(name, data)
        console.log(`/api/device-property${name}/${data.id}`)
        return http.request("put", `/api/device-property/${name}/${data.id}`, { data }, {headers: {"Token": token}})
    },
    // 删除设备类型规则
    deleteDetail: (name?:string, index?:number) => {
        return http.request("delete", `/api/device-property/${name}/${index}`,{}, {headers: {"Token": token}})
    },
    // 添加设备类型规则
    addDetail: (name?:string, data?:object) => {
        return http.request("post", `/api/device-property/${name}`, { data }, {headers: {"Token": token}})
    }
};

//设备列表
export const deviceListApi = {
    page: (params?: object) => {
        return http.request("get", "/api/device/list", { params }, {headers: {"Token": token} });
    },
    // 获取设备列表
    list: (data?:object) => {
        return http.request("get", "/api/devices", { data }, {headers: {"Token": token}})
    },
    // 更新设备信息
    edit: (data?:object) => {
        return http.request("put", `/api/device`, { data }, {headers: {"Token": token}})
    },
    // 删除设备
    delete: (name?:string) => {
        return http.request("delete", `/api/device/${name}`,{}, {headers: {"Token": token}})
    },
    // 添加设备
    add: (data?:object) => {
        return http.request("post", "/api/device", { data }, {headers: {"Token": token}})
    },
}
