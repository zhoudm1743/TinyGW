// import { tableData } from "../data";
import { clone, delay } from "@pureadmin/utils";
import { ref, onMounted, reactive, watchEffect } from "vue";
import type { PaginationProps, LoadingConfig, Align } from "@pureadmin/table";
import { collectorApi } from "@/api/basics"
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog/index";
import save from "./save.vue";
import { log } from "console";

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
            const res = await collectorApi.page({
                page: pagination.currentPage,
                pageSize: pagination.pageSize
            })
            console.log(res)
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
            contentRenderer: () => save,
            props: {
                // 赋默认值
                formInline: {
                    name: '',
                    type: '',
                    serial: {
                        name: '',
                        deviceName: '/dev/ttymxc3',
                        baudRate: 9600,
                        check: 'N',
                        dataBit: 8,
                        stopBit: '1',
                    },
                    tcpClient: {
                        name: '',
                        ip: '192.168.100.21',
                        port: 7000,
                    },
                    tcpServer: {
                        name: '',
                        port: 51483,
                    },
                    mqtt: {
                        name: '',
                    },
                    channel: {
                        name: '',
                    },
                    timeout: 3000,
                    interval: 200,
                },
                disabled: false,
            },
            closeCallBack: ({ options, args }) => {
                const { formInline } = options.props;
                console.log(formInline)
                if (args?.command === "cancel") {
                    // 您点击了取消按钮
                    message(`已取消新增`, { type: "warning" })
                } else if (args?.command === "sure") {
                    collectorApi.add(formInline).then(() => {
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
                    collectorApi.edit(row).then(() => {
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
                    const res = await collectorApi.delete(row.name);
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

    /** 表格列配置 */
    const columns: TableColumnList = [
        {
            type: "selection",
            align: "left",
            reserveSelection: true,
            hide: () => (select.value === "no" ? true : false)
        },
        {
            label: "接口名称",
            prop: "name",
            hide: () => (hideVal.value === "hideName" ? true : false)
        },
        {
            label: "接口类型",
            prop: "type",
            hide: () => (hideVal.value === "hideAddress" ? true : false)
        },
        {
            label: "参数",
            prop: "params",
            align: "left",
            formatter: (row) => {
                if (row.type === 'Serial') {
                    return '物理接口：' + row.serial.deviceName + ' 波特率：' + row.serial.baudRate + ' 数据位：' + row.serial.dataBit + ' 停止位：' + row.serial.stopBit + ' 校验位：' + row.serial.check;
                } else if (row.type === 'Mqtt') {
                    return 'MQTT名称：' + row.mqtt.name;
                } else if (row.type === 'Channel') {
                    return '透传：' + row.channel.name;
                } else {
                    return '目标IP：' + row.tcpClient.ip + ' 端口：' + row.tcpClient.port;
                }
            }
        },
        {
            label: "超时时间",
            prop: "timeout",
            hide: () => (hideVal.value === "hideAddress" ? true : false)
        },
        {
            label: "通讯间隔",
            prop: "interval",
            hide: () => (hideVal.value === "hideAddress" ? true : false)
        },
        {
            label: "操作",
            align: "center",
            fixed: "right",
            cellRenderer: ({ index, row }) => (
                <>
                    <el-text style="padding:0 8px" text onClick={() => handleShow(index + 1, row)}>
                        查看
                    </el-text>
                    <el-text style="padding:0 8px" text type="primary" onClick={() => handleEdit(index + 1, row)}>
                        编辑
                    </el-text>
                    <el-text style="padding:0 8px" text type="danger" onClick={() => handleDelete(index + 1, row)}>
                        删除
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
        // await delay(600). then(() => {
        //   const newList = [];
        //   Array.from({ length: 1 }).forEach(() => {
        //     newList.push(clone(tableData.value.data, true));
        //   });
        //   newList.flat(Infinity).forEach((item, index) => {
        //     dataList.value.push({ id: index, ...item });
        //   });
        //   loading.value = false;
        // });
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
