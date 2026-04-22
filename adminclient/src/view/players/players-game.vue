 <template>
  <div>
    <Card>
      <h2 style="margin-bottom: 10px">基本信息</h2>
      <tables :columns="columns1" v-model="userInfo" />
      <br />
      <h2 style="margin-bottom: 10px">对局详情</h2>
      <Card>
        <label>对局时间</label>&nbsp;
        <DatePicker
          :options="startDateRestrict"
          type="datetime"
          style="width: 175px"
          v-model="startTime"
          placeholder="选择开始日期"
        ></DatePicker
        >&emsp;—&emsp;
        <DatePicker
          :options="startDateRestrict"
          type="datetime"
          style="width: 175px"
          v-model="endTime"
          placeholder="选择结束日期"
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
            :value="items.number"
            :key="items.number"
            :label="items.label"
            >{{ items.label }}</Option
          >
        </Select>
        <div style="right: 20px; top: 18px; position: absolute">
          <Button type="primary" @click="searchAction">搜索</Button>
        </div>
        &nbsp; <label>注单号：</label>&nbsp;
        <Input
          v-model="officeNumber"
          placeholder="请输入"
          style="width: 180px"
        />
      </Card>
      <br />
      <tables ref="tables" editable v-model="tableData" :columns="columns" />
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

      <Modal v-model="viewSettlementDetail" width="65%" title="游戏详情">
        <div>
          <iframe
            :src="this.viewGameDetailUrl"
            width="100%"
            height="500"
            frameborder="0"
          ></iframe>
        </div>
        <div slot="footer"></div>
      </Modal>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import Detial from './players-game-detial.vue'
import { getDate } from '@/libs/tools'
import { setting } from '@/config'
import {
  getPlayerInfoData,
  getPlayerFwDetailData,
  getGameData2,
  getGameServers
} from '@/api/data'

export default {
  name: 'players-game',
  components: {
    Tables,
    Detial
  },
  props: ['id'],
  data () {
    return {
      startDateRestrict: {
        disabledDate (date) {
          if (date.getYear() != new Date().getYear()) {
            return true
          }
          if (date.getMonth() - new Date().getMonth() < -1) {
            return true
          }
          if (date.getMonth() - new Date().getMonth() > 0) {
            return true
          }
        }
      },
      columns1: [
        {
          title: '玩家ID',
          key: 'id'
        },
        {
          title: '玩家昵称',
          key: 'nickName'
        },
        {
          title: '站点',
          key: 'webName'
        },
        {
          title: '所属代理',
          key: 'agentName'
        },
        {
          title: '最近登录时间',
          key: 'logTime',
          width: 180,
          render (h, params) {
            return <span>{getDate(params.row.logTime * 1000)}</span>
          }
        },
        {
          title: '最近登录IP',
          key: 'logIp'
        },
        {
          title: '局数',
          key: 'totalNumber'
        },
        {
          title: '时长',
          key: 'times'
        },
        {
          title: '有效下注',
          key: 'totalEffBet'
        },
        {
          title: '状态',
          key: 'state',
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
      columns: [
        {
          title: '游戏id',
          key: 'gameId',
          width: 100
        },
        {
          title: '游戏名称',
          key: 'gameName'
        },
        { title: '局号', key: 'roundID', width: 200 },
        {
          title: '对局时间',
          width: 180,
          key: 'playedDate',
          render (h, params) {
            return <span>{getDate(params.row.playedDate)}</span>
          }
        },
        {
          title: '货币',
          key: 'currency',
          width: 100
        },
        { title: '有效下注', key: 'bet', width: 200 },
        {
          title: '局总盈亏',
          key: 'win',
          width: 200,
          render (h, params) {
            return params.row.win > 0 ? (
              <span style="color:green">
                {Number(params.row.win).toFixed(2)}
              </span>
            ) : (
              <span style="color:red">{Number(params.row.win).toFixed(2)}</span>
            )
          }
        },
        {
          title: '对局详情',
          align: 'center',
          width: 100,
          render: (h, params) => {
            return h(
              'a',
              {
                on: {
                  click: () => {
                    this.imgClick(params.row)
                  }
                }
              },
              '查看'
            )
          }
        },
        {
          title: '流水查询',
          align: 'center',
          width: 100,
          render: (h, params) => {
            return h(
              'a',
              {
                on: {
                  click: () => {
                    let routeData = this.$router.resolve({
                      path:
                        '/players-record-' +
                        params.row.userId +
                        '?on=' +
                        params.row.roundID,
                      query: { agent: params.row.agentId }
                    })
                    window.open(routeData.href, '_blank')
                  }
                }
              },
              '查询'
            )
          }
        }
      ],
      startTime: '',
      endTime: '',
      officeNumber: '',
      totalProfitLoss: '',
      gameId: { value: '', option: [] },
      tableData: [],
      viewSettlementDetail: false,
      viewGameDetailUrl: '',
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
    imgClick (row) {
      this.viewGameDetailUrl = this.replays[0] + '/share/' + row.hash
      this.viewSettlementDetail = true
    },
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    },
    handleSearch () {
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { userId: this.id },
        { officeNumber: this.officeNumber },
        {
          startTime: this.startTime ? getDate(this.startTime) : ''
        },
        {
          endTime: this.endTime ? getDate(this.endTime) : ''
        },
        { gameId: this.gameId.value },
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

      getPlayerFwDetailData(Data).then((res) => {
        res.data.data.data.map((d) => {
          if (d.detail && !d.detail.details) {
            d.detail = JSON.parse(d.detail)
          }
        })
        this.tableData = []
        this.spinShow = false
        this.tableData.push(...res.data.data.data)
        this.pageData.current = res.data.data.total
        this.totalProfitLoss = res.data.data.totalProfitLoss
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
    getGameServers().then((data) => {
      console.log('======>', data)
      if (data.data.data) this.replays = data.data.data.data.replays
    })
    getGameData2().then((res) => {
      this.gameId.option.push({ id: '', name: '全部', nameZH: '', number: 0 })
      this.gameId.option.push(...Object.assign(res.data.data))
      this.gameId.option.map((item) => {
        if (item.nameZH == '' || item.nameZH === undefined) {
          item.label = item.name
        } else {
          item.label = item.name + ' [' + item.nameZH + ']'
        }
      })
    })
    let data = {
      id: this.id,
      agentId: this.$route.query.agent
    }
    getPlayerInfoData(data).then((res) => {
      this.userInfo.push(res.data.data)
    })

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
