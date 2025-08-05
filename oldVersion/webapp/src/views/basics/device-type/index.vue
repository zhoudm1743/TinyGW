<script setup lang="ts">
import deviceType from './deviceType.vue'
import detail from './detail.vue'
</script>

<template>
    <el-row :gutter="0">
        <el-col :span="8">
            <deviceType />
        </el-col>
        <el-col :span="16">
            <detail />
        </el-col>
    </el-row>
</template>

<style scoped lang="scss">

</style>
