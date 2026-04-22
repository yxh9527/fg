<template>
  <div v-loading="loading">
    <el-row>
      <el-col :span="24">
        <el-table
          :data="table_data_list"
          stripe
          :header-cell-style="{background:'#e2e6eb'}"
          class="hunting-record-header J_Record"
        >
          <el-table-column
            type="index"
            :index="indexMethod(this.index)"
            :label="$t('message.no')"
            :empty-text="$t('message.no_data')"
            align="center"
            width="60"
          ></el-table-column>
          <el-table-column prop="id" :label="$t('message.scene_id')" align="center" min-width="15%">
            <!--场景编号-->
          </el-table-column>
          <el-table-column
            prop="game_name"
            :label="$t('message.game_name')"
            align="center"
            min-width="15%"
          >
            <!--游戏名称-->
          </el-table-column>
          <el-table-column
            prop="game_model"
            :label="$t('message.game_mode')"
            align="center"
            min-width="16%"
          >
            <!--游戏模式-->
          </el-table-column>
          <el-table-column
            prop="start_time"
            :label="$t('message.start_time') + timeLabel()"
            align="center"
            min-width="16%"
          >
            <!--开始时间-->
          </el-table-column>
          <el-table-column
            prop="end_time"
            :label="$t('message.end_time') + timeLabel()"
            align="center"
            min-width="16%"
          >
            <!--结束时间-->
          </el-table-column>
          <el-table-column
            prop="bullet_count"
            :label="$t('message.bullets')"
            align="center"
            min-width="8%"
          >
            <!--子弹数量-->
          </el-table-column>
          <el-table-column
            prop="bullet_chips"
            :label="$t('message.bullet_total_price')"
            align="center"
            min-width="12%"
          >
            <!--子弹总价值-->
          </el-table-column>
          <el-table-column prop="bpay" :label="$t('message.income')" align="center" min-width="12%">
            <!--输赢-->
          </el-table-column>
          <el-table-column
            prop="detail"
            :label="$t('message.detail')"
            align="center"
            min-width="6%"
          >
            <!--操作-->
            <template slot-scope="scope">
              <el-button
                @click.native.prevent="queryRowDetail(scope.row)"
                type="text"
                size="small"
              >{{ $t("message.detail") }}</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-table :data="count_data" :show-header="false">
          <el-table-column type="data1" align="center" width="60"></el-table-column>
          <el-table-column prop="data2" align="center" min-width="46%"></el-table-column>
          <el-table-column prop="data3" align="center" min-width="16%"></el-table-column>
          <el-table-column prop="data4" align="center" min-width="8%"></el-table-column>
          <el-table-column prop="data5" align="center" min-width="12%"></el-table-column>
          <el-table-column prop="data6" align="center" min-width="12%"></el-table-column>
          <el-table-column prop="data7" align="center" min-width="6%"></el-table-column>
        </el-table>
      </el-col>
    </el-row>

    <el-row>
      <el-col :span="24" align="center">
        <div>
          <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            :current-page="current_page"
            :page-sizes="[20, 50, 100]"
            :page-size="query_data.game_size"
            layout="total, sizes, prev, pager, next, jumper"
            :total="total"
          ></el-pagination>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style lang="scss">
.hunting-record-header .el-table__header-wrapper {
  position: absolute;
  z-index: 100;
}
.hunting-record-header .el-table__body-wrapper {
  margin-top: 48px;
}
.el-table__body tr {
  height: 40px;
}
.el-table td {
  padding: 3px 0;
}
.el-row {
  margin-bottom: 20px;
  &:last-child {
    margin-bottom: 0;
  }
}
.el-table .cell {
  padding-left: 5px;
  padding-right: 5px;
}
.el-button--primary {
  color: #fff;
  background-color: #4d75a1;
  border-color: #4d75a1;
}
.grid-content {
  border-radius: 0;
  min-height: 36px;
}
.left {
  float: left;
  display: inline-flex;
  align-items: center;
  margin-left: 10px;
}
</style>

