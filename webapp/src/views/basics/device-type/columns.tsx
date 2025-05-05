import { clone, delay } from "@pureadmin/utils";
import { ref, onMounted, reactive, watchEffect } from "vue";
import type { PaginationProps, LoadingConfig, Align } from "@pureadmin/table";
import {deviceApi} from "@/api/basics"
import { message } from "@/utils/message";
import {addDialog, dialogStore} from "@/components/ReDialog/index";
import save from "./save.vue";
import detailSave from "./detailSave.vue";
import upload from "./upload.vue";

const deailList = ref([]);

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
            const res = await deviceApi.page({
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
            width: "35%",
            title: "新增",
            top: "20vh",
            contentRenderer: () => save,
            props: {
                // 赋默认值
                formInline: {
                    name: '',   //类型名称
                },
                disabled: false,
            },
            closeCallBack: ({ options, args }) => {
                const { formInline } = options.props;
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消新增`, { type: "warning" })
                } else if (args?.command === "sure") {
                    deviceApi.add(formInline).then(() => {
                        message(`新增成功`, {type: "success"})
                        fetchData().then(r => {
                            message(`刷新成功`, {type: "success"})
                        })
                    })
                }
            }
        });
    }

    /** 查看 */
    function handleShow(param: any, row: undefined) {
        addDialog({
            width: "35%",
            title: "查看",
            top: "20vh",
            contentRenderer: () => save,
            props: {
                formInline: row,
                disabled: true,
                disabledLock: true
            },
            closeCallBack: ({ options, args }) => {
                // options.props 是响应式的
                const { formInline } = options.props;
                console.log(formInline.value)
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                } else if (args?.command === "sure") {
                }
            }
        });

    }

    /** 编辑 */
    function handleEdit(param: any, row: undefined) {
        addDialog({
            width: "35%",
            title: "编辑",
            top: "20vh",
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
                    deviceApi.edit(row).then(() => {
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
                    const res = await deviceApi.delete(row.name);
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

    /** 上传驱动 */
    function handleUpload(param: any, row: undefined) {
        addDialog({
            width: "35%",
            title: "上传驱动",
            top: "20vh",
            hideFooter: true,
            props: {
                formInline: {
                    name: row.name,
                },
            },
            contentRenderer: () => upload,
            closeCallBack: ({ options, args }) => {
                // options.props 是响应式的
                const { formInline } = options.props;
                console.log(formInline.value)
                if (args?.command === "cancel") {
                    fetchData().then(r => {})
                } else if (args?.command === "sure") {
                    fetchData().then(r => {})
                }else {
                    fetchData().then(r => {})
                }
            }
        })
    }

    /** 弹出层关闭回调 */


    /** 表格列配置 */
    const columns: TableColumnList = [
        {
            type: "selection",
            align: "left",
            reserveSelection: true,
            hide: () => (select.value === "no")
        },
        {
            label: "设备",
            prop: "name",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "关联插件",
            prop: "driver",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "操作",
            align: "center",
            fixed: "right",
            width: "180px",
            cellRenderer: ({ index, row }) => (
                <div style="display: flex;justify-content: space-between;">
                    <el-text style="" type="primary" text="primary" onClick={() => handleEdit(index + 1, row)}>
                        编辑
                    </el-text>
                    <el-text style="" type="danger" text="danger" onClick={() => handleDelete(index + 1, row)}>
                        删除
                    </el-text>
                    <el-text style="" type="warning" text="warning" onClick={() => handleUpload(index + 1, row)}>
                        上传协议
                    </el-text>
                </div>
            )
        },
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

    /** 获取当前行信息 */
    function onRowClick(row, column, event) {
        deailList.value = row;
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
        refreshTable,
        onRowClick
    };
}

export function detail() {
    const dataList = ref([]);
    const loading = ref(false)
    const select = ref("no");
    const hideVal = ref("nohide");
    const tableSize = ref("default");
    const paginationAlign = ref("right");

    const tableData = ref([])

    /** 刷新表格 */
    const refreshTable = () => {
        if (deailList.value.length === 0) {
            message("请选择设备", { type: "warning" });
            return;
        }
        loading.value = true;
        delay(200).then(() => {
            newData().then(r => {
                message("刷新成功", { type: "success" });
                fetchData().then(r => {
                    delay(200).then(() => {
                        loading.value = false;
                    });
                });
            });
        });
    };

    /** 获取新数据 */
    const newData = async () => {
        const name = deailList.value.name
        const res = await deviceApi.page()
        if (res.code === 200) {
            let lists = res.data.lists; 
            for (let i = 0; i < lists; i++) {
                if (lists[i].name === name) {
                    deailList.value[i] = lists[i]
                }
            }
        }
        console.log(deailList.value)
    }

    /** 新增 */
    function handleAdd() {
        if (deailList.value.length === 0) {
            message("请选择设备", { type: "warning" });
            return;
        }
        addDialog({
            width: "35%",
            title: "新增",
            top: "10vh",
            contentRenderer: () => detailSave,
            props: {
                // 赋默认值
                formInline: {
                    name: '',           //规则名称
                    description: '',    //描述
                    type: 'double',           //数据类型
                    length: 13,          //数据长度
                    decimal: 2,         //小数位数
                    unit: '',           //计量单位
                    value: 0,           //值
                    reported: true,     //是否上报
                    isAlarm: false,     //是否报警
                    threshold: 0,   //预警值
                    autoCalc: false,    //是否自动计算
                },
                disabled: false,
            },
            closeCallBack: ({ options, args }) => {
                const { formInline } = options.props;
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消新增`, { type: "warning" })
                } else if (args?.command === "sure") {
                    const name = deailList.value.name
                    deviceApi.addDetail(name, formInline).then(() => {
                        message(`新增成功`, {type: "success"})
                        refreshTable()
                    })
                }
            }
        });
    }

    /** 查看 */
    function handleShow(param: any, row: undefined) {
        addDialog({
            width: "35%",
            title: "查看",
            top: "10vh",
            contentRenderer: () => detailSave,
            props: {
                formInline: row,
                disabled: true,
                disabledLock: true
            },
            closeCallBack: ({ options, args }) => {
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                } else if (args?.command === "sure") {
                }
            }
        });

    }

    /** 编辑 */
    function handleEdit(param: any, row: undefined) {
        addDialog({
            width: "35%",
            title: "编辑",
            top: "10vh",
            contentRenderer: () => detailSave,
            props: {
                // 赋默认值
                formInline: row,
                disabled: false,
                disabledLock: true
            },
            closeCallBack: ({ options, args }) => {
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消编辑`, { type: "warning" })
                    refreshTable()
                } else if (args?.command === "sure") {
                    const name = deailList.value.name

                    deviceApi.editDetail(name, row).then(() => {
                        message(`编辑成功`, {type: "success"})
                        refreshTable()
                    })
                }else {
                    refreshTable()
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
                    const name = deailList.value.name

                    const res = await deviceApi.deleteDetail(name,row.id);
                    if (res.code === 200) {
                        message("删除成功", { type: "success" })
                        await refreshTable()
                    }
                } else {
                    message("您取消了删除操作", { type: "warning" })
                }
            },
            contentRenderer: () => <p>即将删除数据：{ row.description }</p>
        });
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
            label: "属性名称",
            prop: "name",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "描述",
            prop: "description",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "数据类型",
            prop: "type",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "数据长度",
            prop: "length",
            hide: () => (hideVal.value === "hideAddress")
        },

        {
            label: "小数位数",
            prop: "decimal",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "计量单位",
            prop: "unit",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "最新数值",
            prop: "value",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "是否上报",
            prop: "reported",
            hide: () => (hideVal.value === "hideAddress"),
            cellRenderer: ({ index, row }) => (
                <el-checkbox v-model={row.reported} disabled></el-checkbox>
            )
        },
        {
            label: "预警值",
            prop: "threshold",
            hide: () => (hideVal.value === "hideAddress")
        },
        {
            label: "警报",
            prop: "isAlarm",
            hide: () => (hideVal.value === "hideAddress"),
            cellRenderer: ({ index, row }) => (
                <el-checkbox v-model={row.isAlarm} disabled></el-checkbox>
            )
        },
        {
            label: "操作",
            align: "center",
            fixed: "right",
            width: "130px",
            cellRenderer: ({ index, row }) => (
                <div style="display: flex;justify-content: space-between;">
                    <el-text size="default" onClick={() => handleShow(index + 1, row)}>
                        查看
                    </el-text>
                    <el-text size="default" type="primary" onClick={() => handleEdit(index + 1, row)}>
                        编辑
                    </el-text>
                    <el-text size="default" type="danger" onClick={() => handleDelete(index + 1, row)}>
                        删除
                    </el-text>
                </div>
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
        // fetchData().then(r => {
            delay(500).then(() => {
                loading.value = false;
            });
        // })
    }

    watchEffect(() => {
        pagination.align = paginationAlign.value as Align;
    });

    /** 获取表格数据 */
    const fetchData = async () => {
        if (deailList.value.length === 0) return;

        tableData.value = properties.value
        pagination.total = properties.value?.length;
        const newList = [];
        dataList.value = [];
        Array.from({ length: 1 }).forEach(() => {
            newList.push(clone(tableData.value, true));
        });
        newList.flat(Infinity).forEach((item, index) => {
            dataList.value.push({ id: index + 1, ...item });
        });
    };

    /** 监听deailList值的变化 */
    const properties = ref([])

    watchEffect(() => {
        if (deailList.value) {
            properties.value = deailList.value.properties
            fetchData()
        }
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
        refreshTable,
    };
}
