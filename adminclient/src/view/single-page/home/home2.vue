<template>
<!-- 这是新改版的 看板页面 接口还没有弄出来 2020-10-14-->
  <div>
    <div style="display: flex;">
      <h2 style="margin-bottom: 20px;">
      游戏总数据
    </h2>
     <div class="inlineForm" style="margin-left: 10px;">
        <div>站点选择：</div>
        <div>
          <Select v-model="site" @on-change="setSite" style="width:200px" filterable  filter-by-label>
            <Option v-for="item in siteOption" :value="item.id" :key="item.id" :label="item.name">{{ item.name }}</Option>
          </Select>
        </div>
        <div style="padding-left:20px;">代理选择：</div>
        <div>
          <Select v-model="agent" style="width:200px" filterable  filter-by-label>
            <Option v-for="item in agentOption" :value="item.id" :key="item.id" :label="item.name">{{ item.name }}</Option>
          </Select>
        </div>
      </div>
      <span style="margin-left:35px;font-weight: normal;">
        <RadioGroup
          name="time-radio"
          type="button"
          v-model="timeVal"
          size="small"
          @on-change="changeTimeData"
        >
          <Radio
            v-for="items in timeList"
            :key="items.label"
            :label="items.value"
            class="radio-style"
            >{{ items.label }}
          </Radio>
        </RadioGroup>
        <span>时间：</span>
        <DatePicker v-model="date" type="month" placeholder="请选择"></DatePicker>
      </span>
      <Button @click="selectIndex" style="margin-left:20px;" type="primary">搜索</Button>
    </div>
      <Row :gutter="20" style="font-size: 16px">
        <i-col
          :xs="12"
          :md="8"
          :lg="4"
          v-for="(infor, i) in inforCardData"
          :key="`infor-${i}`"
          style="padding-bottom: 10px;"
          v-if="infor.show"
        >
          <Card >
            <div style="padding:6px 0px">
              <p class="count-parent">
                <Icon :type="infor.icon" :color="infor.color" size="24" />
                {{ timeList[timeVal - 1].label }}{{ infor.title }}
                <div v-if="infor.key==='killNumber'" style="font-size: 26px">{{infor.count}}</div>
                <count-to
                  :end="infor.count"
                  separator
                  :decimals="infor.decimals"
                  count-class="count-style"
                  suffix="%"
                  v-if="infor.key!=='killNumber'"
                >
                  <span
                    v-if="infor.key == 'odds' || infor.key == 'netClaims'"
                    style="font-size:22px;"
                    slot="right"
                    >&nbsp;%</span
                  >
                </count-to>
              </p>
            </div>
            <!--  -->
            <template
              v-if="
                infor.subTitle != '本月日均玩家净赔付率' &&
                  infor.subTitle != '本月日均玩家赔付率'
              "
            >
              <p class="sub-count-parent">
                  <span
                    v-if="infor.key == 'odds' || infor.key == 'netClaims'"
                    style="font-size:16px;"
                    slot="right"
                    >&nbsp;%</span
                  >
                </count-to>
              </p>
            </template>
            <div v-else style="height: 21px;"></div>
          </Card>
        </i-col>
      </Row>

      <Row :gutter="20" v-for="item in chartData" :key="item.id">
        <i-col :md="24" :lg="16" style="margin-bottom: 20px;">
          <Card shadow>
            <example
              v-if="item.show"
              style="height: 380px;"
              :type="item.type"
              :legend="item.legend"
              :columns="item.columns"
              :barData="item.barData"
              :text="item.barText"
              :color="item.color"
            />
          </Card>
        </i-col>
        <i-col :md="24" :lg="8" style="margin-bottom: 20px;">
          <Card>
            <chart-pie
              v-if="item.pieShow && item.pieText != '盈亏分布'"
              style="height: 380px;"
              :value="item.pieData"
              :text="item.pieText"
            ></chart-pie>
            <chart-bar
              v-else-if="item.pieShow && item.pieText == '盈亏分布'"
              style="height: 380px;"
              :value="item.pieData"
              :text="item.pieText"
            />
          </Card>
        </i-col>
      </Row>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import Tables from '_c/tables'
