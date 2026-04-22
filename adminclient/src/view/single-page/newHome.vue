<template>
  <div>
    <Card>
      <div class="inlineForm">
        <div>时间选择：</div>
        <div style="display: flex; align-items: center">
          <DatePicker
            v-model="startDate"
            :options="startDateRestrict"
            placeholder="起始"
            style="width: 120px"
            :clearable="false"
          ></DatePicker>
          <div style="padding: 0 10px">—</div>
          <DatePicker
            v-model="endDate"
            :options="endDateRestrict"
            placeholder="结束"
            style="width: 120px"
            :clearable="false"
          ></DatePicker>
        </div>
        <div style="margin-left: 10px">站点选择：</div>
        <div>
          <Select
            v-model="site"
            @on-change="siteChanged"
            style="width: 200px"
            filterable
            filter-by-label
          >
            <Option
              v-for="item in siteOption"
              :value="item.id"
              :key="item.id"
              :label="item.name"
              >{{ item.name }}
            </Option>
          </Select>
        </div>
        <div style="padding-left: 20px">代理选择：</div>
        <div>
          <Select
            v-model="agent"
            @on-change="agentSearch"
            style="width: 200px"
            filterable
            filter-by-label
          >
            <Option
              v-for="item in agentOption"
              :value="item.id"
              :key="item.id"
              :label="item.name"
              >{{ item.name }}
            </Option>
          </Select>
        </div>
        <Button @click="agentSearch" style="margin-left: 20px" type="primary"
          >搜索</Button
        >
        <div style="font-size: 12px; margin-left: 50px">
          <Button type="success" @click="viewDetail2" style="margin-right: 10px"
            >代理业绩</Button
          >
          <Button type="success" @click="toWeekIndicatro">指标对比</Button>
        </div>
      </div>
    </Card>
    <Card style="margin: 10px 0; padding: 20px 0px 45px 0px">
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
      <div v-if="agent != 9999999" style="font-size: 12px">
        <Button
          type="primary"
          style="margin-right: 20px"
          @click="viewGame(agentData[0])"
          >查询游戏</Button
        >
        <Button type="primary" @click="viewDetail(agentData[0])"
          >查询明细</Button
        >
      </div>
      <Row class="countBox" span="24">
        <Col class="countItem" span="2.5">
          <span>{{ parseInt(agentData[0].rangeRegUser) || 0 }}</span>
          <br />
          <b>新增会员</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span>{{ parseInt(agentData[0].totalRegUser) || 0 }}</span>
          <br />
          <b>总会员</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span>{{
            parseFloat(agentData[0].chipsTotal).toFixed(2) || 0
          }}</span>
          <br />
          <b>总打码</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span
            v-if="(parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal)) >= 0"
            >{{ (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal)).toFixed(2) }}</span
          >
          <span
            v-if="(parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal)) < 0"
            style="color: #ffa805"
            >{{ (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal)).toFixed(2) }}</span
          >
          <br />
          <b>盈利额</b>
        </Col>
        <!-- <Col class="countItem">
          <span>-</span>
          <br />
          <b>税收</b>
        </Col> -->
        <Col class="countItem" span="2.5">
          <span>{{ parseFloat(agentData[0].score_up).toFixed(2) || 0 }}</span>
          <br />
          <b>上分总额</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span>{{ parseFloat(agentData[0].score_down).toFixed(2) || 0 }}</span>
          <br />
          <b>下分总额</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span>{{
            parseFloat(agentData[0].revenueTotal).toFixed(2) || 0
          }}</span>
          <br />
          <b>税收</b>
        </Col>
        <Col class="countItem" span="2.5">
          <span v-if="parseFloat(agentData[0].effectiveBetsTotal) == 0">{{
            0
          }}</span>
          <span
            v-if="
              (
                (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal)) /parseFloat(agentData[0].chipsTotal)
              ).toFixed(3) >= 0
            "
            >{{
              (
                (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal))  /parseFloat(agentData[0].chipsTotal)
              ).toFixed(3)
            }}</span
          >
          <span
            v-if="
              (
                (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal))  /parseFloat(agentData[0].chipsTotal)
              ).toFixed(3) < 0
            "
            style="color: #ffa805"
            >{{
              (
                (parseFloat(agentData[0].effectiveBetsTotal)-parseFloat(agentData[0].profitLossTotal))  /parseFloat(agentData[0].chipsTotal)
              ).toFixed(3)
            }}</span
          >
          <br />
          <b>整体杀率</b>
        </Col>
      </Row>
    </Card>
    <Card>
      <div class="inlineForm">
        <div style="margin-left: 10px">游戏选择：</div>
        <div>
          <Select
            v-model="paramgame"
            style="width: 500px"
            filterable
            filter-by-label
          >
            <Option
              v-for="item in gameOptions"
              :value="item.number"
              :key="item.number"
              :label="item.label"
            >
              {{ item.label }}</Option
            >
          </Select>
        </div>
        <div style="margin-left: 20px">日期选择：</div>
        <div>
          <DatePicker
            v-model="betInfoStartDate"
            :options="betInfoStartDateRestrict"
            placeholder="查看日期"
            style="width: 220px"
            :clearable="false"
          ></DatePicker>
        </div>
        <Button
          @click="checkBetInfoByDate"
          style="margin-left: 20px"
          type="primary"
          >搜索</Button
        >
      </div>

      <div class="betInfoChartDom" id="betInfoChart1Dom"></div>
      <div class="betInfoChartDom" id="betInfoChart2Dom"></div>
      <div class="betInfoChartDom" id="betInfoChart3Dom"></div>
      <div class="betInfoChartDom" id="betInfoChart4Dom"></div>
    </Card>
  </div>
