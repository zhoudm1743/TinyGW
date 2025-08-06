<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NButton, NSpace, NDataTable, NCard, NInput, NSelect, NSwitch, useMessage, NModal, NForm, NFormItem, NInputNumber, NTabs, NTabPane } from 'naive-ui'
import { getCollectorList, deleteCollector, createCollector, updateCollector, CollectorInfo, CollectorRequest, Serial, TcpClient, TcpServer } from '@/service/api/collector'

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

// 采集器类型选项
const collectorTypeOptions = [
  { label: '全部', value: '' },
  { label: '串口', value: 'Serial' },
  { label: 'TCP客户端', value: 'TcpClient' },
  { label: 'TCP服务端', value: 'TcpServer' },
  { label: 'MQTT', value: 'Mqtt' },
  { label: '通道', value: 'Channel' },
  { label: '4G GPRS', value: 'FourGPRS' }
]

// 采集器类型选项（不含全部）
const collectorTypeFormOptions = [
  { label: '串口', value: 'Serial' },
  { label: 'TCP客户端', value: 'TcpClient' },
  { label: 'TCP服务端', value: 'TcpServer' },
  { label: 'MQTT', value: 'Mqtt' },
  { label: '通道', value: 'Channel' },
  { label: '4G GPRS', value: 'FourGPRS' }
]

// 搜索条件
const searchParams = reactive({
  type: '',
  enable: null as boolean | null
})

// 表格数据
const tableData = ref<CollectorInfo[]>([])

// 表单对话框
const showModal = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formData = reactive<CollectorRequest>({
  name: '',
  type: 'Serial',
  address: '',
  serial: {
    name: '',
    deviceName: '',
    baudRate: 9600,
    dataBit: 8,
    stopBit: '1',
    check: 'none'
  },
  tcpClient: {
    name: '',
    ip: '',
    port: 502
  },
  tcpServer: {
    name: '',
    port: 502
  },
  mqtt: {
    name: ''
  },
  channel: {
    name: ''
  },
  fourGPRS: {
    name: ''
  },
  timeout: 5000,
  interval: 5000,
  enable: true
})

// 表单规则
const rules = {
  name: {
    required: true,
    message: '请输入采集器名称',
    trigger: 'blur'
  },
  type: {
    required: true,
    message: '请选择采集器类型',
    trigger: 'change'
  },
  address: {
    required: true,
    message: '请输入采集器地址',
    trigger: 'blur'
  },
  timeout: {
    required: true,
    message: '请输入超时时间',
    trigger: 'blur'
  },
  interval: {
    required: true,
    message: '请输入间隔时间',
    trigger: 'blur'
  }
}

// 串口波特率选项
const baudRateOptions = [
  { label: '1200', value: 1200 },
  { label: '2400', value: 2400 },
  { label: '4800', value: 4800 },
  { label: '9600', value: 9600 },
  { label: '19200', value: 19200 },
  { label: '38400', value: 38400 },
  { label: '57600', value: 57600 },
  { label: '115200', value: 115200 }
]

// 数据位选项
const dataBitOptions = [
  { label: '5', value: 5 },
  { label: '6', value: 6 },
  { label: '7', value: 7 },
  { label: '8', value: 8 }
]

// 停止位选项
const stopBitOptions = [
  { label: '1', value: '1' },
  { label: '1.5', value: '1.5' },
  { label: '2', value: '2' }
]

// 校验位选项
const checkOptions = [
  { label: '无校验', value: 'none' },
  { label: '奇校验', value: 'odd' },
  { label: '偶校验', value: 'even' }
]

