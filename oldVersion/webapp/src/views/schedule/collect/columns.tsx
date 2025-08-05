import { clone, delay } from "@pureadmin/utils";
import { ref, onMounted, reactive, watchEffect } from "vue";
import type { PaginationProps, LoadingConfig, Align } from "@pureadmin/table";
import { collect } from "@/api/schedule"
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog/index";
import save from "./save.vue";

export function useColumns() {
    const dataList = ref([]);
    const loading = ref(true);
    const select = ref("no");
    const hideVal = ref("nohide");
    const tableSize = ref("default");
    const paginationAlign = ref("right");

    const tableData = ref([])

    /** 刷新表格 */
    const refreshTable = () => {
        loading.value = true;
        delay(600).then(() => {
            fetchData().then(r => {
                message("刷新成功", { type: "success" });
                loading.value = false
            });
        });
    };

    /** 获取表格数据 */
    const fetchData = async () => {
        try {
            const res = await collect.page({
                page: pagination.currentPage,
                pageSize: pagination.pageSize
            })
            tableData.value = res.data
            pagination.total = tableData.value.count;
            const newList = [];
            dataList.value = [];
            Array.from({ length: 1 }).forEach(() => {
                newList.push(clone(tableData.value.lists, true));
            });
            newList.flat(Infinity).forEach((item, index) => {
                dataList.value.push({ id: index, ...item });
            });
        } catch (error) {
            message(`数据加载失败: ${error.message}`, { type: "error" })
        }
    };


    /** 新增 */
    function handleAdd() {
        addDialog({
            width: "auto",
            title: "新增",
            top: "25vh",
            contentRenderer: () => save,
            props: {
                // 赋默认值
                formInline: {
                    name: '', //任务名称
                    cron: '', //定时策略
                },
                disabled: false,
            },
            closeCallBack: ({ options, args }) => {
                const { formInline } = options.props;
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消新增`, { type: "warning" })
                } else if (args?.command === "sure") {
                    collect.add(formInline).then(() => {
                        message(`新增成功`, {type: "success"})
                        fetchData().then(r => {
                            message(`刷新成功`, {type: "success"})
                        })
                    })
                }
            }
        });
    }

    /** 编辑 */
    function handleEdit(param: any, row: undefined) {
        addDialog({
            width: "auto",
            title: "编辑",
            top: "25vh",
            contentRenderer: () => save,
            props: {
                // 赋默认值
                formInline: row,
                disabled: false,
                disabledLock: true
            },
            closeCallBack: ({ options, args }) => {
                // options.props 是响应式的
                const { formInline } = options.props;
                console.log(formInline.value)
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消编辑`, { type: "warning" })
                    fetchData().then(r => {})
                } else if (args?.command === "sure") {
                    collect.edit(row).then(() => {
                        message(`编辑成功`, {type: "success"})
                        fetchData().then(r => {
                            message(`刷新成功`, {type: "success"})
                        })
                    })
                }else {
                    fetchData().then(r => {})
                }
            }
        });
    }

    /** 删除 */
    async function handleDelete(param: any, row: undefined) {
        addDialog({
            title: "确定要删除吗？",
            top: "35vh",
            width: "30%",
            showClose: false,
            closeCallBack: async ({ args }) => {
                if (args?.command === "cancel") {
                    message("您取消了删除操作", { type: "warning" })
                } else if (args?.command === "sure") {
                    const res = await collect.delete(row.name);
                    if (res.code === 200) {
                        message("删除成功", { type: "success" })
                        await fetchData();
                    }
                } else {
                    message("您取消了删除操作", { type: "warning" })
                }
            },
            contentRenderer: () => <p>即将删除数据：{ row.name }</p>
        });
    }

    /** 启动/停止 */
    async function handleStart(param: any, row: undefined) {
        if (row.status === 1) {
            const res = await collect.stop(row.name)
            if (res.code === 200) {
                fetchData().then(r => {
                    message(`停止成功`, {type: "success"})
                })
            }
        } else {
            const res = await collect.start(row.name)
            if (res.code === 200) {
                fetchData().then(r => {
                    message(`启动成功`, {type: "success"})
                })
            }
        }
    }

    /** 表格列配置 */
    const columns: TableColumnList = [
        {
            type: "selection",
            align: "left",
            reserveSelection: true,
            hide: () => (select.value === "no")
        },
        {
            label: "任务名称",
            prop: "name",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "定时策略",
            prop: "cron",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "状态",
            prop: "status",
            formatter: (row) => (row.status === 1 ? "已启动" : "未启动"),
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "操作",
            align: "center",
            fixed: "right",
            width: "180px",
            cellRenderer: ({ index, row }) => (
                <>
                    <el-text style="padding:0 8px"  type="primary" onClick={() => handleEdit(index + 1, row)}>
                        编辑
                    </el-text>
                    <el-text style="padding:0 8px"  type="danger" onClick={() => handleDelete(index + 1, row)}>
                        删除
                    </el-text>
                    <el-text style="padding:0 8px"  type="warning" onClick={() => handleStart(index + 1, row)}>
                        {row.status === 1 ? "停止" : "启动"}
                    </el-text>
                </>
            )
        }
    ];

    /** 分页配置 */
    const pagination = reactive<PaginationProps>({
        pageSize: 10,
        currentPage: 1,
        pageSizes: [10, 15, 20],
        total: 0,
        align: "right",
        background: true,
        size: "default"
    });

    /** 加载动画配置 */
    const loadingConfig = reactive<LoadingConfig>({
        text: "正在加载第一页...",
        viewBox: "-10, -10, 50, 50",
        spinner: `
        <path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>
      `
        // svg: "",
        // background: rgba()
    });

    function onChange(val) {
        pagination.size = val;
    }

    /** 选择分页大小 */
    function onSizeChange(val) {
        fetchData().then(r => {});
    }

    /** 选择分页 */
    function onCurrentChange(val) {
        loadingConfig.text = `正在加载第${val}页...`;
        loading.value = true;

        pagination.currentPage = val;
        fetchData().then(r => {
            delay(500).then(() => {
                loading.value = false;
            });
        })
    }

    watchEffect(() => {
        pagination.align = paginationAlign.value as Align;
    });

    onMounted(async () => {
        await fetchData().then(() => {
            delay(500). then(() => {
                loading.value = false;
            });
        });
    });

    return {
        loading,
        columns,
        dataList,
        select,
        hideVal,
        tableSize,
        pagination,
        loadingConfig,
        paginationAlign,
        onChange,
        onSizeChange,
        onCurrentChange,
        handleAdd,
        refreshTable
    };
}
