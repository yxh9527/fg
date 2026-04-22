<template>
  <el-dialog
    ref="dialog"
    :visible.sync="detail_visible"
    v-loading="loading"
    custom-class="content-box"
    append-to-body
    :show-close="false"
    :before-close="close"
    :width="default_width"
    :fullscreen="is_douyu"
  >
    <el-row type="flex" justify="center">
      <el-col :span="22" type="flex" align="middle" style="margin-top: 30px;" v-if="!loading">
        <el-button
          v-if="play_video_url"
          type="primary"
          size="small"
          class="playback"
          @click.native.prevent="playBack"
        >{{ $t("message.video_playback") }}</el-button>
        <el-col :span="24" class="bg-purple-dark">
          <p style="font-weight: bold;">{{ $t("message.game_details") }}</p>
        </el-col>
        <el-table :data="tableData" :show-header="false" :border="true" stripe>
          <el-table-column prop="key" align="center"></el-table-column>
          <el-table-column prop="value" align="center"></el-table-column>
        </el-table>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center" v-if="!loading && detailRows.length">
      <el-col :span="22" style="margin-top: 20px;">
        <el-col :span="24" class="bg-purple-dark">
          <p style="font-weight: bold;">{{ $t("message.game_results") }}</p>
        </el-col>
        <el-table :data="detailRows" :border="true" stripe>
          <el-table-column prop="key" align="center" width="220"></el-table-column>
          <el-table-column prop="value" align="left"></el-table-column>
        </el-table>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center" v-if="!loading && detailJson">
      <el-col :span="22" style="margin-top: 20px;">
        <el-col :span="24" class="bg-purple-dark">
          <p style="font-weight: bold;">JSON</p>
        </el-col>
        <pre class="detail-json">{{ detailJson }}</pre>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center">
      <el-col :span="22" type="flex" align="middle" style="margin-top: 30px; margin-bottom: 30px;" v-if="!is_douyu">
        <el-button @click.native.prevent="close" style="font-weight: bold;">{{ $t("message.close") }}</el-button>
      </el-col>
    </el-row>
  </el-dialog>
</template>

<style>
.el-dialog__header {
  padding: 0;
}

.el-dialog__body {
  padding: 0;
}

.bg-purple-dark {
  background: #e2e6eb;
}

.content-box {
  width: 100%;
  margin: auto;
  background: rgb(240, 240, 240);
  position: relative;
}

.playback {
  position: relative;
  right: 10px;
  top: 40px;
  z-index: 1;
  float: right;
}

.detail-json {
  margin: 0;
  padding: 16px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-top: 0;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.6;
}
</style>

<script>
import http from '../../assets/js/http'

export default {
  data() {
    return {
      timmer: 0,
      loading: true,
      is_douyu: false,
      default_width: '80%',
      douyu_url: '',
      detail_visible: false,
      recordData: {},
      tableData: [],
      detailRows: [],
      detailJson: '',
      play_video_url: ''
    }
  },
  methods: {
    timeLabel() {
      const utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    playBack() {
      if (this.play_video_url) {
        window.open(this.play_video_url)
      }
    },
    toDisplayValue(value) {
      if (value === null || value === undefined || value === '') {
        return '-'
      }
      if (typeof value === 'number') {
        return value
      }
      if (typeof value === 'string') {
        return value
      }
      try {
        return JSON.stringify(value)
      } catch (error) {
        return String(value)
      }
    },
    buildTable(info) {
      return [
        { key: this.$t('message.game_time') + this.timeLabel(), value: this.toDisplayValue(info.time || this.recordData.time) },
        { key: this.$t('message.game_number'), value: this.toDisplayValue(info.id || this.recordData.id) },
        { key: this.$t('message.game_round_id'), value: this.toDisplayValue(info.round_id) },
        { key: this.$t('message.game_name'), value: this.toDisplayValue(info.game_name || this.recordData.game_name) },
        { key: this.$t('message.room_type'), value: this.toDisplayValue(info.room_type) },
        { key: this.$t('message.total_revenue'), value: this.toDisplayValue(info.all_profit) }
      ].filter(item => item.value !== '-')
    },
    buildDetailRows(info) {
      const entries = []
      const skipKeys = new Set([
        'time',
        'id',
        'round_id',
        'game_name',
        'room_type',
        'all_profit',
        'user_list',
        'seats',
        'banker',
        'replay_data',
        'cards',
        'card',
        'left',
        'right',
        'public'
      ])

      Object.keys(info || {}).forEach(key => {
        if (skipKeys.has(key)) {
          return
        }
        const value = info[key]
        if (Array.isArray(value) || (value && typeof value === 'object')) {
          return
        }
        entries.push({
          key,
          value: this.toDisplayValue(value)
        })
      })

      return entries
    },
    open() {
      this.timmer = setTimeout(() => {
        if (this.loading) {
          bus.$message({
            message: this.$t('message.request_timeout'),
            duration: 0,
            type: 'warning'
          })
        }
      }, 5000)

      this.detail_visible = true
      this.recordData = Lockr.get('chess_detail') || {}

      const requestUrl = this.is_douyu ? this.douyu_url : this.recordData.dt_new_url
      if (!requestUrl) {
        this.loading = false
        this.tableData = []
        this.detailRows = []
        this.detailJson = ''
        return
      }

      this.apiGet(requestUrl).then((res) => {
        this.handelResponse(res, (payload) => {
          const data = this.is_douyu ? payload.data : payload
          const info = data.info || {}

          this.tableData = this.buildTable(info)
          this.detailRows = this.buildDetailRows(info)
          this.detailJson = JSON.stringify(data, null, 2)
          this.play_video_url = data.play_video_url || ''
          this.loading = false
        })
      })
    },
    close() {
      clearTimeout(this.timmer)
      Lockr.rm('chess_detail')
      this.loading = true
      this.detail_visible = false
      this.recordData = {}
      this.tableData = []
      this.detailRows = []
      this.detailJson = ''
      this.play_video_url = ''
    }
  },
  created() {
    const data = this.$route.query
    if (JSON.stringify(data) !== '{}') {
      document.title = this.$t('message.game_details')
      this.is_douyu = true
      this.default_width = '100%'
      if (data.token) {
        window.axios.defaults.headers.Authorization = 'Bearer ' + data.token
      }
      if (data.token && data.round_id && data.agent_id) {
        this.douyu_url = _g.dyUrl() + '?round_id=' + data.round_id + '&agent_id=' + data.agent_id
      }
      this.open()
    }
  },
  mixins: [http]
}
</script>
