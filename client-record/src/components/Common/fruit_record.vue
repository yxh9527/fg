<template>
  <div v-loading="loading">
    <el-row>
      <el-col :span="24">
        <el-table
          :data="table_data_list"
          stripe
          :header-cell-style="{background:'#e2e6eb'}"
          class="fruit-record-header J_Record"
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
            prop="game_name"
            :label="$t('message.game_name')"
            align="center"
            min-width="15%"
          ></el-table-column>
          <el-table-column
            prop="id"
            :label="$t('message.transaction_id')"
            align="center"
            min-width="10%"
          ></el-table-column>
          <el-table-column
            prop="time"
            :label="$t('message.operation_time') + timeLabel()"
            align="center"
            min-width="20%"
          ></el-table-column>
          <el-table-column
            prop="game_type"
            :label="$t('message.type')"
            align="center"
            min-width="10%"
          ></el-table-column>
          <el-table-column
            prop="all_bets"
            :label="$t('message.total_bets')"
            align="center"
            min-width="10%"
          ></el-table-column>
          <el-table-column
            prop="all_bonus"
            :label="$t('message.win')"
            align="center"
            min-width="10%"
          ></el-table-column>
          <el-table-column
            prop="all_bpay"
            :label="$t('message.income')"
            align="center"
            min-width="10%"
          ></el-table-column>
          <el-table-column
            prop="detail"
            :label="$t('message.detail')"
            align="center"
            min-width="10%"
          >
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
          <el-table-column prop="data2" align="center" min-width="45%"></el-table-column>
          <el-table-column prop="data3" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data4" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data5" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data6" align="center" min-width="10%"></el-table-column>
          <el-table-column prop="data7" align="center" min-width="10%"></el-table-column>
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
.fruit-record-header .el-table__header-wrapper {
  position: absolute;
  z-index: 100;
}
.fruit-record-header .el-table__body-wrapper {
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
    /** 中奖金额 加上 JP大奖；输赢加上JP大奖 */
    calcSum() {
      if (this.table_data_list) {
        for (var key in this.table_data_list) {
          let bonus = (parseFloat(this.table_data_list[key].all_bonus) + parseFloat(this.table_data_list[key].jp_bonus)).toFixed(2)
          let bpay = (parseFloat(this.table_data_list[key].all_bpay) + parseFloat(this.table_data_list[key].jp_bonus)).toFixed(2)
          this.table_data_list[key].all_bonus = bonus == 0 ? 0 : bonus
          this.table_data_list[key].all_bpay = bpay == 0 ? 0 : bpay
        }
      }
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
      Lockr.set('fruit_detail', row)
      bus.$emit('fruit_detail', row)
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

      this.apiGet(_g.apiUrl() + 'fruit_log', form_data).then(res => {
        this.handelResponse(res, data => {
          this.loading = false
          this.table_data_list = data.info.data
          this.calcSum()
          this.total = data.info.total
          this.count_data.forEach((val, index) => {
            if (index === 0) {
              val.data4 = data.info.s_data.x_all_bets
              val.data5 = data.info.s_data.x_all_bonus
              val.data6 = data.info.s_data.x_all_bpay
            } else if (index === 1) {
              val.data4 = data.info.s_data.all_bets
              val.data5 = data.info.s_data.all_bonus
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
// fruit_record.vue?27ac3614