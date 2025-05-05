<script setup lang="ts">
import { detail } from "./columns";

const {
    loading,            // 加载状态
    columns,            // 表格列
    dataList,           // 表格数据
    select,             // 选中行
    hideVal,            // 隐藏列
    tableSize,          // 表格大小
    pagination,         // 分页配置
    loadingConfig,      // 加载配置
    paginationAlign,    // 分页对齐方式
    onChange,           // 选中行变化
    onSizeChange,       // 分页切换每页条数
    onCurrentChange,    // 分页切换页码
    handleAdd,          // 新增按钮
    refreshTable        // 刷新表格
} = detail();
</script>

<template>
    <el-card style="margin: 20px;border-radius: 10px">
        <el-card shadow="never" class="searchBar-1" style="margin-bottom: 10px;border: none;background: white">
<!--            <el-button type="primary" class="float-right" @click="handleAdd">新增</el-button>-->
            <el-form style="display: flex;align-items: center;">
<!--                <el-form-item class="mb-0 mr-5">-->
<!--                    <el-input placeholder="请输入接口名称"/>-->
<!--                </el-form-item>-->
                <el-form-item class="mb-0 mr-5">
<!--                    <el-button type="primary">查询</el-button>-->
                    <el-button type="primary" class="float-right" @click="handleAdd">新增</el-button>
                    <el-button @click="refreshTable">刷新</el-button>
                </el-form-item>
            </el-form>

<!--            <el-button type="primary" class="float-right" @click="handleAdd" style="margin-left: 10px">新增</el-button>-->
<!--            <el-button class="float-right" @click="refreshTable">刷新</el-button>-->
        </el-card>

        <pure-table
            stripe
            row-key="id"
            alignWhole="center"
            showOverflowTooltip
            :size="tableSize as any"
            :loading="loading"
            :loading-config="loadingConfig"
            :height="tableSize === 'large' ? 600 : 540"
            :data="dataList"
            :columns="columns"
            :pagination="pagination"
            @page-size-change="onSizeChange"
            @page-current-change="onCurrentChange"
        />
    </el-card>
</template>

<style scoped>
.mb-0{
    margin-bottom: 0;
}
</style>
