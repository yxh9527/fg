<template>
  <div>
    <Card>
      <div v-for="item in req1" :key="item.key" class="label-style">
        <div v-if="item.type == 'datetime'">
          <label style="margin-right: 5px">{{ item.label }}</label>
          <DatePicker
            @on-open-change="clearTimeDatess"
            v-model="item.value"
            :options="startDateRestrict"
            :placeholder="'请选择' + item.label"
          ></DatePicker>
        </div>
        <div v-if="item.type == 'radio'" style="margin-left: 20px">
          <RadioGroup
            type="button"
            border="true"
            v-for="items in item.option"
            :key="items.label"
            @on-change="reresetDatePiker"
            size="small"
            v-model="item.value"
          >
            <Radio :label="items.label" class="radio-style">
              {{ items.title }}
            </Radio>
          </RadioGroup>
          <br />
        </div>
      </div>
      <div class="select-style">
        <label>站点选择</label>
        <Select
          v-model="webId"
          @on-change="setSite"
          style="width: 210px"
          filterable
          filter-by-label
        >
          <Option
            v-for="item in siteOption"
            :value="item.id"
            :key="item.id"
            :label="item.name"
            >{{ item.name }}</Option
          >
        </Select>
      </div>
      <div class="select-style" style="margin-left: 10px">
        <label>代理选择</label>
        <Select
          v-model="agentId"
          style="width: 210px"
          filterable
          filter-by-label
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
      <div v-for="item in req" :key="item.number" class="label-style">
        <div class="select-style" style="margin-left: 0px 0px">
          <label>{{ item.label }}</label>
          <Select
            v-model="item.value"
            style="width: 400px"
            filterable
            filter-by-label
          >
            <Option
              v-for="item in item.option"
              :value="item.number"
              :key="item.number"
              :label="item.label"
              >{{ item.label }}</Option
            >
          </Select>
        </div>
      </div>
      <div class="select-style">
        <Button
          type="primary"
          style="margin-right: 15px"
          @click="
            pageData.page = 1;
            handleSearch();
          "
          >搜索</Button
        >
        <Button @click="handleAllSearch">重置</Button>
      </div>

      <div class="select-style">
        <Button
          type="primary"
          style="margin-right: 15px"
          @click="exportAgentDataWithTime()"
          >导出代理数据(注单统计)</Button
        >
        <Spin fix v-if="loading">
          <Icon type="ios-loading" class="demo-spin-icon-load" size="18"></Icon>
          <div>正在导出...</div>
        </Spin>
      </div>
    </Card>

    <Card style="margin: 10px 0">
      <ul>
        <span v-for="(item, index) in req1" :key="item.id">
          <li v-if="index == 0" class="label-style" style="display: block">
            <template v-if="item.value">
              <div>
                日期范围：{{ formatterTime(item.value) }} —
                {{ formatterTime(req1[1].value) }}
              </div>
            </template>
            <template v-else>
              <div>
                日期范围：
                {{
                  req1[2].value
                    ? req1[2].option.find((x) => x.label == req1[2].value).title
                    : "所有时间"
                }}
              </div>
            </template>
          </li>
        </span>

        <span v-for="item in req" :key="item.id">
          <li class="label-style">
            <div>
              {{ item.label }}：{{
                item.option.find((x) => x.id == item.value) &&
                item.option.find((x) => x.id == item.value).name
              }}
            </div>
          </li>
        </span>
        <li
          class="label-style"
          v-for="item in betList.filter(
            (x) => x.label != '总下注' && x.label != '税收' && x.label != '盈亏'
          )"
          :key="item.label"
        >
          <b>{{ item.label }}：{{ item.value }}</b>
        </li>
        <li class="label-style">
          <b>盈亏：{{ Number(betList[3].value).toFixed(2) }}</b>
        </li>
        <li class="label-style">
          <b>税收：{{ betList[4].value }}</b>
        </li>
        <li class="label-style">
          <b
            >杀数：{{
              isNaN(
                (Number(betList[3].value) / Number(betList[1].value)).toFixed(3)
              )
                ? 0
                : (Number(betList[3].value) / Number(betList[1].value)).toFixed(
                    3
                  )
            }}</b
          >
        </li>
      </ul>
    </Card>
    <Card>
      <tables ref="tables" v-model="tableData" :columns="columns" />
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
import { getDate } from '@/libs/tools'
import { exportExcel } from '@/libs/excel'
import JSZip from 'jszip'
import FileSaver from 'file-saver'

