<template>
  <div v-loading="loading">
    <el-row>
      <el-col :span="24">
        <el-table
          :data="table_data_list"
          stripe
          style="width: 100%;"
          :header-cell-style="{background:'#e2e6eb'}"
          class="chess-account-header"
        >
          <el-table-column
            type="index"
            :index="indexMethod(this.index)"
            :label="$t('message.no')"
            :empty-text="$t('message.no_data')"
            align="center"
            width="80"
          ></el-table-column>
          <el-table-column
            prop="id"
            :label="$t('message.transaction_id')"
            align="center"
            min-width="15%"
          >
            <!--单号-->
          </el-table-column>
          <el-table-column
            prop="epoch"
            :label="$t('message.operation_time') + timeLabel()"
            align="center"
            min-width="15%"
          >
            <!--操作时间-->
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
            prop="type"
            :label="$t('message.transaction_type')"
            align="center"
            min-width="18%"
          >
            <!--操作类型-->
          </el-table-column>
          <el-table-column prop="result" :label="$t('message.gain')" align="center" min-width="15%">
            <!--收支-->
          </el-table-column>
          <el-table-column
            prop="start_chips"
            :label="$t('message.pre_balance')"
            align="center"
            min-width="18%"
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
            min-width="18%"
          >
            <!--结束筹码-->
            <template slot-scope="scope">
              <i style="font-style:normal;">{{(scope.row.end_chips / 100).toFixed(2)}}</i>
            </template>
          </el-table-column>
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

<style>
.chess-account-header .el-table__header-wrapper {
  position: absolute;
  z-index: 100;
}
/* .chess-account-header .el-table__body-wrapper{
  margin-top: 48px;
} */
.el-row {
  margin-bottom: 20px;
  &:last-child {
    margin-bottom: 0;
  }
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
.el-button--primary {
  color: #fff;
  background-color: #4d75a1;
  border-color: #4d75a1;
}
</style>

<script>
import http from '../../assets/js/http'
export default {
  data() {
    return {
      loading: true,
      current_size: 0,
      query_data: [],
      table_data_list: [],
      current_page: 1,
      total: 1,
      index: 1
    }
  },
  computed: {},
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
      this.getaccountData(this.query_data)
    },
    handleCurrentChange(val) {
      this.current_page = val
      this.getaccountData(this.query_data)
    },
    getaccountData(form_data) {
      const data = {
        params: {
          game_id: form_data.game_id,
          start_time: form_data.game_time[0],
          end_time: form_data.game_time[1],
          page: this.current_page,
          page_size: form_data.game_size,
          language: Lockr.get('language'),
          m_type: 1
        }
      }
      this.apiGet(_g.apiUrl() + 'bull_statement', data).then(res => {
        this.handelResponse(res, data => {
          this.loading = false
          this.table_data_list = data.info.list
          this.total = data.info.total
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
    this.query_data = Lockr.get('query_data')
    this.getaccountData(this.query_data)
  },
  mixins: [http]
}
</script>


// WEBPACK FOOTER //
// chess_account.vue?5e03e3ce