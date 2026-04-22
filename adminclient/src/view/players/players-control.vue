<template>
  <div>
    <Card>
      <div class="inlineForm">
        <ul>
          <li
            class="label-style"
            v-for="item in req"
            :key="item.label"
            style="margin-right: 6px"
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

            <Checkbox
              v-if="item.type == 'Checkbox'"
              :true-value="1"
              :false-value="0"
              v-model="item.value"
              >是否受控</Checkbox
            >
          </li>

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
          <div class="label-style">
            <Button
              type="primary"
              @click="searchAction"
              style="margin-right: 15px"
              >搜索</Button
            >
            <Button @click="handleAllSearch" style="margin-right: 15px"
              >重置</Button
            >
            <Button
              type="primary"
              @click="showAutoSingleControlSettingPage"
              style="margin-left: 10px"
              >自动单控条件设置</Button
            >
          </div>
        </ul>
      </div>

      <Modal
        width="1100px"
        v-model="showAutoSingleControlSetting"
        title="玩家自动控制配置"
        @on-ok="saveAutoSingleControlParams"
        okText="保存"
      >
        <Card>
          <ul style="list-style: none">
            <li style="display: inline; width: 180px; margin-right: 6px">
              <span style="margin-right: 6px">盈利金额:</span>
              <InputNumber
                style="width: 100px"
                type="number"
                v-model.number="currentAutoSingleControlParams.totalProfLoss"
                :min="0"
                :step="100"
              />
            </li>
            <li style="display: inline; width: 180px; margin-right: 6px">
              <span style="margin-right: 6px">盈利百分比:</span>
              <InputNumber
                style="width: 100px"
                type="number"
                v-model.number="
                  currentAutoSingleControlParams.totalProfLossRate
                "
                :min="0"
                :max="100"
                :step="1"
              />
            </li>
            <li style="display: inline; width: 180px; margin-right: 6px">
              <span style="margin-right: 6px">打码量:</span>
              <InputNumber
                style="width: 100px"
                type="number"
                v-model.number="currentAutoSingleControlParams.totalEffect"
                :min="0"
                :step="100"
              />
            </li>
            <li style="display: inline; width: 180px; margin-right: 6px">
              <span style="margin-right: 6px">控制概率:</span>
              <InputNumber
                style="width: 100px"
                type="number"
                v-model.number="currentAutoSingleControlParams.controlRate"
                :min="-1"
                :step="0.01"
              />
            </li>
            <li style="display: inline; width: 180px; margin-right: 6px">
              <span style="margin-right: 6px">控制分数:</span>
              <InputNumber
                style="width: 100px"
                type="number"
                v-model.number="currentAutoSingleControlParams.score"
                :value="0"
                :min="0"
                :step="100"
              />
            </li>
            <li
              style="
                display: inline;
                margin-right: 6px;
                width: 180px;
                padding-left: 20px;
              "
            >
              <Button type="primary" @click="addAutoSingleControlParams"
                >添加</Button
              >
            </li>
          </ul>
        </Card>

        <Card style="margin: 10px 0">
          <Table
            :height="350"
            :columns="singleControlConfig"
            :data="this.AutoSingleControlParams"
          ></Table>
        </Card>
      </Modal>
    </Card>
    <br />

    <Card>
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
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'

