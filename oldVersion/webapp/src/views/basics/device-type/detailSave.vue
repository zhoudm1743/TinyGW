<script setup lang="ts">
import {onMounted, ref} from "vue";

// 声明 props 类型
export interface FormProps {
    disabledLock?: boolean;            // 是否禁用锁定按钮
    disabled?: boolean;                 // 是否禁用表单项
    formInline: {
        name: string;
        description: string;
        type: string;
        length: number;
        decimal: number;
        unit: string;
        value: number;
        reported: boolean;
        isAlarm: boolean;
        threshold: number;
        autoCalc: boolean;
    };
}
// 声明 props 默认值
const props = withDefaults(defineProps<FormProps>(), {
    //查看不可编辑
    disabled: false,
    disabledLock: false,
    //数据初始化，防止父组件传递的数据为空时，子组件报错
    formInline: () => ({
        name: '',           //规则名称
        description: '',    //描述
        type: 'double',     //数据类型
        length: 13,         //数据长度
        decimal: 2,         //小数位数
        unit: '',           //计量单位
        value: 0,           //最新数值
        reported: true,     //是否上报
        isAlarm: false,     //是否报警
        threshold: 0,       //预警值
        autoCalc: false,    //是否自动计算
    })
});
const newFormInline = ref(props.formInline);

/** 属性名称 **/
const selectOptions = [
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

/** 选择后赋值 */
const lock = ref(false);

const handleSelect = (name: string) => {
    if (name === 'dev_consumption' || name === 'dev_flow') {
        newFormInline.value.autoCalc = true
        lock.value = true
    }else {
        newFormInline.value.autoCalc = false
        lock.value = false
    }

    //根据name赋值dataForm.description
    selectOptions.forEach(item => {
        if (item.value === name) {
            newFormInline.value.description = item.label
            newFormInline.value.unit = item.unit
        }
    })
}
onMounted(() => {
    if (props.disabled) {
        lock.value = true;
    }else if (props.formInline.name === 'dev_consumption' || props.formInline.name === 'dev_flow') {
        lock.value = true;
    }
})


</script>

<template>
    <el-form :model="newFormInline">
        <el-form-item label="规则名称" prop="name">
            <el-select v-model="newFormInline.name" placeholder="请选择规则名称" :disabled="disabled" @change="handleSelect">
                <el-option
                        v-for="item in selectOptions"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value">
                </el-option>
            </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description" label-width="68px">
            <el-input v-model="newFormInline.description" placeholder="请输入描述" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item label="数据类型" prop="type">
            <el-input v-model="newFormInline.type" placeholder="请输入数据类型" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item label="数据长度" prop="length">
            <el-input v-model="newFormInline.length" placeholder="请输入数据长度" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item label="小数位数" prop="decimal">
            <el-input v-model="newFormInline.decimal" placeholder="请输入小数位数" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item label="计量单位" prop="unit">
            <el-input v-model="newFormInline.unit" placeholder="请输入计量单位" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item label="是否上报" prop="reported">
            <el-switch v-model="newFormInline.reported" active-color="#13ce66" inactive-color="#ff4949"
                      :disabled="disabled"></el-switch>
        </el-form-item>
        <el-form-item label="预警值" prop="threshold" label-width="68px">
            <el-input-number v-model="newFormInline.threshold" :min="0" :disabled="disabled" label="预警值"></el-input-number>
        </el-form-item>
        <el-form-item label="是否报警" prop="isAlarm">
            <el-switch v-model="newFormInline.isAlarm" active-color="#13ce66" inactive-color="#ff4949"
                      :disabled="disabled"></el-switch>
        </el-form-item>
        <el-form-item label="自动计算" prop="autoCalc">
            <el-switch v-model="newFormInline.autoCalc" active-color="#13ce66" inactive-color="#ff4949"
                      :disabled="lock"></el-switch>
        </el-form-item>
    </el-form>
</template>

<style scoped lang="scss">

</style>
