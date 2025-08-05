<script setup lang="ts">
import { ref } from "vue";

// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;                     // 接口名称
        type: string;                     // 接口类型
        // register_code: string;            // 设备码
        serial: {
            name: string;                   // 串口名称
            deviceName: string;             // 串口设备名称
            baudRate: number;               // 波特率
            check: string;                  // 校验位
            dataBit: number;                // 数据位
            stopBit: string;                // 停止位
        };
        tcpClient: {
            name: string;                   // 名称
            ip: string;                     // IP地址
            port: number;                   // 端口
        };
        tcpServer: {
            name: string;                   // 名称
            port: number;                   // 端口
        };
        mqtt: {
            name: string;                   // 名称
        };
        channel: {
            name: string;                   // 名称
        };
        timeout: number;                  // 超时时间
        interval: number;                 // 通讯间隔
    };
}

// 声明 props 默认值
const props = withDefaults(defineProps<FormProps>(), {
    //查看不可编辑
    disabled: false,
    disabledLock: false,
    //数据初始化，防止父组件传递的数据为空时，子组件报错
    formInline: () => ({
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
    })
});
const newFormInline = ref(props.formInline);

</script>

<template>
    <el-form :model="newFormInline">
        <el-form-item label="接口名称">
            <el-input
                    v-model="newFormInline.name"
                    class=""
                    :disabled="props.disabledLock"
                    placeholder="请输入接口名称"
            />
        </el-form-item>
        <el-form-item label="接口类型">
            <el-select
                    v-model="newFormInline.type"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请选择接口类型"
            >
                <el-option label="Channel" value="Channel" />
                <el-option label="Mqtt" value="Mqtt" />
                <el-option label="Serial" value="Serial" />
                <el-option label="TcpClient" value="TcpClient" />
                <el-option label="TcpServer" value="TcpServer" />
            </el-select>
        </el-form-item>

        <el-form-item v-if="newFormInline.type === 'Channel'" label="名称" label-width="68px">
            <el-input
                    v-model="newFormInline.channel.name"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入名称"
            />
        </el-form-item>

        <el-form-item v-if="newFormInline.type === 'Mqtt'" label="名称" label-width="68px">
            <el-input
                    v-model="newFormInline.mqtt.name"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入名称"
            />
        </el-form-item>

        <el-form-item v-if="newFormInline.type === 'Serial'" label="串口名称">
            <el-input
                    v-model="newFormInline.serial.name"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入串口名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'Serial'" label="设备名称">
            <el-input
                    v-model="newFormInline.serial.deviceName"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入串口设备名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'Serial'" label="波特率" label-width="68px">
            <el-input
                    v-model="newFormInline.serial.baudRate"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入波特率"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'Serial'" label="数据位" label-width="68px">
            <el-input
                    v-model="newFormInline.serial.dataBit"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入数据位"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'Serial'" label="停止位" label-width="68px">
            <el-input
                    v-model="newFormInline.serial.stopBit"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入停止位"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'Serial'" label="校验位" label-width="68px">
            <el-input
                    v-model="newFormInline.serial.check"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入校验位"
            />
        </el-form-item>

        <el-form-item v-if="newFormInline.type === 'TcpClient'" label="名称" label-width="68px">
            <el-input
                    v-model="newFormInline.tcpClient.name"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'TcpClient'" label="IP地址" label-width="68px">
            <el-input
                    v-model="newFormInline.tcpClient.ip"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入IP地址"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'TcpClient'" label="端口" label-width="68px">
            <el-input
                    v-model="newFormInline.tcpClient.port"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入端口"
            />
        </el-form-item>

        <el-form-item v-if="newFormInline.type === 'TcpServer'" label="名称" label-width="68px">
            <el-input
                    v-model="newFormInline.tcpServer.name"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入名称"
            />
        </el-form-item>
        <el-form-item v-if="newFormInline.type === 'TcpServer'" label="端口" label-width="68px">
            <el-input
                    v-model="newFormInline.tcpServer.port"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入端口"
            />
        </el-form-item>

        <el-form-item label="超时时间">
            <el-input
                    v-model="newFormInline.timeout"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入超时时间"
            >
                <template #append>毫秒</template>
            </el-input>
        </el-form-item>
        <el-form-item label="通讯时间">
            <el-input
                    v-model="newFormInline.interval"
                    class=""
                    :disabled="props.disabled"
                    placeholder="请输入通讯时间"
            >
                <template #append>毫秒</template>
            </el-input>
        </el-form-item>
    </el-form>
</template>

<style scoped lang="scss">

</style>