// 表格列定义
const columns = [
  {
    title: '采集器名称',
    key: 'name'
  },
  {
    title: '类型',
    key: 'type',
    render(row: CollectorInfo) {
      const typeMap: Record<string, string> = {
        'Serial': '串口',
        'TcpClient': 'TCP客户端',
        'TcpServer': 'TCP服务端',
        'Mqtt': 'MQTT',
        'Channel': '通道',
        'FourGPRS': '4G GPRS'
      }
      return typeMap[row.type] || row.type
    }
  },
  {
    title: '地址',
    key: 'address'
  },
  {
    title: '超时(秒)',
    key: 'timeout'
  },
  {
    title: '间隔(秒)',
    key: 'interval'
  },
  {
    title: '状态',
    key: 'enable',
    render(row: CollectorInfo) {
      return h(NSwitch, {
        value: row.enable,
        disabled: true
      })
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render(row: CollectorInfo) {
      return new Date(row.createdAt).toLocaleString()
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row: CollectorInfo) {
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
          h(
            NButton,
            {
              size: 'small',
              type: 'info',
              onClick: () => handleView(row)
            },
            { default: () => '查看' }
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
      type: searchParams.type,
      enable: searchParams.enable === null ? undefined : searchParams.enable
    }
    const { isSuccess, data } = await getCollectorList(params)
    if (isSuccess && data) {
      tableData.value = data.data
      pagination.itemCount = data.total
    }
  } catch (error) {
    console.error('加载采集器列表失败', error)
    message.error('加载采集器列表失败')
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
  searchParams.type = ''
  searchParams.enable = null
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
function handleEdit(row: CollectorInfo) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    name: row.name,
    type: row.type,
    address: row.address,
    serial: row.serial,
    tcpClient: row.tcpClient,
    tcpServer: row.tcpServer,
    mqtt: row.mqtt,
    channel: row.channel,
    fourGPRS: row.fourGPRS,
    timeout: row.timeout,
    interval: row.interval,
    enable: row.enable
  })
  showModal.value = true
}

// 查看采集器
function handleView(row: CollectorInfo) {
  console.log('查看采集器', row)
}

// 删除采集器
async function handleDelete(row: CollectorInfo) {
  try {
    const { isSuccess } = await deleteCollector(row.name)
    if (isSuccess) {
      message.success('删除成功')
      loadData()
    }
  } catch (error) {
    console.error('删除采集器失败', error)
    message.error('删除采集器失败')
  }
}

// 重置表单
function resetForm() {
  formData.name = ''
  formData.type = 'Serial'
  formData.address = ''
  formData.serial = {
    name: '',
    deviceName: '',
    baudRate: 9600,
    dataBit: 8,
    stopBit: '1',
    check: 'none'
  }
  formData.tcpClient = {
    name: '',
    ip: '',
    port: 502
  }
  formData.tcpServer = {
    name: '',
    port: 502
  }
  formData.mqtt = {
    name: ''
  }
  formData.channel = {
    name: ''
  }
  formData.fourGPRS = {
    name: ''
  }
  formData.timeout = 5000
  formData.interval = 5000
  formData.enable = true
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
        // 编辑采集器
        const { isSuccess } = await updateCollector(formData)
        if (isSuccess) {
          message.success('更新成功')
          showModal.value = false
          loadData()
        }
      } else {
        // 新增采集器
        const { isSuccess } = await createCollector(formData)
        if (isSuccess) {
          message.success('创建成功')
          showModal.value = false
          loadData()
        }
      }
    } catch (error) {
      console.error('保存采集器失败', error)
      message.error('保存采集器失败')
    }
  })
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-4">
    <NCard title="采集器管理" class="mb-4">
      <template #header-extra>
        <NButton type="primary" @click="handleAdd">
          新增采集器
        </NButton>
      </template>
      <NSpace vertical>
        <NSpace>
          <NSelect v-model:value="searchParams.type" :options="collectorTypeOptions" placeholder="采集器类型" style="width: 200px" />
          <NSelect 
            v-model:value="searchParams.enable" 
            :options="[
              { label: '全部', value: null },
              { label: '启用', value: true },
              { label: '禁用', value: false }
            ]" 
            placeholder="状态" 
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
    <NModal v-model:show="showModal" :title="isEdit ? '编辑采集器' : '新增采集器'" preset="card" style="width: 700px">
      <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
        <NFormItem label="采集器名称" path="name">
          <NInput v-model:value="formData.name" placeholder="请输入采集器名称" :disabled="isEdit" />
        </NFormItem>
        <NFormItem label="采集器类型" path="type">
          <NSelect v-model:value="formData.type" :options="collectorTypeFormOptions" placeholder="请选择采集器类型" />
        </NFormItem>
        <NFormItem label="采集器地址" path="address">
          <NInput v-model:value="formData.address" placeholder="请输入采集器地址" />
        </NFormItem>
        
        <!-- 根据采集器类型显示不同的表单 -->
        <NTabs v-model:value="formData.type" type="line" animated>
          <!-- 串口配置 -->
          <NTabPane name="Serial" tab="串口配置">
            <NFormItem label="串口名称" path="serial.name">
              <NInput v-model:value="formData.serial.name" placeholder="请输入串口名称" />
            </NFormItem>
            <NFormItem label="设备名称" path="serial.deviceName">
              <NInput v-model:value="formData.serial.deviceName" placeholder="请输入设备名称" />
            </NFormItem>
            <NFormItem label="波特率" path="serial.baudRate">
              <NSelect v-model:value="formData.serial.baudRate" :options="baudRateOptions" placeholder="请选择波特率" />
            </NFormItem>
            <NFormItem label="数据位" path="serial.dataBit">
              <NSelect v-model:value="formData.serial.dataBit" :options="dataBitOptions" placeholder="请选择数据位" />
            </NFormItem>
            <NFormItem label="停止位" path="serial.stopBit">
              <NSelect v-model:value="formData.serial.stopBit" :options="stopBitOptions" placeholder="请选择停止位" />
            </NFormItem>
            <NFormItem label="校验位" path="serial.check">
              <NSelect v-model:value="formData.serial.check" :options="checkOptions" placeholder="请选择校验位" />
            </NFormItem>
          </NTabPane>
          
          <!-- TCP客户端配置 -->
          <NTabPane name="TcpClient" tab="TCP客户端">
            <NFormItem label="名称" path="tcpClient.name">
              <NInput v-model:value="formData.tcpClient.name" placeholder="请输入名称" />
            </NFormItem>
            <NFormItem label="IP地址" path="tcpClient.ip">
              <NInput v-model:value="formData.tcpClient.ip" placeholder="请输入IP地址" />
            </NFormItem>
            <NFormItem label="端口" path="tcpClient.port">
              <NInputNumber v-model:value="formData.tcpClient.port" placeholder="请输入端口" />
            </NFormItem>
          </NTabPane>
          
          <!-- TCP服务端配置 -->
          <NTabPane name="TcpServer" tab="TCP服务端">
            <NFormItem label="名称" path="tcpServer.name">
              <NInput v-model:value="formData.tcpServer.name" placeholder="请输入名称" />
            </NFormItem>
            <NFormItem label="端口" path="tcpServer.port">
              <NInputNumber v-model:value="formData.tcpServer.port" placeholder="请输入端口" />
            </NFormItem>
          </NTabPane>
          
          <!-- MQTT配置 -->
          <NTabPane name="Mqtt" tab="MQTT">
            <NFormItem label="名称" path="mqtt.name">
              <NInput v-model:value="formData.mqtt.name" placeholder="请输入名称" />
            </NFormItem>
          </NTabPane>
          
          <!-- 通道配置 -->
          <NTabPane name="Channel" tab="通道">
            <NFormItem label="名称" path="channel.name">
              <NInput v-model:value="formData.channel.name" placeholder="请输入名称" />
            </NFormItem>
          </NTabPane>
          
          <!-- 4G GPRS配置 -->
          <NTabPane name="FourGPRS" tab="4G GPRS">
            <NFormItem label="名称" path="fourGPRS.name">
              <NInput v-model:value="formData.fourGPRS.name" placeholder="请输入名称" />
            </NFormItem>
          </NTabPane>
        </NTabs>
        
        <NFormItem label="超时(毫秒)" path="timeout">
          <NInputNumber v-model:value="formData.timeout" placeholder="请输入超时时间" />
        </NFormItem>
        <NFormItem label="间隔(毫秒)" path="interval">
          <NInputNumber v-model:value="formData.interval" placeholder="请输入间隔时间" />
        </NFormItem>
        <NFormItem label="是否启用" path="enable">
          <NSwitch v-model:value="formData.enable" />
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