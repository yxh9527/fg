<template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.key">
          <label>{{ item.label }}</label>
          <Select style="width:180px" v-model="item.value">
            <Option
              v-for="items in item.option"
              :value="items.id"
              :key="items.id"
              >{{ items.name }}</Option
            >
          </Select>
        </li>
        <div class="search">
          <Button type="primary" @click="searchAction">搜索</Button>
          <Button @click="handleAllSearch">重置</Button>
        </div>
      </ul>
    </Card>
    <br />
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
import {
  // getClassList,
  // getTypeList,
  getSelectGames,
  getControlGameData,
  setControlGameProb
} from '@/api/data'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  props: ['id'],
  inject: ['handleLogOut'],
  data () {
    return {
      req: [
        {
          label: '游戏名称',
          key: 'gameId',
          value: '',
          option: []
        },
        {
          label: '游戏平台',
          key: 'platformId',
          value: '',
          option: []
        },
        {
          label: '游戏类型',
          key: 'typeId',
          value: '',
          option: []
        },
        {
          label: '控制状态',
          key: 'contType',
          value: '',
          option: [
            { name: '控制', id: 1 },
            { name: '未控制', id: 2 }
          ]
        }
      ],
      columns: [
        { title: '游戏名称', key: 'name', align: 'center' },
        { title: '房间名', key: 'difficultyName', align: 'center' },
        { title: '游戏平台', key: 'platformName', align: 'center' },
        { title: '游戏分类', key: 'typeName', align: 'center' },
        {
          title: '房间总盈亏',
          key: 'totalProfitLoss',
          align: 'right',
          render (h, params) {
            return params.row.totalProfitLoss >= 0 ||
              JSON.stringify(params.row.totalProfitLoss).indexOf('-') == -1 ? (
                <span style="color:green">{params.row.totalProfitLoss}</span>
              ) : (
                <span style="color:red">{params.row.totalProfitLoss} </span>
              )
          }
        },
        {
          title: '今日房间盈亏',
          align: 'right',
          key: 'profitLoss',
          render (h, params) {
            return params.row.profitLoss >= 0 ||
              JSON.stringify(params.row.profitLoss).indexOf('-') == -1 ? (
                <span style="color:green">{params.row.profitLoss}</span>
              ) : (
                <span style="color:red">{params.row.profitLoss} </span>
              )
          }
        },
        {
          title: '房间胜率控制',
          key: 'isControl',
          align: 'center',
          render (h, params) {
            return params.row.isControl == true ? (
              <span style="color:green">{params.row.winProb}%</span>
            ) : (
              <span>不控制</span>
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          button: [
            (h, params) => {
              if (!params.row.controlId) {
                // 新增
                return h(
                  'Button',
                  {
                    props: {
                      type: 'info',
                      size: 'small'
                    },
                    on: {
                      click: () => {
                        this.changeWinProb(params.row)
                      }
                    }
                  },
                  '新增'
                )
              }

              if (params.row.controlId && params.row.winProb != 0) {
                // 修改
                return h(
                  'Button',
                  {
                    props: {
                      type: 'warning',
                      size: 'small'
                    },
                    on: {
                      click: () => {
                        this.changeWinProb(params.row)
                      }
                    }
                  },
                  '修改'
                )
              }

              if (params.row.controlId && params.row.winProb == 0) {
                // 修改
                return h(
                  'Button',
                  {
                    props: {
                      type: 'info',
                      size: 'small'
                    },
                    on: {
                      click: () => {
                        this.changeWinProb(params.row, '新增')
                      }
                    }
                  },
                  '新增'
                )
              }
            }
          ]
        }
      ],
      winProb: 0,
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
    changeWinProb (params, title) {
      if (params.controlId != '') {
        this.winProb = params.winProb
      }

      this.$Modal.confirm({
        title:
          params.controlId != '' && title != '新增'
            ? '修改游戏胜率'
            : '新增游戏胜率',
        okText: '修改胜率',
        closable: true,
        onOk: () => {
          let Data = {
            prob: this.winProb,
            agentId: params.agentId
          }
          params.controlId != ''
            ? Object.assign(Data, { id: params.controlId })
            : Object.assign(Data, {
              agentId: params.agentId,
              gameId: params.id,
              difficulty: params.difficulty
            })
          setControlGameProb(Data).then(res => {
            if (res.data.code == 200) this.$Message.success('修改胜率成功')
            this.$nextTick(() => {
              this.handleSearch()
            })
          })
        },

        render: h => {
          return [
            h(
              'span',
              {
                style: { fontSize: '18px' }
              },
              '游戏胜率：'
            ),
            h('Slider', {
              props: {
                value: params.winProb,
                showInput: true,
                min: -100,
                max: 100
              },
              on: {
                input: val => {
                  this.winProb = val
                }
              }
            })
          ]
        }
      })
    },

    handleSearch () {
      this.spinShow = true
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { agentId: this.id }
      ]
      this.req.map(item => {
        if (item.value) {
          Data.push({
            [item.key]: item.value
          })
        }
      })

      getControlGameData(Data)
        .then(res => {
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
              }
            })
          }
        })
        .catch(err => {
          this.spinShow = false
        })
    },

    handleAllSearch () {
      this.spinShow = true
      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { agentId: this.id }
      ]
      getControlGameData(Data).then(res => {
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

    this.req.map(item => {
      // if (item.key == "platformId") {
      //   getClassList(this.id).then(res => {
      //     item.option.push(...res.data.data);
      //   });
      // }
      // if (item.key == "typeId") {
      //   getTypeList(this.id).then(res => {
      //     item.option.push(...res.data.data);
      //   });
      // }
      if (item.key == 'gameId') {
        getSelectGames(this.id).then(res => {
          item.option.push(...res.data.data)
        })
      }
    })
  }
}
</script>

<style>
.ivu-input-number.ivu-input-number-large {
  margin-top: -7px;
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
