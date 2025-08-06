<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NButton, NSpace, NDataTable, NCard, NInput, useMessage, NModal, NForm, NFormItem, NTabs, NTabPane, NSelect, NInputNumber, NSwitch, NPopconfirm, NIcon, NDivider } from 'naive-ui'
import { getDeviceTypeList, createDeviceType, updateDeviceType, deleteDeviceType, DeviceTypeInfo, DeviceTypeRequest, DeviceProperty } from '@/service/api/deviceType'
import NovaIcon from '@/components/common/NovaIcon.vue'

const message = useMessage()

// 定义设备类型接口
interface DeviceType {
  id: string
  name: string
  protocol: string
  description: string
  createdAt: number
}

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
  name: ''
})

// 表格数据
const tableData = ref<DeviceTypeInfo[]>([])

// 表单对话框
const showModal = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formData = reactive<DeviceTypeRequest>({
  id: '',
  name: '',
  protocol: '',
  description: '',
  driver: '',
  properties: []
})

// 数据类型选项
const dataTypeOptions = [
  { label: '整数', value: 'int' },
  { label: '长整数', value: 'long' },
  { label: '浮点数', value: 'double' },
  { label: '字符串', value: 'string' }
]

// 常用属性选项
const commonPropertyOptions = [
  { label: '总电能', value: 'dev_consumption', unit: 'kwh' },
  { label: '剩余金额', value: 'dev_remain_amt', unit: '元' },
  { label: '剩余电量', value: 'dev_remain_elec', unit: 'kwh' },
  { label: '欠费金额', value: 'dev_owe_amt', unit: '元' },
  { label: '欠费电量', value: 'dev_owe_elec', unit: 'kwh' },
  { label: '瞬时流量', value: 'dev_instant_flow', unit: 'm³' },
  { label: '总流量', value: 'dev_flow', unit: 'm³' },
  { label: '温度值', value: 'dev_temp', unit: '℃' },
  { label: '湿度值', value: 'dev_rh', unit: '%' },
  { label: 'PM2.5', value: 'dev_pm25', unit: 'mg/m³' },
  { label: 'CO2', value: 'dev_co2', unit: 'ppm' },
  { label: 'CH2O', value: 'dev_ch2o', unit: 'mg/m³' },
  { label: 'VOC', value: 'dev_voc', unit: 'mg/m³' },
  { label: 'O2', value: 'dev_o2', unit: '%' },
  { label: 'CO', value: 'dev_co', unit: 'mg/m³' },
  { label: '上下行', value: 'dev_updown', unit: 'int类型' },
  { label: '运行状态', value: 'dev_status', unit: 'int类型' },
  { label: '检修状态', value: 'dev_fix', unit: 'int类型' },
  { label: '故障状态', value: 'dev_err', unit: 'int类型' },
  { label: '停泊状态', value: 'dev_stp', unit: 'int类型' }
]

// 表单规则
const rules = {
  id: {
    required: true,
    message: '请输入类型ID',
    trigger: 'blur'
  },
  name: {
    required: true,
    message: '请输入类型名称',
    trigger: 'blur'
  },
  protocol: {
    required: true,
    message: '请输入协议',
    trigger: 'blur'
  },
  driver: {
    required: true,
    message: '请输入驱动程序',
    trigger: 'blur'
  }
}

