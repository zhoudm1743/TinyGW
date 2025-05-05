<script setup lang="ts">
import {onMounted, ref} from "vue";
import { collectorApi, deviceApi } from "@/api/basics"

//设备类型列表
const deviceTypeList = ref<any[]>([])
//采集接口列表
const collectorList = ref<any[]>([])

//初始化数据
onMounted(() => {
    getDeviceTypeList()
    getCollectorList()
})

// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;                       // 设备名称
        type: {
            name: string;                     // 设备类型名称
            driver: string;                   // 驱动名称
            properties: any[];                // 属性列表
        };
        address: string;                    // 通讯地址
        collector: {
            name: string;                     // 采集接口名称
        };
        alone: boolean;                     // 独立开关
        initialVal: number;                 // 初始值
        serial: {
            name: string;                     // 串口名称
            deviceName: string;               // 串口设备名称
            baudRate: number;                 // 波特率
            check: string;                    // 校验位
            dataBit: number;                  // 数据位
            stopBit: string;                  // 停止位
        };
    };
}
// 声明 props 默认值
const props = withDefaults(defineProps<FormProps>(), {
    //查看不可编辑
    disabled: false,
    disabledLock: false,
    //数据初始化，防止父组件传递的数据为空时，子组件报错
    formInline: () => ({
        name: '',                       //设备名称
        type: {                         //设备类型
            name: '',
            driver: '',
            properties: [],
        },
        address: '',                    //通讯地址
        collector: {
            name: '',
        },        //采集接口
        alone: false,                   //独立开关
        initialVal: 0,                  //初始值
        serial: {
            name: '',                     // 串口名称
            deviceName: '/dev/ttymxc3',   // 串口设备名称
            baudRate: 9600,               // 波特率
            check: 'N',                   // 校验位
            dataBit: 8,                   // 数据位
            stopBit: '1',                 // 停止位
        },
    })
});
const newFormInline = ref(props.formInline);

//获取设备类型列表
const getDeviceTypeList = async () => {
    const res = await deviceApi.list({});
    if (res.code === 200) {
        deviceTypeList.value = res.data;
    }
};
//获取采集接口列表
const getCollectorList = async () => {
    const res = await collectorApi.list({});
    if (res.code === 200) {
        collectorList.value = res.data;
    }
};

const handleDevice = (row:any) => {
    newFormInline.value.type = row
};

</script>

<template>
    <el-form :model="newFormInline">
        <el-form-item label="设备名称">
            <el-input
                v-model="newFormInline.name"
                class=""
                :disabled="props.disabledLock"
                placeholder="请输入接口名称"
            />
        </el-form-item>

        <el-form-item label="设备类型">
            <el-select
                v-model="newFormInline.type.name"
                class=""
                :disabled="props.disabled"
                placeholder="请选择设备类型"
                @change="handleDevice"
            >
                <el-option v-for="item in deviceTypeList" :key="item.id" :label="item.name" :value="item">
                    <span style="float: left">{{ item.name }}</span>
                    <span style="float: right; color: #8492a6; font-size: 13px">{{ item.driver }}</span>
                </el-option>
            </el-select>
        </el-form-item>

        <el-form-item label="通讯地址">
            <el-input
                v-model="newFormInline.address"
                class=""
                :disabled="props.disabled"
                placeholder="请输入通讯地址"
            />
        </el-form-item>

        <el-form-item label="采集接口">
            <el-select
                v-model="newFormInline.collector.name"
                class=""
                :disabled="props.disabled"
                placeholder="请选择采集接口"
            >
                <el-option v-for="item in collectorList" :key="item.id" :label="item.name" :value="item.name" />
            </el-select>
        </el-form-item>

        <el-form-item label="独立开关">
            <el-switch
                v-model="newFormInline.alone"
                :disabled="props.disabled"
                active-text="开"
                inactive-text="关"
            />
        </el-form-item>

        <el-form-item label="初始值" label-width="68px">
            <el-input-number
                v-model="newFormInline.initialVal"
                :disabled="props.disabled"
                controls-position="right"
                placeholder="请输入初始值"
            />
        </el-form-item>

        <el-form-item v-if="newFormInline.alone === true" label="串口名称">
            <el-input
                v-model="newFormInline.serial.name"
                class=""
                :disabled="props.disabled"
                placeholder="请输入串口名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="串口设备" label-width="68px">
            <el-input
                v-model="newFormInline.serial.deviceName"
                class=""
                :disabled="props.disabled"
                placeholder="请输入串口设备名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="波特率" label-width="68px">
            <el-input-number
                v-model="newFormInline.serial.baudRate"
                :disabled="props.disabled"
                controls-position="right"
                placeholder="请输入波特率"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="校验位" label-width="68px">
            <el-input
                v-model="newFormInline.serial.check"
                class=""
                :disabled="props.disabled"
                placeholder="请输入校验位"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="数据位" label-width="68px">
            <el-input-number
                v-model="newFormInline.serial.dataBit"
                :disabled="props.disabled"
                controls-position="right"
                placeholder="请输入数据位"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="停止位" label-width="68px">
            <el-input
                v-model="newFormInline.serial.stopBit"
                class=""
                :disabled="props.disabled"
                placeholder="请输入停止位"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.alone === true" label="校验位" label-width="68px">
            <el-input
                v-model="newFormInline.serial.check"
                class=""
                :disabled="props.disabled"
                placeholder="请输入校验位"
            />
        </el-form-item>
    </el-form>
</template>

<style scoped lang="scss">

</style>
