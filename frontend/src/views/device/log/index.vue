<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NButton, NSpace, NDataTable, NCard, NInput, NSelect, NDatePicker, useMessage, NTag, NModal, NDescriptions, NDescriptionsItem, NCode } from 'naive-ui'
import { getDeviceLogList, deleteDeviceLog, clearDeviceLogs, DeviceLogInfo } from '@/service/api/deviceLog'

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

// 日志状态选项
const statusOptions = [
  { label: '全部', value: undefined },
  { label: '成功', value: 0 },
  { label: '失败', value: 1 }
]

// 搜索条件
const searchParams = reactive({
  deviceID: '',
  status: undefined as number | undefined,
  property: '',
  timeRange: null as [number, number] | null
})

// 表格数据
const tableData = ref<DeviceLogInfo[]>([])

// 详情对话框
const showDetailModal = ref(false)
const currentLog = ref<DeviceLogInfo | null>(null)

// 表格列定义
const columns = [
  {
    title: '设备ID',
    key: 'deviceID'
  },
  {
    title: '设备名称',
    key: 'deviceName'
  },
  {
    title: '属性',
    key: 'property'
  },
  {
    title: '值',
    key: 'value'
  },
  {
    title: '状态',
    key: 'status',
    render(row: DeviceLogInfo) {
      const statusMap = {
        0: { text: '成功', type: 'success' },
        1: { text: '失败', type: 'error' }
      }
      const status = statusMap[row.status as keyof typeof statusMap] || { text: '未知', type: 'warning' }
      return h(NTag, { type: status.type as 'success' | 'error' | 'warning' }, { default: () => status.text })
    }
  },
  {
    title: '错误信息',
    key: 'error',
    render(row: DeviceLogInfo) {
      return row.error || '-'
    }
  },
  {
    title: '时间',
    key: 'time',
    render(row: DeviceLogInfo) {
      return new Date(row.time).toLocaleString()
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row: DeviceLogInfo) {
      return h(NSpace, { justify: 'center' }, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              type: 'info',
              onClick: () => handleView(row)
            },
            { default: () => '查看详情' }
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
    const params: any = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      deviceID: searchParams.deviceID || undefined,
      status: searchParams.status,
      property: searchParams.property || undefined
    }
    
    if (searchParams.timeRange) {
      params.startTime = searchParams.timeRange[0]
      params.endTime = searchParams.timeRange[1]
    }
    
    const { isSuccess, data } = await getDeviceLogList(params)
    if (isSuccess && data) {
      tableData.value = data.data
      pagination.itemCount = data.total
    }
  } catch (error) {
    console.error('加载设备日志列表失败', error)
    message.error('加载设备日志列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
function handleSearch() {
  pagination.page = 1
  loadData()
}

// 重置搜索
function handleReset() {
  searchParams.deviceID = ''
  searchParams.status = undefined
  searchParams.property = ''
  searchParams.timeRange = null
  pagination.page = 1
  loadData()
}

// 查看日志详情
function handleView(row: DeviceLogInfo) {
  currentLog.value = row
  showDetailModal.value = true
}

// 删除日志
async function handleDelete(row: DeviceLogInfo) {
  try {
    const { isSuccess } = await deleteDeviceLog(row.id)
    if (isSuccess) {
      message.success('删除成功')
      loadData()
    }
  } catch (error) {
    console.error('删除日志失败', error)
    message.error('删除日志失败')
  }
}

// 清除历史日志
async function handleClearLogs() {
  try {
    // 默认清除7天前的日志
    const sevenDaysAgo = Date.now() - 7 * 24 * 60 * 60 * 1000
    const { isSuccess } = await clearDeviceLogs(sevenDaysAgo)
    if (isSuccess) {
      message.success('历史日志清除成功')
      loadData()
    }
  } catch (error) {
    console.error('清除历史日志失败', error)
    message.error('清除历史日志失败')
  }
}

// 格式化JSON字符串
function formatJSON(str: string) {
  try {
    if (!str) return ''
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch (e) {
    return str
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-4">
    <NCard title="设备日志管理" class="mb-4">
      <template #header-extra>
        <NButton type="warning" @click="handleClearLogs">
          清除历史日志
        </NButton>
      </template>
      <NSpace vertical>
        <NSpace>
          <NInput v-model:value="searchParams.deviceID" placeholder="设备ID" />
          <NInput v-model:value="searchParams.property" placeholder="属性" />
          <NSelect 
            v-model:value="searchParams.status" 
            :options="statusOptions" 
            placeholder="状态" 
            style="width: 120px" 
          />
          <NDatePicker 
            v-model:value="searchParams.timeRange" 
            type="datetimerange" 
            clearable 
            placeholder="选择时间范围"
            style="width: 300px"
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

    <!-- 详情对话框 -->
    <NModal v-model:show="showDetailModal" title="日志详情" preset="card" style="width: 700px">
      <NDescriptions v-if="currentLog" bordered>
        <NDescriptionsItem label="设备ID">
          {{ currentLog.deviceID }}
        </NDescriptionsItem>
        <NDescriptionsItem label="设备名称">
          {{ currentLog.deviceName }}
        </NDescriptionsItem>
        <NDescriptionsItem label="属性">
          {{ currentLog.property }}
        </NDescriptionsItem>
        <NDescriptionsItem label="值">
          {{ currentLog.value }}
        </NDescriptionsItem>
        <NDescriptionsItem label="状态">
          <NTag :type="currentLog.status === 0 ? 'success' : 'error'">
            {{ currentLog.status === 0 ? '成功' : '失败' }}
          </NTag>
        </NDescriptionsItem>
        <NDescriptionsItem label="时间">
          {{ new Date(currentLog.time).toLocaleString() }}
        </NDescriptionsItem>
        <NDescriptionsItem label="错误信息" v-if="currentLog.error">
          {{ currentLog.error }}
        </NDescriptionsItem>
        <NDescriptionsItem label="原始命令" span="2">
          <NCode :code="currentLog.originCmd" language="json" />
        </NDescriptionsItem>
        <NDescriptionsItem label="原始响应" span="2">
          <NCode :code="formatJSON(currentLog.originRes)" language="json" />
        </NDescriptionsItem>
        <NDescriptionsItem label="响应结果" span="2">
          <NCode :code="formatJSON(currentLog.res)" language="json" />
        </NDescriptionsItem>
        <NDescriptionsItem label="创建时间">
          {{ new Date(currentLog.createdAt).toLocaleString() }}
        </NDescriptionsItem>
        <NDescriptionsItem label="更新时间">
          {{ new Date(currentLog.updatedAt).toLocaleString() }}
        </NDescriptionsItem>
      </NDescriptions>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showDetailModal = false">关闭</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template> 