<template>
  <div class="tickets">
    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">新增工单</el-button>
        <el-select v-model="filter.type" placeholder="工单类型" clearable style="width: 150px; margin-left: 10px">
          <el-option label="故障报修" value="fault" />
          <el-option label="变更申请" value="change" />
          <el-option label="资源申请" value="resource" />
        </el-select>
        <el-select v-model="filter.status" placeholder="状态" clearable style="width: 150px; margin-left: 10px">
          <el-option label="待处理" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
          <el-option label="已关闭" value="closed" />
        </el-select>
        <el-select v-model="filter.priority" placeholder="优先级" clearable style="width: 150px; margin-left: 10px">
          <el-option label="紧急" value="urgent" />
          <el-option label="重要" value="important" />
          <el-option label="普通" value="normal" />
        </el-select>
      </div>

      <el-table :data="tickets" border style="margin-top: 20px" @row-click="handleRowClick">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            {{ getTypeText(row.type) }}
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)">{{ getPriorityText(row.priority) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="applicant" label="申请人" width="100">
          <template #default="{ row }">{{ row.applicant?.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="handler" label="处理人" width="100">
          <template #default="{ row }">{{ row.handler?.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="700px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="工单类型">
          <el-select v-model="form.type">
            <el-option label="故障报修" value="fault" />
            <el-option label="变更申请" value="change" />
            <el-option label="资源申请" value="resource" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="form.priority">
            <el-option label="紧急" value="urgent" />
            <el-option label="重要" value="important" />
            <el-option label="普通" value="normal" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联资产">
          <el-select v-model="form.asset_ids" multiple placeholder="选择资产">
            <el-option v-for="asset in assets" :key="asset.id" :label="`${asset.hostname} (${asset.ip})`" :value="asset.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="工单详情" width="800px">
      <el-descriptions :column="2" border v-if="currentTicket">
        <el-descriptions-item label="ID">{{ currentTicket.id }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ getTypeText(currentTicket.type) }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ currentTicket.title }}</el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="getPriorityType(currentTicket.priority)">{{ getPriorityText(currentTicket.priority) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTicket.status)">{{ getStatusText(currentTicket.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请人">{{ currentTicket.applicant?.username }}</el-descriptions-item>
        <el-descriptions-item label="处理人">{{ currentTicket.handler?.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentTicket.created_at }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ currentTicket.description }}</el-descriptions-item>
        <el-descriptions-item label="关联资产" :span="2">
          <el-tag v-for="asset in currentTicket.assets" :key="asset.id" style="margin-right: 5px">{{ asset.hostname }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审批记录" :span="2">
          <el-table v-if="currentTicket.approvals?.length" :data="currentTicket.approvals" size="small">
            <el-table-column prop="approver.username" label="审批人" />
            <el-table-column prop="level" label="层级" />
            <el-table-column prop="result" label="结果">
              <template #default="{ row }">{{ row.result === 'approved' ? '通过' : '拒绝' }}</template>
            </el-table-column>
            <el-table-column prop="comment" label="意见" />
            <el-table-column prop="created_at" label="时间" />
          </el-table>
          <span v-else>暂无审批记录</span>
        </el-descriptions-item>
      </el-descriptions>

      <template #footer v-if="canApprove">
        <el-input v-model="approveComment" placeholder="审批意见" style="margin-bottom: 10px" />
        <el-button type="success" @click="handleApprove('approved')">通过</el-button>
        <el-button type="danger" @click="handleApprove('rejected')">拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const tickets = ref([])
const assets = ref([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const dialogTitle = ref('新增工单')
const isEdit = ref(false)
const currentId = ref(null)
const currentTicket = ref(null)
const approveComment = ref('')

const filter = ref({ type: '', status: '', priority: '' })
const form = ref({
  type: 'fault',
  title: '',
  description: '',
  priority: 'normal',
  asset_ids: []
})

const canApprove = computed(() => {
  if (!currentTicket.value) return false
  const role = authStore.user?.role
  return (role === 'leader' || role === 'manager') && currentTicket.value.status === 'pending'
})

const fetchTickets = async () => {
  const params = {}
  if (filter.value.type) params.type = filter.value.type
  if (filter.value.status) params.status = filter.value.status
  if (filter.value.priority) params.priority = filter.value.priority

  const { data } = await api.get('/tickets', { params })
  tickets.value = data
}

const fetchAssets = async () => {
  const { data } = await api.get('/assets')
  assets.value = data
}

watch(filter, fetchTickets)

const handleAdd = () => {
  form.value = { type: 'fault', title: '', description: '', priority: 'normal', asset_ids: [] }
  dialogTitle.value = '新增工单'
  isEdit.value = false
  dialogVisible.value = true
}

const handleRowClick = async (row) => {
  const { data } = await api.get(`/tickets/${row.id}`)
  currentTicket.value = data
  detailVisible.value = true
}

const handleSave = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/tickets/${currentId.value}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/tickets', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchTickets()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  }
}

const handleApprove = async (result) => {
  try {
    await api.post(`/tickets/${currentTicket.value.id}/approve`, { result, comment: approveComment.value })
    ElMessage.success('审批成功')
    detailVisible.value = false
    approveComment.value = ''
    fetchTickets()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '审批失败')
  }
}

const getTypeText = (type) => ({ fault: '故障报修', change: '变更申请', resource: '资源申请' }[type] || type)
const getPriorityType = (p) => ({ urgent: 'danger', important: 'warning', normal: 'info' }[p] || '')
const getPriorityText = (p) => ({ urgent: '紧急', important: '重要', normal: '普通' }[p] || p)
const getStatusType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger', closed: 'info' }[s] || '')
const getStatusText = (s) => ({ pending: '待处理', approved: '已通过', rejected: '已拒绝', closed: '已关闭' }[s] || s)

onMounted(() => {
  fetchTickets()
  fetchAssets()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
}
</style>