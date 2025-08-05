<template>
    <el-row :gutter="20" style="margin: 20px;">
        <el-col :span="24" style="height: 180px">
            <el-card>
                <div style="display: flex;align-items: center;margin-bottom: 10px">
                    <el-avatar shape="square" :size="60" :src="avatar" />
                    <div class="ml-10">
                        <span class="text-20 opacity-80">欢迎进入西奥网关</span>
                        <span class="ml-4 opacity-50">今日事，今日毕。</span>
                    </div>
                </div>
                <div>
                    <p class="text-14">{{ quote }}</p>
                    <p class="text-right text-12">{{ author }}</p>
                </div>
            </el-card>
        </el-col>
        <el-col :span="8">
            <el-card title="网关信息" class="w-30%">
                <template #header>
                    <div class="card-header flex justify-between items-center">
                        <span>网关信息</span>
                        <el-popconfirm
                            @confirm="restart"
                            title="确定重启吗？"
                        >
                            <template #reference>
                                <el-button type="primary" size="small">重启</el-button>
                            </template>
                        </el-popconfirm>
                    </div>
                </template>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>网关编号:</p>
                        <p>{{ state.info.Name }}</p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>硬件版本:</p>
                        <p>{{ state.info.HardVer }}</p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>软件版本:</p>
                        <p>{{ state.info.SoftVer }}</p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>系统时间:</p>
                        <p>
                            {{ state.info.SystemRTC }}
                            <el-tag type="success" size="default" @click="timing">校准</el-tag>
                        </p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>系统运行:</p>
                        <p>{{ state.info.RunTime }}</p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>串口功能:</p>
                        <p>{{ state.info.OpenEveryTime }}</p>
                    </div>
                </el-card>
                <el-card shadow="never" class="border-none">
                    <div style="display: flex;justify-content: space-between">
                        <p>软件更新:</p>
                        <p>{{ state.info.SN }}</p>
                    </div>
                </el-card>
            </el-card>
        </el-col>
        <el-col :span="16">
            <el-card>
                <template #header>
                    <div class="card-header flex justify-between items-center">
                        <span>系统资源使用情况</span>
                    </div>
                </template>
                <el-row :gutter="20">
                    <el-col :span="8" style="max-height: 449px;height: 449px">
                        <div ref="diskChart" style="width: 100%; height: 400px;margin-top: 50px"></div>
                    </el-col>
                    <el-col :span="8" style="max-height: 449px;height: 449px">
                        <div ref="memChart" style="width: 100%; height: 400px;margin-top: 50px"></div>
                    </el-col>
                    <el-col :span="8" style="max-height: 449px;height: 449px">
                        <div ref="deviceChart" style="width: 100%; height: 400px;margin-top: 50px"></div>
                    </el-col>
                </el-row>
            </el-card>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { systemInfoApi } from "@/api/system"
import {onMounted, reactive, ref} from "vue"
import { message } from "@/utils/message"
import { useEcharts } from "@/plugins/echarts"
import * as echarts from "echarts/core";
import {isAllEmpty} from "@pureadmin/utils";
import {useUserStoreHook} from "@/store/modules/user";
import Avatar from "@/assets/user.jpg";

const avatar = isAllEmpty(useUserStoreHook()?.avatar)
    ? Avatar
    : useUserStoreHook()?.avatar;

/**声明接口类型*/
//网关信息类型
interface SystemInfo {
    Name: string;
    HardVer: string;
    SoftVer: string;
    SystemRTC: string;
    RunTime: string;
    OpenEveryTime: string;
    SN: string;

    DiskUse: string;        //磁盘使用率
    MemUse: string;         //内存使用率
    DeviceOnline: string;   //设备在线数
    DeviceTotal: string;    //设备总数
}
//query接口返回类型
interface ApiResponse {
    code: number;
    data: SystemInfo;
}
const state = reactive({
    info: {} as SystemInfo,
})

const diskChart = ref(null);
const memChart = ref(null);
const deviceChart = ref(null);

