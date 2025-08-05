<script setup lang="ts">
import {useColumns} from "./columns";

const {
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
} = useColumns();
</script>

<template>
  <el-card style="margin: 20px;border-radius: 10px">
    <el-card shadow="never" class="searchBar" style="margin-bottom: 40px;border: none;background: white">
      <el-button type="primary" class="float-right" @click="handleAdd">新增</el-button>
      <el-form style="display: flex;align-items: center;">
        <el-form-item class="mb-0 mr-5">
          <el-input placeholder="请输入服务名称"/>
        </el-form-item>
        <el-form-item class="mb-0 mr-5">
          <el-button type="primary">查询</el-button>

          <el-button @click="refreshTable">刷新</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <pure-table
      stripe
      row-key="id"
      alignWhole="center"
      showOverflowTooltip
      :size="tableSize as any"
      :loading="loading"
      :loading-config="loadingConfig"
      :height="tableSize === 'large' ? 600 : 530"
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
