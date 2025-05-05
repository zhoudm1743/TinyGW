import { http } from "@/utils/http";
import Cookies from "js-cookie";
import {TokenKey} from "@/utils/auth";

const token = JSON.parse(Cookies.get(TokenKey)).accessToken;

//采集接口
export const systemInfoApi = {
    // 网关信息
    systemStatus: (data?: object) => {
        return http.request("post", "/api/debug/system-status", { data }, {headers: {"Token": token} });
    },
    // 时间校准
    timing: (data?: object) => {
        return http.request("post", "/api/debug/system-ntp", { data }, {headers: {"Token": token} });
    },
    // 重启系统
    reboot: (data?: object) => {
        return http.request("post", "/api/debug/system-reboot", { data }, {headers: {"Token": token} });
    },
};