const quote = ref("今天，你打算做什么？")
const author = ref("西奥")
/** 随机励志语 */
const quoteAuthorPairs = [
    {
        quote: "一个人几乎可以在任何他怀有无限热忱的事情上成功。",
        author: "查尔斯·史考伯"
    },
    {
        quote: "成功并非终点，勇气才是继续前行的持续意愿。",
        author: "埃德温·阿灵顿·罗宾逊"
    },
    {
        quote: "只有那些敢于大步跨越的人，才能发现自己的潜力。",
        author: "安东尼·罗宾斯"
    },
    {
        quote: "成功的秘诀在于坚持不懈奋斗。",
        author: "查尔斯·狄更斯"
    },
    {
        quote: "无论你做什么，都要充满热情。",
        author: "阿尔伯特·爱因斯坦"
    },
    {
        quote: "生活不是等待风暴过去，而是学会在雨中起舞。",
        author: "弗朗西斯科·德·昆卡"
    },
    {
        quote: "不要等待机会，而要创造机会。",
        author: "乔治·伯纳德·肖"
    },
    {
        quote: "生活最大的危险在于一个空虚的心。",
        author: "王尔德"
    },
]
const updateRandomQuote = () => {
    const randomIndex = Math.floor(Math.random() * quoteAuthorPairs.length);
    const result = quoteAuthorPairs[randomIndex];
    quote.value = `"${result.quote}"`;
    author.value = result.author;
}

/** 获取系统信息 */
const getSystemInfo = async () => {
    const res = await systemInfoApi.systemStatus() as ApiResponse
    if (res.code === 200) {
        state.info = res.data
        initDiskChart();
        initMemChart();
        initDeviceChart();
    }
}

/** 重启 */
const restart = async () => {
    const res = await systemInfoApi.reboot() as ApiResponse
    if (res.code === 200) {
        message('重启成功', { type: 'success' })
    }
}

/** 校准时间 */
const timing = async () => {
    const res = await systemInfoApi.timing() as ApiResponse
    if (res.code === 200) {
        state.info.SystemRTC = res.data
    }
}

/** 初始化磁盘使用率图表 */
const initDiskChart = () => {
    if (diskChart.value) {
        const myChart = echarts.init(diskChart.value);
        myChart.setOption({
            title: {
                text: '磁盘使用率',
                left: 'center'
            },
            tooltip: {
                trigger: 'item'
            },
            series: [
                {
                    name: '磁盘使用率',
                    type: 'pie',
                    radius: '50%',
                    data: [
                        { value: parseFloat(state.info.DiskUse), name: '已使用' },
                        { value: 100 - parseFloat(state.info.DiskUse), name: '未使用' }
                    ],
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)'
                        }
                    }
                }
            ]
        });
    }
}

/** 初始化内存使用率图表 */
const initMemChart = () => {
    if (memChart.value) {
        const myChart = echarts.init(memChart.value);
        myChart.setOption({
            title: {
                text: '内存使用率',
                left: 'center'
            },
            tooltip: {
                trigger: 'item'
            },
            series: [
                {
                    name: '内存使用率',
                    type: 'pie',
                    radius: '50%',
                    data: [
                        { value: parseFloat(state.info.MemUse), name: '已使用' },
                        { value: 100 - parseFloat(state.info.MemUse), name: '未使用' }
                    ],
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)'
                        }
                    }
                }
            ]
        });
    }
}

/** 初始化设备在线数图表 */
const initDeviceChart = () => {
    if (deviceChart.value) {
        const myChart = echarts.init(deviceChart.value);
        const online = parseInt(state.info.DeviceOnline);
        const total = parseInt(state.info.DeviceTotal);
        const offline = total - online;

        myChart.setOption({
            title: {
                text: '设备在线数',
                left: 'center'
            },
            tooltip: {
                trigger: 'item'
            },
            series: [
                {
                    name: '设备在线数',
                    type: 'pie',
                    radius: '50%',
                    data: [
                        { value: online, name: '在线' },
                        { value: offline, name: '离线' }
                    ],
                    emphasis: {
                        itemStyle: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)'
                        }
                    }
                }
            ]
        });
    }
}

onMounted(() => {
    getSystemInfo()
    updateRandomQuote()
})
</script>

<style scoped>
.border-none{
    border: none;
}
</style>
