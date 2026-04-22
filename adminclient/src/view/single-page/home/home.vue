<template>
  <div>
    <div style="display: flex;">
      <h2 style="margin-bottom: 20px;">
      游戏总数据
      <span style="margin-left:35px;font-weight: normal;">
        <RadioGroup
          name="time-radio"
          type="button"
          v-model="timeVal[0]"
          size="small"
          @on-change="changeTimeData(1, timeVal[0])"
        >
          <Radio
            v-for="items in timeList"
            :key="items.label"
            :label="items.value"
            class="radio-style"
            >{{ items.label }}
          </Radio>
        </RadioGroup>
      </span>
    </h2>
    <div class="inlineForm">
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
        <Button @click="selectIndex" style="margin-left:20px;" type="primary">搜索</Button>
      </div>
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
              {{ timeList[timeVal[0] - 1].label }}{{ infor.title }}
              <count-to
                :end="infor.count"
                separator
                :decimals="infor.decimals"
                count-class="count-style"
                suffix="%"
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

    <Row :gutter="20" v-for="(item, index) in setExample" :key="item.id">
      <i-col :md="24" :lg="16" style="margin-bottom: 20px;">
        <Card shadow>
          <span style="position: absolute;left: 145px;top: 19px;z-index:9">
            <RadioGroup
              type="button"
              :name="'time-radio-' + index"
              v-model="timeVal[index + 1]"
              size="small"
              @on-change="
                changeTimeData(index + 2, timeVal[index + 1], gameVal[index])
              "
            >
              <Radio
                v-for="items in timeList"
                :key="items.value"
                :label="items.value"
                class="radio-style"
              >
                {{ items.label }}</Radio
              >
            </RadioGroup>
          </span>

          <span style="position: absolute;left: 546px;top: 15px;z-index:9">
            <Select
              v-model="gameVal[index]"
              :name="'game-radio-' + index"
              @on-change="
                changeTimeData(index + 2, timeVal[index + 1], gameVal[index])
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
            style="height: 380px;"
            :type="item.type"
            :legend="item.legend"
            :columns="item.columns"
            :barData="item.barData"
            :text="item.barText"
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

    <!-- <Row :gutter="20" style="margin-top: 10px;">
      <Card>
        <ul style="margin-bottom:5px">
          <li class="label-style" v-for="item in tableLabel" :key="item.key">
            <label>{{ item.label }}</label>
            <Select v-model="item.value" style="width: 180px">
              <Option
                v-for="items in item.option"
                :value="items.id"
                :key="items.id"
                >{{ items.name }}</Option
              >
            </Select>
          </li>
          <Button
            type="primary"
            shape="circle"
            icon="ios-search"
            @click="handleSearch({ clearShortcut: true })"
          ></Button>
          <label style="margin-left:20px;margin-right:5px">日期选择</label>
          <DatePicker
            @on-open-change="clearTimeDatess"
            ref="isclearDate"
            type="date"
            v-model="tableDate.time"
            placeholder="请选择日期"
            size="small"
          >
          </DatePicker>
          <span style="margin-left:10px;font-weight: normal;">
            <RadioGroup
              type="button"
              v-model="tableDate.timeType"
              size="small"
              border="true"
              @on-change="handleSearch({ clearDate: true })"
            >
              <Radio
                v-for="items in timeList"
                :key="items.label"
                :label="items.value"
                class="radio-style"
              >
                {{ items.label }}</Radio
              >
            </RadioGroup>
          </span>
        </ul>
        <tables :columns="columnsTable" v-model="tableData" />
        <Spin fix v-if="spinShow">
          <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
          <div>数据加载中</div>
        </Spin>
      </Card>
    </Row> -->
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import Tables from '_c/tables'
import {
  getHomeData,
  getHomeGameData,
  getHomeGameAvgData,
  getLinkageList,
  // getClassList,
  getSelectGames
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
      timeVal: [1, 1, 1, 1, 1, 1],

      timeList: [
        {
          label: '今日',
          value: 1
        },
        {
          label: '昨日',
          value: 2
        },
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
          title: '新增玩家',
          key: 'userNumber',
          icon: 'md-person-add',
          count: 0,
          decimals: 0,
          color: '#444444',
          subTitle: '本月日均新增玩家',
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
          title: '抽水分数',
          key: 'pump',
          icon: 'logo-yen',
          count: 0,
          decimals: 2,
          color: '#ff9900',
          subTitle: '本月日均抽水分数',
          subCount: 0,
          show: true
        },
        {
          title: '玩家赔付率',
          key: 'odds',
          icon: 'ios-ribbon',
          count: 0,
          decimals: 2,
          color: '#996600',
          subTitle: '本月日均玩家赔付率',
          subCount: 0,
          show: false
        },
        {
          title: '玩家赔付率',
          key: 'netClaims',
          icon: 'ios-cash',
          count: 0,
          decimals: 2,
          color: '#66aa11',
          subTitle: '本月日均玩家净赔付率',
          subCount: 0,
          show: true
        }
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

      setExample: [
        {
          barText: '活跃玩家详情',
          show: true,
          pieShow: true,
          legend: ['活跃玩家'],
          columns: 1,
          barData: {
            sum: []
          },
          type: 'line',
          pieSet: 0,
          pieData: [],
          pieText: '活跃玩家分布'
        },
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
          barData: {
            sum: [],
            bet: [],
            vs: [],
            fish: []
          },
          type: 'line',
          pieData: [],
          pieText: '有效下注分布'
        },
        {
          barText: '盈亏分数详情',
          show: true,
          pieShow: true,
          columns: 1,
          type: 'line',
          barData: {
            sum: [],
            bet: [],
            vs: [],
            fish: []
          },
          pieData: [],
          pieText: '盈亏分布'
        }
      ],

      gameSelectList: [],
      gameVal: [0, 0, 0, 0, 0, 0],
      spinShow: false,

      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],

      /**
       * 查询参数
       */
      params: {
        time: null,
        timeType: null,
        agentId: null,
        webId: null,
        gameId: null
      }
    }
  },
  methods: {
    ...mapActions(['handleLogOut']),
    searchAll () {
      for (let i = 1; i <= 5; i++) {
        this.searchIndex(i, 1)
        this.searchGameData(i, 1)
      }
      this.timeVal = [1, 1, 1, 1, 1, 1]
    },

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
    changeTimeData (type, time, gameId) {
      this.params = {
        timeType: time,
        dataType: type,
        agentId: this.agent,
        webId: this.site,
        gameId: gameId
      }
      this.searchIndex(type, time)
      this.searchIndex(type, time, gameId)
      this.searchGameData(type, time, gameId)
    },
    changeNum (num) {
      return Number(String(num).replace(/\,/g, ''))
    },

    /**
     * 搜索图表：折线图
     */
    searchIndex (type, time, gameId) {
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
          })
          .catch((err) => {})
      } else {
        /**
         * 折线图表，2-5 为列表Type
         */
        for (let index = 2; index <= 5; index++) {
          if (type == index) {
            getHomeData(Data)
              .then((res) => {
                this.setExample[type - 2].show = false
                this.setExample[type - 2].columns = time
                this.setExample[type - 2].barData.sum = Object.assign(
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
                if (type - 2 == 0) {
                  this.setExample[0].pieShow = false
                  this.setExample[0].pieData = res.data.data
                    .map((item) => {
                      return {
                        value: item.number,
                        name: item.key
                      }
                    })
                    .filter((item) => item)
                  this.$nextTick(() => {
                    this.setExample[0].pieShow = true
                  })
                }
                this.$nextTick(() => {
                  this.setExample[type - 2].show = true
                })
              })
              .catch((err) => {})
          }
        }
      }
    },
    searchGameData (type, time, gameId) {
      let Data = [
        {
          timeType: time,
          webId: this.site,
          agentId: this.agent === 9999999 ? null : this.agent
        }
      ]
      if (type >= 3) {
        getHomeGameData(Data)
          .then((res) => {
            if (res.data.code == 200) {
              let option = []
              this.setExample[type - 2].pieShow = false
              this.setExample[type - 2].pieData = res.data.data
                .map((item) => {
                  item.profitLoss = String(-Number(item.profitLoss))
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
                    return {
                      [item.name]: option
                    }
                  }
                })
                .filter((item) => item)

              this.$nextTick(() => {
                this.setExample[type - 2].pieShow = true
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
        if (!this.agentOption.find((x) => x.id == 9999999)) {
          this.agentOption.unshift({
            name: '全部',
            id: 9999999
          })
        }
        this.agentOption.map((item) => {
          item.label = item.name
        })
      }
    },

    /**
     * 查询
     */
    async selectIndex () {
      this.searchIndex(1, this.params.timeType)
      // this.searchIndex(2, this.params.timeType, this.params.gameId)
    }
  },
  async mounted () {
    getLinkageList()
      .then((res) => {
        window.windowData_getLinkageList = res.data.data
        let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
        let siteOption = JSON.parse(
          sessionStorage.getItem('siteOption') || '[]'
        ) // 获取当前session存储的站点列表数据
        this.siteOption = siteOption
        let agent = sessionStorage.getItem('agentVal')
        this.site = sid * 1
        // 把选择的站点赋值到页面选中
        siteOption &&
          siteOption.map((item, index) => {
            if (item.id == sid) {
              this.agentOption = item.agentList
              // this.agentOption.unshift({
              //   name: "全部",
              //   id: 9999999,
              // });
              this.agent = agent * 1
              this.agentOption.map((item) => {
                item.label = item.name
              })
            }
          })
        // this.siteOption = res.data.data;
        // this.site = res.data.data[0].id;
        // res.data.data[0].agentList.unshift({
        //   name: "全部",
        //   id: 9999999,
        // });
        // this.agentOption = res.data.data[0].agentList;
        // this.agent = 9999999;

        for (let i = 1; i <= 5; i++) {
          this.searchIndex(i, 1, 0)
          this.searchGameData(i, 1, 0)
        }

        this.handleSearch()
        this.siteOption.map((item) => {
          item.label = item.name
        })

        let temptimer = setInterval(() => {
          if (sessionStorage.getItem('agentVal')) {
            this.tableLabel.map((item) => {
              if (item.key == 'gameId') {
                getSelectGames(sessionStorage.getItem('agentVal')).then(
                  (res) => {
                    item.option.push(
                      {
                        id: 0,
                        name: '全部'
                      },
                      ...res.data.data
                    )
                    this.gameSelectList.push(
                      {
                        id: 0,
                        name: '全部'
                      },
                      ...res.data.data
                    )
                  }
                )
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
      })
      .catch((err) => {})
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
.count-to-wrapper{
    font-size: 26px;
}
</style>
