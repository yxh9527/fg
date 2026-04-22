 <template>
  <div>
    <Card>
      <h2 style="margin-bottom:10px;">基本信息</h2>
      <tables :columns="columns1" v-model="userInfo" />
      <br />
      <h2 style="margin-bottom:10px;">对局详情</h2>
      <Card>
        <label>对局时间</label>&nbsp;
        <DatePicker type="datetime" style="width:175px" v-model="startTime" placeholder="选择开始日期"></DatePicker>&emsp;—&emsp;
        <DatePicker type="datetime" style="width:175px" v-model="endTime" placeholder="选择结束日期"></DatePicker>&nbsp;
        <Select v-model="gameId.value" placeholder="选择游戏" style="width: 180px">
          <Option v-for="items in gameId.option" :value="items.id" :key="items.id">{{ items.name }}</Option>
        </Select>
        <div style="right: 20px;top: 18px;position: absolute;">
          <label style="margin-right:50px;">
            盈利分数：
            <span style="color:red;font-size:18px;">{{totalProfitLoss}}</span>
          </label>
          <Button type="primary" @click="searchAction">搜索</Button>
        </div>
      </Card>
      <br />
      <tables ref="tables" editable v-model="tableData" :columns="columns" />
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
import Detial from './cardGame-detial.vue'
import {
  getPlayerFwDetailData,
  getSelectGames
} from '@/api/data'
export default {
  name: 'cardGame-player-info',
  components: {
    Tables,
    Detial
  },
  props: ['id'],
  data () {
    return {
      columns1: [
        {
          title: '玩家ID',
          key: 'userId',
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
            return (
              <span>
                {new Date(params.row.logTime * 1000).toLocaleString('chinese', {
                  hour12: false
                })}
              </span>
            )
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
            return params.row.state == 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        }
      ],
      userInfo: [],
      gameId: { value: '', option: [] },
      columns: [
        {
          title: '游戏名称',
          key: 'gameName',
          width: 150,
          render (h, params) {
            return params.row.settType == 2 ? (
              <span>搏一搏</span>
            ) : (
              <span>{params.row.gameName}</span>
            )
          }
        },
        { title: '游戏平台', key: 'platFormName', width: 120 },
        {
          title: '游戏房间',
          key: 'difficultyName',
          width: 150,
          render (h, params) {
            return params.row.settType == 2 ? (
              ''
            ) : (
              <span>{params.row.difficultyName}</span>
            )
          }
        },
        { title: '局号', key: 'officeNumber', width: 230 },
        {
          title: '开局时间',
          key: 'beginTime',
          render (h, params) {
            return (
              <span>
                {new Date(params.row.beginTime * 1000).toLocaleString(
                  'chinese',
                  { hour12: false }
                )}
              </span>
            )
          }
        },
        { title: '有效下注', key: 'effectiveBets' },
        {
          title: '局总返奖',
          key: 'profitLoss',
          render (h, params) {
            return params.row.profitLoss >= 0 ? (
              <span style="color:green">{params.row.profitLoss}</span>
            ) : (
              <span style="color:red">{params.row.profitLoss}</span>
            )
          }
        },
        {
          title: '对局详情',
          type: 'expand',
          align: 'center',
          render: (h, params) => {
            if (params.row.detail) {
              return h(Detial, {
                props: {
                  rowInfo: params.row.detail.details,
                  rowId: params.row.userId
                }
              })
            } else {
              return (
                <div style="text-align:center">
                  <Icon size="24" type="logo-freebsd-devil" />
                  没有详情数据
                </div>
              )
            }
          }
        },
        {
          title: '流水查询',
          align: 'center',
          render: (h, params) => {
            return h(
              'Button',
              {
                props: {
                  type: 'info',
                  size: 'small',
                  to:
                    '/players-record-' +
                    params.row.userId +
                    '?on=' +
                    params.row.officeNumber
                },
                style: {
                  marginRight: '5px'
                }
              },
              '查询'
            )
          }
        }
      ],
      startTime: '',
      endTime: '',
      totalProfitLoss: '',
      tableData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
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
        {
          startTime: getDate(this.startTime)
        },
        {
          endTime: getDate(this.endTime)
        },
        { gameId: this.gameId.value }
      ]
      getPlayerFwDetailData(Data).then(res => {
        this.tableData = []
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
    getSelectGames(sessionStorage.getItem('agentVal')).then(res => {
      this.gameId.option.push(...Object.assign(res.data.data))
    })

    // getPlayerInfoData(this.id).then(res => {
    //   this.userInfo.push(res.data.data);
    // });

    // this.handleSearch();
  }
}
</script>

<style>
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
