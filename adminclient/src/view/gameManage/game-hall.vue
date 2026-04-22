 <template>
  <div>
    <Card>
      <h2 style="margin:15px 10px 20px;">大厅设置</h2>
      <Table
        :columns="columns"
        :data="tableData"
        style="margin:15px auto"
        class="hallTable"
        stripe
        highlight-row
        draggable
        @on-drag-drop="dragdrop"
      ></Table>
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
    </Card>
  </div>
</template>

<script>
import {
  getGameHallList,
  setGameHallUpSort,
  setGameHallUpData,
  changeGameHallUpState,
  changeGameHallUpMaintain,
  getGameHallTypeList
} from '@/api/data'
import { setting } from '@/config'
let self = {}

export default {
  name: 'game_add',
  inject: ['handleLogOut'],
  data () {
    return {
      columns: [
        { title: '排序', type: 'index', width: 70, align: 'center' },
        { title: '游戏名称', key: 'name', width: 150, align: 'center' },
        {
          title: '火',
          align: 'center',
          minWidth: 70,
          render (h, params) {
            if (params.row.showType == 1) {
              return <Icon type="md-flame" color="red" size="30" />
            } else {
              return h('Radio', {
                props: {
                  size: 'large',
                  name: 'name' + params.row.id
                },
                on: {
                  'on-change': val => {
                    if (val) {
                      self.changeGameState(params.row.id, 1)
                    }
                  }
                }
              })
            }
          }
        },
        {
          title: '新',
          align: 'center',
          minWidth: 70,
          render (h, params) {
            if (params.row.showType == 2) {
              return <Icon type="md-star" color="orange" size="30" />
            } else {
              return h('Radio', {
                props: {
                  size: 'large',
                  name: 'name' + params.row.id
                },
                on: {
                  'on-change': val => {
                    if (val) {
                      self.changeGameState(params.row.id, 2)
                    }
                  }
                }
              })
            }
          }
        },
        {
          title: '敬请期待',
          align: 'center',
          minWidth: 95,
          render (h, params) {
            if (params.row.showType == 3) {
              return <Icon type="md-rose" color="green" size="30" />
            } else {
              return h('Radio', {
                props: {
                  size: 'large',
                  name: 'name' + params.row.id
                },
                on: {
                  'on-change': val => {
                    if (val) {
                      self.changeGameState(params.row.id, 3)
                    }
                  }
                }
              })
            }
          }
        },
        {
          title: '操作',
          align: 'right',
          width: 250,
          render (h, params) {
            return [
              params.row.showType > 0
                ? h(
                  'Button',
                  {
                    props: {
                      name: 'name' + params.row.id,
                      size: 'small'
                    },
                    on: {
                      click: val => {
                        if (val) {
                          self.changeGameState(params.row.id, 0)
                        }
                      }
                    }
                  },
                  '移除标识'
                )
                : '',
              h(
                'Poptip',
                {
                  props: {
                    transfer: true,
                    confirm: true,
                    placement: 'left',
                    width: '300',
                    title:
                      '您确定要' +
                      (params.row.state == 1 ? '维护' : '恢复') +
                      '游戏吗?'
                  },
                  style: { textAlign: 'left', marginLeft: '10px' },
                  on: {
                    'on-ok': () => {
                      let data = {
                        agentId: sessionStorage.getItem('agentVal'),
                        id: params.row.id,
                        state: params.row.state == 1 ? 2 : 1
                      }
                      changeGameHallUpMaintain(data).then(res => {
                        self.$nextTick(() => {
                          self.handleSearch()
                        })
                      })
                    }
                  }
                },
                [
                  h(
                    'Button',
                    {
                      props: {
                        type: params.row.state == 1 ? 'warning' : 'success',
                        size: 'small'
                      },
                      style: {
                        marginRight: '5px'
                      }
                    },
                    params.row.state == 1 ? '维护游戏' : '恢复游戏'
                  )
                ]
              ),
              h(
                'Button',
                {
                  props: {
                    type: 'primary',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      let Data = {
                        id: params.row.id,
                        agentId: sessionStorage.getItem('agentVal'),
                        gameIco: params.row.gameIcon,
                        fmImg: params.row.fmBgImg,
                        fmMusic: params.row.fmMusic,
                        djMusic: params.row.djMusic,
                        hallType: params.row.hallType
                      }
                      self.$Modal.confirm({
                        title: '大厅音乐及背景设置',
                        closable: true,
                        align: 'center',
                        width: 500,
                        render: h => {
                          return [
                            h(
                              'Input',
                              {
                                props: {
                                  type: 'url',
                                  placeholder: '请输入URL地址',
                                  value: Data.gameIco
                                },
                                style: {
                                  marginTop: '15px',
                                  marginBottom: '10px'
                                },
                                on: {
                                  input: val => {
                                    Data.gameIco = val
                                  }
                                }
                              },
                              [
                                h(
                                  'span',
                                  {
                                    slot: 'prepend'
                                  },
                                  '图标地址'
                                )
                              ]
                            ),
                            h(
                              'Input',
                              {
                                props: {
                                  type: 'url',
                                  placeholder: '请输入URL地址',
                                  value: Data.fmImg
                                },
                                style: {
                                  marginBottom: '10px'
                                },
                                on: {
                                  input: val => {
                                    Data.fmImg = val
                                  }
                                }
                              },
                              [
                                h(
                                  'span',
                                  {
                                    slot: 'prepend'
                                  },
                                  '房门背景'
                                )
                              ]
                            ),
                            h(
                              'Input',
                              {
                                props: {
                                  type: 'url',
                                  placeholder: '请输入URL地址',
                                  value: Data.fmMusic
                                },
                                style: {
                                  marginBottom: '10px'
                                },
                                on: {
                                  input: val => {
                                    Data.fmMusic = val
                                  }
                                }
                              },
                              [
                                h(
                                  'span',
                                  {
                                    slot: 'prepend'
                                  },
                                  '房门音乐'
                                )
                              ]
                            ),
                            h(
                              'Input',
                              {
                                props: {
                                  type: 'url',
                                  placeholder: '请输入URL地址',
                                  value: Data.djMusic
                                },
                                style: {
                                  marginBottom: '10px'
                                },
                                on: {
                                  input: val => {
                                    Data.djMusic = val
                                  }
                                }
                              },
                              [
                                h(
                                  'span',
                                  {
                                    slot: 'prepend'
                                  },
                                  '对局音乐'
                                )
                              ]
                            ),
                            h(
                              'div',
                              {
                                class: 'edit-checkbox',
                                style: { position: 'initial' }
                              },
                              [
                                h('label', {}, '游戏类型'),
                                h(
                                  'Select',
                                  {
                                    props: {
                                      type: 'url',
                                      placeholder: '请输入类型',
                                      value: Data.hallType
                                    },
                                    style: {
                                      width: '250px'
                                    },
                                    on: {
                                      input: val => {
                                        Data.hallType = val
                                      }
                                    }
                                  },
                                  [
                                    h(
                                      'Option',
                                      { props: { value: 1 } },
                                      '经典棋牌'
                                    ),
                                    h(
                                      'Option',
                                      { props: { value: 2 } },
                                      '街机电游'
                                    ),
                                    h(
                                      'Option',
                                      { props: { value: 3 } },
                                      '百人大战'
                                    ),
                                    h(
                                      'Option',
                                      { props: { value: 4 } },
                                      '新游推荐'
                                    )
                                  ]
                                )
                              ]
                            )
                          ]
                        },
                        onOk: () => {
                          setGameHallUpData(Data).then(res => {
                            if (res.data.code == 200) {
                              self.$Message.success('修改设置成功')
                              self.$nextTick(self.handleSearch())
                            } else {
                              self.$Message.error(
                                res.data.code + '：' + res.data.msg
                              )
                            }
                          })
                        }
                      })
                    }
                  }
                },
                '设置'
              )
            ]
          }
        },
        {
          title: '状态',
          key: 'state',
          align: 'center',
          minWidth: 80,
          render (h, params) {
            return params.row.state == 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">维护</span>
            )
          }
        },
        { title: '图标地址', key: 'gameIcon', align: 'center', minWidth: 95 },
        { title: '房门背景', key: 'fmBgImg', align: 'center', minWidth: 95 },
        { title: '房门音乐', key: 'fmMusic', align: 'center', minWidth: 95 },
        { title: '对局音乐', key: 'djMusic', align: 'center', minWidth: 95 },
        {
          title: '游戏类型',
          key: 'hallType',
          align: 'center',
          minWidth: 95,
          render (h, pramas) {
            return (
              <span>
                {
                  {
                    1: '经典棋牌',
                    2: '街机电游',
                    3: '百人大战',
                    4: '新游推荐'
                  }[pramas.row.hallType]
                }
              </span>
            )
          }
        }
      ],
      tableData: [],
      huo: false,
      sortData: [],
      typeList: [],
      spinShow: false
    }
  },
  created () {
    self = this
  },
  methods: {
    dragdrop (a, b) {
      this.tableData.splice(
        b,
        1,
        ...this.tableData.splice(a, 1, this.tableData[b])
      )
      let Data = []
      for (let i = 0; i < this.tableData.length; i++) {
        Data.push({
          id: this.tableData[i].id,
          weight: this.tableData.length - i
        })
      }
      this.$nextTick(() => {
        if (Data.length > 0) {
          setGameHallUpSort(Data).then(res => {
            this.$Message.success(res.data.msg)
          })
        }
      })
    },
    handleSearch () {
      this.spinShow = true
      this.tableData = []
      getGameHallList({ agentId: sessionStorage.getItem('agentVal') }).then(
        res => {
          this.spinShow = false
          if (res.data.code == 200) {
            this.tableData.push(...res.data.data)
          } else {
            // 判断响应状态是否为Token失效，如果失效则执行退出函数并刷新页面。
            this.$nextTick(() => {
              if (setting.arrStatus.indexOf(res.data.code) != -1) {
                this.$Message.error(
                  res.data.code + ' ：&nbsp;' + res.data.msg + '请重新登录'
                )
                // this.handleLogOut();
              } else {
                this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
              }
            })
          }
        }
      )
    },
    changeGameState (id, state) {
      const Data = {
        agentId: sessionStorage.getItem('agentVal'),
        id: id,
        state: state
      }
      this.$Spin.show({
        render: h => {
          return h('div', [
            h('Icon', {
              class: 'demo-spin-icon-load',
              props: {
                type: 'ios-loading',
                size: 60
              }
            })
          ])
        }
      })
      changeGameHallUpState(Data).then(res => {
        if (res.data.code == 200) {
          this.$Spin.hide()
          this.$nextTick(() => {
            this.$Message.success('修改成功')
            this.tableData.forEach(item => {
              if (item.id == id) {
                item.showType = state
              }
            })
          })
        }
      })
    }
  },
  mounted () {
    this.handleSearch()
    getGameHallTypeList().then(res => {
      this.typeList = []
      this.typeList.push(...res.data.data)
    })
  }
}
</script>

<style>
.demo-spin-icon-load {
  animation: ani-demo-spin 1s linear infinite;
}
.edit-checkbox {
  position: relative;
}
.edit-checkbox > label {
  display: inline-block;
  margin-right: 10px;
  text-align: center;
  padding: 4px 7px;
  line-height: 22px;
  color: #515a6e;
  background-color: #f8f8f9;
  border: 1px solid #dcdee2;
  border-radius: 4px;
}
.hallTable thead tr th.ivu-table-column-right {
  text-align: center;
}
.ivu-spin-fullscreen .ivu-spin-fix {
  background: none;
}
.ivu-modal-confirm-head > div {
  display: inline-block;
}
</style>
