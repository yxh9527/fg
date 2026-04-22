 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in labelList" :key="item.label">
          <label>{{item.label}}</label>
          <Select v-model="item.value" style="width: 180px">
            <Option
              v-for="items in item.option"
              :value="items.value"
              :key="items.label"
            >{{ items.label }}</Option>
          </Select>
        </li>
        <div class="search">
          <Button type="primary" @click="searchAction">搜索</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <Button type="primary" style="margin-bottom: 10px" to="/game-message-add">添加消息</Button>
      <tables ref="tables" v-model="tableData" :columns="columns" />
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
import {
  getGameMsgData,
  editGameMsgData,
  deleteGameMsgData
} from '@/api/data'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
import Edits from '_c/edit-data'
export default {
  name: 'gameMessage',
  components: {
    Tables,
    Edits
  },
  inject: ['handleLogOut'],
  data () {
    return {
      labelList: [
        {
          label: '邮件类型',
          key: 'msgType',
          value: '',
          option: [
            // { label: "全部", value: "0" },
            { label: '活动消息', value: '1' },
            { label: '维护公告', value: '2' }
          ]
        }
      ],
      // startTime: "",
      // endTime: "",
      columns: [
        { title: '序号', key: 'number', width: 120, align: 'center', required: true, showRedPoint: true},
        { title: '消息标题', key: 'title', align: 'center', showRedPoint: true},
        { title: '消息内容', key: 'info', width: 280, align: 'center', showRedPoint: true },
        {
          title: '发布时间',
          key: 'startTime',
          width: 180,
          align: 'center',
          render (h, params) {
            return (
              <span>
                {params.row.startTime
                  ? getDate(params.row.startTime * 1000)
                  : '未设定'}
              </span>
            )
          }
        },
        {
          title: '结束时间',
          key: 'endTime',
          width: 180,
          align: 'center',
          render (h, params) {
            return (
              <span>
                {' '}
                {params.row.endTime !== null
                  ? getDate(params.row.endTime * 1000)
                  : '未设定'}
              </span>
            )
          }
        },
        {
          title: '类型',
          key: 'msgType',
          align: 'center',
          render (h, params) {
            return params.row.msgType == 1 ? (
              <span style="color:green">活动消息</span>
            ) : (
              <span style="color:orange">维护公告</span>
            )
          }
        },
        { title: '备注', key: 'remarks', align: 'center' },
        {
          title: '接收点',
          key: 'receive',
          align: 'center',
          render (h, params) {
            return h(
              'Poptip',
              {
                props: {
                  transfer: true,
                  placement: 'left',
                  width: 300,
                  title: '接收点'
                }
              },
              [
                h(
                  'div',
                  {
                    props: {
                      className: 'api'
                    },
                    class: 'titleBox',
                    slot: 'content'
                  },
                  [
                    h('p', [h('span', '站点：'), params.row.webName]),
                    h('p', [h('span', '代理：'), params.row.agentName]),
                    h('p', [h('span', '游戏分类：'), params.row.gameTypeName]),
                    h('p', [h('span', '游戏名称：'), params.row.gameName]),
                    h('p', [h('span', '游戏平台：'), params.row.platFromName])
                  ]
                ),
                h(
                  'Button',
                  {
                    props: {
                      size: 'small'
                    }
                  },
                  '查看'
                )
              ]
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 180,
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
                      this.isFormValid = false
                      this.$Modal.confirm({
                        title: '修改数据',
                        closable: true,
                        width: 800,
                        onOk: () => {
                          this.editData.agentId = sessionStorage.getItem(
                            'agentVal'
                          )

                          if (!this.isFormValid) { //
                            if (Object.keys(this.editData).length == 1) { //
                              this.$Message.error('未修改')//
                              return//
                            }//

                            editGameMsgData(this.editData).then(res => {
                              if (res.data.code == 200) {
                                this.editData = []
                                this.$nextTick(() => {
                                  this.$Message.success('编辑游戏消息成功')
                                  this.handleSearch()
                                })
                              }
                            })
                          } else { //
                            this.$Message.error('请检查必填项')//
                          }//
                        },
                        render: h => {
                          let that = this
                          return h(Edits, {
                            props: {
                              row: params.row,
                              columns: this.columns
                            },
                            on: {
                              sendError (e) { //
                                that.isFormValid = true//
                                that.$Message.error(e)//
                              }, //
                              sendEditData: data => {
                                that.isFormValid = false //
                                data = data.map(item => {
                                  if (item.value != '') {
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
                    }
                  }
                },
                '修改'
              )
            },

            (h, params) => {
              return h(
                'Poptip',
                {
                  props: {
                    transfer: true,
                    confirm: true,
                    placement: 'left',
                    title:
                      '是否确认删除此条消息？ （消息删除后无法恢复，且停止滚播，请谨慎操作）'
                  },
                  on: {
                    'on-ok': () => {
                      let data = {
                        id: params.row.id,
                        agentId: sessionStorage.getItem('agentVal')
                      }
                      deleteGameMsgData(data).then(res => {
                        if (res.data.code == 200) {
                          this.$nextTick(() => {
                            this.$Message.success('消息删除成功')
                            this.handleSearch()
                          })
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
                        type: 'error',
                        size: 'small'
                      }
                    },
                    '删除'
                  )
                ]
              )
            }
          ]
        }
      ],
      tableData: [],
      editData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
    handleSearch () {
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      this.labelList.map(item => {
        Data.push({ [item.key]: item.value })
      })
      getGameMsgData(Data)
        .then(res => {
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
.titleBox {
  font-size: 16px;
}
.titleBox p span {
  display: inline-block;
  width: 100px;
  text-align: right;
}
</style>