</template>

<script>
import { getLinkageList } from '@/api/data'
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import * as dayjs from 'dayjs'
import * as echarts from 'echarts'

export default {
  components: {},
  data () {
    let _this = this
    return {
      betInfoStartDate: dayjs(dayjs().format('YYYY-MM-DD 00:00:00')).toDate(),
      paramgame: null,
      gameOptions: [],
      games: JSON.parse(sessionStorage.getItem('games') || '[]'),
      spinShow: false,
      showBtn: false,
      /**
       * 查询参数
       */
      startDate: dayjs(
        dayjs().startOf('month').format('YYYY-MM-DD 00:00:00')
      ).toDate(),
      betInfoStartDateRestrict: {
        disabledDate (date) {
          if (date.getTime() > dayjs().unix() * 1000) {
            return true
          }

          if (
            date.getTime() <
            dayjs().subtract(3, 'month').startOf('month').unix() * 1000
          ) {
            return true
          }
        }
      },
      startDateRestrict: {
        disabledDate (date) {
          if (date.getTime() > dayjs(_this.endDate).unix() * 1000) {
            return true
          }
          if (date.getTime() < Date.now() - 1000 * 60 * 60 * 24 * 30 * 6) {
            return true
          }
        }
      },
      endDate: dayjs(dayjs().format('YYYY-MM-DD 23:59:59')).toDate(),
      endDateRestrict: {
        disabledDate (date) {
          if (date.getTime() < dayjs(_this.startDate).unix() * 1000) {
            return true
          }
          if (date.getTime() > Date.now()) {
            return true
          }
        }
      },
      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],

      /**
       * 表格配置
       */
      agentColumns: [
        {
          title: '站1点',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '代理',
          align: 'center',
          key: 'nickName'
        },
        {
          align: 'center',
          title: '代理ID',
          key: 'agentId'
        },
        {
          title: '玩家数量',
          align: 'center',
          key: 'userTotal'
        },
        {
          align: 'center',
          title: '注单数量',
          key: 'eNumber'
        },
        {
          align: 'center',
          title: '期间有效下注',
          key: 'eTotal',
          render (h, params) {
            let jsx = <span>{params.row.eTotal.toFixed(2)}</span>
            return jsx
          }
        },
        {
          align: 'center',
          title: '盈亏',
          key: 'pTotal',
          render (h, params) {
            let jsx = (
              <div style="color:red">
                {(params.row.pTotal && params.row.pTotal.toFixed(2)) || 0}
              </div>
            )
            if (Number(params.row.pTotal) > 0) {
              jsx = (
                <div style="color:green">{params.row.pTotal.toFixed(2)}</div>
              )
            }
            return jsx
          }
        },
        {
          title: '新增会员',
          align: 'center',
          key: 'rangeRegUser'
        },
        {
          title: '总会员',
          align: 'center',
          key: 'totalRegUser'
        },
        {
          title: '上分总额',
          align: 'center',
          key: 'score_up',
          render (h, params) {
            return (
              <span>
                {(params.row.score_up && params.row.score_up.toFixed(2)) || 0}
              </span>
            )
          }
        },
        {
          title: '下分总额',
          align: 'center',
          key: 'score_down',
          render (h, params) {
            return (
              <span>
                {(params.row.score_down && params.row.score_down.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '整体杀数',
          align: 'center',
          key: 'kill',
          render (h, params) {
            return (
              <span>
                {(params.row.pTotal / params.row.eTotal &&
                  (params.row.pTotal / params.row.eTotal).toFixed(3)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '游戏',
          align: 'center',
          slot: 'game'
          // render(h, params) {
          //   return h(
          //     "Button",
          //     {
          //       props: {
          //         type: "primary",
          //         size: "small",
          //         to:
          //           "/newHome-agentGamesStatistic/" +
          //           params.row.agentId +
          //           convertTime(),
          //         target: "_blank",
          //       },
          //     },
          //     "查询"
          //   );
          // },
        },
        {
          align: 'center',
          title: '明细',
          slot: 'detail'
          // render(h, params) {
          //   return h(
          //     "Button",
          //     {
          //       props: {
          //         type: "primary",
          //         size: "small",
          //         to: "/newHome-agentOrderStatistic-" + params.row.agentId,
          //         target: "_blank",
          //       },
          //     },
          //     "查询"
          //   );
          // },
        }
      ],

      /**
       * 表格数据
       */
      agentData: [
        {
          rangeRegUser: 0,
          totalRegUser: 0,
          effectiveBetsTotal: 0,
          profitLossTotal: 0,
          score_up: 0,
          score_down: 0,
          revenueTotal: 0,
          chipsTotal: 0
        }
      ]
    }
  },
  methods: {
    async checkBetInfoByDate () {
      if (typeof this.betInfoStartDate === 'string') {
        this.betInfoStartDate = dayjs(this.betInfoStartDate).toDate()
      }
      let currentParams = {
        token: getToken(),
        gameId: this.paramgame == 99999 ? null : this.paramgame,
        startTime: dayjs(this.betInfoStartDate.getTime()).startOf('day').unix(),
        endTime: dayjs(this.betInfoStartDate.getTime()).endOf('day').unix()
      }
      let currentData = await axios.request({
        url: 'v2/govern/userAndGameDataByHour',
        method: 'get',
        params: currentParams
      })

      currentParams.startTime = dayjs(this.betInfoStartDate.getTime())
        .subtract(1, 'day')
        .startOf('day')
        .unix()
      currentParams.endTime = dayjs(this.betInfoStartDate.getTime())
        .subtract(1, 'day')
        .endOf('day')
        .unix()
      let previousData = await axios.request({
        url: 'v2/govern/userAndGameDataByHour',
        method: 'get',
        params: currentParams
      })

      currentData = currentData.data.data
      previousData = previousData.data.data
      let newCurrentData = {
          active: [],
          cnt: [],
          eff: [],
          pro: []
        },
        newPreviousData = {
          active: [],
          cnt: [],
          eff: [],
          pro: []
        }

      for (let index = 0; index < 24; index++) {
        let currentDate =
          dayjs(this.betInfoStartDate.getTime()).startOf('day').unix() +
          3600 * index
        let previousDate =
          dayjs(this.betInfoStartDate.getTime())
            .subtract(1, 'day')
            .startOf('day')
            .unix() +
          3600 * index
        Array.from(['active', 'cnt', 'eff', 'pro']).map((ikey) => {
          if (!currentData[ikey]) return
          let currentDataExist = currentData[ikey].find(
            (x) => {
              if (ikey == 'pro') {
                x.pro = (Number(x.eff) - Number(x.pro)).toFixed(2)
              }
              return x.record_time == currentDate
            }
          )
          let previousDataExist = previousData[ikey].find(
            (x) => {
              if (ikey == 'pro') {
                x.pro = (Number(x.eff) - Number(x.pro)).toFixed(2)
              }
              return x.record_time == previousDate
            }
          )

          if (currentDataExist) {
            currentDataExist[ikey == 'cnt' ? 'cnt' : ikey] = Number(
              currentDataExist[ikey == 'cnt' ? 'cnt' : ikey]
            )
          }
          if (previousDataExist) {
            previousDataExist[ikey == 'cnt' ? 'cnt' : ikey] = Number(
              previousDataExist[ikey == 'cnt' ? 'cnt' : ikey]
            )
          }

          if (ikey == 'pro') {
            if (currentDataExist) {
              currentDataExist['pro'] = Number(
                currentDataExist['pro'].toFixed(2)
              )
            }
            if (previousDataExist) {
              previousDataExist['pro'] = Number(
                previousDataExist['pro'].toFixed(2)
              )
            }
          }
          newCurrentData[ikey].push({
            record_time: currentDate,
            [ikey == 'count' ? 'cnt' : ikey]: currentDataExist
              ? currentDataExist[ikey == 'cnt' ? 'cnt' : ikey]
              : 0
          })
          newPreviousData[ikey].push({
            record_time: previousDate,
            [ikey == 'count' ? 'cnt' : ikey]: previousDataExist
              ? previousDataExist[ikey == 'cnt' ? 'cnt' : ikey]
              : 0
          })
        })
      }
      [newCurrentData.pro, newPreviousData.pro].map((x) => {
        x.map((xx) => {
          xx.pro = -xx.pro
        })
      })

      Array.from(['active', 'cnt', 'eff', 'pro']).map((ikey, index) => {
        let dom = document.querySelector(`#betInfoChart${index + 1}Dom`)
        if (this[`myChart${index}`]) {
          this[`myChart${index}`].clear()
        }
        setTimeout(() => {
          if (!this[`myChart${index}`]) {
            this[`myChart${index}`] = echarts.init(dom)
          }
          // 绘制图表
          this[`myChart${index}`].setOption({
            title: {
              text: {
                active: '时间段人数',
                cnt: '时间段注单量',
                eff: '时间段投注量',
                pro: '时间段盈亏'
              }[ikey],
              top: 10,
              left: 10
            },
            grid: {
              left: '100',
              right: '50',
              bottom: '10%',
              top: '70'
            },
            legend: {
              data: ['当日', '昨日']
            },
            tooltip: {
              trigger: 'axis',
              formatter: function (params, ticket, callback) {
                let htmlCode = ''
                params.map((series) => {
                  let line = series.seriesName + ': ' + series.value + '<br>'
                  htmlCode += line
                })
                return htmlCode
              }
            },
            xAxis: {
              type: 'category',
              boundaryGap: false,
              data: newCurrentData[ikey].map((x) =>
                dayjs(x.record_time * 1000).format('HH-mm')
              )
            },
            yAxis: {
              type: 'value'
            },
            series: [
              {
                name: '当日',
                data: newCurrentData[ikey].map(
                  (x) => x[ikey == 'cnt' ? 'cnt' : ikey]
                ),
                type: 'line',
                areaStyle: {}
              },
              {
                name: '昨日',
                data: newPreviousData[ikey].map(
                  (x) => x[ikey == 'cnt' ? 'cnt' : ikey]
                ),
                type: 'line',
                areaStyle: {}
              }
            ]
          })
        }, 500)
      })
    },
    /**
     * 游戏列表
     */
    async initGamesData () {
      let newArray = this.games
      let a = []
      a.push(...newArray)
      a.unshift({
        id: 99999,
        number: 99999,
        name: '所有游戏',
        nameZH: ''
      })
      a.map((item) => {
        if (item.nameZH.trim().length > 0) {
          item.label = item.name + '  [' + item.nameZH + ']'
        } else {
          item.label = item.name
        }
        return item
      })
      this.paramgame = a[0].number
      this.gameOptions = a
    },

    /**
     * 查询游戏
     */
    viewGame (row) {
      let id, st, et
      id = this.agent
      // if(typeof this.startDate==="string"){
      //   st = dayjs(
      //     dayjs(this.startDate.substring(0,10) + " 00:00:00")
      //   ).unix();
      // }else{
      //   st = dayjs(
      //     dayjs(this.startDate.getTime()).format("YYYY-MM-DD 00:00:00")
      //   ).unix();
      // }

      st = dayjs(
        dayjs(this.startDate.getTime()).format('YYYY-MM-DD 00:00:00')
      ).unix()

      // if(typeof this.endDate==="string"){
      //   et = dayjs(
      //     dayjs(this.endDate.substring(0,10) + " 23:59:59")
      //   ).unix();
      // }else{
      //   et = dayjs(
      //     dayjs(this.endDate.getTime()).format("YYYY-MM-DD 23:59:59")
      //   ).unix();
      // }

      et = dayjs(
        dayjs(this.endDate.getTime()).format('YYYY-MM-DD 23:59:59')
      ).unix()
      const route = {
        name: 'newHome-agentGamesStatistic/:id/:st/:et',
        params: {
          id,
          st,
          et
        }
      }

      let routeUrl = this.$router.resolve(route)
      window.open(routeUrl.href, '_blank')
    },

    /**
     * 打开单周指标
     */
    toWeekIndicatro () {
      this.$router.push({
        name: 'newHome-weekIndicator'
      })
    },

    /**
     * 查询代理业绩
     */
    viewDetail2 () {
      const route = {
        name: 'newHome-agentPerformanceStatistic/:id',
        params: {
          id: this.site
        }
      }

      let routeUrl = this.$router.resolve(route)
      window.open(routeUrl.href, '_blank')
    },

    /**
     * 查询游戏
     */
    viewDetail (row) {
      let id, st, et
      id = this.agent
      st = dayjs(
        dayjs(this.startDate.getTime()).format('YYYY-MM-DD 00:00:00')
      ).unix()
      et = dayjs(
        dayjs(this.endDate.getTime()).format('YYYY-MM-DD 23:59:59')
      ).unix()
      const route = {
        name: 'newHome-agentOrderStatistic/:id/:st/:et',
        params: {
          id,
          st,
          et
        }
      }
      let routeUrl = this.$router.resolve(route)
      window.open(routeUrl.href, '_blank')
    },

    /**
     * 初始化代理数据
     */
    async initAgentData () {
      let data = await getLinkageList()
      if (data && data.data && data.data.code == 200) {
        this.siteOption = data.data.data
        this.site = data.data.data[0].id
        data.data.data[0].agentList.unshift({
          name: '全部',
          id: 9999999
        })
        this.agentOption = data.data.data[0].agentList
        this.agent = 9999999
        this.siteOption.map((item) => {
          item.label = item.name
        })
        this.agentOption.map((item) => {
          item.label = item.name
        })
        // 默认查询
        this.agentSearch()
      }
    },

    /**
     * 切换站点
     */
    siteChanged (siteId) {
      this.agent = 9999999
      sessionStorage.setItem('siteVal', siteId)
      this.agentOption = this.siteOption.find(
        (site) => site.id == siteId
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
    },

    /**
     * 代理查询
     */
    async agentSearch () {
      this.spinShow = true
      // 转换到月初月末时间
      let startTime, endTime
      startTime = dayjs(
        dayjs(this.startDate.getTime()).format('YYYY-MM-DD 00:00:00')
      ).unix()
      endTime = dayjs(
        dayjs(this.endDate.getTime()).format('YYYY-MM-DD 23:59:59')
      ).unix()

      let params = {
        token: getToken(),
        webId: this.site,
        agentId: this.agent,
        startTime: startTime,
        endTime: endTime
      }

      if (this.agent == 9999999) {
        params.agentId = null
      }

      let data = await axios.request({
        url: 'v2/stat/agent/info',
        method: 'get',
        params
      })

      if (data && data.data && data.data.code == 200) {
        if (data.data.data && data.data.data.length > 0) {
          let list = data.data.data
          let obj = {
            rangeRegUser: 0,
            totalRegUser: 0,
            effectiveBetsTotal: 0,
            profitLossTotal: 0,
            score_up: 0,
            score_down: 0,
            revenueTotal: 0,
            chipsTotal: 0
          }
          list.forEach((item) => {
            obj.rangeRegUser += item.rangeRegUser
            obj.totalRegUser += item.totalRegUser
            obj.effectiveBetsTotal += Number(item.effectiveBetsTotal)
            obj.profitLossTotal += Number(item.profitLossTotal)
            obj.score_up += item.score_up
            obj.score_down += item.score_down
            obj.agentId = item.agentId
            obj.revenueTotal += Number(item.revenueTotal)
            obj.chipsTotal += Number(item.chipsTotal.value)
          })
          this.agentData = [obj]
          this.showBtn = true
        } else {
          this.agentData = [
            {
              rangeRegUser: 0,
              totalRegUser: 0,
              effectiveBetsTotal: 0,
              profitLossTotal: 0,
              score_up: 0,
              score_down: 0,
              revenueTotal: 0,
              chipsTotal: 0
            }
          ]
          this.showBtn = false
        }
        this.spinShow = false
      } else {
        this.spinShow = false
      }
    }
  },
  mounted () {
    this.initGamesData()
    this.initAgentData()
  }
}
</script>

<style scoped lang="less">
.betInfoChartDom {
  height: 400px;
}

.inlineForm {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.countBox {
  // display: flex;
  // justify-content: space-between;
  height: 60px;
  clear: both;
  margin-top: 15px;

  .countItem {
    text-align: center;
    background: #efeeee;
    padding: 10px 40px;
    float: left;
    margin-left: 20px;
    background: url(../../assets/images/a2.png);
    font-size: 15px;
    border-radius: 5px;
    -webkit-box-shadow: 0px 1px 2px 0px #ccc4c4;
    box-shadow: 0px 1px 2px 0px #ccc4c4;
    background-size: cover;
    background-position: center;
    color: #fff;
    height: 80px;
    line-height: 25px;

    span {
      font-weight: 600;
    }

    b {
      font-size: 14px;
    }
  }
}

.countBox .countItem:first-child {
  margin-left: 0px;
}
</style>
