<template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{ item.label }}</label>
          <Select
            style="width: 180px"
            v-model="item.value"
            v-if="item.type == 'select'"
          >
            <Option
              v-for="items in item.option[0]"
              :value="items.id"
              :key="items.id"
              >{{ items.name }}</Option
            >
          </Select>
          <Input
            v-model="item.value"
            v-if="item.type == 'text'"
            placeholder="请输入"
            style="width: 180px"
          />
          <DatePicker
            type="datetime"
            v-if="item.type == 'datetime'"
            v-model="item.value"
            :placeholder="'请选择' + item.label"
          ></DatePicker>
        </li>

        <div class="search" style="width: 200px">
          <Button type="primary" @click="handleSearch">搜索</Button>
          <Button @click="handleAllSearch">重置</Button>
        </div>
        <div style="width: 300px; float: right">
          <Button type="primary" to="/agent-add">创建代理商</Button>
          <Button type="primary" style="margin-left: 20px" to="/agent-domain"
            >代理域名设置</Button
          >
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <tables
        ref="tables"
        editable
        @on-select="selectedAgent"
        v-model="tableData"
        :columns="columns"
        :rowClassName="rowClassName"
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
import Tables from '_c/tables'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import {
  getAgentData,
  editAgentData,
  getAgentInfo,
  agentScoreLog
} from '@/api/data'
import Edits from '_c/edit-data'
let self = {}
export default {
  name: 'agent',
  components: {
    Tables,
    Edits
  },
  inject: ['viewAccess', 'handleLogOut', 'reFreshSiteAgentList'],
  data () {
    return {
      selectionAgents: [],
      req: [{ label: '代理名', type: 'text', key: 'name', value: '' }],
      columns: [
        {
          title: '序号',
          key: 'id',
          width: 80,
          align: 'center',
          isdisabled: true
        },
        {
          title: '代理名',
          key: 'nickName',
          align: 'center',
          minWidth: 100,
          showRedPoint: true
        },
        { title: '所属站点', key: 'webName', align: 'right', minWidth: 100 },
        {
          title: 'Key',
          key: '',
          align: 'center',
          minWidth: 100,
          render (h, params) {
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
                    self.showInfoKey(params.row)
                  }
                }
              },
              '查看'
            )
          }
        },
        { title: '备注', key: 'remarks', align: 'center', minWidth: 100 },
        {
          title: '操作',
          key: 'handle',
          width: 180,
          align: 'center',
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
                      getAgentInfo({ id: params.row.id }).then((res) => {
                        params.row = res.data.data
                        this.$Modal.confirm({
                          title: '修改数据',
                          closable: true,
                          width: 800,
                          onOk: () => {
                            editAgentData(this.editData).then((res) => {
                              if (res.data.code == 200) {
                                this.editData = []
                                this.$nextTick(() => {
                                  this.$Message.success('编辑代理成功')
                                  this.reFreshSiteAgentList()
                                  this.handleSearch()
                                })
                              } else {
                                this.$Message.error(
                                  '编辑失败：' + res.data.msg
                                )
                              }
                            })
                          },
                          render: (h) => {
                            return h(Edits, {
                              props: {
                                row: params.row,
                                columns: this.columns
                                  .filter((x) => x.title != 'Key')
                                  .filter((x) => x.title != '上级代理')
                              },
                              on: {
                                sendEditData: (data) => {
                                  data = data.map((item) => {
                                    if (item.value != null) {
                                      if (item.key.indexOf('Time') != -1) {
                                        if (item.value == 'Invalid Date') {
                                          item.value = ''
                                        } else {
                                          item.value = getDate(item.value)
                                        }
                                      }
                                      return { [item.key]: item.value }
                                    }
                                  })
                                  this.editData = Object.assign(...data, {
                                    id: params.row.id
                                  })
                                }
                              }
                            })
                          }
                        })
                      })
                    }
                  }
                },
                '编辑'
              )
            },
            (h, params) => {
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
                        (params.row.isFrozen == 0 ? '冻结' : '解冻') +
                        '代理吗?'
                    },
                    style: {
                      textAlign: 'left',
                      marginRight: '5px',
                      zIndex: '99'
                    },
                    on: {
                      'on-ok': () => {
                        let data = {
                          id: params.row.id,
                          isFrozen: params.row.isFrozen == 0 ? 1 : 0,
                          isFrozenType: 1
                        }

                        editAgentData(data).then((res) => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
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
                          type: params.row.isFrozen == 0 ? 'error' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.isFrozen == 0 ? '冻结' : '解冻'
                    )
                  ]
                )
              ]
            }
            // (h, params) => {
            //   return h(
            //     "Button",
            //     {
            //       props: {
            //         type: "info",
            //         size: "small"
            //       },
            //       style: {
            //         marginRight: "5px"
            //       },
            //       on: {
            //         click: () => {
            //           generateCustom(sessionStorage.getItem("agentVal")).then(
            //             res => {
            //               if(res.status == 200){
            //                   this.$Message.success("操作成功");
            //               };
            //             }
            //           );
            //         }
            //       }
            //     },
            //     "生成配置"
            //   );
            // }
          ]
        }
      ],
      tableData: [],
      editData: [],
      showMsg: false,
      addPoint: 0,
      columnsPoint: [
        {
          title: '时间',
          key: 'createTime',
          width: 180,
          align: 'center',
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        { title: '上分人', key: 'adminName', align: 'center' },
        {
          title: '来源',
          key: 'logType',
          align: 'center',
          render (h, parmas) {
            return parmas.row.logType == 1 ? (
              <span>增加点数</span>
            ) : (
              <span>{parmas.row.logType}</span>
            )
          }
        },
        { title: '点数', key: 'point', align: 'center' },
        { title: '增加前点数', key: 'originalPoint', align: 'center' },
        { title: '增加后点数', key: 'nowPoint', align: 'center' }
      ],
      dataPoint: [],
      columnsScore: [
        { title: '流水号', key: 'logOn', align: 'center' },
        { title: '时间', key: 'createTime', align: 'center' },
        { title: '玩家ID', key: 'userId', align: 'center' },
        { title: '第三方代理玩家ID', key: 'openId', align: 'center' },
        { title: '帐变前金额', key: 'score', align: 'center' },
        { title: '上下分', key: 'logType', align: 'center' },
        { title: '帐变后金额', key: 'userScore', align: 'center' }
      ],
      dataScore: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      spinShow: false
    }
  },

  methods: {
    selectedAgent (selection, row) {
      this.selectionAgents = selection
    },

    stopAllAgent () {
      let _this = this
      this.$Modal.confirm({
        title: '关闭系统',
        content: '确定一键关闭系统吗？',
        async onOk () {
          let data = await axios.request({
            url: 'v1/agent/stop',
            method: 'get',
            params: {
              token: getToken()
            }
          })
          if (data.data.code == 200) {
            _this.$Message.success(data.data.msg)
            _this.handleSearch()
          }
        },
        onCancel () {}
      })
    },

    async startAllAgent () {
      let _this = this
      this.$Modal.confirm({
        title: '开启系统',
        content: '确定一键开启系统吗？',
        async onOk () {
          let data = await axios.request({
            url: 'v1/agent/start',
            method: 'get',
            params: {
              token: getToken()
            }
          })

          if (data.data.code == 200) {
            _this.$Message.success(data.data.msg)
            _this.handleSearch()
          }
        },
        onCancel () {}
      })
    },

    rowClassName (row, index) {
      if (Number(row.point) <= Number(row.warningPoint)) {
        return 'demo-table-error-row'
      }
      return ''
    },
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    },
    handleSearch (cb) {
      this.selectionAgents = []
      this.spinShow = true
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      this.req.map((item) => {
        if (item.value) {
          if (item.type == 'datetime') {
            Data.push({
              [item.key]: getDate(item.value)
            })
          } else {
            Data.push({
              [item.key]: item.value
            })
          }
        }
      })
      getAgentData(Data).then((res) => {
        this.spinShow = false
        this.tableData = []
        if (res.data.code == 200) {
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
    handleAllSearch () {
      this.spinShow = true
      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      getAgentData(Data).then((res) => {
        this.spinShow = false
        this.tableData = []
        this.pageData.page = 1
        if (res.data.code == 200) {
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
    // 显示代理游戏
    showInfo (data) {
      getAgentInfo({ id: data.id }).then((res) => {
        data = res.data.data
        let game = data.gameList.map((item) => {
          // let title = { 1: "乐游", 2: "开元", 3: "红包" }[item.id] || item.id;
          let content = item ? item.nameList : ''
          return "<p class='gameList'>" + content + '</p>'
        })
        self.$Modal.info({
          title: '代理的游戏',
          closable: true,
          content: JSON.stringify(game)
            .replace(/\[|]|"/g, '')
            .replace(/,/g, '&emsp;'),
          onOk: () => {}
        })
      })
    },

    // 显示三个Key
    showInfoKey (data) {
      getAgentInfo({ id: data.id }).then((res) => {
        data = res.data.data

        let html = `
        <h3>aesKey</h3><p class='gameList'>${data.aesKey}</p>
        <h3>agentKey</h3><p class='gameList'>${data.agentKey}</p>
        <h3>md5Key</h3><p class='gameList'>${data.md5Key}</p>
        `
        self.$Modal.info({
          title: '代理的游戏',
          closable: true,
          content: html,
          width: 500,
          onOk: () => {}
        })
      })
    },

    // 显示代理控分记录
    showScoreLog (id) {
      this.$Modal.confirm({
        title: '代理控分记录',
        closable: true,
        align: 'center',
        width: 1200,
        render: (h) => {
          return [
            h('Table', {
              props: {
                columns: this.columnsScore,
                data: this.dataScore
              }
            })
          ]
        }
      })
      let Data = { id: id }
      agentScoreLog(Data).then((res) => {
        this.dataScore = []
        // this.dataScore.push(...res.data.data.data);
        // totalPoint = res.data.data.totalPoint;
      })
    }
  },
  created () {
    self = this
    this.columns.forEach((item, index) => {
      if (item.key == 'handle') {
        if (!this.viewAccess) {
          this.columns.splice(index, 1)
        }
      }
    })
  },
  mounted () {
    this.$nextTick(() => {
      this.handleSearch()
    })
    if (sessionStorage.getItem('siteOption').length) {
      this.req.forEach((item) => {
        if (item.key == 'webId') {
          item.option.push(
            Object.assign(JSON.parse(sessionStorage.getItem('siteOption')))
          )
          if (sessionStorage.getItem('siteVal')) {
            this.req[0].value = Number(sessionStorage.getItem('siteVal'))
          }
        }
      })
    } else {
      this.$Message.error('无法获取列表数据，请刷新页面')
    }
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
