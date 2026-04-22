<template>
  <div>
    <Card>
      <tables v-model="tableData" :columns="columns" />
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
import {
  getAgentData,
  addControlAgentPomp
} from '@/api/data'
import { setting } from '@/config'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      columns: [
        { title: '序号', width: 80, key: 'id', align: 'center' },
        { title: '站点', key: 'webName', align: 'center' },
        { title: '代理', key: 'nickName', align: 'center' },
        { title: '剩余点数', key: 'point', align: 'right' },
        // {
        //   title: "总消耗",
        //   key: "totalConsume",
        //   align: "right",
        //   render(h, params) {
        //     return params.row.totalConsume >= 0 ? (
        //       <span style="color:green">{params.row.totalConsume}</span>
        //     ) : (
        //       <span style="color:red">{params.row.totalConsume} </span>
        //     );
        //   }
        // },
        // {
        //   title: "当日消耗",
        //   key: "consume",
        //   align: "right",
        //   render(h, params) {
        //     return params.row.consume >= 0 ? (
        //       <span style="color:green">{params.row.consume}</span>
        //     ) : (
        //       <span style="color:red">{params.row.consume} </span>
        //     );
        //   }
        // },
        // {
        //   title: "剩余库存点数",
        //   align: "right",
        //   key: "stock"
        // },
        {
          title: '操作',
          key: 'handle',
          width: 450,
          align: 'center',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to:
                      '/oper-manage-user/' +
                      params.row.id +
                      '/' +
                      params.row.id
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  },
                  on: {
                    click: () => {}
                  }
                },
                '玩家单控'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/oper-manage-game-' + params.row.id
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '游戏控制'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                    // to: "/oper-manage-agent-" + params.row.id
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  },
                  on: {
                    click: () => {
                      this.$Modal.confirm({
                        title: '代理总控',
                        render: (h) => {
                          return [
                            h(
                              'div',
                              {
                                style: {
                                  marginTop: '30px',
                                  marginBottom: '10px',
                                  fontSize: '18px'
                                }
                              },
                              '站点：' + params.row.webName
                            ),
                            h(
                              'div',
                              {
                                style: {
                                  marginBottom: '10px',
                                  fontSize: '18px'
                                }
                              },
                              '代理：' + params.row.nickName
                            ),
                            h(
                              'label',
                              {
                                style: {
                                  fontSize: '18px',
                                  marginBottom: '10px'
                                }
                              },
                              '抽水设置：'
                            ),
                            h('InputNumber', {
                              props: {
                                value: params.row.pomp,
                                max: 100,
                                min: 0,
                                placeholder: '范围：0%-100%'
                              },
                              style: {
                                width: '180px',
                                fontSize: '18px',
                                marginBottom: '10px'
                              },
                              on: {
                                input: (val) => {
                                  params.row.pomp = val
                                }
                              }
                            }),
                            h(
                              'span',
                              {
                                style: {
                                  fontSize: '18px'
                                }
                              },
                              ' %'
                            )
                          ]
                        },
                        closable: true,
                        onOk: () => {
                          let Data = {}
                          Data = {
                            id: params.row.id,
                            pomp: params.row.pomp
                          }
                          addControlAgentPomp(Data).then((res) => {
                            if (res.data.code == 200) {
                              this.$Message.success('修改代理控制成功')
                              this.handleSearch()
                            }
                          })
                        }
                      })
                    }
                  }
                },
                '代理总控'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/oper-manage-log/' + params.row.id
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '日志'
              )
            }
          ]
        }
      ],
      tableData: [],
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
    handleSearch () {
      this.spinShow = true
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      if (sessionStorage.getItem('siteVal')) {
        Data = [{ webId: Number(sessionStorage.getItem('siteVal')) }, ...Data]
      }
      getAgentData(Data).then((res) => {
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
.item-input p {
  /* text-align: center; */
  margin: 10px;
  font-size: 16px;
}
.item-input p span {
  min-width: 100px;
  text-align: right;
  display: inline-block;
}
div.agent-oper div {
  font-size: 16px;
  line-height: 32px;
}
div.agent-oper label {
  display: inline-block;
  width: 80px;
  text-align: right;
  margin-right: 20px;
}
div.agent-oper input {
  width: 120px;
}
</style>
