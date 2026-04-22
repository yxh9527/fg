<template>
  <div>
    <Card>
      <h2 style="margin-bottom: 10px">基本信息</h2>
      <tables :columns="columns1" v-model="userInfo" />
      <br />
      <h2 style="margin-bottom: 10px">看板详情</h2>
      <Row :gutter="20" v-for="(item, index) in setExample" :key="item.id">
        <i-col :md="24" :lg="16" style="margin-bottom: 20px">
          <Card shadow>
            <span
              style="position: absolute; left: 145px; top: 19px; z-index: 9"
            >
              <RadioGroup
                type="button"
                :name="'time-radio-' + index"
                v-model="timeVal[index + 1]"
                size="small"
                @on-change="
                  changeTimeData(
                    index + 3,
                    timeVal[index + 1],
                    gameVal[index],
                    true
                  )
                "
              >
                <Radio
                  v-for="items in timeList"
                  :key="items.value"
                  :label="items.value"
                  class="radio-style"
                  >{{ items.label }}</Radio
                >
              </RadioGroup>
            </span>
            <span
              style="position: absolute; left: 530px; top: 15px; z-index: 9"
            >
              <Select
                v-model="gameVal[index]"
                @on-change="
                  changeTimeData(index + 3, timeVal[index + 1], gameVal[index])
                "
                style="width: 150px"
              >
                <Option
                  v-for="items in gameSelectList"
                  :value="items.id"
                  :key="items.id"
                  >{{ items.name }}</Option
                >
              </Select>
            </span>
            <example
              v-if="item.show"
              style="height: 380px"
              :type="item.type"
              :legend="item.legend"
              :columns="item.columns"
              :barData="item.barData"
              :text="item.barText"
            />
          </Card>
        </i-col>
        <i-col :md="24" :lg="8" style="margin-bottom: 20px">
          <Card>
            <chart-pie
              v-if="item.pieShow && item.pieText != '盈亏分布'"
              style="height: 380px"
              :value="item.pieData"
              :text="item.pieText"
            ></chart-pie>
            <chart-bar
              v-else-if="item.pieShow && item.pieText == '盈亏分布'"
              style="height: 380px"
              :value="item.pieData"
              :text="item.pieText"
            />
          </Card>
        </i-col>
      </Row>
    </Card>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import Tables from '_c/tables'
import Example from '@/view/single-page/home/example.vue'
import {
  getPlayerInfoData,
  getHomeGameData,
  getSelectGames,
  getHomeData
} from '@/api/data'
import { ChartPie, ChartBar } from '_c/charts'
export default {
  name: 'players-record',
  components: {
    Tables,
    Example,
    ChartPie,
    ChartBar
  },
  props: ['id'],
  data () {
    return {
      // 是否刷新圆饼图
      isreloadPie: true,
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
            return params.row.state <= 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        }
      ],
      userInfo: [],
      timeVal: [1, 1, 1, 1, 1, 1],
      timeList: [
        { label: '今日', value: 1 },
        { label: '昨日', value: 2 },
        { label: '本周', value: 3 },
        { label: '上周', value: 4 },
        { label: '本月', value: 5 },
        { label: '上月', value: 6 }
      ],
      startTime: '',
      endTime: '',
      setExample: [
        {
          barText: '玩家局数详情',
          barData: {
            sum: []
          },
          show: true,
          pieShow: true,
          legend: ['总局数'],
          columns: 1,
          type: 'bar',
          pieData: [],
          pieText: '玩家局数分布'
        },
        {
          barText: '有效下注详情',
          show: true,
          pieShow: true,
          columns: 1,
          legend: ['总点数'],
          barData: {
            sum: []
          },
          type: 'line',
          pieData: [],
          pieText: '有效下注分布'
        },
        {
          barText: '盈亏分数详情',
          show: true,
          pieShow: true,
          legend: ['总点数'],
          columns: 1,
          type: 'line',
          barData: {
            sum: []
          },
          pieData: [],
          pieText: '盈亏分布'
        }
      ],
      tableData: [],
      gameSelectList: [],
      gameVal: [0, 0, 0]
    }
  },
  methods: {
    ...mapActions(['handleLogOut']),
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    },
    changeTimeData (type, time, gameId, isreloadPie) {
      this.isreloadPie = isreloadPie
      this.searchIndex(type, time, gameId)
      this.searchGameData(type, time)
    },
    changeNum (num) {
      return Number(String(num).replace(/\,/g, ''))
    },
    /**
     * 搜索图表：折线图
     */
    searchIndex (type, time, gameId) {
      let Data = [
        { timeType: time },
        { dataType: type },
        { gameId: gameId },
        { userId: this.id }
      ]
      for (let index = 0; index < 3; index++) {
        if (type - 3 == index) {
          getHomeData(Data)
            .then((res) => {
              this.setExample[type - 3].show = false
              this.setExample[type - 3].columns = time

              this.setExample[type - 3].barData.sum = Object.assign(
                res.data.data.map((item) => {
                  if (
                    time == 1 &&
                    type != 3 &&
                    Number(item.key.split(':')[0]) > new Date().getHours()
                  ) {

                  } else {
                    return this.changeNum(item.number)
                  }
                })
              )

              if (this.setExample[type - 3].barText == '盈亏分数详情') {
                this.setExample[type - 3].barData.sum = this.setExample[
                  type - 3
                ].barData.sum.map((x) => -x)
              }
              this.$nextTick(() => {
                this.setExample[type - 3].show = true
              })
            })
            .catch((err) => {})
        }
      }
    },
    searchGameData (type, time) {
      let Data = [{ timeType: time }, { dataType: type }, { userId: this.id }]
      getHomeGameData(Data)
        .then((res) => {
          if (res.data.code == 200) {
            let option = []

            /**
             * 切换游戏类型不刷新圆饼
             */
            if (this.isreloadPie) {
              this.setExample[type - 3].pieShow = false
              this.setExample[type - 3].pieData = res.data.data
                .map((item) => {
                  if (this.setExample[type - 3].barText == '盈亏分数详情') {
                    // item.profitLoss = -Number(item.profitLoss);
                  }
                  option = this.changeNum(
                    {
                      1: item.gameNumber,
                      2: item.effectiveBets,
                      3: item.profitLoss
                    }[type - 2]
                  )
                  if (option > 0 && type != 5) {
                    return {
                      value: option,
                      name: item.name
                    }
                  } else if (type == 5) {
                    return { [item.name]: option }
                  }
                })
                .filter((item) => item)
              this.$nextTick(() => {
                this.setExample[type - 3].pieShow = true
              })
            }
          } else {
            this.$Message.error(res.data.code, res.data.msg)
          }
        })
        .catch((err) => {})
    }
  },
  mounted () {
    this.$nextTick(() => {
      for (let i = 3; i <= 5; i++) {
        this.searchIndex(i, 1, 0)
        this.searchGameData(i, 1, 0)
      }
    })
    let data = {
      id: this.id,
      agentId: sessionStorage.getItem('agentVal')
    }
    getPlayerInfoData(data).then((res) => {
      this.userInfo = []
      this.userInfo.push(res.data.data)
    })
    getSelectGames(sessionStorage.getItem('agentVal')).then((res) => {
      this.gameSelectList.push({ id: 0, name: '全部' }, ...res.data.data)
    })
  }
}
</script>

<style></style>
