<template>
  <div>
    <Card>
      <div class="inlineForm">
        <div>站点选择：</div>
        <div>
          <Select
            placeholder="选择站点"
            filterable
            filter-by-label
            v-model="siteParams.val"
            @on-change="setSiteSession"
            style="width: 200px"
          >
            <Option
              v-for="item in siteParams.option"
              :value="item.id"
              :label="item.label"
              :key="item.id"
            >
              {{ item.name }}</Option
            >
          </Select>
        </div>
        <div style="padding-left: 20px">代理选择：</div>
        <div>
          <Select
            :placeholder="agentParams.msg"
            filterable
            filter-by-label
            v-model="agentParams.val"
            style="width: 200px"
          >
            <Option
              v-for="item in agentParams.option"
              :value="item.id"
              :key="item.id"
              :label="item.name"
            >
              {{ item.name }}</Option
            >
          </Select>
        </div>
        <div style="margin-left: 10px">游戏选择：</div>
        <div>
          <Select
            v-model="paramgame"
            style="width: 400px"
            filterable
            filter-by-label
            clearable
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
        <Button @click="getCompareData" style="margin-left: 20px" type="primary"
          >搜索</Button
        >
      </div>
    </Card>
    <Card v-if="compareInfo" style="margin-top: 20px">
      <div class="compareDataPlane">
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color:
                  compareInfo.week.cur.user[0].registPercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.week.cur.user[0].registPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.user[0].registPercent.toFixed(2) + "%"
              }}
            </div>
            <div
              :style="{
                color:
                  compareInfo.month.cur.user[0].registPercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.month.cur.user[0].registPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.month.cur.user[0].registPercent.toFixed(2) + "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ Number(compareInfo.week.cur.user[0].regist).toFixed(2) }}
            </div>
            <div>
              {{ Number(compareInfo.month.cur.user[0].regist).toFixed(2) }}
            </div>
          </div>
          <h2>注册</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color:
                  compareInfo.week.cur.user[0].new_activePercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.week.cur.user[0].new_activePercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.user[0].new_activePercent.toFixed(2) +
                    "%"
              }}
            </div>
            <div
              :style="{
                color:
                  compareInfo.month.cur.user[0].new_activePercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.month.cur.user[0].new_activePercent.toFixed(0) ==
                "0"
                  ? "/"
                  : compareInfo.month.cur.user[0].new_activePercent.toFixed(2) +
                    "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ Number(compareInfo.week.cur.user[0].new_active).toFixed(2) }}
            </div>
            <div>
              {{ Number(compareInfo.month.cur.user[0].new_active).toFixed(2) }}
            </div>
          </div>
          <h2>新增</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color:
                  compareInfo.week.cur.user[0].all_activePercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.week.cur.user[0].all_activePercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.user[0].all_activePercent.toFixed(2) +
                    "%"
              }}
            </div>
            <div
              :style="{
                color:
                  compareInfo.month.cur.user[0].all_activePercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.month.cur.user[0].all_activePercent.toFixed(0) ==
                "0"
                  ? "/"
                  : compareInfo.month.cur.user[0].all_activePercent.toFixed(2) +
                    "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ Number(compareInfo.week.cur.user[0].all_active).toFixed(2) }}
            </div>
            <div>
              {{ Number(compareInfo.month.cur.user[0].all_active).toFixed(2) }}
            </div>
          </div>
          <h2>活跃</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              v-if="compareInfo.week.prev.game[0].eff < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.week.cur.game[0].effPercent < 0 ? 'red' : 'green',
              }"
            >
              {{
                compareInfo.week.cur.game[0].effPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.game[0].effPercent.toFixed(2) + "%"
              }}
            </div>
            <div
              v-if="compareInfo.month.prev.game[0].eff < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.month.cur.game[0].effPercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.month.cur.game[0].effPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.month.cur.game[0].effPercent.toFixed(2) + "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ Number(compareInfo.week.cur.game[0].eff).toFixed(2) }}
            </div>
            <div>
              {{ Number(compareInfo.month.cur.game[0].eff).toFixed(2) }}
            </div>
          </div>
          <h2>有效下注</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              v-if="compareInfo.week.prev.game[0].pro < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.week.cur.game[0].proPercent < 0 ? 'red' : 'green',
              }"
            >
              {{
                compareInfo.week.cur.game[0].proPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.game[0].proPercent.toFixed(2) + "%"
              }}
            </div>
            <div
              v-if="compareInfo.month.prev.game[0].pro < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.month.cur.game[0].proPercent < 0
                    ? 'red'
                    : 'green',
              }"
            >
              {{
                compareInfo.month.cur.game[0].proPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.month.cur.game[0].proPercent.toFixed(2) + "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ Number(compareInfo.week.cur.game[0].pro).toFixed(2) }}
            </div>
            <div>
              {{ Number(compareInfo.month.cur.game[0].pro).toFixed(2) }}
            </div>
          </div>
          <h2>盈亏</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              v-if="compareInfo.week.prev.game[0].ss < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.week.cur.game[0].ssPercent < 0 ? 'red' : 'green',
              }"
            >
              {{
                compareInfo.week.cur.game[0].ssPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.week.cur.game[0].ssPercent.toFixed(2) + "%"
              }}
            </div>
            <div
              v-if="compareInfo.month.prev.game[0].ss < 0"
              style="color: red; font-size: 16px"
            >
              参考值为负
            </div>
            <div
              v-else
              :style="{
                color:
                  compareInfo.month.cur.game[0].ssPercent < 0 ? 'red' : 'green',
              }"
            >
              {{
                compareInfo.month.cur.game[0].ssPercent.toFixed(0) == "0"
                  ? "/"
                  : compareInfo.month.cur.game[0].ssPercent.toFixed(2) + "%"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{ (Number(compareInfo.week.cur.game[0].ss) * 100).toFixed(4) }}%
            </div>
            <div>
              {{ (Number(compareInfo.month.cur.game[0].ss) * 100).toFixed(4) }}%
            </div>
          </div>
          <h2>杀数</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>本周</div>
            <div>上周</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color:
                  weekAverageCurrent < weekAveragePrevious ? 'red' : 'green',
              }"
            >
              {{ isNaN(weekAverageCurrent) ? "/" : weekAverageCurrent }}
            </div>

            <div
              :style="{
                color:
                  weekAverageCurrent > weekAveragePrevious ? 'red' : 'green',
              }"
            >
              {{ isNaN(weekAveragePrevious) ? "/" : weekAveragePrevious }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div></div>
            <div></div>
          </div>
          <h2>平均次留</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color: averageBetRatio < 0 ? 'red' : 'green',
              }"
            >
              {{ averageBetRatio ? averageBetRatio.toFixed(2) + "%" : "/" }}
            </div>

            <div
              :style="{
                color: averageBetRatiomonth < 0 ? 'red' : 'green',
              }"
            >
              {{
                averageBetRatiomonth
                  ? averageBetRatiomonth.toFixed(2) + "%"
                  : "/"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{
                averageBetCurrent == Infinity
                  ? 0
                  : averageBetCurrent
                  ? averageBetCurrent.toFixed(2)
                  : 0
              }}
            </div>
            <div>
              {{
                averageBetCurrentmonth == Infinity
                  ? 0
                  : averageBetCurrentmonth
                  ? averageBetCurrentmonth.toFixed(2)
                  : 0
              }}
            </div>
          </div>
          <h2>人均流水</h2>
        </div>
        <div class="compareBox">
          <div class="compareLine">
            <div>环比上周</div>
            <div>同比上月</div>
          </div>
          <div class="compareLine comparePercent">
            <div
              :style="{
                color: averageBetCountRatio < 0 ? 'red' : 'green',
              }"
            >
              {{
                averageBetCountRatio
                  ? averageBetCountRatio.toFixed(2) + "%"
                  : "/"
              }}
            </div>

            <div
              :style="{
                color: averageBetCountRatiomonth < 0 ? 'red' : 'green',
              }"
            >
              {{
                averageBetCountRatiomonth
                  ? averageBetCountRatiomonth.toFixed(2) + "%"
                  : "/"
              }}
            </div>
          </div>
          <div class="compareLine compareData">
            <div>
              {{
                averageBetCountCurrent == Infinity
                  ? 0
                  : averageBetCountCurrent
                  ? averageBetCountCurrent.toFixed(2)
                  : 0
              }}
            </div>
            <div>
              {{
                averageBetCountCurrentmonth == Infinity
                  ? 0
                  : averageBetCountCurrentmonth
                  ? averageBetCountCurrentmonth.toFixed(2)
                  : 0
              }}
            </div>
          </div>
          <h2>人均注单数</h2>
        </div>
      </div>
    </Card>

    <Card v-if="false" style="margin-top: 20px">
      <h3>周游戏活跃排行</h3>
      <div class="chartViewport" id="weekRankingChartDom"></div>
    </Card>

    <Card v-if="false" style="margin-top: 20px">
      <h3 style="padding-bottom: 20px">玩家盈利/亏损榜</h3>
      <Tabs value="name1" type="card">
        <TabPane label="玩家盈利榜" name="name1">
          <div class="lossProfitRank" style="align-items: start">
            <div class="lossProfitRankSubPane">
              <div>昨日排行({{ momentYesterday }})</div>
              <div>
                <Table
                  v-if="playerLossProfitData1"
                  :columns="playerLossProfitTableColumn"
                  :data="playerLossProfitData1.yesterday"
                ></Table>
              </div>
            </div>
            <div class="lossProfitRankSubPane">
              <div>上周排行({{ momentLastweek }})</div>
              <div>
                <Table
                  v-if="playerLossProfitData1"
                  :columns="playerLossProfitTableColumn"
                  :data="playerLossProfitData1.last_week"
                ></Table>
              </div>
            </div>
          </div>
        </TabPane>
        <TabPane label="玩家亏损榜" name="name2">
          <div class="lossProfitRank" style="align-items: start">
            <div class="lossProfitRankSubPane">
              <div>昨日排行({{ momentYesterday }})</div>
              <div>
                <Table
                  v-if="playerLossProfitData2"
                  :columns="playerLossProfitTableColumn1"
                  :data="playerLossProfitData2.yesterday"
                ></Table>
              </div>
            </div>
            <div class="lossProfitRankSubPane">
              <div>上周排行({{ momentLastweek }})</div>
              <div>
                <Table
                  v-if="playerLossProfitData2"
                  :columns="playerLossProfitTableColumn1"
                  :data="playerLossProfitData2.last_week"
                ></Table>
              </div>
            </div>
          </div>
        </TabPane>
      </Tabs>
    </Card>

    <Card style="margin-top: 20px">
      <div class="inlineForm" style="margin: 10px 100px">
        <div style="width: 200px">
          <RadioGroup @on-change="switchChartSearch" v-model="switchChartMode">
            <Radio label="按周查询"></Radio>
            <Radio label="按月查询"></Radio>
          </RadioGroup>
        </div>
        <div style="width: 400px">
          时间选择：
          <Select v-model="chartDateParams.val" style="width: 100px">
            <Option
              v-for="item in chartDateParams.option"
              :value="item.value"
              :key="item.value"
              >{{ item.label }}</Option
            >
          </Select>
          <span style="padding-left: 20px">{{ showDateRange }}</span>
        </div>
        <div style="width: 100px">
          <Button @click="renderChart" type="primary">搜索</Button>
        </div>
      </div>
      <div class="chartViewport" id="renderLineChartDom"></div>
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
  data () {
    return {
      averageBetRatio: 0,
      averageBetCurrent: 0,
      averageBetPrevious: 0,
      averageBetCountRatio: 0,
      averageBetCountCurrent: 0,
      averageBetCountPrevious: 0,
      weekAverageCurrent: 0,
      weekAveragePrevious: 0,
      switchChartMode: '按周查询',
      chartDateParams: { val: '', option: [] },
      siteParams: { val: '', option: [] },
      agentParams: { val: '', option: [], msg: '请先选择站点' },
      paramgame: null,
      gameOptions: [],
      games: JSON.parse(sessionStorage.getItem('games') || '[]'),
      compareInfo: null,
      chartSelected: {},
      momentYesterday: dayjs().subtract(1, 'day').format('MM-DD'),
      momentLastweek:
        dayjs()
          .subtract(1, 'week')
          .startOf('week')
          .add(1, 'day')
          .format('MM-DD') +
        ' — ' +
        dayjs().subtract(1, 'week').endOf('week').add(1, 'day').format('MM-DD'),
      data1: [],
      playerLossProfitData1: null,
      playerLossProfitData2: null,
      playerLossProfitTableColumn: [
        {
          title: '排名',
          type: 'index',
          align: 'center'
        },
        {
          title: '用户ID',
          key: 'user_id',
          align: 'center'
        },
        {
          title: '打码量',
          key: 'eff',
          align: 'center'
        },
        {
          title: '盈利金额',
          key: 'pro',
          align: 'center',
          render (h, params) {
            return <div>{params.row.pro}</div>
          }
        },
        {
          title: '比例',
          key: 'test',
          align: 'center',
          render (h, params) {
            return (
              <div>{((params.row.pro / params.row.eff) * 100).toFixed(4)}%</div>
            )
          }
        }
      ],
      playerLossProfitTableColumn1: [
        {
          title: '排名',
          type: 'index',
          align: 'center'
        },
        {
          title: '用户ID',
          key: 'user_id',
          align: 'center'
        },
        {
          title: '打码量',
          key: 'eff',
          align: 'center'
        },
        {
          title: '亏损金额',
          key: 'pro',
          align: 'center',
          render (h, params) {
            return <div>{params.row.pro}</div>
          }
        },
        {
          title: '比例',
          key: 'test',
          align: 'center',
          render (h, params) {
            return (
              <div>{((params.row.pro / params.row.eff) * 100).toFixed(4)}%</div>
            )
          }
        }
      ]
    }
  },
  methods: {
    /**
     * 获取周游戏活跃排行
     */
    async renderWeekRanking () {
      return
      /* eslint-disable no-unreachable */
      let currentParams = {
        token: getToken(),
        agentId: this.agentParams.val == 999999 ? null : this.agentParams.val
      }
      let currentData1 = await axios.request({
        url: 'v2/govern/lastWeekTopGames',
        method: 'get',
        params: currentParams
      })

      // 游戏排序
      //       game_id: 10
      // w_active: 10

      currentData1.data.data.sort((a, b) => b.w_active - a.w_active)
      currentData1.data.data.map((x) => {
        x.gameName = this.games.find((d) => d.number == x.game_id).name
      })

      let dom = document.querySelector('#weekRankingChartDom')
      if (this.myChart_weekRankingChartDom) {
        this.myChart_weekRankingChartDom.clear()
      }
      setTimeout(() => {
        if (!this.myChart_weekRankingChartDom) {
          this.myChart_weekRankingChartDom = echarts.init(dom)
        }
        // 绘制图表
        this.myChart_weekRankingChartDom.setOption({
          grid: {
            left: '88',
            right: '88',
            bottom: '40',
            top: '40'
          },
          xAxis: [
            {
              type: 'category',
              data: currentData1.data.data.map((x) => x.gameName)
            }
          ],
          yAxis: {
            type: 'value'
          },
          tooltip: {
            trigger: 'axis'
          },
          series: [
            {
              data: currentData1.data.data.map((x) => x.w_active),
              type: 'bar'
            }
          ]
        })
      }, 500)
      /* eslint-enable no-unreachable */
    },

    /**
     * 玩家盈亏榜
     */
    async renderPlayerLossProfitRank () {
      return
      /* eslint-disable no-unreachable */
      let currentParams = {
        token: getToken(),
        agentId: this.agentParams.val == 999999 ? null : this.agentParams.val,
        isWin: 1 // 1是盈利
      }
      let currentData1 = await axios.request({
        url: 'v2/govern/playerProRank',
        method: 'get',
        params: currentParams
      })

      currentParams.isWin = 0
      let currentData2 = await axios.request({
        url: 'v2/govern/playerProRank',
        method: 'get',
        params: currentParams
      })

      currentData1.data.data.last_week.map((x) => {
        x.pro = Number(x.pro)
        x.eff = Number(x.eff).toFixed(2)
      })
      currentData1.data.data.last_week =
        currentData1.data.data.last_week.filter((x) => x.pro >= 0)
      currentData1.data.data.last_week.sort((a, b) => b.pro - a.pro)
      currentData1.data.data.yesterday.map((x) => {
        x.pro = Number(x.pro)
        x.eff = Number(x.eff).toFixed(2)
      })
      currentData1.data.data.yesterday =
        currentData1.data.data.yesterday.filter((x) => x.pro >= 0)
      currentData1.data.data.yesterday.sort((a, b) => b.pro - a.pro)

      currentData2.data.data.last_week.map((x) => {
        x.pro = Number(x.pro)
        x.eff = Number(x.eff).toFixed(2)
      })
      currentData2.data.data.last_week.sort((a, b) => a.pro - b.pro)
      currentData2.data.data.yesterday.map((x) => {
        x.pro = Number(x.pro)
        x.eff = Number(x.eff).toFixed(2)
      })
      currentData2.data.data.yesterday.sort((a, b) => a.pro - b.pro)

      this.playerLossProfitData1 = currentData1.data.data
      this.playerLossProfitData2 = currentData2.data.data
      /* eslint-enable no-unreachable */
    },

    /**
     *渲染图表
     */
    async renderChart () {
      let searchDate = this.chartDateParams.option.find(
        (x) => x.value == this.chartDateParams.val
      )

      let currentParams = {
        token: getToken(),
        gameId: this.paramgame == 99999 ? null : this.paramgame,
        agentId: this.agentParams.val == 999999 ? null : this.agentParams.val,
        webId: this.siteParams.val,
        dateRange: `${searchDate.current.dateStartStr} - ${searchDate.current.dateEndStr}`,
        startTime: searchDate.current.dateStart,
        endTime: searchDate.current.dateEnd
      }
      let currentData1 = await axios.request({
        url: 'v2/govern/userAndGameSummaryDetail',
        method: 'get',
        params: currentParams
      })

      let previousParams = {
        token: getToken(),
        gameId: this.paramgame == 99999 ? null : this.paramgame,
        agentId: this.agentParams.val == 999999 ? null : this.agentParams.val,
        webId: this.siteParams.val,
        dateRange: `${searchDate.previous.dateStartStr} - ${searchDate.previous.dateEndStr}`,
        startTime: searchDate.previous.dateStart,
        endTime: searchDate.previous.dateEnd
      }
      let previousData1 = await axios.request({
        url: 'v2/govern/userAndGameSummaryDetail',
        method: 'get',
        params: previousParams
      })

      let chartDates = [],
        chartDatesPrevious = [],
        currentData = { game: [], user: [] },
        previousData = { game: [], user: [] },
        days = searchDate.label.includes('周') ? 7 : 31

      for (let index = 0; index < days; index++) {
        let date, date2
        if (days == 7) {
          date = dayjs(searchDate.current.dateStartStr)
            .add(index, 'day')
            .format('YYYY-MM-DD')
          chartDates.push(date)
          date2 = dayjs(searchDate.previous.dateStartStr)
            .add(index, 'day')
            .format('YYYY-MM-DD')
          chartDatesPrevious.push(date2)
        } else {
          let oldDateStr = dayjs(searchDate.current.dateStartStr).format(
            'YYYY-MM-DD'
          )
          oldDateStr = oldDateStr.split('-')
          oldDateStr[2] = Number(oldDateStr[2]) + index
          date = oldDateStr.join('-')
          chartDates.push(date)

          let oldDateStr1 = dayjs(searchDate.previous.dateStartStr).format(
            'YYYY-MM-DD'
          )
          oldDateStr1 = oldDateStr1.split('-')
          oldDateStr1[2] = Number(oldDateStr1[2]) + index
          date2 = oldDateStr1.join('-')
          chartDatesPrevious.push(date2)
        }

        [
          currentData.game,
          previousData.game,
          currentData.user,
          previousData.user
        ].map((ddd, dddindex) => {
          let dateStr = dayjs(searchDate.current.dateStartStr)
            .add(index, 'day')
            .unix()
          let datePrev = dayjs(searchDate.previous.dateStartStr)
            .add(index, 'day')
            .unix()

          if (dddindex < 2) {
            let remotData =
              dddindex == 0
                ? currentData1.data.data.game
                : previousData1.data.data.game
            let remoteDate = dddindex == 0 ? dateStr : datePrev
            let existData = remotData.find((x) => x.record_time == remoteDate)
            if (existData) {
              ddd.push(existData)
            } else {
              ddd.push({
                eff: 0,
                pro: 0,
                record_time: remoteDate,
                ss: 0
              })
            }
          } else {
            let remotData =
              dddindex == 2
                ? currentData1.data.data.user
                : previousData1.data.data.user
            let remoteDate = dddindex == 2 ? dateStr : datePrev
            let existData = remotData.find((x) => x.record_time == remoteDate)
            if (existData) {
              ddd.push(existData)
            } else {
              ddd.push({
                regist: 0,
                all_active: 0,
                record_time: remoteDate,
                new_active: 0
              })
            }
          }
        })
      }

      let maxlength = Math.max(
        currentData.game.length,
        currentData.user.length,
        previousData.game.length,
        previousData.user.length
      );
      [
        currentData.game,
        previousData.game,
        currentData.user,
        previousData.user
      ].map((data, dataIndex) => {
        let templateData =
          dataIndex <= 1
            ? {
              eff: 0,
              pro: 0,
              ss: 0
            }
            : {
              all_active: 0,
              new_active: 0,
              regist: 0
            }
        let dataBoundary = maxlength - data.length
        if (dataBoundary > 0) {
          for (let index = 0; index < dataBoundary; index++) {
            data.push(templateData)
          }
        }
      })

      let chartSeries = [];

      // 折叠数据
      ['eff', 'pro', 'ss'].map((ikey) => {
        [currentData, previousData].map((tenseData, dindex) => {
          let chartDatas = []
          tenseData.game.map((gameData) => {
            if (ikey == 'ss') {
              chartDatas.push(-Number(Number(gameData[ikey]).toFixed(4)))
            } else if (ikey == 'pro') {
              chartDatas.push(-Number(Number(gameData[ikey]).toFixed(4)))
            } else {
              chartDatas.push(Number(Number(gameData[ikey]).toFixed(2)))
            }
          })
          chartSeries.push({
            name: {
              eff: '有效下注' + (dindex == 1 ? '(上期)' : ''),
              pro: '盈利' + (dindex == 1 ? '(上期)' : ''),
              ss: '杀数' + (dindex == 1 ? '(上期)' : '')
            }[ikey],
            data: chartDatas,
            type: 'line'
          })
        })
      });

      ['all_active', 'new_active', 'regist'].map((ikey) => {
        [currentData, previousData].map((tenseData, dindex) => {
          let chartDatas = []
          tenseData.user.map((userData) => {
            chartDatas.push(Number(Number(userData[ikey]).toFixed(2)))
          })
          chartSeries.push({
            name: {
              all_active: '活跃' + (dindex == 1 ? '(上期)' : ''),
              new_active: '新增' + (dindex == 1 ? '(上期)' : ''),
              regist: '注册' + (dindex == 1 ? '(上期)' : '')
            }[ikey],
            data: chartDatas,
            type: 'line'
          })
        })
      })

      let dom = document.querySelector('#renderLineChartDom')
      if (this.myChart) {
        this.myChart.clear()
      }
      setTimeout(() => {
        if (!this.myChart) {
          this.myChart = echarts.init(dom)
        }

        this.myChart.on('legendselectchanged', (params) => {
          this.chartSelected = params.selected
        })

        // 绘制图表
        this.myChart.setOption({
          grid: {
            left: '100',
            right: '150',
            bottom: '10%',
            top: '40'
          },
          xAxis: [
            {
              type: 'category',
              data: chartDates,
              position: 'top'
            },
            {
              type: 'category',
              data: chartDatesPrevious,
              position: 'bottom'
            }
          ],
          yAxis: {
            type: 'value'
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
          dataZoom: [
            {
              bottom: 0
            }
          ],
          legend: {
            type: 'scroll',
            orient: 'vertical',
            right: 10,
            top: 'center',
            icon: 'rect',
            selected: JSON.parse(JSON.stringify(this.chartSelected)),
            data: [
              '有效下注',
              '有效下注(上期)',
              '盈利',
              '盈利(上期)',
              '杀数',
              '杀数(上期)',
              '活跃',
              '活跃(上期)',
              '新增',
              '新增(上期)',
              '注册',
              '注册(上期)'
            ]
          },
          series: chartSeries
        })
      }, 500)
    },

    /**
     * 数据查询
     */
    async getCompareData () {
      let params = {
        token: getToken(),
        gameId: this.paramgame == 99999 ? null : this.paramgame,
        agentId: this.agentParams.val == 999999 ? null : this.agentParams.val,
        webId: this.siteParams.val
      }
      let data = await axios.request({
        url: 'v2/govern/userAndGameSummary',
        method: 'get',
        params
      })
      console.log('=====>', data.data)
      if (data && data.data && data.data.code == 200) {
        let newData = data.data.data;
        // 添加默认数据
        ['month', 'week'].map((dateType) => {
          ['cur', 'prev'].map((tense) => {
            ['game', 'user'].map((type) => {
              // 空数据添加0
              if (newData[dateType][tense][type].length == 0) {
                if (type == 'game') {
                  newData[dateType][tense][type].push({
                    active_cout: 0,
                    agent_id: 0,
                    bets_count: 0,
                    eff: 0,
                    game_id: 0,
                    keep_alive: 0,
                    not_keep_alive: 0,
                    pro: 0,
                    record_time: 0,
                    ss: 0,
                    web_id: 1
                  })
                } else {
                  newData[dateType][tense][type].push({
                    regist: 0,
                    all_active: 0,
                    new_active: 0
                  })
                }
              }

              if (type == 'game') {
                newData[dateType][tense][type].map((x, dindex) => {
                  // 添加人均注单数
                  x.averageBetCount =
                    x.bets_count / (x.active_cout ? x.active_cout : 0)
                  if (
                    isNaN(x.averageBetCount) ||
                    x.averageBetCount == Infinity
                  ) {
                    x.averageBetCount = 0
                  }
                  // 每日人均流水=当日总投注÷当日有产生有效投注的人数
                  x.averageBet = x.eff / (x.active_cout ? x.active_cout : 0)
                  if (isNaN(x.averageBet) || x.averageBet == Infinity) {
                    x.averageBet = 0
                  }

                  // 累加有效下注和盈亏
                  x.eff = Number(x.eff)
                  x.pro = -Number(x.pro)
                  if (dindex > 0) {
                    newData[dateType][tense].game[0].eff += x.eff
                    newData[dateType][tense].game[0].pro += x.pro
                  }
                })

                // 计算周月的杀数
                newData[dateType][tense].game[0].ss =
                  newData[dateType][tense].game[0].pro /
                  newData[dateType][tense].game[0].eff
                if (
                  isNaN(newData[dateType][tense].game[0].ss) ||
                  Infinity == newData[dateType][tense].game[0].ss
                ) {
                  newData[dateType][tense].game[0].ss = 0
                }
              } else {
                // 累加注册和活跃用户
                newData[dateType][tense][type].map((d, dindex) => {
                  if (dindex > 0) {
                    newData[dateType][tense].user[0].regist += d.regist
                    newData[dateType][tense].user[0].all_active += d.all_active
                    newData[dateType][tense].user[0].new_active += d.new_active
                  }
                })
              }
            })
          })
        })

        // 上月天数和上上月天数
        let curDays = dayjs().subtract(1, 'month').daysInMonth()
        let prevDays = dayjs().subtract(2, 'month').daysInMonth()

        // 计算人均注单数
        this.averageBetCountCurrent =
          newData.week.cur.game.reduce((a, b) => a + b.averageBetCount, 0) / 7
        this.averageBetCountPrevious =
          newData.week.prev.game.reduce((a, b) => a + b.averageBetCount, 0) / 7
        this.averageBetCountRatio =
          ((this.averageBetCountCurrent - this.averageBetCountPrevious) /
            this.averageBetCountPrevious) *
          100
        this.averageBetCountCurrentmonth =
          newData.month.cur.game.reduce((a, b) => a + b.averageBetCount, 0) /
          curDays
        this.averageBetCountPreviousmonth =
          newData.month.prev.game.reduce((a, b) => a + b.averageBetCount, 0) /
          prevDays
        this.averageBetCountRatiomonth =
          ((this.averageBetCountCurrentmonth -
            this.averageBetCountPreviousmonth) /
            this.averageBetCountPreviousmonth) *
          100

        // 计算人均流水
        this.averageBetCurrent =
          newData.week.cur.game.reduce((a, b) => a + b.averageBet, 0) / 7
        this.averageBetPrevious =
          newData.week.prev.game.reduce((a, b) => a + b.averageBet, 0) / 7
        this.averageBetRatio =
          ((this.averageBetCurrent - this.averageBetPrevious) /
            this.averageBetPrevious) *
          100
        this.averageBetCurrentmonth =
          newData.month.cur.game.reduce((a, b) => a + b.averageBet, 0) /
          curDays
        this.averageBetPreviousmonth =
          newData.month.prev.game.reduce((a, b) => a + b.averageBet, 0) /
          prevDays
        this.averageBetRatiomonth =
          ((this.averageBetCurrentmonth - this.averageBetPreviousmonth) /
            this.averageBetPreviousmonth) *
          100

        // 计算本周和上周平均次留
        newData.week.next.game.map((x) => {
          x.aliveRatio = x.keep_alive / (x.keep_alive + x.not_keep_alive)
          if (isNaN(x.aliveRatio)) {
            x.aliveRatio = 0
          }
        })
        this.weekAverageCurrent = (
          newData.week.next.game.reduce((a, b) => a + b.aliveRatio, 0) / 7
        ).toFixed(2)
        newData.week.cur.game.map((x) => {
          x.aliveRatio = x.keep_alive / (x.keep_alive + x.not_keep_alive)
          if (isNaN(x.aliveRatio)) {
            x.aliveRatio = 0
          }
        })
        this.weekAveragePrevious = (
          newData.week.cur.game.reduce((a, b) => a + b.aliveRatio, 0) / 7
        ).toFixed(2);

        // 计算周月的数值比率
        ['month', 'week'].map((dateType) => {
          ['eff', 'pro', 'ss', 'regist', 'all_active', 'new_active'].map(
            (ikey, keyIndex) => {
              let currentData, previousData
              if (keyIndex <= 2) {
                currentData = newData[dateType].cur.game[0]
                previousData = newData[dateType].prev.game[0]
              } else {
                currentData = newData[dateType].cur.user[0]
                previousData = newData[dateType].prev.user[0]
              }

              if (
                Number(currentData[ikey]) == 0 ||
                Number(previousData[ikey]) == 0
              ) {
                currentData[ikey + 'Percent'] = 0
              } else {
                currentData[ikey + 'Percent'] =
                  ((Number(currentData[ikey]) - Number(previousData[ikey])) /
                    Number(previousData[ikey])) *
                  100
              }
            }
          )
        })

        this.compareInfo = newData
        this.renderChart()
        this.renderWeekRanking()
        this.renderPlayerLossProfitRank()
      }
    },

    /**
     * 切换图标查询日期模式
     */
    switchChartSearch (val) {
      let dateArray = []
      if (val == '按周查询') {
        dateArray.push({
          label: `本周`,
          value: 0,
          current: {
            dateEndStr: dayjs().format('YYYY-MM-DD'),
            dateEnd: dayjs().unix(),
            dateStartStr: dayjs()
              .startOf('week')
              .add(1, 'day')
              .format('YYYY-MM-DD'),
            dateStart: dayjs().startOf('week').add(1, 'day').unix()
          },
          previous: {
            dateEndStr: dayjs().startOf('week').format('YYYY-MM-DD'),
            dateEnd: dayjs().startOf('week').unix(),
            dateStartStr: dayjs()
              .subtract(1, 'week')
              .startOf('week')
              .add(1, 'day')
              .format('YYYY-MM-DD'),
            dateStart: dayjs()
              .subtract(1, 'week')
              .startOf('week')
              .add(1, 'day')
              .unix()
          }
        })

        for (let index = 0; index < 12; index++) {
          dateArray.push({
            label: `前${index + 1}周`,
            value: index + 1,
            current: {
              dateEndStr: dayjs()
                .subtract(index, 'week')
                .startOf('week')
                .format('YYYY-MM-DD'),
              dateEnd: dayjs().subtract(index, 'week').startOf('week').unix(),
              dateStartStr: dayjs()
                .subtract(index + 1, 'week')
                .startOf('week')
                .add(1, 'day')
                .format('YYYY-MM-DD'),
              dateStart: dayjs()
                .subtract(index + 1, 'week')
                .startOf('week')
                .add(1, 'day')
                .unix()
            },
            previous: {
              dateEndStr: dayjs()
                .startOf('week')
                .subtract(index + 1, 'week')
                .format('YYYY-MM-DD'),
              dateEnd: dayjs()
                .startOf('week')
                .subtract(index + 1, 'week')
                .unix(),
              dateStartStr: dayjs()
                .subtract(index + 2, 'week')
                .startOf('week')
                .add(1, 'day')
                .format('YYYY-MM-DD'),
              dateStart: dayjs()
                .subtract(index + 2, 'week')
                .startOf('week')
                .add(1, 'day')
                .unix()
            }
          })
        }
      } else {
        dateArray.push({
          label: `本月`,
          value: 0,
          current: {
            dateEndStr: dayjs().format('YYYY-MM-DD'),
            dateEnd: dayjs().unix(),
            dateStartStr: dayjs().startOf('month').format('YYYY-MM-DD'),
            dateStart: dayjs().startOf('month').unix()
          },
          previous: {
            dateEndStr: dayjs()
              .startOf('month')
              .subtract(1, 'day')
              .format('YYYY-MM-DD'),
            dateEnd: dayjs().startOf('month').subtract(1, 'day').unix(),
            dateStartStr: dayjs()
              .subtract(1, 'month')
              .startOf('month')
              .format('YYYY-MM-DD'),
            dateStart: dayjs().subtract(1, 'month').startOf('month').unix()
          }
        })
        for (let index = 0; index < 3; index++) {
          dateArray.push({
            label: `前${index + 1}月`,
            value: index + 1,
            current: {
              dateEndStr: dayjs()
                .subtract(index, 'month')
                .startOf('month')
                .subtract(1, 'day')
                .format('YYYY-MM-DD'),
              dateEnd: dayjs()
                .subtract(index, 'month')
                .startOf('month')
                .subtract(1, 'day')
                .unix(),
              dateStartStr: dayjs()
                .subtract(index + 1, 'month')
                .startOf('month')
                .endOf(1, 'day')
                .format('YYYY-MM-DD'),
              dateStart: dayjs()
                .subtract(index + 1, 'month')
                .startOf('month')
                .endOf(1, 'day')
                .unix()
            },
            previous: {
              dateEndStr: dayjs()
                .subtract(index + 1, 'month')
                .startOf('month')
                .subtract(1, 'day')
                .format('YYYY-MM-DD'),
              dateEnd: dayjs()
                .subtract(index + 1, 'month')
                .startOf('month')
                .subtract(1, 'day')
                .unix(),
              dateStartStr: dayjs()
                .subtract(index + 2, 'month')
                .startOf('month')
                .endOf(1, 'day')
                .format('YYYY-MM-DD'),
              dateStart: dayjs()
                .subtract(index + 2, 'month')
                .startOf('month')
                .endOf(1, 'day')
                .unix()
            }
          })
        }
      }
      this.chartDateParams.option = dateArray
      this.chartDateParams.val = dateArray[0].value
    },

    /**
     * 设置站点
     */
    setSiteSession (val) {
      this.agentParams.val = ''
      if (val > 0) {
        this.agentParams.option = []
        for (const key in this.siteParams.option) {
          if (this.siteParams.option[key].id == val) {
            this.agentParams.option.push(
              ...this.siteParams.option[key].agentList
            )
            this.agentParams.msg = '选择代理'
          }
        }
        this.agentParams.option.map((item) => {
          item.label = item.name
        })
        this.agentParams.option.unshift({
          id: 999999,
          label: '所有代理',
          name: '所有代理'
        })
        this.agentParams.val = this.agentParams.option[0].id
      }
    },

    /**
     * 游戏列表
     */
    async initAgentData () {
      let newArray = this.games
      let a = []
      a.push(...newArray)
      a.unshift({
        id: 99999,
        number: 99999,
        name: '所有游戏',
        nameZH: '',
        symbol: ''
      })
      this.gameOptions = a
      this.gameOptions.map((item) => {
        if (item.nameZH == '' || item.nameZH === undefined) {
          item.label = item.name
        } else {
          item.label = item.name + ' [' + item.nameZH + ']'
        }
      })
      this.paramgame = a[0].number
      console.log('=====>', this.gameOptions)
    },

    /**
     * 获取站点
     */
    getSite () {
      getLinkageList().then((res) => {
        this.siteParams.option = []
        this.siteParams.option.push(...Object.assign(res.data.data))
        this.siteParams.option.map((item) => {
          item.label = item.name
          item.value = item.id
        })
        this.siteParams.val = this.siteParams.option[0].value
        this.setSiteSession(this.siteParams.val)
        setTimeout(() => {
          this.getCompareData()
        }, 50)
      })
    }
  },
  computed: {
    showDateRange () {
      let data = this.chartDateParams.option.find(
        (x) => x.value == this.chartDateParams.val
      )
      if (data) {
        return data.current.dateStartStr + ' — ' + data.current.dateEndStr
      } else {
        return ''
      }
    }
  },
  async mounted () {
    // 设置基本时间
    let currentParams = {
      token: getToken()
    }
    let currentData1 = await axios.request({
      url: 'v2/site/serverTime',
      method: 'get',
      params: currentParams
    })
    let diff = currentData1.data.data.timestamp * 1000 - Date.now()
    dayjs.now = function () {
      return Date.now ? Date.now() + diff : new Date().getTime() + diff
    }

    this.initAgentData()
    this.getSite()
    this.switchChartSearch('按周查询')
  }
}
</script>

<style scoped lang="less">
.chartViewport {
  height: 500px;
  margin-top: 20px;
}

.inlineForm {
  display: flex;
  align-items: center;
}

.compareDataPlane {
  display: flex;
  flex-wrap: wrap;

  > .compareBox {
    &:nth-of-type(3n + 0) {
      border: none;
    }

    padding: 10px;
    margin-bottom: 20px;
    box-sizing: border-box;
    border-right: 1px solid #ccc;
    flex-shrink: 0;
    width: 33.3%;

    h2 {
      text-align: center;
    }

    .compareLine {
      display: flex;
      flex-wrap: wrap;

      > div {
        width: 50%;
        text-align: center;
      }
    }

    .comparePercent {
      font-size: 26px;
    }

    .compareData {
      font-size: 18px;
      color: gray;
    }
  }
}

.lossProfitRank {
  display: flex;
  .lossProfitRankSubPane {
    width: 50%;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    & > div {
      text-align: center;
      width: 100%;
      flex-shrink: 0;
      flex-grow: 1;
    }
  }
}
</style>
