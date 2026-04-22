<template>
  <div v-loading="loading">
    <el-row>
      <el-col :span="24">
        <el-table
          :data="table_data_list"
          stripe
          :header-cell-style="{background:'#e2e6eb'}"
          class="chess-record-header J_Record"
        >
          <el-table-column
            type="index"
            :index="indexMethod(this.index)"
            :label="$t('message.no')"
            :empty-text="$t('message.no_data')"
            align="center"
            width="100"
          ></el-table-column>
          <el-table-column
            prop="round_id"
            :label="$t('message.gambling_id')"
            align="center"
            min-width="10%"
          >
            <!--牌局ID-->
          </el-table-column>
          <el-table-column
            prop="game_name"
            :label="$t('message.game_name')"
            align="center"
            min-width="10%"
          >
            <!--游戏名称-->
          </el-table-column>
          <el-table-column
            prop="epoch"
            :label="$t('message.operation_time') + timeLabel()"
            align="center"
            min-width="20%"
          >
            <!--操作时间-->
          </el-table-column>
          <el-table-column
            prop="all_bets"
            :label="$t('message.total_bets')"
            align="center"
            min-width="10%"
          >
            <!--总下注-->
            <template slot-scope="scope">
              <i
                style="font-style:normal;"
              >{{(scope.row.all_bets / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
          <el-table-column prop="result" :label="$t('message.gain')" align="center" min-width="10%">
            <!--收支-->
            <template slot-scope="scope">
              <i style="font-style:normal;">{{(scope.row.result / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
          <el-table-column
            prop="effective_dama"
            :label="$t('message.effective_dama')"
            align="center"
            min-width="10%"
          >
            <!--有效下注-->
            <template slot-scope="scope">
              <i style="font-style:normal;">{{(scope.row.effective_dama / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
          <el-table-column
            prop="start_chips"
            :label="$t('message.pre_balance')"
            align="center"
            min-width="12%"
          >
            <!--开始筹码-->
            <template slot-scope="scope">
              <i style="font-style:normal;">{{(scope.row.start_chips / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
          <el-table-column
            prop="end_chips"
            :label="$t('message.post_balance')"
            align="center"
            min-width="12%"
          >
            <!--结束筹码-->
            <template slot-scope="scope">
              <i style="font-style:normal;">{{(scope.row.end_chips / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
          <el-table-column
            prop="dt_url"
            :label="$t('message.detail')"
            align="center"
            min-width="10%"
          >
            <!--详情-->
            <template slot-scope="scope">
              <el-button
                @click.native.prevent="queryRowDetail(scope.row)"
                type="text"
                size="small"
              >{{ $t("message.query") }}</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-table :data="count_data" :show-header="false">
          <el-table-column type="data1" align="center" width="100"></el-table-column>
          <el-table-column prop="data2" align="center" min-width="20%"></el-table-column>
          <el-table-column prop="data3" align="center" min-width="15%"></el-table-column>
          <el-table-column prop="data4" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data5" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data6" align="center" min-width="30%"></el-table-column>
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
.chess-record-header .el-table__header-wrapper {
  position: absolute;
  z-index: 100;
}
.chess-record-header .el-table__body-wrapper {
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
      count_data: [{ data1: '', data2: '', data3: '', data4: 0, data5: 0, data6: '' }, { data1: '', data2: '', data3: '', data4: 0, data5: 0, data6: '' }],
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
      Lockr.set('chess_detail', row)
      bus.$emit('chess_detail', row)
    },
    getRecordData(form) {
      const form_data = {
        params: {
          game_id: form.game_id,
          page_size: form.game_size,
          page: this.current_page,
          start_time: form.game_time[0],
          end_time: form.game_time[1],
          language: Lockr.get('language'),
          is_game: 0
        }
      }

      this.apiGet(_g.apiUrl() + 'bull_log', form_data).then(res => {
        this.handelResponse(res, data => {
          this.loading = false
          this.table_data_list = data.info.list
          this.total = data.info.total
          this.count_data.forEach((val, index) => {
            if (index === 0) {
              val.data4 = data.info.s_data.x_all_bets
              val.data5 = data.info.s_data.x_all_bpay
            } else if (index === 1) {
              val.data4 = data.info.s_data.all_bets
              val.data5 = data.info.s_data.all_bpay
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
// chess_record.vue?5f10548c