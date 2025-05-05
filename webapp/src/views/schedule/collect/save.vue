<script setup lang="ts">
import { ref } from "vue";


// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;
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
        name: '', //任务名称
        cron: '', //定时策略
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
