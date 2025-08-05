<script setup lang="ts">
import { ref } from "vue";


// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;
        reportName: string;
        ip: string;
        port: string;
        clientID: string;
        username: string;
        password: string;
        cron: string;
    };
}
// 声明 props 默认值
const props = withDefaults(defineProps<FormProps>(), {
    //查看不可编辑
    disabled: false,
    disabledLock: false,
    //数据初始化，防止父组件传递的数据为空时，子组件报错
    formInline: () => ({
        name: '',        //服务名称
        reportName: '',  //平台名称
        ip: '',          //ip地址
        port: '',        //端口号
        clientID: '',    //网关编号
        username: '',    //用户名
        password: '',    //密码
        cron: '',        //定时策略
    })
});
const newFormInline = ref(props.formInline);

</script>

<template>
    <el-form :model="newFormInline">
        <el-form-item label="任务名称">
            <el-input
                v-model="newFormInline.name"
                class=""
                :disabled="props.disabledLock"
                placeholder="请输入任务名称"
            />
        </el-form-item>
        <el-form-item label="平台名称">
            <el-input
                v-model="newFormInline.reportName"
                class=""
                :disabled="props.disabledLock"
                placeholder="请输入平台名称"
            />
        </el-form-item>
        <el-form-item label="ip地址" label-width="68px">
            <el-input
                v-model="newFormInline.ip"
                class=""
                :disabled="props.disabled"
                placeholder="请输入ip地址"
            />
        </el-form-item>
        <el-form-item label="端口号" label-width="68px">
            <el-input
                v-model="newFormInline.port"
                class=""
                :disabled="props.disabled"
                placeholder="请输入端口号"
            />
        </el-form-item>
        <el-form-item label="网关编号">
            <el-input
                v-model="newFormInline.clientID"
                class=""
                :disabled="props.disabled"
                placeholder="请输入网关编号"
            />
        </el-form-item>
        <el-form-item label="用户名" label-width="68px">
            <el-input
                v-model="newFormInline.username"
                class=""
                :disabled="props.disabledLock"
                placeholder="请输入用户名"
            />
        </el-form-item>
        <el-form-item label="密码" label-width="68px">
            <el-input
                v-model="newFormInline.password"
                class=""
                :disabled="props.disabledLock"
                placeholder="请输入密码"
            />
        </el-form-item>
        <el-form-item label="定时策略">
            <el-input
                v-model="newFormInline.cron"
                class=""
                :disabled="props.disabled"
                placeholder="请输入定时策略"
            />
        </el-form-item>
    </el-form>
</template>

<style scoped lang="scss">

</style>
