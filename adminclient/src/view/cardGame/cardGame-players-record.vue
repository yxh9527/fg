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
          type="datetime"
          v-model="startTime"
          style="width: 175px"
          placeholder="选择开始日期"
        ></DatePicker
        >&emsp;—&emsp;
        <DatePicker
          type="datetime"
          v-model="endTime"
          style="width: 175px"
          placeholder="选择结束日期"
        ></DatePicker
        >&nbsp;
        <Select
          v-model="gameId.value"
          placeholder="选择游戏"
          style="width: 180px"
        >
          <Option
            v-for="items in gameId.option"
            :value="items.id"
            :key="items.id"
            >{{ items.name }}</Option
          >
        </Select>
        <label style="margin: 0 10px">交易类型</label>
        <Select v-model="dealType.value" style="width: 120px">
          <Option
            v-for="items in dealType.option"
            :value="items.id"
            :key="items.id"
            >{{ items.name }}</Option
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
      <Tables ref="tables" editable v-model="tableData" :columns="columns" />
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
import { getPlayerFwData, getSelectGames } from '@/api/data'
import { getDate } from '@/libs/tools'
export default {
  name: 'players-record',
  components: {
    Tables
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
            return params.row.state == 1 ? (
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
      dealType: {
        value: '',
        option: [
          { id: 1, name: '充值' },
          { id: 2, name: '上分' },
          { id: 3, name: '其他' }
        ]
      },
      columns: [
        { title: '流水号', key: 'flowingWaterOn', width: 220 },
        {
          title: '时间',
          key: 'createTime',
          align: 'center',
          width: 180,
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        { title: '游戏名称', key: 'gameName', align: 'center' },
        { title: '游戏平台', key: 'gameClassName', align: 'center' },
        { title: '游戏房间', key: 'difficultyName', align: 'center' },
        {
          title: '交易类型',
          key: 'fwType',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {{
                  1: '下注',
                  2: '返奖',
                  3: '返奖',
                  4: '下注',
                  6: '返奖',
                  7: '返奖'
                }[params.row.fwType] || ''}
              </span>
            )
          }
        },
        { title: '备注', key: 'explain', align: 'center' },
        {
          title: '账变前金额',
          key: 'score',
          align: 'center',
          render (h, params) {
            let sum = ''
            params.row.fwType == 1
              ? (sum = Number(params.row.userScore) + Number(params.row.bets))
              : (sum = Number(params.row.userScore) - Number(params.row.bets))
            return <span>{sum.toFixed(2)}</span>
          }
        },
        {
          title: '账变金额',
          key: 'bets',
          align: 'center',
          render (h, params) {
            return params.row.fwType == 2 || params.row.fwType == 3 ? (
              <span style="color:green">{params.row.bets.toFixed(2)}</span>
            ) : params.row.class != 1 ? (
              <span style="color:red">{params.row.bets.toFixed(2)} </span>
            ) : (
              ''
            )
          }
        },
        {
          title: '账变后金额',
          key: 'userScore',
          align: 'center',
          render (h, params) {
            return params.row.userScore >= 0 ? (
              <span style="color:green">
                {Number(params.row.userScore).toFixed(2)}
              </span>
            ) : (
              <span style="color:red">
                {Number(params.row.userScore).toFixed(2)}{' '}
              </span>
            )
          }
        }
      ],
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
      const url = location.search
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        {
          startTime: getDate(this.startTime)
        },
        {
          endTime: getDate(this.endTime)
        },
        { userId: this.id },
        { gameId: this.gameId.value }
      ]
      if (url.indexOf('?') != -1) {
        var str = url.split('=')
        Data.push({ officeNumber: str[1] })
      }
      getPlayerFwData(Data).then((res) => {
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
    getSelectGames(sessionStorage.getItem('agentVal')).then((res) => {
      this.gameId.option.push(...Object.assign(res.data.data))
    })
  }
}
</script>

<style>
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
