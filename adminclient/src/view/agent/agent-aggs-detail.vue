<template>
  <div>
    <Card>
      <div class="select-style" style="margin-left: 10px">
        <label>代理选择:</label>
        <Select
          v-model="agentId"
          style="width: 210px"
          filterable
          filter-by-label
          @on-change="refreshPage"
        >
          <Option
            v-for="item in agentOption"
            :value="item.id"
            :key="item.id"
            :label="item.name"
            >{{ item.name }}</Option
          >
        </Select>
      </div>
    </Card>
    <Card style="margin-top: 10px">
      <div style="height: 800px; overflow: scroll">
        <tables ref="tables" v-model="tableData" :columns="columns" />
        <Spin fix v-if="spinShow">
          <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
          <div>数据加载中</div>
        </Spin>
      </div>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getAgentGameDataAggs } from '@/api/data'
export default {
  name: 'agent-aggs-detail',
  components: {
    Tables
  },
  inject: ['viewAccess', 'handleLogOut', 'reFreshSiteAgentList'],
  data () {
    return {
      tableData: [],
      agentOption: [],
      agentId: 0,
      columns: [
        {
          title: '代理ID',
          key: 'agentId',
          width: 80,
          align: 'center',
          isdisabled: true
        },
        {
          title: '代理名称',
          key: 'agentName',
          width: 150,
          align: 'center',
          isdisabled: true
        },
        {
          title: '游戏名称',
          key: 'gameName',
          width: 450,
          align: 'center',
          isdisabled: true
        },
        {
          title: 'Symbol',
          key: 'symbol',
          width: 150,
          align: 'center',
          isdisabled: true
        },
        {
          title: '人次',
          key: 'userNumber',
          align: 'center',
          minWidth: 100,
          showRedPoint: true
        },
        { title: '局数', key: 'gameNumber', align: 'right', minWidth: 100 },
        {
          title: '有效下注',
          key: 'effeBetsScore',
          align: 'center',
          minWidth: 100,
          render (h, params) {
            return (
              <span>
                {(params.row.effeBetsScore &&
                  params.row.effeBetsScore.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '有效打码',
          key: 'chips',
          align: 'right',
          minWidth: 100,
          render (h, params) {
            return (
              <span>
                {(params.row.chipsTotal && params.row.chipsTotal.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '赔付',
          key: 'profitLossTotal',
          align: 'right',
          minWidth: 100,
          render (h, params) {
            if (params.row.profitLossTotal) {
              let sum
              sum = Number(
                String(params.row.profitLossTotal).replace(/,/g, '')
              )
              sum = sum.toFixed(2)
              return <span>{String(sum)}</span>
            } else {
              return <span style="color:#000">0</span>
            }
          }
        },
        {
          title: '盈亏',
          key: 'yk',
          align: 'right',
          minWidth: 100,
          render (h, params) {
            let sum, effect
            sum = Number(String(params.row.profitLossTotal).replace(/,/g, ''))
            effect = Number(
              String(params.row.effectiveBetsTotal).replace(/,/g, '')
            )
            return effect - sum >= 0 ? (
              <span style="color:green">
                {String((effect - sum).toFixed(2))}
              </span>
            ) : (
              <span style="color:red">{String((effect - sum).toFixed(2))}</span>
            )
          }
        },
        {
          title: '税收',
          key: 'revenueTotal',
          align: 'right',
          minWidth: 80,
          render (h, params) {
            return (
              <span>
                {(params.row.revenueTotal &&
                  params.row.revenueTotal.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '杀数',
          key: '',
          align: 'right',
          minWidth: 80,
          render (h, params) {
            let sum = Number(
              String(params.row.profitLossTotal).replace(/,/g, '')
            )
            let effect = Number(
              String(params.row.effectiveBetsTotal).replace(/,/g, '')
            )
            return (
              <span>
                {isNaN(
                  ((effect - sum) / params.row.effectiveBetsTotal).toFixed(3)
                )
                  ? 0
                  : ((effect - sum) / params.row.effectiveBetsTotal).toFixed(3)}
              </span>
            )
          }
        }
      ],
      spinShow: false
    }
  },
  methods: {
    getAgentGameAggs () {
      let Data = [
        { agentId: this.$route.query.agent },
        { startTime: this.$route.query.startTime },
        { endTime: this.$route.query.endTime }
      ]
      getAgentGameDataAggs(Data)
        .then((res) => {
          this.tableData = res.data.data.data
        })
        .catch((err) => {
          console.log(err)
        })
    },
    refreshPage () {
      let Data = [
        { agentId: this.agentId },
        { startTime: this.$route.query.startTime },
        { endTime: this.$route.query.endTime },
        { webId: this.$route.query.webId }
      ]
      getAgentGameDataAggs(Data)
        .then((res) => {
          this.tableData = res.data.data.data
        })
        .catch((err) => {
          console.log(err)
        })
    }
  },
  mounted () {
    let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
    this.agentOption = siteOption.find(
      (site) => site.id == this.$route.query.webId
    ).agentList
    this.agentOption.map((item) => {
      item.label = item.name
    })
    this.getAgentGameAggs()
  }
}
</script>

<style lang="less">
* {
  -webkit-tap-highlight-color: rgba(255, 255, 255, 0);
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
.gameList {
  margin-bottom: 10px;
  margin-top: 10px;
  padding-bottom: 5px;
  border-bottom: 1px solid #ccc;
  font-size: 16px;
}
.message-box {
  display: inline-table;
  margin: 40px 20px 30px 0;
  font-size: 16px;
  text-align: left;
  > div {
    padding: 5px 0;
  }
  button {
    margin-top: 10px;
    margin-right: 20px;
  }
}
</style>
