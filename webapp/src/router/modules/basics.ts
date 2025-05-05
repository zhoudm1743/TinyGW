export default {
  path: "/basics",
  redirect: "/basics/index",
  meta: {
    icon: "ri:rhythm-fill",
    title: "基础数据",
    rank: 8
  },
  children: [
    {
      path: "/basics/collectors",
      name: "collectors",
      component: () => import("@/views/basics/collectors/index.vue"),
      meta: {
        title: "采集接口",
        icon: "ri:list-check"
      }
    },
    {
      path: "/basics/device-type",
      name: "device-type",
      component: () => import("@/views/basics/device-type/index.vue"),
      meta: {
        title: "设备类型",
        icon: "ri:function-line"
      }
    },
    {
      path: "/basics/device-list",
      name: "device-list",
      component: () => import("@/views/basics/device-list/index.vue"),
      meta: {
        title: "设备列表",
        icon: "ri:list-check"
      }
    }
  ]
} satisfies RouteConfigsTable;
