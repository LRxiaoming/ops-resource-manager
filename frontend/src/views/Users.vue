<template>
  <div class="users">
    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">新增用户</el-button>
      </div>

      <el-table :data="users" border style="margin-top: 20px">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="role" label="角色" width="150">
          <template #default="{ row }">
            <el-tag>{{ getRoleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="200" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="密码" v-if="!isEdit">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item label="新密码" v-if="isEdit">
          <el-input v-model="form.password" type="password" placeholder="留空则不修改" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role">
            <el-option label="系统管理员" value="admin" />
            <el-option label="申请人" value="applicant" />
            <el-option label="组长" value="leader" />
            <el-option label="经理" value="manager" />
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const users = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const isEdit = ref(false)
const currentId = ref(null)

const form = ref({
  username: '',
  password: '',
  role: 'applicant'
})

const fetchUsers = async () => {
  const { data } = await api.get('/users')
  users.value = data
}

const handleAdd = () => {
  form.value = { username: '', password: '', role: 'applicant' }
  dialogTitle.value = '新增用户'
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  form.value = { username: row.username, password: '', role: row.role }
  dialogTitle.value = '编辑用户'
  isEdit.value = true
  currentId.value = row.id
  dialogVisible.value = true
}

const handleSave = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/users/${currentId.value}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/users', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该用户？', '提示', { type: 'warning' })
    await api.delete(`/users/${row.id}`)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

const getRoleText = (role) => ({ admin: '系统管理员', applicant: '申请人', leader: '组长', manager: '经理' }[role] || role)

onMounted(fetchUsers)
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
}
</style>