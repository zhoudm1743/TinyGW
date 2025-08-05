<template>
    <el-card>
        <el-upload
            class="upload-demo"
            drag
            :action="url + '/api/debug/upgrade'"
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
        <div>
            <el-button :disabled="!fileListLength" @click="handleClick" style="margin-top: 10px; font-weight: 500">
                上传文件
            </el-button>
        </div>
    </el-card>
</template>

<script setup lang="ts">
import Cookies from "js-cookie";
import { TokenKey } from "@/utils/auth";
import { UploadFilled } from '@element-plus/icons-vue';
import { ref } from "vue";
import type { UploadInstance, UploadFile, UploadRawFile } from 'element-plus';
import { message } from "@/utils/message";


const url = import.meta.env.VITE_APP_BASE_API;
const token = JSON.parse(Cookies.get(TokenKey) || '{}').accessToken;

const uploadRef = ref<UploadInstance | null>(null);
const fileListLength = ref(0);

const headers = {
    Token: token,
};

const handleChange = (file: UploadFile, fileList: UploadFile[]) => {
    fileListLength.value = fileList.length;
};

const handleClick = () => {
    uploadRef.value?.submit();
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
</script>

<style scoped lang="scss">
.upload-demo {
    margin-bottom: 20px;
}

.el-upload__text {
    margin-top: 10px;
}

.el-upload__tip {
    color: #999;
}
</style>
