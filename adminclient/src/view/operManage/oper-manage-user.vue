<template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{ item.label }}</label>
          <Input
            v-model="item.value"
            :maxlength="50"
            placeholder="请输入"
            style="width: 180px"
          />
        </li>

        <div class="search">
          <!-- <Button type="primary" @click="historyControl()">历史控制</Button> -->
          <Button type="primary" @click="searchAction">搜索</Button>
          <Button @click="resetParams">重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <div style="text-align: right;;margin-bottom: 10px;">
      <Button
        type="primary"
        style="margin-right: 60px;"
        @click="changeWinProb(0)"
        >批量新增</Button
      >
      <Button type="primary" style @click="changeWinProb('cancle')"
        >批量取消</Button
      >
    </div>
    <Card>
      <tables
        ref="tables"
        v-model="tableData"
        :columns="columns"
        @on-selection-change="onSelect"
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
    <div></div>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import {
  getControlUserData,
  addControlUserProb,
  removeControlUserProb,
  getHistoryControlUserData
} from '@/api/data'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  props: ['agentId', 'siteId'],
  data () {
    return {
      req: [{ label: '玩家ID', key: 'userId', value: '' }],
      columns: [
        { title: '', type: 'selection', width: 60, align: 'center' },
        { title: '玩家ID', key: 'id', align: 'center' },
        { title: '玩家昵称', key: 'nickName', align: 'center' },
        { title: '代理ID', key: 'agentId', align: 'center', width: 80 },
        {
          title: '胜率',
          key: 'winProb',
          align: 'center'
        },
        // { title: "设置输赢", key: "score", align: "center" },
        // { title: "剩余输赢", key: "incScore", align: "center" },
        { title: '控次', key: 'controlNumber', align: 'center' },
        // { title: "累计盈亏", key: "profitLoss", align: "center" },
        { title: '操作人', key: 'AdminName', align: 'center' },
        {
          title: '操作时间',
          key: 'createTime',
          width: 180,
          align: 'center',
          render (h, params) {
            if (params.row.createTime) {
              return (
                <span>
                  {new Date(
                    params.row.createTime * 1000
                  ).toLocaleString('chinese', { hour12: false })}
                </span>
              )
            }
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: params.row.controlId ? 'warning' : 'info',
                    size: 'small'
                  },
                  on: {
                    click: () => {
                      this.changeWinProb(params.row.controlId, params.row.id)
                    }
                  }
                },
                params.row.controlId ? '取消' : '新增'
              )
            }
          ]
        }
      ],
      winProb: {
        ids: [],
        controlIds: [],
        winProb: 0,
        score: 0
      },
      tableData: [],
      historyColumns: [
        { title: '控制ID', key: 'id', align: 'center', width: 80 },
        {
          title: '用户ID',
          key: 'userId',
          align: 'center'
        },
        {
          title: '胜率',
          key: 'winProb',
          align: 'center'
        },
        {
          title: '设置的输赢分',
          key: 'score',
          align: 'center'
        },
        {
          title: '操作人',
          key: 'adminName',
          align: 'center'
        },
        {
          title: '操作时间',
          key: 'createTime',
          width: 180,
          align: 'center',
          render (h, params) {
            if (params.row.createTime) {
              return (
                <span>
                  {new Date(
                    params.row.createTime * 1000
                  ).toLocaleString('chinese', { hour12: false })}
                </span>
              )
            }
          }
        }
      ],
      historyTableData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      historyPageData: {
        total: 0,
        current: 1,
        pageSize: 10
      },
      spinShow: false
    }
  },
  methods: {
    onSelect (sel) {
      this.winProb.ids = []
      this.winProb.ids = sel.map(item => {
        return item.id
      })

      this.winProb.controlIds = []
      this.winProb.controlIds = sel.map(item => {
        return item.controlId
      })
    },
    historyControl () {
      this.$nextTick(() => {
        let Data = {
          page: 1,
          pageSize: 10,
          agentId: this.agentId
        }
        this.handleHistorySearch(Data)
        this.$Modal.info({
          title: '历史控制记录',
          width: '1280px',
          closable: true,
          render: h => {
            return [
              h('Input', {
                props: {
                  search: true,
                  enterButton: true,
                  placeholder: '搜索玩家ID'
                },
                style: {
                  margin: '10px',
                  width: '250px',
                  position: 'absolute',
                  right: '30px',
                  top: '-50px'
                },
                on: {
                  'on-search': val => {
                    Data = {
                      keys: val,
                      page: 1,
                      pageSize: 10,
                      agentId: this.agentId
                    }
                    this.handleHistorySearch(Data)
                  }
                }
              }),
              h('Table', {
                props: {
                  columns: this.historyColumns,
                  data: this.historyTableData
                }
              }),
              h('Page', {
                ref: 'pageData',
                props: {
                  total: this.historyPageData.total,
                  current: this.historyPageData.current,
                  pageSize: this.historyPageData.pageSize,
                  showTotal: true
                },
                style: { float: 'left', margin: '20px 0' },
                on: {
                  'on-change': val => {
                    Data.page = val
                    this.handleHistorySearch(Data)
                  }
                }
              })
            ]
          }
        })
      })
    },
    changeWinProb (params, singleId) {
      if (!params) {
        this.$Modal.confirm({
          title: '新增玩家胜率',
          okText: '新增胜率',
          closable: true,
          onOk: () => {
            let Data = {}
            Data = {
              agentId: this.agentId,
              userId: singleId
                ? `[${singleId}]`
                : JSON.stringify(this.winProb.ids),
              winProb: this.winProb.winProb,
              score: this.winProb.score
            }

            let dataFormatting = JSON.parse(Data.userId)
            if (
              dataFormatting.length == 0 ||
              dataFormatting.find(x => !x) != undefined
            ) {
              this.$Message.error('请选择未添加控制玩家')
              return
            }

            addControlUserProb(Data).then(res => {
              if (res.data.code == 200) {
                this.winProb.ids = []
                this.winProb.controlIds = []
                this.$Message.success('修改胜率成功')
                this.$nextTick(() => {
                  this.handleSearch()
                })
              } else {
                this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
              }
            })
          },

          render: h => {
            return [
              h(
                'label',
                {
                  style: { fontSize: '16px' }
                },
                '胜率'
              ),
              h('Slider', {
                props: {
                  value: this.winProb.winProb,
                  showInput: true,
                  min: -100,
                  max: 100
                },
                on: {
                  input: val => {
                    this.winProb.winProb = val
                  }
                }
              })
              // h(
              //   "span",
              //   {
              //     style: { fontSize: "16px" }
              //   },
              //   "输赢分数"
              // ),
              // h("Input", {
              //   props: {
              //     value: this.winProb.score
              //   },
              //   on: {
              //     input: val => {
              //       this.winProb.score = val;
              //     }
              //   }
              // })
            ]
          }
        })
      } else {
        this.$Modal.confirm({
          title: '取消胜率控制',
          okText: '移除控制',
          onOk: () => {
            let Data = {
              id: `[${params}]`,
              agentId: this.agentId
            }
            if (params == 'cancle') {
              Data.id = JSON.stringify(this.winProb.controlIds)
            }

            let dataFormatting = JSON.parse(Data.id)
            if (
              dataFormatting.length == 0 ||
              dataFormatting.find(x => !x) != undefined
            ) {
              this.$Message.error('请选择已有控制玩家')
              return
            }

            removeControlUserProb(Data).then(res => {
              this.winProb.ids = []
              this.winProb.controlIds = []
              if (res.data.code == 200) {
                this.$Message.success('移除控制成功')
                this.$nextTick(() => {
                  this.handleSearch()
                })
              } else {
                this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
              }
            })
          }
        })
      }
    },

    handleSearch () {
      this.spinShow = true
      let Data = [
        { agentId: this.agentId },
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      this.req.map(item => {
        if (item.value) {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      // debugger; // $$ debug
      getControlUserData(Data).then(res => {
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
    },
    resetParams () {
      this.req[0].value = ''
    },
    handleAllSearch () {
      this.spinShow = true

      let Data = [
        {agentId: sessionStorage.getItem('agentVal'), page: this.pageData.page, pageSize: this.pageData.pageSize }
      ]

      getControlUserData(Data).then(res => {
        this.spinShow = false
        this.tableData = []
        this.tableData.push(...res.data.data.data)
        this.pageData.current = res.data.data.total
      })

      this.req[0].value = ''
    },
    handleHistorySearch (Data) {
      getHistoryControlUserData(Data).then(res => {
        if (res.data.code == 200) {
          this.historyTableData = []
          this.historyPageData.total = res.data.data.total
          this.historyTableData.push(...res.data.data.data)
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
    changeHistoryPage (index) {
      this.pageData.page = index
      this.handleSearch()
    },
    searchAction () {
      this.pageData.page = 1
      this.handleSearch()
    }
  },
  mounted () {
    this.handleSearch()
  }
}
</script>

<style>
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
