<template>
  <div class="assets">
    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">新增资产</el-button>
        <el-select v-model="filter.type" placeholder="资产类型" clearable style="width: 150px; margin-left: 10px">
          <el-option label="服务器" value="server" />
          <el-option label="网络设备" value="network" />
        </el-select>
        <el-select v-model="filter.status" placeholder="状态" clearable style="width: 150px; margin-left: 10px">
          <el-option label="在线" value="online" />
          <el-option label="离线" value="offline" />
          <el-option label="维护中" value="maintenance" />
        </el-select>
      </div>

      <el-table :data="assets" border style="margin-top: 20px">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            {{ row.type === 'server' ? '服务器' : '网络设备' }}
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="150" />
        <el-table-column prop="hostname" label="主机名" width="150" />
        <el-table-column prop="cpu" label="CPU" width="120" />
        <el-table-column prop="memory" label="内存" width="100" />
        <el-table-column prop="disk" label="磁盘" width="100" />
        <el-table-column prop="location" label="机房位置" width="150" />
        <el-table-column prop="responsible" label="责任人" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="资产类型">
          <el-select v-model="form.type">
            <el-option label="服务器" value="server" />
            <el-option label="网络设备" value="network" />
          </el-select>
        </el-form-item>
        <el-form-item label="IP地址">
          <el-input v-model="form.ip" />
        </el-form-item>
        <el-form-item label="主机名">
          <el-input v-model="form.hostname" />
        </el-form-item>
        <el-form-item label="CPU">
          <el-input v-model="form.cpu" />
        </el-form-item>
        <el-form-item label="内存">
          <el-input v-model="form.memory" />
        </el-form-item>
        <el-form-item label="磁盘">
          <el-input v-model="form.disk" />
        </el-form-item>
        <el-form-item label="机房位置">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="责任人">
          <el-input v-model="form.responsible" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护中" value="maintenance" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const assets = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增资产')
const isEdit = ref(false)
const currentId = ref(null)

const filter = ref({ type: '', status: '' })
const form = ref({
  type: 'server',
  ip: '',
  hostname: '',
  cpu: '',
  memory: '',
  disk: '',
  location: '',
  responsible: '',
  status: 'online'
})

const fetchAssets = async () => {
  const params = {}
  if (filter.value.type) params.type = filter.value.type
  if (filter.value.status) params.status = filter.value.status

  const { data } = await api.get('/assets', { params })
  assets.value = data
}

watch(filter, fetchAssets)

const handleAdd = () => {
  form.value = { type: 'server', ip: '', hostname: '', cpu: '', memory: '', disk: '', location: '', responsible: '', status: 'online' }
  dialogTitle.value = '新增资产'
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  form.value = { ...row }
  dialogTitle.value = '编辑资产'
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

const handleSave = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/assets/${currentId.value}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/assets', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchAssets()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该资产？', '提示', { type: 'warning' })
    await api.delete(`/assets/${row.id}`)
    ElMessage.success('删除成功')
    fetchAssets()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

const getStatusType = (status) => {
  const map = { online: 'success', offline: 'danger', maintenance: 'warning' }
  return map[status] || ''
}

const getStatusText = (status) => {
  const map = { online: '在线', offline: '离线', maintenance: '维护中' }
  return map[status] || status
}

onMounted(fetchAssets)
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
}
</style>