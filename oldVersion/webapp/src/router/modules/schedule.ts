export default {
  path: "/schedule",
  redirect: "/schedule/collect",
  meta: {
    icon: "ri:server-fill",
    // showLink: false,
    title: "任务计划",
    rank: 10
  },
  children: [
    {
      path: "/schedule/collect",
      name: "collect",
      component: () => import("@/views/schedule/collect/index.vue"),
      meta: {
        title: "采集任务",
        icon: "ri:task-line"
      }
    },
    {
      path: "/schedule/report",
      name: "report",
      component: () => import("@/views/schedule/report/index.vue"),
      meta: {
        title: "上报任务",
        icon: "ri:contacts-book-upload-line"
      }
    },
  ]
} satisfies RouteConfigsTable;
