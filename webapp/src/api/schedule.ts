import { http } from "@/utils/http";
import Cookies from "js-cookie";
import {TokenKey} from "@/utils/auth";

const token = JSON.parse(Cookies.get(TokenKey)).accessToken;
export const collect = {
    page: (params?: object) => {
        return http.request("get", "/api/collect-task/list", { params }, { headers: { "Token": token } });
    },
    //获取采集任务列表
    list: (data?:object) => {
        return http.request("get", "/api/collect-tasks", { data }, { headers: { "Token": token } });
    },
    //更新采集任务
    edit: (data?:object) => {
        return http.request("put", `/api/collect-task`, { data }, { headers: { "Token": token } });
    },
    //删除采集任务
    delete: (name?:string) => {
        return http.request("delete", `/api/collect-task/${name}`, {}, { headers: { "Token": token } });
    },
    //添加采集任务
    add: (data?:object) => {
        return http.request("post", "/api/collect-task", { data }, { headers: { "Token": token } });
    },
    //启动
    start: (name?:string) => {
        return http.request("get", `/api/collect-task/start/${name}`, {}, { headers: { "Token": token } });
    },
    //停止
    stop: (name?:string) => {
        return http.request("get", `/api/collect-task/stop/${name}`, {}, { headers: { "Token": token } });
    },
}

export const report = {
    page: (params?: object) => {
        return http.request("get", "/api/report-task/list", { params }, { headers: { "Token": token } });
    },
    //获取上报任务列表
    list: (data?:object) => {
        return http.request("get", "/api/report-tasks", { data }, { headers: { "Token": token } });
    },
    //更新上报任务
    edit: (data?:object) => {
        return http.request("put", `/api/report-task`, { data }, { headers: { "Token": token } });
    },
    //删除上报任务
    delete: (name?:string) => {
        return http.request("delete", `/api/report-task/${name}`,{}, { headers: { "Token": token } });
    },
    //添加上报任务
    add: (data?:object) => {
        return http.request("post", "/api/report-task", { data }, { headers: { "Token": token } });
    },
    //启动
    start: (name?:string) => {
        return http.request("get", `/api/report-task/start/${name}`,{}, { headers: { "Token": token } });
    },
    //停止
    stop: (name?:string) => {
        return http.request("get", `/api/report-task/stop/${name}`,{}, { headers: { "Token": token } });
    },
}
