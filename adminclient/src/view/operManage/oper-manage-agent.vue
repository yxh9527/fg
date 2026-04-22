 <template>
  <div>
    <Card>
      <tables :key="1" :columns="columns1" v-model="agentIonfo" />
      <br />
      <div class style="text-align: right;margin-bottom: 10px;margin-right: 15px;">
        <h2 style="left: 15px;position: absolute;">控制规则</h2>
        <Button type="primary" style="margin-left:10px;" @click="addControlAgent">增加新控制</Button>
      </div>
      <tables ref="tables" editable v-model="tableData" :columns="columns" />
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
    <Modal
      :title="modalEdit.type?'修改代理控制':'新增代理控制'"
      :closable="true"
      v-model="modalEdit.state"
      class="modalEditBox"
      width="600px"
      @on-ok="handleControlAgent(modalEdit.type)"
    >
      <br />
      <div>
        <label style="width:200px;text-align: right;">触发条件</label>&emsp;
        <Select v-model="modalEdit.triggerType" style="width:380px">
          <Option v-for="item in modalEdit.trigger" :value="item.id" :key="item.id">{{item.title}}</Option>
        </Select>
        <div class="modalStyle">
          <InputNumber
            :min="0"
            :max="100"
            v-model="modalEdit.triggerProb"
            style="width:80px"
            placeholder="X值"
          />&emsp;%
        </div>
      </div>
      <br />
      <div>
        <label style="width:200px;text-align: right;">控制方式</label>&emsp;
        <Select v-model="modalEdit.probType" style="width:380px">
          <Option v-for="item in modalEdit.action" :value="item.id" :key="item.id">{{item.title}}</Option>
        </Select>
        <div class="modalStyle">
          <InputNumber
            :min="-100"
            :max="100"
            v-model="modalEdit.winProb"
            style="width:80px"
            placeholder="X值"
          />&emsp;%
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
import Tables from '_c/tables'
import {
  getControlAgentData,
  addControlAgentPomp,
  editControlAgentProb,
  delControlAgentProb,
  getAgentInfo
} from '@/api/data'
import { setting } from '@/config'
export default {
  name: 'website',
  components: {
    Tables
  },
  props: ['id'],
  inject: ['handleLogOut'],
  data () {
    return {
      columns1: [
        {
          title: '站点',
          key: 'webId'
        },
        {
          title: '代理昵称',
          key: 'nickName'
        },
        {
          title: '剩余点数',
          key: 'point'
        },
        {
          title: '总盈亏',
          key: 'totalProfLoss'
        },
        {
          title: '当日盈亏',
          key: 'profitLoss'
        },
        {
          title: '剩余库存点数',
          key: 'stock'
        },
        {
          title: '总玩家',
          key: 'totalUserNumber'
        },
        {
          title: '当日下注玩家',
          key: 'userNumber'
        },
        {
          title: '总有效下注',
          key: 'totalEffBet'
        },
        {
          title: '总抽水分数',
          key: 'totalPump'
        }
      ],
      agentIonfo: [],
      columns: [
        { title: '序号', key: 'id', width: 80, align: 'center' },
        {
          title: '触发条件',
          width: 450,
          align: 'center',
          render (h, pramas) {
            return (
              <span>
                当代理&nbsp;
              {pramas.row.triggerType == 1 ? <b>增加</b> : <b>消耗</b>}&nbsp;
                分数大于初始分数的&nbsp;<b>{pramas.row.triggerProb}%</b>&nbsp;
                时
              </span>
            )
          }
        },
        { title: '胜率', key: 'winProb', align: 'center' },
        { title: '操作时间', key: 'createTime', align: 'center' },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 300,
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
                      this.modalEdit.state = true
                      this.modalEdit.type = true
                      this.modalEdit.id = params.row.id
                      this.modalEdit.triggerType = params.row.triggerType
                      this.modalEdit.triggerProb = params.row.triggerProb
                      this.modalEdit.winProb = params.row.winProb
                      this.modalEdit.updateTime = params.row.updateTime
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
                    title: '是否确认删除该控制？'
                  },
                  on: {
                    'on-ok': () => {
                      let data = {
                        id: params.row.id,
                        agentId: this.id
                      }
                      delControlAgentProb(data).then(res => {
                        if (res.data.code == 200) {
                          this.$nextTick(() => {
                            this.$Message.success('控制删除成功')
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
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      modalEdit: {
        type: false,
        state: false,
        agentId: 0,
        id: null,
        triggerType: 1,
        trigger: [
          { id: 1, title: '增加分数大于初始分数的 X %' },
          { id: 2, title: '消耗分数大于初始分数的 X %' }
        ],
        triggerProb: null,
        winProb: null,
        probType: 1,
        action: [{ id: 1, title: '玩家胜率变为 X %' }]
      }
    }
  },
  methods: {
    handleAgentInfo () {
      getAgentInfo({ id: this.id })
        .then(res => {
          if (res.data.code == 200) {
            this.agentIonfo = []
            this.agentIonfo[0] = res.data.data
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
    handleSearch () {
      let Data = { page: this.pageData.page, agentId: 0 }
      getControlAgentData(Data).then(res => {
        if (res.data.code == 200) {
          this.tableData = []
          this.tableData.push(...res.data.data.data)
          this.pageData.current = res.data.data.total
        }
      })
    },
    handleControlAgent (type) {
      if (type) {
        let Data = {
          id: this.modalEdit.id,
          agentId: this.id,
          winProb: this.modalEdit.winProb ? this.modalEdit.winProb : 0,
          triggerType: this.modalEdit.triggerType,
          triggerProb: this.modalEdit.triggerProb
            ? this.modalEdit.triggerProb
            : 0
        }
        editControlAgentProb(Data)
          .then(res => {
            if (res.data.code == 200) {
              this.$Message.success('修改代理抽水成功')
              this.handleSearch()
            }
          })
      } else {
        let Data = {
          agentId: this.id,
          winProb: this.modalEdit.winProb ? this.modalEdit.winProb : 0,
          triggerType: this.modalEdit.triggerType,
          triggerProb: this.modalEdit.winProb ? this.modalEdit.winProb : 0
        }
        addControlAgentPomp(Data)
          .then(res => {
            if (res.data.code == 200) {
              this.$Message.success('新增代理抽水成功')
              this.handleSearch()
            }
          })
      }
    },
    addControlAgent () {
      this.modalEdit.type = false
      this.modalEdit.state = true
      this.modalEdit.triggerType = 1
      this.modalEdit.triggerProb = null
      this.modalEdit.winProb = null
      // this.modalEdit.xzWinProb = null;
      this.modalEdit.probType = 1
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
    this.handleAgentInfo()
    this.handleSearch()
  }
}
</script>

<style lang="less">
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
.maintain {
  font-size: 15px;
  padding: 0 10px;
  > div {
    text-align: left;
    margin-bottom: 20px;
    label {
      margin-right: 10px;
    }
    span.ivu-input-suffix {
      margin: 5px;
    }
  }
  > button {
    margin-right: 20px;
  }
}
.modalEditBox div.ivu-modal-content div.ivu-modal-body {
  min-height: 75px !important;
}
.modalStyle {
  text-align: center;
  margin: 10px;
}
</style>
