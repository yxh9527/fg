<template>
  <div>
    <Card>
      <ul style="display: inline-block">
        <li
          class="label-style"
          v-for="item in req"
          :key="item.label"
          style="float: left; margin-right: 6px"
        >
          <label>{{ item.label }}</label>
          <Input
            v-model="item.value"
            placeholder="请输入"
            v-if="item.type == 'text'"
            style="width: 120px"
            :maxlength="50"
          />
          <Select
            v-model="item.value"
            v-if="item.type == 'select'"
            @on-change="setLinkage($event, item.key)"
            style="width: 120px"
            filterable
            filter-by-label
          >
            <Option
              v-for="items in item.option"
              :value="items.id"
              :key="items.id"
              :label="item.name"
              >{{ items.name }}</Option
            >
          </Select>
        </li>
        <div
          style="
            float: left;
            margin-top: 5px;
            margin-right: 6px;
            display: flex;
            align-items: center;
          "
        >
          <div>区间投注/盈亏：</div>
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
        </div>
        <div style="float: left; margin-top: 5px; margin-right: 6px">
          <span>站点选择：</span>
          <Select
            v-model="site"
            @on-change="siteChanged"
            style="width: 150px"
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
        <div style="float: left; margin-top: 5px; margin-right: 6px">
          <span>代理选择：</span>
          <Select
            v-model="agent"
            style="width: 150px"
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
        <div class="label-style" style="margin-left: 20px">
          <Button
            type="primary"
            @click="searchAction"
            style="margin-right: 15px"
            >搜索</Button
          >
          <Button @click="handleAllSearch" style="margin-right: 15px"
            >重置</Button
          >
          <Button type="primary" @click="viewDetail2">单控管理</Button>
          <Button type="primary" style="margin-left: 15px" @click="viewClearPlayerGameState">清理玩家游戏状态</Button>
        </div>
      </ul>
      <Modal v-model="bClearPlayerGameState" width="600" title="清理玩家游戏状态" @on-ok="fClearPlayerGameState">
        <div>玩家id:
          <Input v-model="iClearPlayerGameId"   placeholder=""/>
        </div>
        <div>
          货币类型：
          <Select
            v-model="currency"
            filterable
            filter-by-label
          >
            <Option
              v-for="item in gameCurrency"
              :value="item"
              :key="item"
              :label="item"
              >{{ item }}</Option
            >
          </Select>
        </div>
        <div>游戏:
          <Select
            v-model="gameId"
            clearable
            filterable
            filter-by-label
          >
            <Option
              v-for="item in gameOptions"
              :value="item.number"
              :key="item.number"
              :label="item.label"
              >{{ item.label }}
            </Option>
          </Select>
          <p style="color:red;font-size:11pt;margin-top:20px">清理状态之前需要让玩家停止游戏<br>待执行清理操作后<b>10分钟</b>才能进入游戏，否则清理失败</p>
        </div>
      </Modal>
    </Card>
    <Card style="margin-top: 5px">
      <div
        class
        style="text-align: right; margin-bottom: 10px; margin-right: 15px"
        v-if="userInfo && userInfo.userType != 2"
      >
        <Button type="primary" size="small" @click="editBatchState(2)"
          >批量冻结</Button
        >
        <Button
          type="primary"
          size="small"
          style="margin-left: 10px"
          @click="editBatchState(1)"
          >批量解冻</Button
        >
      </div>
      <tables
        ref="tables"
        v-model="tableData"
        width="100%"
        :columns="columns"
        @on-selection-change="onSelect"
        @on-sort-change="sortProfit"
      />
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
import Cookies from 'js-cookie'
import * as dayjs from 'dayjs'
import Tables from '_c/tables'
import { setting } from '@/config'
import { getPlayerData, editPlayerState, agentScoreLog, clearPlayerGameState, getGameCurrency } from '@/api/data'
export default {
  name: 'players',
  components: {
    Tables
  },
  inject: ['viewAccess', 'handleLogOut'],
  data () {
    let _this = this
    return {
      startDate: dayjs(
        dayjs().startOf('month').format('YYYY-MM-DD 00:00:00')
      ).toDate(),
      startDateRestrict: {
        disabledDate (date) {
          if (date.getTime() > dayjs(_this.endDate).unix() * 1000) {
            return true
          }
          // 两个月以前
          if (date.getTime() < Date.now() - 1000 * 60 * 60 * 24 * 30 * 2) {
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

      remarks: undefined,
      order: '-totalProfLoss',
      userInfo: null,
      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],
      modal1: false,

      req: [
        { label: '三方ID:', key: 'userId', type: 'text', value: '' },
        { label: '昵称：', key: 'name', type: 'text', value: '' },
        { label: 'ID:', key: 'id', type: 'text', value: '' }
      ],
      columns: [
        {
          title: '',
          type: 'selection',
          width: 60,
          align: 'center',
          fixed: 'left'
        },
        {
          title: 'ID',
          key: 'id',
          width: 85,
          align: 'center',
          tooltip: true
        },
        {
          title: '昵称',
          key: 'nickName',
          width: 140,
          align: 'center',
          tooltip: true
        },
        {
          title: '账号',
          key: 'userId',
          width: 140,
          align: 'center',
          tooltip: true
        },
        {
          title: '试玩',
          key: 'isTourist',
          width: 80,
          align: 'center',
          render (h, params) {
            return params.row.isTourist>0 ? (<span style="color:red;">是</span>) : (<span style="color:green;">否</span>)
          }
        },
        {
          title: '代理',
          key: 'agentId',
          align: 'center',
          width: 100,
          render (h, params) {
            let name = ''
            _this.siteOption.map((x) => {
              x.agentList.map((y) => {
                if (y.id == params.row.agentId) {
                  name = y.name
                }
              })
            })

            return <span>{name}</span>
          }
        },

        {
          title: '最近登录时间',
          key: 'logTime',
          width: 170,
          sortable: true,
          sortType: 'desc',
          align: 'center',
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
        { title: '局数', key: 'totalNumber', align: 'center', width: 70 },
        {
          title: '有效下注',
          key: 'totalEffBet',
          width: 110,
          align: 'center',
          render (h, params) {
            return (
              <span>
                {(params.row.totalEffBet &&
                  params.row.totalEffBet.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '总盈利',
          key: 'totalProfLoss',
          width: 120,
          align: 'center',
          sortable: 'custom',
          render (h, params) {
            return (
              <span>
                {(params.row.totalProfLoss &&
                  (params.row.totalProfLoss).toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '区间局数',
          key: 'month_docCount',
          width: 130,
          render (h, params) {
            return <span>{params.row.month_docCount || 0}</span>
          }
        },
        {
          title: '区间有效投注',
          key: 'month_effectiveBets',
          width: 140,
          render (h, params) {
            return (
              <span>
                {(params.row.month_effectiveBets &&
                  params.row.month_effectiveBets.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '区间盈利',
          key: 'month_profitLoss',
          width: 140,
          render (h, params) {
            return (
              <span>
                {(params.row.month_profitLoss &&
                  (params.row.month_profitLoss).toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '区间杀数',
          key: 'month_profitLoss',
          width: 140,
          render (h, params) {
            if (
              params.row.month_profitLoss &&
              params.row.month_profitLoss != 0 &&
              params.row.month_effectiveBets &&
              params.row.month_effectiveBets != 0
            ) {
              return (
                <span>
                  {(
                    (params.row.month_profitLoss - params.row.month_effectiveBets) / params.row.month_effectiveBets
                  ).toFixed(3)}
                </span>
              )
            } else {
              return <span>0</span>
            }
          }
        },
        {
          title: '周期投注',
          key: 'cycle_effectiveBets',
          width: 140,
          render (h, params) {
            return (
              <span>
                {(params.row.cycle_effectiveBets &&
                  params.row.cycle_effectiveBets.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '周期盈利',
          key: 'cycle_profitLoss',
          width: 140,
          render (h, params) {
            return (
              <span>
                {(params.row.cycle_profitLoss &&
                  (params.row.cycle_profitLoss + params.row.cycle_effectiveBets).toFixed(2)) ||
                  0}
              </span>
            )
          }
        },
        {
          title: '状态',
          key: 'state',
          align: 'center',
          width: 80,
          render (h, params) {
            return params.row.state <= 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 250,
          fixed: 'right',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      let routeData = this.$router.resolve({
                        path: '/players-record-' + params.row.id,
                        query: { agent: params.row.agentId }
                      })
                      window.open(routeData.href, '_blank')
                    }
                  }
                },
                '流水'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      this.$Modal.info({
                        title: '上下分记录',
                        closable: true,
                        align: 'center',
                        width: 1200,
                        render: (h) => {
                          return [
                            h('Table', {
                              props: {
                                columns: [
                                  {
                                    title: '流水号',
                                    key: 'logOn',
                                    align: 'center',
                                    width: 200
                                  },
                                  {
                                    title: '时间',
                                    key: 'createtime',
                                    align: 'center',
                                    width: 200,
                                    render (h, params) {
                                      return (
                                        <span>
                                          {dayjs(
                                            params.row.createtime * 1000
                                          ).format('YYYY-MM-DD HH:mm:ss')}
                                        </span>
                                      )
                                    }
                                  },
                                  {
                                    title: '玩家ID',
                                    key: 'userId',
                                    align: 'center'
                                  },
                                  {
                                    title: '三方平台ID',
                                    key: 'openId',
                                    align: 'center'
                                  },
                                  {
                                    title: '帐变前金额',
                                    key: 'score',
                                    align: 'center'
                                  },
                                  {
                                    title: '上下分',
                                    key: 'logType',
                                    align: 'center',
                                    render (h, params) {
                                      return (
                                        <span>
                                          {params.row.logType == 1
                                            ? '上分'
                                            : '下分'}
                                        </span>
                                      )
                                    }
                                  },
                                  {
                                    title: '帐变后金额',
                                    key: 'userScore',
                                    align: 'center'
                                  }
                                ],
                                data: this.dataScore
                              }
                            }),
                            h('Page', {
                              style: {
                                marginTop: '10px'
                              },
                              props: {
                                total: this.newtotal,
                                pageSize: this.newpagesize
                              },
                              on: {
                                'on-change': (page) => {
                                  this.newpage = page
                                  let Data = {
                                    id: params.row.id,
                                    page: this.newpage,
                                    agentId: params.row.agentId
                                  }
                                  agentScoreLog(Data).then((res) => {
                                    this.dataScore = []
                                    this.dataScore.push(...res.data.data.data)
                                    this.newtotal = res.data.data.total
                                    this.newpagesize = res.data.data.page_size
                                  })
                                }
                              }
                            })
                          ]
                        }
                      })
                      let Data = {
                        id: params.row.id,
                        page: this.newpage,
                        agentId: params.row.agentId
                      }
                      agentScoreLog(Data).then((res) => {
                        this.dataScore = []
                        this.dataScore.push(...res.data.data.data)
                        this.newtotal = res.data.data.total
                        this.newpagesize = res.data.data.page_size
                      })
                    }
                  }
                },
                '上下分'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'warning',
                    size: 'small',
                    // to: "/players-game-" + params.row.id,
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      const params2 = { agent: params.row.agentId }
                      let routeData = this.$router.resolve({
                        path: '/players-game-' + params.row.id,
                        query: params2
                      })
                      window.open(routeData.href, '_blank')
                    }
                  }
                },
                '战绩'
              )
            },
            (h, params) => {
              if (this.viewAccess) {
                return [
                  h(
                    'Poptip',
                    {
                      props: {
                        transfer: true,
                        confirm: true,
                        placement: 'left',
                        title:
                          '您确定要' +
                          (params.row.state == 2 ? '解冻' : '冻结') +
                          '此玩家吗?'
                      },
                      style: { textAlign: 'left', zIndex: '99' },
                      on: {
                        'on-ok': () => {
                          let data = {
                            agentId: sessionStorage.getItem('agentVal'),
                            id: JSON.stringify([params.row.id]),
                            state: params.row.state == 2 ? 1 : 2
                          }
                          editPlayerState(data).then((res) => {
                            if (res.data.code == 200) {
                              this.$nextTick(() => {
                                this.userSelect = []
                                this.handleSearch()
                                this.$Message.success(res.data.msg)
                              })
                            } else {
                              this.$Message.error(res.data.msg)
                            }
                          })
                        }
                      }
                    },
                    [
                      h(
                        'Button',
                        {
                          props: {
                            type: params.row.state <= 1 ? 'error' : 'success',
                            size: 'small'
                          }
                        },
                        params.row.state <= 1 ? '冻结' : '解冻'
                      )
                    ]
                  )
                ]
              }
            }
          ]
        }
      ],
      tableData: [],
      userSelect: [],
      dataScore: [],
      newpage: 1,
      newtotal: 1,
      newpagesize: 1,
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      spinShow: false,
      bClearPlayerGameState: false,
      iClearPlayerGameId: '',
      gameId: 0,
      gameOptions: [],
      games: JSON.parse(sessionStorage.getItem('games') || '[]'),
      gameCurrency: [],
      currency: 'CNY'
    }
  },
  methods: {
    fClearPlayerGameState () {
      let Data = {
        playerId: this.iClearPlayerGameId,
        gameId: this.gameId,
        currency: this.currency
      }
      clearPlayerGameState(Data).then((res) => {
        if (res.data.code == 200) {
          this.$Message.success('清理成功')
        } else {
          this.$Message.error('清理失败')
        }
      })
    },
    /**
     * 查询代理业绩
     */
    viewDetail2 () {
      const route = {
        name: 'players-control'
      }

      let routeUrl = this.$router.resolve(route)
      window.open(routeUrl.href, '_blank')
    },
    viewClearPlayerGameState () {
      this.bClearPlayerGameState = true
    },
    onSelect (sel) {
      this.userSelect = sel.map((item) => {
        return item.id
      })
    },
    editBatchState (status) {
      if (!this.userSelect || this.userSelect.length == 0) {
        return
      }
      let Data = {
        agentId: sessionStorage.getItem('agentVal'),
        id: JSON.stringify(this.userSelect),
        state: status
      }
      editPlayerState(Data).then((res) => {
        if (res.data.code == 200) this.$Message.success('批量修改状态成功')
        this.$nextTick(() => {
          this.userSelect = []
          this.handleSearch()
        })
      })
    },
    setLinkage (i, key) {
      if (key == 'webId') {
        let [k, m] = [null, null]
        for (const j in this.req) {
          if (this.req[j].key == 'webId') k = j
          if (this.req[j].key == 'agentId') m = j
        }
        this.req[m].option = []
        this.req[m].value =
          sessionStorage.getItem('agentVal') != null
            ? Number(sessionStorage.getItem('agentVal'))
            : ''
        this.req[k].option.forEach((item) => {
          if (item.agentList.length > 0 && item.id == i) {
            this.req[m].option.push(...item.agentList)
          }
        })
      }
    },
    sortProfit (e) {
      this.order = e.order == 'desc' ? `-${e.key}` : e.key
      this.handleSearch()
    },

    handleSearch () {
      this.spinShow = true
      let st, et

      st = dayjs(
        dayjs(this.startDate.getTime()).format('YYYY-MM-DD 00:00:00')
      ).unix()

      et = dayjs(
        dayjs(this.endDate.getTime()).format('YYYY-MM-DD 23:59:59')
      ).unix()

      let Data = [
        { order: this.order },
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { webId: this.site },
        { agentId: this.agent === 9999999 ? undefined : this.agent },
        { startTime: st },
        { endTime: et }
      ]
      if (sessionStorage.getItem('siteVal')) {
        Data = [{ webId: Number(sessionStorage.getItem('siteVal')) }, ...Data]
      }
      if (sessionStorage.getItem('agentVal')) {
        Data = [
          { agentId: Number(sessionStorage.getItem('agentVal')) },
          ...Data
        ]
      }
      this.req.map((item) => {
        if (item.value !== '') {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      getPlayerData(Data)
        .then((res) => {
          this.spinShow = false
          if (res.data.code == 200) {
            this.tableData = []
            this.tableData.push(...res.data.data.data)
            this.pageData.current = res.data.data.total
          } else {
            // 判断响应状态是否为Token失效，如果失效则执行退出函数并刷新页面。
            this.$nextTick(() => {
              if (setting.arrStatus.indexOf(res.data.code) != -1) {
                this.$Message.error(
                  res.data.code + ' ：&nbsp;' + res.data.msg + '请重新登录'
                )
                this.handleLogOut()
              } else {
                this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
              }
            })
          }
        })
        .catch((err) => {
          this.spinShow = false
        })
    },
    handleAllSearch () {
      this.spinShow = true
      for (const i in this.req) {
        if (this.req[i].type === 'Checkbox') {
          this.req[i].value = 0
        } else {
          this.req[i].value = null
        }
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { webId: this.site }
      ]
      this.agent = null
      getPlayerData(Data).then((res) => {
        this.spinShow = false
        if (res.data.code == 200) {
          this.tableData = []
          this.pageData.page = 1
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
      if (this.agentOption.length > 1) {
        this.agentOption.map((item) => {
          item.label = item.name
        })
      }
    }

    // clearPlayerLock() {
    //   let _this = this;
    //   this.$Modal.confirm({
    //     title: "清理玩家锁",
    //     render: (h) => {
    //       return (
    //         <div>
    //           <Input
    //             value={_this.value19}
    //             onInput={(e) => {
    //               _this.value19 = e;
    //             }}
    //           >
    //             <span slot="prepend">用户账号:</span>
    //           </Input>
    //           <div style="color:red;margin-top:10px">
    //             <b>说明:</b>用户账号在这里指的是<b>第三方平台的账号</b>，多个账号之间用<b>英文逗号(,)</b>分割
    //           </div>
    //         </div>
    //       );
    //     },
    //     async onOk() {
    //       if (!_this.value19) {
    //         return;
    //       }

    //       let params = {
    //         token: getToken(),
    //         ids: _this.value19,
    //       };
    //       let data = await axios.request({
    //         url: "v2/govern/clearLock",
    //         method: "get",
    //         params,
    //       });
    //       if (data && data.data && data.data.code == 200) {
    //         _this.$Message.info("设置成功");
    //       }
    //     },
    //     onCancel() {},
    //   });
    // },
  },
  mounted () {
    window.testis = this
    let mountedfunc = () => {
      this.userInfo = JSON.parse(Cookies.get('userInfo'))
      let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
      let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
      this.siteOption = siteOption
      this.site = sid * 1
      // 把选择的站点赋值到页面选中开始
      siteOption &&
        siteOption.map((item, index) => {
          if (item.id == sid) {
            this.agentOption = item.agentList || []
            this.agentOption.map((item) => {
              item.value = item.id
              item.label = item.name
            })
            this.agent = 9999999
            this.agentOption.unshift({
              name: '全部',
              id: 9999999,
              value: 9999999,
              label: '全部'
            })
            this.agentOption.map((item) => {
              item.label = item.name
            })
          }
        })
      let newArray = this.games
      let a = []
      a.push(...newArray)
      a.map((item) => {
        if (item.nameZH.length > 0) {
          item.label = item.name + ' [' + item.nameZH + ']'
        } else {
          item.label = item.name
        }
        return item
      })
      this.gameOptions = a
      this.siteOption.map((item) => {
        item.label = item.name
      })
      // 把选择的站点赋值到页面选中结束
      this.handleSearch()
      this.req.map((item) => {
        if (item.key == 'webId') {
          item.option.push(
            ...Object.assign(JSON.parse(sessionStorage.getItem('siteOption')))
          )
          if (sessionStorage.getItem('siteVal')) {
            item.value = Number(sessionStorage.getItem('siteVal'))
            this.$nextTick(() => {
              this.setLinkage(item.value, 'webId')
            })
          }
        }
        if (item.key == 'agentId' && sessionStorage.getItem('agentVal')) {
          item.value = Number(sessionStorage.getItem('agentVal'))
        }
        if (item.key == 'gameType') {
          item.option.push(
            ...Object.assign(JSON.parse(sessionStorage.getItem('typeOption')))
          )
        }
      })
    }

    let timer = setInterval(() => {
      if (sessionStorage.getItem('siteOption')) {
        clearInterval(timer)
        mountedfunc()
      }
    }, 500)

    getGameCurrency().then((res) => {
      if (res.data.code == 200) {
        let tmp = res.data.data
        if (typeof tmp === 'string') {
          tmp = JSON.parse(tmp)
        }
        if (tmp && tmp.currency) {
          this.gameCurrency = Object.keys(tmp.currency)
        }
      }
    })
  }
}
</script>

<style lang="less">
.search {
  display: inline;
  > button {
    margin-right: 10px;
    padding-left: 20px;
    padding-right: 20px;
  }
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
