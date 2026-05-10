<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.total_assets }}</div>
          <div class="stat-label">资产总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.total_tickets }}</div>
          <div class="stat-label">工单总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.pending_tickets }}</div>
          <div class="stat-label">待处理工单</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.online_assets }}</div>
          <div class="stat-label">在线资产</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <div ref="assetChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <div ref="ticketChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card>
          <div ref="trendChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import api from '../api'

const stats = ref({
  total_assets: 0,
  total_tickets: 0,
  pending_tickets: 0,
  online_assets: 0
})

const assetChartRef = ref()
const ticketChartRef = ref()
const trendChartRef = ref()
let assetChart, ticketChart, trendChart

const fetchStats = async () => {
  try {
    const { data } = await api.get('/dashboard/stats')
    stats.value = {
      total_assets: data.total_assets || 0,
      total_tickets: data.total_tickets || 0,
      pending_tickets: (data.ticket_by_status && data.ticket_by_status.pending) || 0,
      online_assets: (data.asset_by_status && data.asset_by_status.online) || 0
    }
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

const fetchCharts = async () => {
  try {
    const { data } = await api.get('/dashboard/stats')
    const chartsRes = await api.get('/dashboard/charts')

    if (assetChart) {
      assetChart.setOption({
        title: { text: '资产类型分布' },
        tooltip: {},
        series: [{
          type: 'pie',
          radius: '60%',
          data: Object.entries(data.asset_by_type || {}).map(([name, value]) => ({ name, value }))
        }]
      })
    }

    if (ticketChart) {
      ticketChart.setOption({
        title: { text: '工单状态分布' },
        tooltip: {},
        series: [{
          type: 'pie',
          radius: '60%',
          data: Object.entries(data.ticket_by_status || {}).map(([name, value]) => ({ name, value }))
        }]
      })
    }

    if (trendChart) {
      const trends = chartsRes.data.trends || []
      trendChart.setOption({
        title: { text: '工单趋势（最近30天）' },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: trends.map(t => t.date) },
        yAxis: { type: 'value' },
        series: [{ data: trends.map(t => t.count), type: 'line', smooth: true }]
      })
    }
  } catch (error) {
    console.error('Failed to fetch charts:', error)
  }
}

onMounted(async () => {
  await fetchStats()

  assetChart = echarts.init(assetChartRef.value)
  ticketChart = echarts.init(ticketChartRef.value)
  trendChart = echarts.init(trendChartRef.value)

  await fetchCharts()

  window.addEventListener('resize', () => {
    assetChart?.resize()
    ticketChart?.resize()
    trendChart?.resize()
  })
})

onUnmounted(() => {
  assetChart?.dispose()
  ticketChart?.dispose()
  trendChart?.dispose()
})
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 20px;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
}

.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 10px;
}
</style>