// 表格列定义
const columns = [
  {
    title: '类型ID',
    key: 'id'
  },
  {
    title: '类型名称',
    key: 'name'
  },
  {
    title: '协议',
    key: 'protocol'
  },
  {
    title: '驱动程序',
    key: 'driver'
  },
  {
    title: '属性数量',
    key: 'propertyCount',
    render(row: DeviceTypeInfo) {
      return row.properties ? row.properties.length : 0
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render(row: DeviceTypeInfo) {
      return new Date(row.createdAt).toLocaleString()
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row: DeviceTypeInfo) {
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

// 属性表格列定义
const propertyColumns = [
  {
    title: '名称',
    key: 'name',
    width: 120
  },
  {
    title: '描述',
    key: 'description',
    width: 120
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
    render(row: DeviceProperty) {
      const typeMap: Record<string, string> = {
        'int': '整数',
        'long': '长整数',
        'double': '浮点数',
        'string': '字符串'
      }
      return typeMap[row.type] || row.type
    }
  },
  {
    title: '单位',
    key: 'unit',
    width: 80
  },
  {
    title: '长度',
    key: 'length',
    width: 60
  },
  {
    title: '小数位',
    key: 'decimal',
    width: 60
  },
  {
    title: '倍率',
    key: 'scale',
    width: 60
  },
  {
    title: '上报',
    key: 'reported',
    width: 60,
    render(row: DeviceProperty) {
      return h(NSwitch, {
        value: row.reported,
        size: 'small',
        onUpdateValue: (value) => {
          row.reported = value
        }
      })
    }
  },
  {
    title: '报警',
    key: 'isAlarm',
    width: 60,
    render(row: DeviceProperty) {
      return h(NSwitch, {
        value: row.isAlarm,
        size: 'small',
        onUpdateValue: (value) => {
          row.isAlarm = value
        }
      })
    }
  },
  {
    title: '预警值',
    key: 'threshold',
    width: 80
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render(row: DeviceProperty, index: number) {
      return h(
        NPopconfirm,
        {
          onPositiveClick: () => {
            formData.properties.splice(index, 1)
          }
        },
        {
          default: () => '确定删除该属性吗？',
          trigger: () => h(
            NButton,
            {
              size: 'small',
              type: 'error'
            },
            { default: () => '删除' }
          )
        }
      )
    }
  }
]

// 加载数据
async function loadData() {
  loading.value = true
  try {
    // 这里需要实现设备类型的API调用
    // 暂时使用模拟数据
    setTimeout(() => {
      tableData.value = [
        {
          id: '2014F453-33',
          name: '电表类型1',
          protocol: 'Modbus',
          description: '电能表设备类型',
          driver: 'modbus/meter.lua',
          properties: [
            {
              name: 'dev_consumption',
              description: '总电能',
              type: 'double',
              length: 4,
              decimal: 2,
              unit: 'kwh',
              value: null,
              reported: true,
              isAlarm: false,
              threshold: 0,
              used: 0,
              autoCalc: false,
              scale: 1
            }
          ],
          createdAt: Date.now(),
          updatedAt: Date.now()
        },
        {
          id: '2018F214-33',
          name: '电表类型2',
          protocol: 'DLT645',
          description: '国家电网标准电表',
          driver: 'dlt645/meter.lua',
          properties: [],
          createdAt: Date.now(),
          updatedAt: Date.now()
        }
      ]
      pagination.itemCount = 2
      loading.value = false
    }, 500)
    
    // 实际API调用
    /*
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      name: searchParams.name || undefined
    }
    const { isSuccess, data } = await getDeviceTypeList(params)
    if (isSuccess && data) {
      tableData.value = data.data
      pagination.itemCount = data.total
    }
    */
  } catch (error) {
    console.error('加载设备类型列表失败', error)
    message.error('加载设备类型列表失败')
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
  searchParams.name = ''
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
function handleEdit(row: DeviceTypeInfo) {
  isEdit.value = true
  resetForm()
  Object.assign(formData, {
    id: row.id,
    name: row.name,
    protocol: row.protocol,
    description: row.description,
    driver: row.driver,
    properties: JSON.parse(JSON.stringify(row.properties || []))
  })
  showModal.value = true
}

// 查看设备类型
function handleView(row: DeviceTypeInfo) {
  console.log('查看设备类型', row)
}

// 删除设备类型
async function handleDelete(row: DeviceTypeInfo) {
  try {
    // 实际API调用
    /*
    const { isSuccess } = await deleteDeviceType(row.id)
    if (isSuccess) {
      message.success('删除成功')
      loadData()
    }
    */
    message.success('删除成功')
    loadData()
  } catch (error) {
    console.error('删除设备类型失败', error)
    message.error('删除设备类型失败')
  }
}

// 重置表单
function resetForm() {
  formData.id = ''
  formData.name = ''
  formData.protocol = ''
  formData.description = ''
  formData.driver = ''
  formData.properties = []
}

// 添加属性
function addProperty() {
  formData.properties.push({
    name: '',
    description: '',
    type: 'double',
    length: 4,
    decimal: 2,
    unit: '',
    value: null,
    reported: true,
    isAlarm: false,
    threshold: 0,
    used: 0,
    autoCalc: false,
    scale: 1
  })
}

// 添加常用属性
function addCommonProperty(option: { label: string; value: string; unit: string }) {
  formData.properties.push({
    name: option.value,
    description: option.label,
    type: option.unit.includes('int') ? 'int' : 'double',
    length: 4,
    decimal: 2,
    unit: option.unit.replace('int类型', ''),
    value: null,
    reported: true,
    isAlarm: false,
    threshold: 0,
    used: 0,
    autoCalc: false,
    scale: 1
  })
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
        // 编辑设备类型
        // 实际API调用
        /*
        const { isSuccess } = await updateDeviceType(formData)
        if (isSuccess) {
          message.success('更新成功')
          showModal.value = false
          loadData()
        }
        */
        message.success('更新成功')
        showModal.value = false
        loadData()
      } else {
        // 新增设备类型
        // 实际API调用
        /*
        const { isSuccess } = await createDeviceType(formData)
        if (isSuccess) {
          message.success('创建成功')
          showModal.value = false
          loadData()
        }
        */
        message.success('创建成功')
        showModal.value = false
        loadData()
      }
    } catch (error) {
      console.error('保存设备类型失败', error)
      message.error('保存设备类型失败')
    }
  })
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-4">
    <NCard title="设备类型管理" class="mb-4">
      <template #header-extra>
        <NButton type="primary" @click="handleAdd">
          新增设备类型
        </NButton>
      </template>
      <NSpace vertical>
        <NSpace>
          <NInput v-model:value="searchParams.name" placeholder="类型名称" />
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
    <NModal v-model:show="showModal" :title="isEdit ? '编辑设备类型' : '新增设备类型'" preset="card" style="width: 900px">
      <NTabs type="line" animated>
        <NTabPane name="basic" tab="基本信息">
          <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="80">
            <NFormItem label="类型ID" path="id">
              <NInput v-model:value="formData.id" placeholder="请输入类型ID" :disabled="isEdit" />
            </NFormItem>
            <NFormItem label="类型名称" path="name">
              <NInput v-model:value="formData.name" placeholder="请输入类型名称" />
            </NFormItem>
            <NFormItem label="协议" path="protocol">
              <NInput v-model:value="formData.protocol" placeholder="请输入协议" />
            </NFormItem>
            <NFormItem label="驱动程序" path="driver">
              <NInput v-model:value="formData.driver" placeholder="请输入驱动程序路径" />
            </NFormItem>
            <NFormItem label="描述" path="description">
              <NInput v-model:value="formData.description" placeholder="请输入描述" type="textarea" />
            </NFormItem>
          </NForm>
        </NTabPane>
        <NTabPane name="properties" tab="属性配置">
          <div class="mb-4">
            <NSpace>
              <NButton type="primary" @click="addProperty">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:add" />
                </template>
                添加属性
              </NButton>
              <NSelect
                placeholder="添加常用属性"
                :options="commonPropertyOptions"
                style="width: 200px"
                @update:value="(value, option) => addCommonProperty(option)"
              />
            </NSpace>
          </div>
          <NDataTable
            :columns="propertyColumns"
            :data="formData.properties"
            :bordered="true"
            :single-line="false"
          />
          <div v-if="formData.properties.length === 0" class="py-4 text-center text-gray-400">
            暂无属性配置，请添加属性
          </div>
          <div v-for="(property, index) in formData.properties" :key="index" class="mb-4 p-4 border rounded">
            <NSpace vertical>
              <NSpace>
                <NFormItem label="名称" :path="`properties[${index}].name`">
                  <NInput v-model:value="property.name" placeholder="请输入属性名称" />
                </NFormItem>
                <NFormItem label="描述" :path="`properties[${index}].description`">
                  <NInput v-model:value="property.description" placeholder="请输入属性描述" />
                </NFormItem>
              </NSpace>
              <NSpace>
                <NFormItem label="类型" :path="`properties[${index}].type`">
                  <NSelect v-model:value="property.type" :options="dataTypeOptions" placeholder="请选择数据类型" />
                </NFormItem>
                <NFormItem label="单位" :path="`properties[${index}].unit`">
                  <NInput v-model:value="property.unit" placeholder="请输入单位" />
                </NFormItem>
                <NFormItem label="长度" :path="`properties[${index}].length`">
                  <NInputNumber v-model:value="property.length" placeholder="请输入长度" />
                </NFormItem>
                <NFormItem label="小数位" :path="`properties[${index}].decimal`">
                  <NInputNumber v-model:value="property.decimal" placeholder="请输入小数位" />
                </NFormItem>
              </NSpace>
              <NSpace>
                <NFormItem label="倍率" :path="`properties[${index}].scale`">
                  <NInputNumber v-model:value="property.scale" placeholder="请输入倍率" />
                </NFormItem>
                <NFormItem label="是否上报" :path="`properties[${index}].reported`">
                  <NSwitch v-model:value="property.reported" />
                </NFormItem>
                <NFormItem label="是否报警" :path="`properties[${index}].isAlarm`">
                  <NSwitch v-model:value="property.isAlarm" />
                </NFormItem>
                <NFormItem label="预警值" :path="`properties[${index}].threshold`">
                  <NInputNumber v-model:value="property.threshold" placeholder="请输入预警值" />
                </NFormItem>
              </NSpace>
              <div class="text-right">
                <NPopconfirm @positive-click="() => formData.properties.splice(index, 1)">
                  <template #trigger>
                    <NButton type="error" size="small">
                      <template #icon>
                        <NovaIcon icon="icon-park-outline:delete" />
                      </template>
                      删除此属性
                    </NButton>
                  </template>
                  确定删除该属性吗？
                </NPopconfirm>
              </div>
            </NSpace>
            <NDivider v-if="index < formData.properties.length - 1" />
          </div>
        </NTabPane>
      </NTabs>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">确定</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template> 