<script>
import http from '../../assets/js/http'
export default {
  data() {
    return {
      loading: true,
      current_size: 0,
      query_data: {},
      table_data_list: [],
      count_data: [
        { data1: '', data2: '', data3: '', data4: 0, data5: 0, data6: 0, data7: '' },
        { data1: '', data2: '', data3: '', data4: 0, data5: 0, data6: 0, data7: '' }
      ],
      current_page: 1,
      total: 0,
      index: 1
    }
  },
  methods: {
    timeLabel() {
      let utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    indexMethod(index) {
      return (this.current_page - 1) * this.query_data.game_size + index++
    },
    handleSizeChange(val) {
      this.query_data.game_size = val
      this.current_size = val
      this.getRecordData(this.query_data)
    },
    handleCurrentChange(val) {
      this.current_page = val
      this.getRecordData(this.query_data)
    },
    queryRowDetail(row) {
      if (row.type == 2) {
        this.$confirm(this.$t('message.confirm_JP_rewards'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        })
      } else if (row.type == 15) {
        this.$confirm(this.$t('message.confirm_bonus_rewards'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        })
      } else if (row.type == 18) {
        this.$confirm(this.$t('message.confirm_lucky_rewards'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        })
      } else if (row.type == 23) {
        this.$confirm(this.$t('message.boss_ranking_list'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        }).catch(() => {})
      } else if (row.type == 24) {
        this.$confirm(this.$t('message.boss_last_blow'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        }).catch(() => {})
      } else if (row.type == 77) {
        this.$confirm(this.$t('message.mysterious_treasure'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        }).catch(() => {})
      } else if (row.type == 78) {
        this.$confirm(this.$t('message.daily_rewards'), {
          showCancelButton: false,
          showConfirmButton: false,
          showClose: false,
          center: true
        }).catch(() => {})
      } else {
        Lockr.set('hunting_detail', row)
        bus.$emit('hunting_detail', row)
      }
    },
    getRecordData(form) {
      const form_data = {
        params: {
          game_id: form.game_id,
          size: form.game_size,
          page: this.current_page,
          start_time: form.game_time[0],
          end_time: form.game_time[1],
          language: Lockr.get('language'),
          is_game: 0
        }
      }

      this.apiGet(_g.apiUrl() + 'fish_log', form_data).then(res => {
        this.handelResponse(res, data => {
          this.loading = false
          this.table_data_list = data.info.data
          this.total = data.info.total
          this.count_data.forEach((val, index) => {
            if (index === 0) {
              val.data4 = data.info.s_data.x_bullet_count
              val.data5 = data.info.s_data.x_bullet_chips
              val.data6 = data.info.s_data.x_all_bpay
            } else if (index === 1) {
              val.data4 = data.info.s_data.bullet_count
              val.data5 = data.info.s_data.bullet_chips
              val.data6 = data.info.s_data.all_bpay
            }
          })
        })
      })
    }
  },
  watch: {
    current_size: function(val) {
      bus.$emit('game_size', val)
    },
    table_data_list: function(val) {
      if (val.length > 0) {
        this.loading = false
      }
    }
  },
  created() {
    setTimeout(() => {
      this.loading = false
    }, 5000)
    this.count_data[0].data3 = this.$t('message.sub_total') + '：'
    this.count_data[1].data3 = this.$t('message.all_total') + '：'
    this.query_data = Lockr.get('query_data')
    this.current_size = this.query_data.game_size
    this.getRecordData(this.query_data)
  },
  mounted() {
    this.$nextTick(() => {
      const table = window.document.querySelector('.J_Record .el-table__body-wrapper')
      const header = window.document.querySelector('.J_Record .el-table__header-wrapper')
      if (table && header) {
        table.style.marginTop = header.clientHeight + 'px'
      }
    })
  },
  mixins: [http]
}
</script>


// WEBPACK FOOTER //
// hunting_record.vue?111eb16c