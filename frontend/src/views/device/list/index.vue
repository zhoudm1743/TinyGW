<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NButton, NSpace, NDataTable, NCard, NInput, NSelect, useMessage, NModal, NForm, NFormItem, NInputNumber } from 'naive-ui'
import { getDeviceList, deleteDevice, createDevice, updateDevice, DeviceInfo, DeviceRequest } from '@/service/api/device'
import { getCollectorList, CollectorInfo } from '@/service/api/collector'

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

// 搜索条件
const searchParams = reactive({
  typeID: '',
  collectorID: ''
})

// 表格数据
const tableData = ref<DeviceInfo[]>([])

// 采集器选项
const collectorOptions = ref<{ label: string; value: string }[]>([])

// 设备类型选项
const deviceTypeOptions = ref<{ label: string; value: string }[]>([
  { label: '电表', value: '2014F453-33' },
  { label: '水表', value: '2018F214-33' },
  { label: '气表', value: '2019F302-33' }
])

// 表单对话框
const showModal = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formData = reactive<DeviceRequest>({
  name: '',
  typeID: '',
  address: '',
  collectorID: '',
  initialVal: 0,
  scale: 1
})

// 表单规则
const rules = {
  name: {
    required: true,
    message: '请输入设备名称',
    trigger: 'blur'
  },
  typeID: {
    required: true,
    message: '请选择设备类型',
    trigger: 'change'
  },
  address: {
    required: true,
    message: '请输入设备地址',
    trigger: 'blur'
  },
  collectorID: {
    required: true,
    message: '请选择采集器',
    trigger: 'change'
  }
}

// 表格列定义
const columns = [
  {
    title: '设备名称',
    key: 'name'
  },
  {
    title: '设备类型',
    key: 'typeID'
  },
  {
    title: '设备地址',
    key: 'address'
  },
  {
    title: '采集器',
    key: 'collectorID'
  },
  {
    title: '在线状态',
    key: 'online',
    render(row: DeviceInfo) {
      return row.online ? '在线' : '离线'
    }
  },
  {
    title: '最近采集时间',
    key: 'collectTime',
    render(row: DeviceInfo) {
      return row.collectTime ? new Date(row.collectTime).toLocaleString() : '-'
    }
  },
  {
    title: '报警状态',
    key: 'alarmStatus',
    render(row: DeviceInfo) {
      return row.alarmStatus ? '报警中' : '正常'
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render(row: DeviceInfo) {
      return new Date(row.createdAt).toLocaleString()
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row: DeviceInfo) {
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
      ...searchParams
    }
    const { isSuccess, data } = await getDeviceList(params)
    if (isSuccess && data) {
      tableData.value = data.data
      pagination.itemCount = data.total
    }
  } catch (error) {
    console.error('加载设备列表失败', error)
    message.error('加载设备列表失败')
  } finally {
    loading.value = false
  }
}

// 加载采集器数据
async function loadCollectors() {
  try {
    const { isSuccess, data } = await getCollectorList({ page: 1, pageSize: 100 })
    if (isSuccess && data) {
      collectorOptions.value = data.data.map((item: CollectorInfo) => ({
        label: item.name,
        value: item.name
      }))
    }
  } catch (error) {
    console.error('加载采集器列表失败', error)
  }
}

// 搜索
function handleSearch() {
  pagination.page = 1
  loadData()
}

// 重置搜索
function handleReset() {
  searchParams.typeID = ''
  searchParams.collectorID = ''
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
function handleEdit(row: DeviceInfo) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    name: row.name,
    typeID: row.typeID,
    address: row.address,
    collectorID: row.collectorID,
    initialVal: row.initialVal,
    scale: row.scale
  })
  showModal.value = true
}

// 查看设备
function handleView(row: DeviceInfo) {
  // 实现查看详情功能
  console.log('查看设备', row)
}

// 删除设备
async function handleDelete(row: DeviceInfo) {
  try {
    const { isSuccess } = await deleteDevice(row.name)
    if (isSuccess) {
      message.success('删除成功')
      loadData()
    }
  } catch (error) {
    console.error('删除设备失败', error)
    message.error('删除设备失败')
  }
}

// 重置表单
function resetForm() {
  formData.name = ''
  formData.typeID = ''
  formData.address = ''
  formData.collectorID = ''
  formData.initialVal = 0
  formData.scale = 1
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
        // 编辑设备
        const { isSuccess } = await updateDevice(formData)
        if (isSuccess) {
          message.success('更新成功')
          showModal.value = false
          loadData()
        }
      } else {
        // 新增设备
        const { isSuccess } = await createDevice(formData)
        if (isSuccess) {
          message.success('创建成功')
          showModal.value = false
          loadData()
        }
      }
    } catch (error) {
      console.error('保存设备失败', error)
      message.error('保存设备失败')
    }
  })
}

onMounted(() => {
  loadData()
  loadCollectors()
})
</script>

<template>
  <div class="p-4">
    <NCard title="设备列表" class="mb-4">
      <template #header-extra>
        <NButton type="primary" @click="handleAdd">
          新增设备
        </NButton>
      </template>
      <NSpace vertical>
        <NSpace>
          <NInput v-model:value="searchParams.typeID" placeholder="设备类型" />
          <NInput v-model:value="searchParams.collectorID" placeholder="采集器" />
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
    <NModal v-model:show="showModal" :title="isEdit ? '编辑设备' : '新增设备'" preset="card" style="width: 600px">
      <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="80">
        <NFormItem label="设备名称" path="name">
          <NInput v-model:value="formData.name" placeholder="请输入设备名称" :disabled="isEdit" />
        </NFormItem>
        <NFormItem label="设备类型" path="typeID">
          <NSelect v-model:value="formData.typeID" :options="deviceTypeOptions" placeholder="请选择设备类型" />
        </NFormItem>
        <NFormItem label="设备地址" path="address">
          <NInput v-model:value="formData.address" placeholder="请输入设备地址" />
        </NFormItem>
        <NFormItem label="采集器" path="collectorID">
          <NSelect v-model:value="formData.collectorID" :options="collectorOptions" placeholder="请选择采集器" />
        </NFormItem>
        <NFormItem label="初始值" path="initialVal">
          <NInputNumber v-model:value="formData.initialVal" placeholder="请输入初始值" />
        </NFormItem>
        <NFormItem label="倍率" path="scale">
          <NInputNumber v-model:value="formData.scale" placeholder="请输入倍率" />
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