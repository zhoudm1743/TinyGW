export default {
    path: "/system",
    redirect: "/system/upgrade",
    meta: {
        icon: "ri:settings-2-line",
        title: "系统管理",
        rank: 20
    },
    children: [
        {
            path: "/system/upgrade",
            name: "upgrade",
            component: () => import("@/views/system/index.vue"),
            meta: {
                title: "系统升级"
            }
        },
    ]
} satisfies RouteConfigsTable;
