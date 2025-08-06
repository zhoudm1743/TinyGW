<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NButton, NSpace, NDataTable, NCard, NInput, NSelect, NTag, useMessage, NModal, NForm, NFormItem, NTransfer } from 'naive-ui'
import { getCollectTaskList, deleteCollectTask, startCollectTask, stopCollectTask, createCollectTask, updateCollectTask, CollectTaskInfo, CollectTaskRequest } from '@/service/api/collectTask'
import { getDeviceList, DeviceInfo } from '@/service/api/device'

const message = useMessage()

// 表格加载状态
const loading = ref(false)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 40],
  onChange: (page: number) => {
    pagination.page = page
    loadData()
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    loadData()
  }
})

// 任务状态选项
const statusOptions = [
  { label: '全部', value: null },
  { label: '停止', value: 0 },
  { label: '运行中', value: 1 }
]

// 搜索条件
const searchParams = reactive({
  status: null as number | null
})

// 表格数据
const tableData = ref<CollectTaskInfo[]>([])

// 表单对话框
const showModal = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formData = reactive<CollectTaskRequest>({
  name: '',
  cron: '*/5 * * * *', // 默认每5分钟执行一次
  status: 0,
  deviceList: []
})

// 设备列表选项
const deviceOptions = ref<{ key: string; label: string }[]>([])
const selectedDeviceKeys = ref<string[]>([])

// 表单规则
const rules = {
  name: {
    required: true,
    message: '请输入任务名称',
    trigger: 'blur'
  },
  cron: {
    required: true,
    message: '请输入Cron表达式',
    trigger: 'blur'
  }
}

// 常用Cron表达式选项
const cronOptions = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每5分钟', value: '*/5 * * * *' },
  { label: '每10分钟', value: '*/10 * * * *' },
  { label: '每30分钟', value: '*/30 * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天凌晨1点', value: '0 1 * * *' },
  { label: '每周一凌晨2点', value: '0 2 * * 1' },
  { label: '每月1日凌晨3点', value: '0 3 1 * *' }
]

// 表格列定义
const columns = [
  {
    title: '任务名称',
    key: 'name'
  },
  {
    title: 'Cron表达式',
    key: 'cron'
  },
  {
    title: '状态',
    key: 'status',
    render(row: CollectTaskInfo) {
      const statusMap = {
        0: { text: '停止', type: 'error' },
        1: { text: '运行中', type: 'success' }
      }
      const status = statusMap[row.status as keyof typeof statusMap] || { text: '未知', type: 'warning' }
      return h(NTag, { type: status.type as 'error' | 'success' | 'warning' }, { default: () => status.text })
    }
  },
  {
    title: '设备数量',
    key: 'deviceCount',
    render(row: CollectTaskInfo) {
      return row.deviceList ? row.deviceList.length : 0
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render(row: CollectTaskInfo) {
      return new Date(row.createdAt).toLocaleString()
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row: CollectTaskInfo) {
      return h(NSpace, { justify: 'center' }, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              type: 'primary',
              onClick: () => handleEdit(row)
            },
            { default: () => '编辑' }
          ),
          row.status === 0 
            ? h(
                NButton,
                {
                  size: 'small',
                  type: 'success',
                  onClick: () => handleStart(row)
                },
                { default: () => '启动' }
              )
            : h(
                NButton,
                {
                  size: 'small',
                  type: 'warning',
                  onClick: () => handleStop(row)
                },
                { default: () => '停止' }
              ),
          h(
            NButton,
            {
              size: 'small',
              type: 'error',
              onClick: () => handleDelete(row)
            },
            { default: () => '删除' }
          )
        ]
      })
    }
  }
]

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      status: searchParams.status === null ? undefined : searchParams.status
    }
    const { isSuccess, data } = await getCollectTaskList(params)
    if (isSuccess && data) {
      tableData.value = data.data
      pagination.itemCount = data.total
    }
  } catch (error) {
    console.error('加载采集任务列表失败', error)
    message.error('加载采集任务列表失败')
  } finally {
    loading.value = false
  }
}

// 加载设备列表
async function loadDevices() {
  try {
    const { isSuccess, data } = await getDeviceList({ page: 1, pageSize: 1000 })
    if (isSuccess && data) {
      deviceOptions.value = data.data.map((item: DeviceInfo) => ({
        key: item.name,
        label: `${item.name} (${item.typeID})`
      }))
    }
  } catch (error) {
    console.error('加载设备列表失败', error)
  }
}

// 搜索
function handleSearch() {
  pagination.page = 1
  loadData()
}

