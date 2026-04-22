 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{item.label}}&emsp;</label>
          <RadioGroup v-model="item.value" @on-change="handleSearch(item.value)">
            <Radio
              v-for="items in item.option"
              :key="items.label"
              :label="items.value"
              border
            >{{items.label}}</Radio>
          </RadioGroup>
        </li>
      </ul>
    </Card>
    <br />
    <Card>
      <Button type="primary" style="margin-bottom: 10px" @click="addBlackWhiteList(1)">增加普通白名单</Button>
      <Button
        type="primary"
        style="margin-bottom: 10px;margin-left: 10px"
        @click="addBlackWhiteList(2)"
      >增加普通黑名单</Button>
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
          @on-change="changePage"
          @on-page-size-change="changePageSize"
        />
      </div>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getControlList, addControlIP, deletControlIP } from '@/api/data'
import Edits from '_c/edit-data'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
export default {
  name: 'gameMessage',
  components: {
    Tables,
    Edits
  },
  inject: ['handleLogOut'],
  data () {
    return {
      req: [
        {
          label: '类型',
          value: 1,
          option: [
            { label: '普通白名单', value: 1 },
            { label: '普通黑名单', value: 2 }
          ]
        }
      ],
      columns: [
        { title: '序号', type: 'index', width: 80 },
        {
          title: '类型',
          key: 'contType',
          align: 'center',
          render (h, params) {
            return params.row.contType == 1 ? (
              <span style="color:green">白名单</span>
            ) : (
              <span style="color:orange">黑名单</span>
            )
          }
        },
        { title: 'IP地址', key: 'ip', align: 'center' },
        {
          title: '创建时间',
          key: 'createTime',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {params.row.createTime
                  ? getDate(params.row.createTime * 1000)
                  : '未知'}
              </span>
            )
          }
        },
        { title: '备注', key: 'remarks', align: 'center' },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          button: [
            (h, params) => {
              return [
                h(
                  'Poptip',
                  {
                    props: {
                      transfer: true,
                      confirm: true,
                      placement: 'left',
                      title: '您确定要删除吗?'
                    },
                    style: { textAlign: 'left', zIndex: '99' },
                    on: {
                      'on-ok': () => {
                        const Data = {
                          id: params.row.id,
                          agentId: sessionStorage.getItem('agentVal')
                        }
                        deletControlIP(Data).then(res => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.$Message.success('成功删除IP')
                              this.handleSearch(params.row.contType)
                            })
                          } else {
                            alert('操作失败：' + res.data.msg)
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
                          disabled: params.row.ip == '127.0.0.1',
                          type: 'error',
                          size: 'small'
                        }
                      },
                      '删除'
                    )
                  ]
                )
              ]
            }
          ]
        }
      ],
      tableData: [],
      ipconfig: '',
      remarks: '',
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
    handleSearch (type = 1) {
      let Data = [
        { page: this.pageData.page },
        { contType: type },
        { agentId: sessionStorage.getItem('agentVal') }
      ]
      getControlList(Data)
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
    addBlackWhiteList (type) {
      this.$Modal.confirm({
        title: type == 1 ? '添加普通白名单' : '添加普通黑名单',
        okText: type == 1 ? '添加白名单' : '添加黑名单',
        closable: true,
        render: h => {
          return [
            h(
              'label',
              {
                class: 'ip-label-style'
              },
              'IP地址'
            ),
            h('Input', {
              props: {
                type: 'text',
                maxlength: 50,
                number: true,
                value: this.ipconfig
              },
              class: 'ip-input-style',
              on: {
                input: val => {
                  this.ipconfig = val
                }
              }
            }),
            h(
              'label',
              {
                style: `font-size: 20px;
    color: red;
    padding-left: 15px;
    transform: translateY(3px);
    position: relative;
    display: inline-block;`
              },
              '*'
            ),
            h('div'),
            h(
              'label',
              {
                class: 'ip-label-style'
              },
              '备注'
            ),
            h('Input', {
              props: {
                value: this.remarks
              },
              class: 'ip-input-style',
              on: {
                input: val => {
                  this.remarks = val
                }
              }
            })
          ]
        },
        onOk: () => {
          let Data = [
            { agentId: sessionStorage.getItem('agentVal') },
            { ip: this.ipconfig },
            { remarks: this.remarks },
            { contType: type }
          ]
          addControlIP(Data).then(res => {
            if (res.data.code == 200) {
              this.$Message.success('成功创建IP')
              this.$nextTick(() => {
                this.handleSearch(type)
                this.req[0].value = type
              })
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
.item-input {
  margin-top: 10px;
}
.item-input label {
  display: inline-block;
  width: 70px;
  margin-left: -15px;
  text-align: right;
}
.item-input input {
  width: 220px;
}
.ip-label-style {
  display: inline-block;
  width: 90px;
  margin-right: 10px;
  font-size: 18px;
  text-align: right;
}
.ip-input-style {
  width: 200px;
  font-size: 18px;
  margin-bottom: 10px;
}
</style>