import {
  getHomeData,
  getHomeGameData,
  getHomeGameAvgData
} from '@/api/data'
import CountTo from '_c/count-to'
import { ChartPie, ChartBar } from '_c/charts'
import Example from './example.vue'
import { getDate } from '@/libs/tools'
export default {
  name: 'home',
  components: {
    Tables,
    CountTo,
    ChartPie,
    ChartBar,
    Example
  },
  data () {
    return {
      timeVal: 1,

      timeList: [
        // {
        //   label: "今日",
        //   value: 1,
        // },
        // {
        //   label: "昨日",
        //   value: 2,
        // },
        {
          label: '本周',
          value: 3
        },
        {
          label: '上周',
          value: 4
        },
        {
          label: '本月',
          value: 5
        },
        {
          label: '上月',
          value: 6
        }
      ],
      inforCardData: [
        {
          title: '杀数', // 杀数前端计算 盈亏/有效下注 保留三位消暑
          key: 'killNumber',
          icon: 'md-person-add',
          count: 0,
          decimals: 0,
          color: '#444444',
          subTitle: '本月日均杀数',
          subCount: 0,
          show: true
        },
        {
          title: '新增有效下注',
          key: 'bets',
          icon: 'logo-usd',
          count: 0,
          decimals: 2,
          color: '#19be6b',
          subTitle: '本月日均有效下注',
          subCount: 0,
          show: true
        },
        {
          title: '盈亏',
          key: 'profitLoss',
          icon: 'md-ionic',
          count: 0,
          decimals: 2,
          color: '#ed3f14',
          subTitle: '本月日均盈亏',
          subCount: 0,
          show: true
        },
        {
          title: '注单数量',
          key: 'aaa',
          icon: 'md-ionic',
          count: 0,
          decimals: 0,
          color: '#ed3f14',
          subTitle: '本月日均注单数量',
          subCount: 0,
          show: true
        },
        {
          title: '登录玩家',
          key: 'bbb',
          icon: 'md-ionic',
          count: 0,
          decimals: 0,
          color: '#ed3f14',
          subTitle: '本月日均登录玩家',
          subCount: 0,
          show: true
        },
        {
          title: '活跃玩家',
          key: 'ccc',
          icon: 'md-ionic',
          count: 0,
          decimals: 0,
          color: '#ed3f14',
          subTitle: '本月日均活跃玩家',
          subCount: 0,
          show: true
        }
        // {
        //   title: "新增玩家",
        //   key: "userNumber",
        //   icon: "md-person-add",
        //   count: 0,
        //   decimals: 0,
        //   color: "#444444",
        //   subTitle: "本月日均新增玩家",
        //   subCount: 0,
        //   show: false,
        // },
        // {
        //   title: "抽水分数",
        //   key: "pump",
        //   icon: "logo-yen",
        //   count: 0,
        //   decimals: 2,
        //   color: "#ff9900",
        //   subTitle: "本月日均抽水分数",
        //   subCount: 0,
        //   show: false,
        // },
        // {
        //   title: "玩家赔付率",
        //   key: "odds",
        //   icon: "ios-ribbon",
        //   count: 0,
        //   decimals: 2,
        //   color: "#996600",
        //   subTitle: "本月日均玩家赔付率",
        //   subCount: 0,
        //   show: false,
        // },
        // {
        //   title: "玩家赔付率",
        //   key: "netClaims",
        //   icon: "ios-cash",
        //   count: 0,
        //   decimals: 2,
        //   color: "#66aa11",
        //   subTitle: "本月日均玩家净赔付率",
        //   subCount: 0,
        //   show: false,
        // },
      ],

      tableLabel: [
        {
          label: '游戏平台',
          value: '',
          key: 'platFrom',
          option: []
        },
        {
          label: '游戏名',
          key: 'gameId',
          value: '',
          option: []
        }
      ],
      tableDate: {
        time: '',
        timeType: ''
      },
      columnsTable: [
        {
          title: '排名',
          type: 'index',
          align: 'center',
          width: 90
        },
        {
          title: '模式',
          key: 'platFromName',
          align: 'center',
          width: 120
        },
        {
          title: '游戏',
          key: 'gameName',
          align: 'center'
        },
        {
          title: '总消耗点数',
          key: 'profitLossAvg',
          align: 'center',
          sortable: true,
          sortMethod: (a, b, type) => {
            if (type == 'asc') {
              return Number(this.changeNum(a)) > Number(this.changeNum(b))
                ? 1
                : -1
            } else {
              return Number(this.changeNum(a)) > Number(this.changeNum(b))
                ? -1
                : 1
            }
          }
        },
        {
          title: '玩家人数',
          key: 'userNumber',
          align: 'center',
          sortable: true
        },
        {
          title: '玩家平均消耗点数',
          key: 'profUserAvg',
          align: 'center',
          width: 180,
          sortable: true,
          sortMethod: (a, b, type) => {
            if (type == 'asc') {
              return Number(this.changeNum(a)) > Number(this.changeNum(b))
                ? 1
                : -1
            } else {
              return Number(this.changeNum(a)) > Number(this.changeNum(b))
                ? -1
                : 1
            }
          }
        },
        {
          title: '总局数',
          key: 'gameNumber',
          align: 'center',
          sortable: true
        }
      ],
      tableData: [],
      chartData: [
        {
          barText: '杀数',
          show: true,
          pieShow: true,
          legend: ['杀数'],
          columns: 1,
          barData: {
            sum: []
          },
          sort: 1,
          type: 'bar',
          pieSet: 0,
          color: '#169BD5',
          pieData: [],
          pieText: '杀数排行'
        },
        {
          barText: '有效下注详情',
          show: true,
          pieShow: true,
          columns: 1,
          barData: {
            sum: [],
            bet: [],
            vs: [],
            fish: []
          },
          sort: 2,
          type: 'bar',
          color: '#169BD5',
          pieData: [],
          pieText: '有效下注分布'
        },
        {
          barText: '盈亏',
          show: true,
          pieShow: true,
          columns: 1,
          sort: 3,
          type: 'line',
          color: '#169BD5',
          barData: {
            sum: [],
            bet: [],
            vs: [],
            fish: []
          },
          pieData: [],
          pieText: '盈亏分布'
        },
        {
          barText: '注单数量',
          show: false,
          pieShow: true,
          legend: ['注单数量'],
          columns: 1,
          sort: 4,
          barData: {
            sum: []
          },
          type: 'bar',
          pieSet: 0,
          pieData: [],
          color: '#169BD5',
          pieText: '注单数量分布'
        },
        {
          barText: '登录玩家数量',
          show: false,
          pieShow: true,
          legend: ['登录玩家'],
          columns: 1,
          sort: 5,
          barData: {
            sum: []
          },
          type: 'bar',
          pieSet: 0,
          color: '#169BD5',
          pieData: [],
          pieText: '登录玩家分布'
        },
        {
          barText: '活跃玩家详情',
          show: false,
          pieShow: true,
          legend: ['活跃玩家'],
          columns: 1,
          sort: 6,
          barData: {
            sum: []
          },
          type: 'bar',
          pieSet: 0,
          color: '#169BD5',
          pieData: [],
          pieText: '活跃玩家分布'
        }
      ],
      gameSelectList: [],
      gameVal: [0, 0, 0, 0, 0, 0],
      spinShow: false,

      site: null, // 站点
      siteOption: [],
      agent: null, // 代理
      agentOption: [],
      // 时间区间
      date: null,

      /**
       * 查询参数
       */
      params: {
        time: 1,
        timeType: null,
        agentId: null,
        webId: null,
        gameId: null
      }
    }
  },
  methods: {
    ...mapActions(['handleLogOut']),

    clearTimeDatess () {
      // 清除快捷日期
      this.tableDate.timeType = null
    },

    handleSearch (param) {
      // 点击快捷日期选择 清空日期输入框
      if (param && param.clearDate) {
        this.$refs.isclearDate.handleClear()
      }

      // 清除快捷日期
      if (param && param.clearShortcut) {
        this.tableDate.timeType = null
      }

      this.spinShow = true
      let Data = [
        {
          time: getDate(this.tableDate.time)
        },
        {
          timeType: this.tableDate.timeType
        },
        {
          agentId: this.agent === 9999999 ? null : this.agent
        },
        {
          webId: this.site
        }
      ]
      this.params = {
        time: getDate(this.tableDate.time),
        timeType: this.tableDate.timeType,
        agentId: this.agent === 9999999 ? null : this.agent,
        webId: this.site
      }

      if (this.tableDate.timeType) {
        Data = [
          {
            timeType: this.tableDate.timeType,
            webId: this.site,
            agentId: this.agent === 9999999 ? null : this.agent
          }
        ]
      }

      Data.push(
        ...this.tableLabel.map((item) => {
          return {
            [item.key]: item.value
          }
        })
      )
      getHomeGameAvgData(Data).then((res) => {
        res.data.data.map(
          (item) => (item.gameNumber = Number(item.gameNumber))
        )
        this.spinShow = false
        this.tableData = []
        this.tableData.push(...res.data.data)
      })
    },
    changeTimeData () {
      this.params = {
        timeType: this.timeVal,
        agentId: this.agent,
        webId: this.site
      }
      this.selectIndex()
    },
    changeNum (num) {
      return Number(String(num).replace(/\,/g, ''))
    },

    /**
     * 搜索图表：折线图
     */
    searchIndex (type, time, gameId) {
      // 查询杀数 0
      // 查询有效下注详情 1
      // 查询盈亏 2
      // 查询注单数量 3
      // 查询登录玩家数量 4
      // 查询活跃玩家数量 5
      // key=>value key 是chartData的下标 value 是dataType参数
      let obj = { 1: 0, 4: 1, 5: 2, 99: 3, 98: 4, 2: 5 }
      let Data = [
        {
          timeType: time
        },
        {
          dataType: type
        },
        {
          gameId: gameId
        },
        {
          webId: this.site
        },
        {
          agentId: this.agent === 9999999 ? null : this.agent
        }
      ]
      sessionStorage.setItem('siteVal', this.site)
      sessionStorage.setItem(
        'agentVal',
        this.agent === 9999999 ? '' : this.agent
      )
      /**
       * 游戏总数据（非图表）
       */
      if (type == 1) {
        getHomeData(Data)
          .then((res) => {
            this.inforCardData.forEach((item, index) => {
              if (item.key == Object.keys(res.data.data.sumData)[index]) {
                item.count = this.changeNum(res.data.data.sumData[item.key])
                let subCount = this.changeNum(
                  res.data.data.thisMonthAvg[item.key]
                )
                item.subCount = subCount || 0
              }
            })
            this.inforCardData[0].count =
              (this.inforCardData[2].count / this.inforCardData[1].count &&
                (
                  this.inforCardData[2].count / this.inforCardData[1].count
                ).toFixed(3) * 1) ||
              0
          })
          .catch((err) => {})
      } else {
        /**
         * 折线图表，2-5 为列表Type
         */
        getHomeData(Data)
          .then((res) => {
            this.chartData[obj[type]].show = false
            this.chartData[obj[type]].columns = time
            this.chartData[obj[type]].barData.sum = Object.assign(
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
            this.chartData[obj[type]].pieShow = false
            this.chartData[obj[type]].pieData = res.data.data
              .map((item) => {
                return {
                  value: item.number,
                  name: item.key
                }
              })
              .filter((item) => item)
            this.$nextTick(() => {
              this.chartData[obj[type]].pieShow = true
            })
            this.$nextTick(() => {
              this.chartData[obj[type]].show = true
            })
          })
          .catch((err) => {})
      }
    },
    searchGameData (type, time, gameId) {
      // 查询杀数 0
      // 查询有效下注详情 1
      // 查询盈亏 2
      // 查询注单数量 3
      // 查询登录玩家数量 4
      // 查询活跃玩家数量 5
      // key=>value key 是chartData的下标 value 是dataType参数
      let obj = { 1: 0, 4: 1, 5: 2, 99: 3, 98: 4, 2: 5 }
      let Data = [
        {
          timeType: time,
          webId: this.site,
          agentId: this.agent === 9999999 ? null : this.agent,
          dataType: type
        }
      ]
      if (type >= 3) {
        getHomeGameData(Data)
          .then((res) => {
            if (res.data.code == 200) {
              let option = []
              this.chartData[obj[type]].pieShow = false
              this.chartData[obj[type]].pieData = res.data.data
                .map((item) => {
                  item.profitLoss = String(-Number(item.profitLoss))
                  option = this.changeNum(
                    {
                      1: item.gameNumber,
                      2: item.effectiveBets,
                      3: item.profitLoss
                    }[obj[type]]
                  )
                  if (option > 0 && type != 5) {
                    return {
                      value: option,
                      name: item.name
                    }
                  } else if (type == 5) {
                    return {
                      [item.name]: option
                    }
                  }
                })
                .filter((item) => item)

              this.$nextTick(() => {
                this.chartData[obj[type]].pieShow = true
              })
            } else {
              if (res.data.data) { this.$Message.error(res.data.data.code, res.data.msg) }
            }
          })
          .catch((err) => {})
      }
    },
    /**
     * set sit
     */
    setSite (val) {
      this.site = val
      if (val > 0) {
        sessionStorage.setItem('siteVal', val)
        this.agentOption = this.siteOption.find(
          (site) => site.id == val
        ).agentList
        this.agentOption.map((item) => {
          item.label = item.name
        })
      }
    },

    /**
     * 查询
     */
    async selectIndex () {
      this.spinShow = false
      // 查询总数据 非图表
      await this.searchIndex(1, this.params.timeType)
      // 查询杀数 0
      // 查询有效下注详情 1
      await this.searchIndex(4, this.params.timeType, this.params.gameId)
      // 查询盈亏 2
      await this.searchIndex(5, this.params.timeType, this.params.gameId)
      await this.searchGameData(5, this.params.timeType, this.params.gameId)
      // 查询注单数量 3
      // 查询登录玩家数量 4
      // 查询活跃玩家数量 5
      await this.searchIndex(2, this.params.timeType, this.params.gameId)
    }
  },
  async mounted () {
    let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
    let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
    this.siteOption = siteOption
    let agent = sessionStorage.getItem('agentVal')
    this.site = sid * 1
    // 把选择的站点赋值到页面选中
    siteOption &&
      siteOption.map((item, index) => {
        if (item.id == sid) {
          this.agentOption = item.agentList
          this.agent = agent * 1
          this.agentOption.map((item) => {
            item.label = item.name
          })
        }
      })
    this.selectIndex()
    let temptimer = setInterval(() => {
      if (sessionStorage.getItem('agentVal')) {
        this.tableLabel.map((item) => {
          if (item.key == 'gameId') {
            // getSelectGames(sessionStorage.getItem("agentVal")).then((res) => {
            item.option.push(
              {
                id: 0,
                name: '全部'
              },
              ...JSON.parse(sessionStorage.getItem('games'))
            )
            this.gameSelectList.push(
              {
                id: 0,
                name: '全部'
              },
              ...JSON.parse(sessionStorage.getItem('games'))
            )
            // });
          }
          if (item.key == 'platFrom') {
            // getClassList(sessionStorage.getItem("agentVal")).then((res) => {
            //   item.option.push(...res.data.data);
            // });
          }
        })
        clearInterval(temptimer)
      }
    }, 200)
  }
}
</script>

<style lang="less" scoped>
.count-parent {
  text-align: left;
  line-height: 25px;

  .count-style {
    display: inline-block;
    margin-top: 5px;
    margin-left: 27px;
    font-size: 30px;
  }
}

.sub-count-parent {
  display: inline-block;
  text-align: center;
  width: 100%;

  .sub-count-style {
    font-size: 20px;
    margin-top: 5px;
    display: inline-block;
  }
}
.inlineForm {
  display: flex;
  align-items: center;
  // padding: 10px 0;
  // margin-top: -20px;
}
.count-to-wrapper {
  font-size: 26px;
}
</style>