// 重置搜索
function handleReset() {
  searchParams.status = null
  pagination.page = 1
  loadData()
}

// 打开新增表单
function handleAdd() {
  isEdit.value = false
  resetForm()
  showModal.value = true
}

// 打开编辑表单
function handleEdit(row: CollectTaskInfo) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    name: row.name,
    cron: row.cron,
    status: row.status,
    deviceList: row.deviceList || []
  })
  selectedDeviceKeys.value = row.deviceList || []
  showModal.value = true
}

// 启动任务
async function handleStart(row: CollectTaskInfo) {
  try {
    const { isSuccess } = await startCollectTask(row.name)
    if (isSuccess) {
      message.success('任务已启动')
      loadData()
    }
  } catch (error) {
    console.error('启动任务失败', error)
    message.error('启动任务失败')
  }
}

// 停止任务
async function handleStop(row: CollectTaskInfo) {
  try {
    const { isSuccess } = await stopCollectTask(row.name)
    if (isSuccess) {
      message.success('任务已停止')
      loadData()
    }
  } catch (error) {
    console.error('停止任务失败', error)
    message.error('停止任务失败')
  }
}

// 删除任务
async function handleDelete(row: CollectTaskInfo) {
  try {
    const { isSuccess } = await deleteCollectTask(row.name)
    if (isSuccess) {
      message.success('删除成功')
      loadData()
    }
  } catch (error) {
    console.error('删除任务失败', error)
    message.error('删除任务失败')
  }
}

// 重置表单
function resetForm() {
  formData.name = ''
  formData.cron = '*/5 * * * *'
  formData.status = 0
  formData.deviceList = []
  selectedDeviceKeys.value = []
}

// 处理设备选择变化
function handleDeviceChange(values: string[]) {
  selectedDeviceKeys.value = values
  formData.deviceList = values
}

// 提交表单
async function handleSubmit(e: MouseEvent) {
  e.preventDefault()
  if (!formRef.value) return
  
  // @ts-ignore
  formRef.value.validate(async (errors: any) => {
    if (errors) return
    
    try {
      if (isEdit.value) {
        // 编辑任务
        const { isSuccess } = await updateCollectTask(formData)
        if (isSuccess) {
          message.success('更新成功')
          showModal.value = false
          loadData()
        }
      } else {
        // 新增任务
        const { isSuccess } = await createCollectTask(formData)
        if (isSuccess) {
          message.success('创建成功')
          showModal.value = false
          loadData()
        }
      }
    } catch (error) {
      console.error('保存任务失败', error)
      message.error('保存任务失败')
    }
  })
}

onMounted(() => {
  loadData()
  loadDevices()
})
</script>

<template>
  <div class="p-4">
    <NCard title="采集任务管理" class="mb-4">
      <template #header-extra>
        <NButton type="primary" @click="handleAdd">
          新增任务
        </NButton>
      </template>
      <NSpace vertical>
        <NSpace>
          <NSelect 
            v-model:value="searchParams.status" 
            :options="statusOptions" 
            placeholder="任务状态" 
            style="width: 120px" 
          />
          <NButton type="primary" @click="handleSearch">搜索</NButton>
          <NButton @click="handleReset">重置</NButton>
        </NSpace>
        <NDataTable
          :columns="columns"
          :data="tableData"
          :loading="loading"
          :pagination="pagination"
          :bordered="false"
          striped
        />
      </NSpace>
    </NCard>

    <!-- 表单对话框 -->
    <NModal v-model:show="showModal" :title="isEdit ? '编辑采集任务' : '新增采集任务'" preset="card" style="width: 700px">
      <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
        <NFormItem label="任务名称" path="name">
          <NInput v-model:value="formData.name" placeholder="请输入任务名称" :disabled="isEdit" />
        </NFormItem>
        <NFormItem label="Cron表达式" path="cron">
          <NInput v-model:value="formData.cron" placeholder="请输入Cron表达式" />
        </NFormItem>
        <NFormItem label="常用表达式">
          <NSelect :options="cronOptions" placeholder="选择常用表达式" @update:value="value => formData.cron = value" />
        </NFormItem>
        <NFormItem label="设备列表">
          <NTransfer 
            v-model:value="selectedDeviceKeys" 
            :options="deviceOptions" 
            virtual-scroll 
            filterable 
            @update:value="handleDeviceChange"
          />
        </NFormItem>
        <NFormItem label="任务状态" path="status">
          <NSelect 
            v-model:value="formData.status" 
            :options="[
              { label: '停止', value: 0 },
              { label: '运行', value: 1 }
            ]" 
            placeholder="请选择任务状态" 
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">确定</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template> 