import Cookies from 'js-cookie'
import * as dayjs from 'dayjs'
import Tables from '_c/tables'
import { setting } from '@/config'
import { getPlayerData, editPlayerState, getUserRecord } from '@/api/data'
export default {
  name: 'players',
  components: {
    Tables
  },
  inject: ['viewAccess', 'handleLogOut'],
  data () {
    let _this = this
    return {
      remarks: undefined,
      order: '-totalProfLoss',
      userInfo: null,
      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],
      modal1: false,

      req: [
        { label: '玩家ID：', key: 'userId', type: 'text', value: '' },
        { label: '玩家昵称：', key: 'name', type: 'text', value: '' },
        {
          label: '',
          key: 'inCtl',
          type: 'Checkbox',
          value: (() => {
            if (window.isctl_temp) {
              window.isctl_temp = 0
              return 1
            } else {
              return 0
            }
          })()
        }
      ],
      singleControlConfig: [
        {
          title: '打码量',
          key: 'totalEffect',
          align: 'center',
          width: 140
        },
        {
          title: '盈亏',
          key: 'totalProfLoss',
          align: 'center'
        },
        {
          title: '盈亏比例',
          key: 'totalProfLossRate',
          align: 'center'
        },
        {
          title: '控制概率',
          key: 'controlRate',
          align: 'center'
        },
        {
          title: '控制分数',
          key: 'score',
          align: 'center'
        },
        {
          title: '操作',
          key: 'gameNumber',
          align: 'center',
          width: 140,
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.$Modal.confirm({
                        title: '确定要删除该配置吗？',
                        async onOk () {
                          let tmp = []
                          _this.AutoSingleControlParams.forEach((item) => {
                            if (
                              params.row.totalEffect == item.totalEffect &&
                              params.row.totalProfLoss == item.totalProfLoss &&
                              params.row.totalProfLossRate ==
                                item.totalProfLossRate &&
                              params.row.score == item.score &&
                              params.row.controlRate == item.controlRate
                            ) {

                            } else {
                              tmp.push(item)
                            }
                          })
                          _this.AutoSingleControlParams = tmp
                          _this.saveAutoSingleControlParams()
                        },
                        onCancel () {}
                      })
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],
      columns: [
        {
          title: '玩家ID',
          key: 'id',
          width: 85,
          align: 'center',
          tooltip: true
        },
        {
          title: '玩家昵称',
          key: 'nickName',
          width: 140,
          align: 'center',
          tooltip: true
        },
        {
          title: '代理',
          key: 'agentId',
          align: 'center',
          width: 100,
          render (h, params) {
            return (
              <span>
                {
                  _this.agentOption.find((x) => x.id == params.row.agentId)[
                    'name'
                  ]
                }
              </span>
            )
          }
        },

        {
          title: '有效下注',
          key: 'totalEffBet',
          width: 120,
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
          title: '总盈亏',
          key: 'totalProfLoss',
          width: 130,
          align: 'center',
          sortable: 'custom',
          render (h, params) {
            return (
              <span>
                {(params.row.totalProfLoss &&
                  params.row.totalProfLoss.toFixed(2)) ||
                  0}
              </span>
            )
          }
        },

        {
          title: '状态',
          key: 'state',
          width: 80,
          align: 'center',
          render (h, params) {
            return params.row.state <= 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        },
        {
          title: '控制系数',
          key: 'rate',
          width: 140,
          render (h, params) {
            return <span>{params.row.rate ? params.row.rate : 0}</span>
          }
        },
        {
          title: '控制分数',
          key: 'rate_score',
          width: 140,
          render (h, params) {
            return (
              <span>{params.row.rate_score ? params.row.rate_score : 0}</span>
            )
          }
        },
        {
          title: '剩余控制分数',
          key: 'left_score',
          render (h, params) {
            return (
              <span>{params.row.left_score ? params.row.left_score : 0}</span>
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 150,
          fixed: 'right',
          button: [
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
                        title: '控制记录',
                        closable: true,
                        align: 'center',
                        width: 700,
                        render: (h) => {
                          return [
                            h('Table', {
                              props: {
                                columns: [
                                  {
                                    title: '时间',
                                    key: 'createtime',
                                    align: 'center',
                                    render (h, params) {
                                      return (
                                        <span>
                                          {dayjs(
                                            params.row.createTime * 1000
                                          ).format('YYYY-MM-DD HH:mm')}
                                        </span>
                                      )
                                    }
                                  },
                                  {
                                    title: '系数',
                                    key: 'rate',
                                    align: 'center'
                                  },
                                  {
                                    title: '设置分数',
                                    key: 'rate_score',
                                    align: 'center'
                                  },
                                  {
                                    title: '控制类型',
                                    key: 'ctrl_type',
                                    align: 'center',
                                    render (h, params) {
                                      return (
                                        <span>
                                          {params.row.ctrl_type == 0
                                            ? '自动控制'
                                            : '系统控制'}
                                        </span>
                                      )
                                    }
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
                                    userId: params.row.id,
                                    page: this.newpage
                                  }
                                  getUserRecord(Data).then((res) => {
                                    this.dataScore = []
                                    this.dataScore.push(...res.data.data.data)
                                    this.newtotal = res.data.data.total
                                    this.newpagesize = res.data.data.pageSize
                                  })
                                }
                              }
                            })
                          ]
                        }
                      })
                      let Data = {
                        userId: params.row.id,
                        page: this.newpage
                      }
                      getUserRecord(Data).then((res) => {
                        this.dataScore = []
                        this.dataScore.push(...res.data.data.data)
                        this.newtotal = res.data.data.total
                        this.newpagesize = res.data.data.pageSize
                      })
                    }
                  }
                },
                '记录'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'warning',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      this.$Modal.info({
                        title: '设置',
                        closable: true,
                        align: 'center',
                        width: 600,
                        render: (h) => {
                          return [
                            <div
                              style={{
                                paddingLeft: '140px'
                              }}
                            >
                              <div
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  paddingRight: '20px'
                                }}
                              >
                                <div>控制系数：</div>
                                <InputNumber
                                  max={1}
                                  min={-1}
                                  step={0.001}
                                  vModel={_this.userControlRate}
                                  placeholder="请输入"
                                  style="width: 120px"
                                />
                              </div>
                              <p style="padding:10px">
                                0为默认不控制，填值范围为-1至1
                              </p>

                              <div
                                style={{
                                  display: 'flex',
                                  alignItems: 'center'
                                }}
                              >
                                <div>控制分数：</div>
                                <InputNumber
                                  max={50000000}
                                  min={0}
                                  step={1}
                                  vModel={_this.userControlRateScore}
                                  placeholder="请输入"
                                  style="width: 120px"
                                />
                              </div>
                            </div>,
                            <div
                              style={{
                                paddingLeft: '140px'
                              }}
                            >
                              <p style="padding:10px">
                                0为默认不控制，填值范围为0-50000000
                              </p>
                            </div>
                          ]
                        },
                        async onOk () {
                          if (
                            _this.userControlRate == 0 &&
                            _this.userControlRateScore == 0
                          ) {
                            let data = {
                              token: getToken(),
                              userId: params.row.id,
                              rate: _this.userControlRate,
                              rate_score: _this.userControlRateScore
                            }

                            let res = await axios.request({
                              url: 'v2/govern/user',
                              method: 'post',
                              params: data
                            })

                            if (res.data.code == 200) {
                              this.$Message.success('取消修改成功')
                              _this.searchAction()
                            }

                            return
                          }

                          if (
                            _this.userControlRate == 0 ||
                            _this.userControlRateScore == 0
                          ) {
                            this.$Message.error('控制系数和分数必须都不为零！')
                            return
                          }

                          let data = {
                            token: getToken(),
                            userId: params.row.id,
                            rate: _this.userControlRate,
                            rate_score: _this.userControlRateScore
                          }

                          let res = await axios.request({
                            url: 'v2/govern/user',
                            method: 'post',
                            params: data
                          })

                          if (res.data.code == 200) {
                            this.$Message.success('设置成功')
                            _this.searchAction()
                          }
                        }
                      })
                    }
                  }
                },
                '设置'
              )
            }
          ]
        }
      ],
      userControlRate: 0,
      userControlRateScore: 0,
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
      showAutoSingleControlSetting: false,
      AutoSingleControlParams: [],
      currentAutoSingleControlParams: {
        totalProfLoss: 0,
        totalProfLossRate: 0,
        totalEffect: 0,
        controlRate: 0,
        score: 0
      }
    }
  },
  methods: {
    showAutoSingleControlSettingPage () {
      this.showAutoSingleControlSetting = true
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
    async addAutoSingleControlParams () {
      let isOk = true
      this.AutoSingleControlParams.forEach((element) => {
        if (
          element.controlRate == this.currentAutoSingleControlParams.controlRate
        ) {
          isOk = false
        }
      })
      if (isOk) {
        this.AutoSingleControlParams.push({
          totalProfLoss: this.currentAutoSingleControlParams.totalProfLoss,
          totalProfLossRate:
            this.currentAutoSingleControlParams.totalProfLossRate,
          totalEffect: this.currentAutoSingleControlParams.totalEffect,
          controlRate: this.currentAutoSingleControlParams.controlRate,
          score: this.currentAutoSingleControlParams.score
        })
      } else {
        this.$Message.error('已经有相同的记录了!')
      }
    },
    async saveAutoSingleControlParams () {
      let params = {
        token: getToken(),
        asc: this.AutoSingleControlParams
      }
      let data = await axios.request({
        url: 'v2/user/saveAutoSingleControl',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.$Message.info('设置成功')
      } else {
        this.$Message.error(data.data.msg)
      }
    },
    async getAutoSingleControlParams () {
      let params = {
        token: getToken(),
        asc: this.AutoSingleControlParams
      }
      let data = await axios.request({
        url: 'v2/user/getAutoSingleControl',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.AutoSingleControlParams = JSON.parse(data.data.data)
      } else {
        this.$Message.error(data.data.msg)
      }
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

      let Data = [
        { order: this.order },
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { webId: this.site },
        { agentId: this.agent === 9999999 ? undefined : this.agent }
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
      this.agentOption.map((item) => {
        item.label = item.name
      })
    }
  },
  mounted () {
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
    this.siteOption.map((item) => {
      item.label = item.name
    })
    this.getAutoSingleControlParams()
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
  },
  watch: {
    userControlRateScore (n, o) {
      if (String(n).includes('.')) {
        setTimeout(() => {
          this.userControlRateScore = Math.floor(n)
        }, 1)
      }
    }
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
