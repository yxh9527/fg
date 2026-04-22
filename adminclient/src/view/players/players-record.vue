 <template>
  <div>
    <Card>
      <h2 style="margin-bottom: 10px">基本信息</h2>
      <tables :columns="columns1" v-model="userInfo" />
      <br />
      <h2 style="margin-bottom: 10px">流水详情</h2>
      <Card>
        <label>流水时间</label>&nbsp;
        <DatePicker
          v-model="startTime"
          placeholder="选择开始日期"
          style="width: 175px"
          :clearable="false"
        ></DatePicker
        >&emsp;—&emsp;
        <DatePicker
          v-model="endTime"
          placeholder="选择结束日期"
          style="width: 175px"
          :clearable="false"
        ></DatePicker
        >&nbsp;
        <Select
          v-model="gameId.value"
          placeholder="选择游戏"
          style="width: 400px"
          filterable
          filter-by-label
          clearable
        >
          <Option
            v-for="items in gameId.option"
            :value="items.symbol"
            :key="items.symbol"
            :label="items.label"
            >{{ items.label }}</Option
          >
        </Select>
        <Button
          type="primary"
          style="right: 20px; top: 18px; position: absolute"
          @click="searchAction"
          >搜索</Button
        >
      </Card>
      <br />
      <Tables
        ref="tables"
        editable
        v-model="tableData"
        :columns="columns"
        style="min-height: 400px"
      />
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
      <div class="pageStyle">
        <Page
          :total="pageData.current"
          :current="pageData.page"
          :page-size="pageData.pageSize"
          :page-size-opts="pageData.pageOpts"
          show-elevator
          show-sizer
          show-total
          title="输入页数后，按Enter键跳转页面"
          @on-change="changePage"
          @on-page-size-change="changePageSize"
        />
      </div>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import { getPlayerInfoData, getPlayerFwData } from '@/api/data'
import { getDate } from '@/libs/tools'
import * as dayjs from 'dayjs'

export default {
  name: 'players-record',
  components: {
    Tables
  },
  props: ['id'],
  data () {
    let _this = this
    return {
      columns1: [
        {
          title: '玩家ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '玩家昵称',
          key: 'nickName',
          align: 'center'
        },
        {
          title: '站点',
          key: 'webName',
          align: 'center'
        },
        {
          title: '所属代理',
          key: 'agentName',
          align: 'center'
        },
        {
          title: '最近登录时间',
          key: 'logTime',
          align: 'center',
          width: 180,
          render (h, params) {
            return <span>{getDate(params.row.logTime * 1000)}</span>
          }
        },
        {
          title: '最近登录IP',
          key: 'logIp',
          align: 'center'
        },
        {
          title: '局数',
          key: 'totalNumber',
          align: 'center'
        },
        {
          title: '时长',
          key: 'times',
          align: 'center'
        },
        {
          title: '有效下注',
          key: 'totalEffBet',
          align: 'center'
        },
        {
          title: '状态',
          key: 'state',
          align: 'center',
          render (h, params) {
            return params.row.state <= 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        }
      ],
      userInfo: [],
      startTime: '',
      endTime: '',
      gameId: { value: '', option: [] },
      columns: [
        { title: '流水号', key: 'flowingWaterOn', width: 300 },
        { title: '局号', key: 'roundId', width: 300 },
        {
          title: '时间',
          key: 'createTime',
          align: 'center',
          width: 180,
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        {
          title: '游戏名称',
          key: 'gameName',
          align: 'center',
          render (h, params) {
            for (let i = 0; i < _this.gameId.option.length; i++) {
              if (params.row.symbol == _this.gameId.option[i].symbol) {
                return <span>{_this.gameId.option[i].label}</span>
              }
            }
            return <span>"未知游戏"</span>
          }
        },
        {
          title: '交易类型',
          key: 'desc',
          align: 'center',
          width: 100,
          render (h, params) {
            return <span>{params.row.desc}</span>
          }
        },
        {
          title: '账变前金额',
          key: 'pScore',
          align: 'center',
          width: 150,
          render (h, params) {
            let sum = Number(params.row.currentScore) - Number(params.row.bet)
            return <span>{sum.toFixed(2)}</span>
          }
        },
        {
          title: '账变金额',
          key: 'bet',
          align: 'center',
          width: 100,
          render (h, params) {
            return <span>{parseFloat(params.row.bet).toFixed(2)}</span>
          }
        },
        {
          title: '账变后金额',
          key: 'currentScore',
          align: 'center',
          width: 150,
          render (h, params) {
            return <b>{Number(params.row.currentScore).toFixed(2)}</b>
          }
        }
      ],
      tableData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      spinShow: false
    }
  },
  methods: {
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    },
    handleSearch () {
      const url = location.search
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        {
          startTime: this.startTime
            ? dayjs(
              dayjs(this.startTime.getTime()).format('YYYY-MM-DD 00:00:00')
            ).unix()
            : ''
        },
        {
          endTime: this.endTime
            ? dayjs(
              dayjs(this.endTime.getTime()).format('YYYY-MM-DD 23:59:59')
            ).unix()
            : ''
        },
        { userId: this.id },
        { symbol: this.gameId.value },
        { agentId: this.$route.query.agent }
      ]

      /**
       * 验证时间范围合法
       */
      if (Data.find((x) => x.startTime || x.endTime)) {
        let startTime = new Date(
          Data.find((x) => x.startTime).startTime
        ).getTime()
        let endTime = new Date(Data.find((x) => x.endTime).endTime).getTime()
        if (endTime - startTime <= 0) {
          this.$Message.error('开始时间不允许大于结束时间')
          return
        }
      }

      this.spinShow = true

      if (url.indexOf('?') != -1) {
        if (url.indexOf('on') != -1) {
          Data.push({ officeNumber: this.$route.query.on })
        }
      }
      getPlayerFwData(Data).then((res) => {
        this.spinShow = false
        this.tableData = []
        if (res.data.code == 200) {
          if (!res.data.data.data.length) this.$Message.warning('暂无数据')
          this.tableData.push(...res.data.data.data)
          this.pageData.current = res.data.data.total
        } else {
          this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
        }
      })
    },
    changePage (index) {
      this.pageData.page = index
      this.handleSearch()
    },
    changePageSize (index) {
      this.pageData.pageSize = index
      this.handleSearch()
    },
    searchAction () {
      this.pageData.page = 1
      this.handleSearch()
    }
  },
  mounted () {
    let data = {
      id: this.id,
      // agentId: sessionStorage.getItem("agentVal"),
      agentId: this.$route.query.agent
    }
    getPlayerInfoData(data).then((res) => {
      this.userInfo = []
      this.userInfo.push(res.data.data)
    })
    // getSelectGames(sessionStorage.getItem("agentVal")).then((res) => {
    this.gameId.option.push({ id: '', number: '', symbol: '', name: '全部' })
    this.gameId.option.push(...JSON.parse(sessionStorage.getItem('games')))
    this.gameId.option.map((item) => {
      if (item.nameZH == '' || item.nameZH === undefined) {
        item.label = item.name
      } else {
        item.label = item.name + ' [' + item.nameZH + ']'
      }
    })
    // });
    this.handleSearch()
  }
}
</script>

<style>
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