import * as dayjs from 'dayjs'
import { getReportData, exportAgentData } from '@/api/data'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    let _this = this
    return {
      loading: false,
      model1: [],
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
      req1: [
        {
          key: 'startTime',
          label: '开始日期',
          type: 'datetime',
          value: ''
        },
        { key: 'endTime', label: '结束日期', type: 'datetime', value: '' },
        {
          key: 'timeType',
          type: 'radio',
          value: '',
          option: [
            { title: '今日', label: 4 },
            { title: '昨日', label: 5 }
          ]
        }
      ],
      req: [
        // {
        //   key: "gameId",
        //   label: "游戏名称",
        //   value: "",
        //   option: [],
        // },
      ],
      betList: [
        { label: '总局数', value: '' },
        { label: '有效下注', value: '' },
        { label: '有效打码', value: '' },
        { label: '盈亏', value: '' },
        { label: '税收', value: '' }
      ],
      selectValue: [],
      columns: [
        { title: '代理', key: 'agentName', align: 'center', minWidth: 100 },
        { title: '人次', key: 'userNumber', width: 100, align: 'center' },
        { title: '局数', key: 'gameNumber', width: 100, align: 'center' },
        {
          title: '有效下注',
          key: 'effeBetsScore',
          align: 'right',
          minWidth: 100,
          render (h, params) {
            return (
              <span>
                {(params.row.effectiveBetsTotal &&
                  params.row.effectiveBetsTotal.toFixed(2)) ||
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
                String(params.row.profitLossTotal).replace(/\,/g, '')
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
            if (params.row.profitLossTotal) {
              let sum, effect
              sum = Number(
                String(params.row.profitLossTotal).replace(/\,/g, '')
              )
              effect = Number(
                String(params.row.effectiveBetsTotal).replace(/\,/g, '')
              )
              return effect - sum >= 0 ? (
                <span style="color:green">
                  {String((effect - sum).toFixed(2))}
                </span>
              ) : (
                <span style="color:red">
                  {String((effect - sum).toFixed(2))}
                </span>
              )
            } else {
              return <span style="color:#000">0</span>
            }
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
              String(params.row.profitLossTotal).replace(/\,/g, '')
            )
            let effect = Number(
              String(params.row.effectiveBetsTotal).replace(/\,/g, '')
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
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 200,
          button: [
            (h, params) => {
              return h(
                'a',
                {
                  on: {
                    click: () => {
                      let routeData = _this.$router.resolve({
                        path: '/agent-aggs-detail',
                        query: {
                          webId: _this.webId,
                          agent: params.row.agentId,
                          startTime: _this.searchData.startTime,
                          endTime: _this.searchData.endTime
                        }
                      })
                      window.open(routeData.href, '_blank')
                    }
                  }
                },
                '详情'
              )
            }
          ]
        }
      ],
      tableData: [],
      dataScore: [],
      newpage: 1,
      newtotal: 1,
      newpagesize: 30,
      searchData: {},
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      spinShow: false,
      /**
       * 补充的查询参数
       */
      webId: null,
      siteOption: [],
      agentId: null,
      agentOption: []
    }
  },
  methods: {
    clearTimeDatess () {
      // 清除快捷日期
      this.req1[2].value = ''
    },
    reresetDatePiker (v) {
      if (v) {
        this.req1[0].value = ''
        this.req1[1].value = ''
      }
    },
    formatterTime (time) {
      if (time) {
        return dayjs(time).format('YYYY-MM-DD')
      } else {
        return ''
      }
    },
    handleSearch () {
      let tmp = []
      this.req1.map((item) => {
        if (item.value) {
          if (item.type == 'datetime') {
            tmp.push({
              [item.key]: getDate(item.value)
            })
          } else {
            tmp.push({
              [item.key]: item.value
            })
          }
        }
      })
      this.req.map((item) => {
        if (item.value !== '') {
          tmp.push({
            [item.key]: item.value
          })
        }
      })
      this.spinShow = true
      // 今日昨日
      let timeRange = tmp.find((x) => x.timeType)
      if (timeRange) {
        switch (timeRange.timeType) {
          case 4:
            this.searchData.startTime = dayjs()
              .subtract(0, 'days')
              .startOf('day')
              .unix()
            this.searchData.endTime = dayjs()
              .subtract(0, 'days')
              .endOf('day')
              .unix()
            break
          case 5:
            this.searchData.startTime = dayjs()
              .subtract(1, 'days')
              .startOf('day')
              .unix()
            this.searchData.endTime = dayjs()
              .subtract(1, 'days')
              .endOf('day')
              .unix()
            break
        }
      } else {
        tmp.filter((params) => {
          if (params.startTime) {
            this.searchData.startTime = dayjs(params.startTime).unix()
          }
          if (params.endTime) {
            this.searchData.endTime = dayjs(params.endTime)
              .subtract(0, 'days')
              .endOf('day')
              .unix()
          }
        })
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { webId: this.webId },
        { agentId: this.agentId === 9999999 ? undefined : this.agentId },
        { startTime: this.searchData.startTime },
        { endTime: this.searchData.endTime }
      ]
      if (this.req1[0].value && this.req1[1].value) {
        this.req1[2].value = ''
      }
      getReportData(Data)
        .then((res) => {
          this.spinShow = false
          if (res.data.code == 200) {
            this.tableData = res.data.data.data
            this.pageData.current = res.data.data.total;
            [
              this.betList[0].value,
              this.betList[1].value,
              this.betList[2].value,
              this.betList[3].value,
              this.betList[4].value
            ] = [
              res.data.data.docCount.toFixed(2),
              res.data.data.effectiveBetsTotal.toFixed(2),
              res.data.data.chipsTotal.toFixed(2),
              (
                Number(res.data.data.effectiveBetsTotal) -
                Number(res.data.data.profitLossTotal)
              ).toFixed(2),
              (res.data.data.revenueTotal &&
                res.data.data.revenueTotal.toFixed(2)) ||
                0
            ]
          }
        })
        .catch((err) => {
          this.spinShow = false
        })
    },
    exportAgentDataWithTime () {
      this.loading = true
      let tmp = []
      this.req1.map((item) => {
        if (item.value) {
          if (item.type == 'datetime') {
            tmp.push({
              [item.key]: getDate(item.value)
            })
          } else {
            tmp.push({
              [item.key]: item.value
            })
          }
        }
      })
      this.req.map((item) => {
        if (item.value !== '') {
          tmp.push({
            [item.key]: item.value
          })
        }
      })
      // 今日昨日
      let timeRange = tmp.find((x) => x.timeType)
      if (timeRange) {
        switch (timeRange.timeType) {
          case 4:
            this.searchData.startTime = dayjs()
              .subtract(0, 'days')
              .startOf('day')
              .unix()
            this.searchData.endTime = dayjs()
              .subtract(0, 'days')
              .endOf('day')
              .unix()
            break
          case 5:
            this.searchData.startTime = dayjs()
              .subtract(1, 'days')
              .startOf('day')
              .unix()
            this.searchData.endTime = dayjs()
              .subtract(1, 'days')
              .endOf('day')
              .unix()
            break
        }
      } else {
        tmp.filter((params) => {
          if (params.startTime) {
            this.searchData.startTime = dayjs(params.startTime).unix()
          }
          if (params.endTime) {
            this.searchData.endTime = dayjs(params.endTime)
              .subtract(0, 'days')
              .endOf('day')
              .unix()
          }
        })
      }
      let Data = [
        { startTime: this.searchData.startTime },
        { endTime: this.searchData.endTime }
      ]
      let _this = this
      exportAgentData(Data).then((res) => {
        const columns = [
          { title: 'Symbol', key: 'symbol', width: 100 },
          { title: '游戏名称', key: 'gameName', width: 80 },
          { title: '注单数量', key: 'doc_count', width: 80 },
          { title: '玩家数量', key: 'userTotal', width: 80 },
          { title: '有效投注', key: 'effectiveBetsTotal', width: 80 },
          { title: '有效打码', key: 'chipsTotal', width: 80 },
          { title: '总返奖', key: 'profitLossTotal', width: 80 },
          { title: '总税收', key: 'revenueTotal', width: 80 }
        ]
        let data = {}
        Object.keys(res.data.data).forEach((key) => {
          let item = res.data.data[key]
          if (data[item.agentId]) {
            data[item.agentId].push(item)
          } else {
            data[item.agentId] = [item]
          }
        })
        const zip = new JSZip()
        Object.keys(data).forEach((key) => {
          (function (tmp) {
            let ec = exportExcel(columns, data[tmp], tmp)
            zip.file(`${tmp}.xlsx`, ec, { binary: true })
          })(key)
        })
        // 生成zip文件并下载
        zip.generateAsync({ type: 'blob' }).then((content) => {
          FileSaver.saveAs(
            content,
            `agent统计[${tmp[0].startTime}-${tmp[1].endTime}].zip`
          )
        })
        _this.loading = false
      })
    },
    handleAllSearch () {
      // 重置时间
      this.req1[0].value = ''
      this.req1[1].value = ''
      this.req1[2].value = ''
      this.agentId = null
      this.webId = this.siteOption[0].id
      sessionStorage.setItem('siteVal', this.siteOption[0].id)

      this.spinShow = true
      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      this.searchData.startTime = 0
      this.searchData.endTime = 0
      getReportData(Data).then((res) => {
        this.spinShow = false
        if (res.data.code == 200) {
          this.pageData.page = 1
          this.tableData = res.data.data.data
          this.pageData.current = res.data.data.total
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
    /**
     * set sit
     */
    setSite (val) {
      this.webId = val
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
     * 根据代理id查询游戏
     */
    selectGameByAgent (val) {
      if (val === 9999999) {
        //  选择的全部则不执行下面的代码
        return
      }
      this.req = [
        // {
        //   key: "gameId",
        //   label: "游戏名称",
        //   value: "",
        //   option: [],
        // },
      ]
    }
  },
  mounted () {
    /**
     * 填充页面的站点参数
     */
    let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
    let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
    this.siteOption = siteOption
    this.webId = sid * 1
    // 把选择的站点赋值到页面选中
    siteOption &&
      siteOption.map((item, index) => {
        if (item.id == sid) {
          this.agentOption = item.agentList
          this.agent = 9999999
          this.agentOption.unshift({
            name: '全部',
            id: 9999999
          })
          this.agentOption.map((item) => {
            item.label = item.name
          })
        }
      })
    this.siteOption.map((item) => {
      item.label = item.name
    })
    this.req.map((item) => {
      if (item.key == 'gameId') {
        let newArray = JSON.parse(sessionStorage.getItem('games'))
        item.option.push(
          { id: 0, number: 0, name: '全部', nameZH: '' },
          ...newArray
        )
        item.option.map((o) => {
          if (o.nameZH.trim().length > 0) {
            o.label = o.name + '  [' + o.nameZH + ']'
          } else {
            o.label = o.name
          }
          return o
        })
      }
      return item
    })
    this.handleSearch()
  }
}
</script>

<style lang="less">
.label-style {
  font-size: 16px;
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
