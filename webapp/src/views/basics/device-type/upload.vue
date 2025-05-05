<template>
        <el-upload
            class="upload-demo"
            drag
            :action="action"
            multiple
            :directory="true"
            accept=".zip"
            :headers="headers"
            name="FileName"
            :limit="1"
            :auto-upload="false"
            :on-change="handleChange"
            :on-success="handleFinish"
            ref="uploadRef"
        >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">点击或者拖动zip文件到该区域来上传</div>
            <template #tip>
                <div class="el-upload__tip">请上传.zip格式的文件</div>
            </template>
        </el-upload>
        <div style="display: flex;justify-content: flex-end;">
            <el-button :disabled="!fileListLength" @click="handleClick" style="margin-top: 10px; font-weight: 500">
                上传文件
            </el-button>
        </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {UploadFilled} from "@element-plus/icons-vue";
import Cookies from "js-cookie";
import { TokenKey } from "@/utils/auth";
import type { UploadInstance, UploadFile, UploadRawFile } from 'element-plus';
import { message } from "@/utils/message";
import { closeDialog, dialogStore } from "@/components/ReDialog/index";

const emit = defineEmits(['refresh'])

// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;
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
    })
});
// const newFormInline = ref(props.formInline.name);

const url = import.meta.env.VITE_APP_BASE_API;
const token = JSON.parse(Cookies.get(TokenKey) || '{}').accessToken;

const uploadRef = ref<UploadInstance | null>(null);
const fileListLength = ref(0);

const headers = {Token: token,};
const action = `${url}/api/device-type/upload/${props.formInline.name}`;

const handleChange = (file: UploadFile, fileList: UploadFile[]) => {
    fileListLength.value = fileList.length;
};

const handleFinish = (response: any, file: UploadRawFile) => {
    console.log(response);
    console.log(file);
    if (response.code === 200) {
        message('上传成功', { type: 'success' });
    }else {
        message(response.msg, { type: 'error' });
    }
};

const handleClick = () => {
    uploadRef.value?.submit();
    closeDialog(dialogStore.value, 0);
    emit('refresh', true);
};
</script>

<style scoped lang="scss">

</style>
