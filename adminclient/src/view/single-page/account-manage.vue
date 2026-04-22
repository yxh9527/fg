<template>
  <div>
    <Card>
      <ul>
        <li
          class="label-style account-radio"
          v-for="item in req"
          :key="item.label"
        >
          <label>{{ item.label }}&emsp;</label>
          <RadioGroup
            v-model="item.value"
            @on-change="
              () => {
                pageData.page = 1;
                handleSearch(item.value);
              }
            "
          >
            <Radio
              v-for="items in item.option"
              :key="items.label"
              :label="items.value"
              border
              >{{ items.label }}</Radio
            >
          </RadioGroup>
        </li>
      </ul>
    </Card>
    <br />
    <Card>
      <Button
        type="primary"
        style="margin-bottom: 10px"
        @click="showCreateAccount()"
        >添加账号</Button
      >
      <tables ref="tables" v-model="tableData" :columns="columns" />
      <div class="pageStyle">
        <Page
          :total="pageData.current"
          :current="pageData.page"
          :page-size="pageData.pageSize"
          :page-size-opts="pageData.pageOpts"
          show-elevator
          show-total
          @on-change="changePage"
          @on-page-size-change="changePageSize"
        />
      </div>
    </Card>
    <!-- 添加账号对话框 -->
    <Modal
      v-model="showAccount"
      :title="modalType.title"
      :loading="accountModalloading"
      @on-ok="addAccount"
      @on-visible-change="changeAccountType"
    >
      <Form
        ref="formValidate"
        :rules="ruleInline"
        style="padding-right: 30px"
        :label-width="100"
      >
        <FormItem
          v-for="item in accountData"
          :key="item.key"
          :label="item.disabled ? '' : item.label"
        >
          <template v-if="!item.disabled">
            <Select
              v-if="item.key == 'uType'"
              v-model="item.value"
              @on-change="showAgentSel"
              :prefix="item.icon"
            >
              <Option
                v-for="items in item.option"
                :value="items.value"
                :key="items.value"
                >{{ items.label }}</Option
              >
            </Select>

            <template v-if="item.key == 'ipLimit'">
              <i-input
                ref="accountInput"
                :maxlength="1024"
                v-model="item.value"
                :disabled="
                  item.disabled || (modalType.type == 2 && item.label == '账号')
                "
              >
                <Icon :type="item.icon" slot="prepend"></Icon>
              </i-input>
            </template>

            <template v-else>
              <i-input
                ref="accountInput"
                type="text"
                :maxlength="50"
                v-if="
                  item.key != 'uType' &&
                  item.key != 'agentId' &&
                  item.key != 'account'
                "
                v-model="item.value"
                :disabled="item.disabled"
                :placeholder="item.des ? item.des : '请输入' + item.label"
              >
                <Icon :type="item.icon" slot="prepend"></Icon>
              </i-input>

              <i-input
                ref="accountInput"
                type="text"
                :maxlength="50"
                :placeholder="item.des ? item.des : '请输入' + item.label"
                v-if="modalType.type == 1 && item.key == 'account'"
                v-model="item.value"
                :disabled="item.disabled"
              >
                <Icon :type="item.icon" slot="prepend"></Icon>
              </i-input>

              <i-input
                ref="accountInput"
                type="text"
                :maxlength="50"
                v-if="modalType.type == 2 && item.key == 'account'"
                v-model="item.value"
                :disabled="true"
              >
                <Icon :type="item.icon" slot="prepend"></Icon>
              </i-input>
            </template>

            <Select
              v-if="item.key == 'agentId'"
              v-model="item.value"
              :disabled="item.disabled"
              :prefix="item.icon"
              clearable
            >
              <Option
                v-for="items in item.option"
                :value="items.value"
                :key="items.value"
                >{{ items.label }}</Option
              >
            </Select>

            <template
              v-if="
                !(modalType.type == 2 && item.label == '账号') &&
                !(modalType.type == 2 && item.label == '密码') &&
                item.label != '账号类型'
              "
            >
              <Badge
                v-if="!['代理', 'IP地址限制'].includes(item.label)"
                style="
                  display: block;
                  position: relative;
                  transform: translate(0px, -16px);
                "
              >
                <Icon
                  type="ios-medical"
                  title="必填项"
                  style="margin-left: 20px"
                  slot="count"
                  color="#ed4014"
                  size="12"
                />
              </Badge>
            </template>
          </template>
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>

<script>
import Tables from '_c/tables'
import {
  getAccountData,
  addAccountData,
  getSelectAgent,
  editAccountData,
  editAccountState,
  deleteAccountState
} from '@/api/data'
import { setting } from '@/config'
export default {
  name: 'gameMessage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      // 对话框等待数据响应
      accountModalloading: true,
      showAccount: false,
      req: [
        {
          label: '类型',
          value: 0,
          option: [
            { label: '全部账号', value: 0 },
            { label: '总控账号', value: 1 },
            { label: '信息账号', value: 2 },
            { label: '代理账号', value: 3 }
          ]
        }
      ],
      columns: [
        {
          title: '类型',
          width: 120,
          align: 'center',
          key: 'uType',
          render (h, params) {
            return params.row.uType == 1 ? (
              <span style="color:green">总控</span>
            ) : params.row.uType == 2 ? (
              <span style="color:black">信息</span>
            ) : params.row.uType == 3 ? (
              <span style="color:orange">代理</span>
            ) : (
              ''
            )
          }
        },
        { title: '持有人', key: 'uName', width: 120, align: 'center' },
        { title: '账号', key: 'account', width: 180, align: 'center' },
        {
          title: 'IP地址限制',
          key: 'ipLimit',
          align: 'center'
        },
        {
          title: '后台域名',
          key: 'realmName',
          align: 'center'
        },
        {
          title: '最后登录时间',
          key: 'loginTime',
          align: 'center',
          width: 180,
          render (h, params) {
            if (params.row.loginTime) {
              return (
                <span>
                  {new Date(
                    params.row.loginTime * 1000
                  ).toLocaleString('chinese', { hour12: false })}
                </span>
              )
            } else {
              return <span>暂无记录</span>
            }
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 250,
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
                      this.accountData[3].value = ''
                      this.showAccount = true
                      this.modalType.title = '编辑账号'

                      this.req[0].value = Number(params.row.uType)
                      this.showAgentSel(params.row.uType)
                      window.testutypt = Number(params.row.uType)
                      this.modalType.type = 2
                      this.modalType.id = params.row.id

                      this.$nextTick(() => {
                        this.handleSearch()
                      })

                      for (const i in Object.keys(params.row)) {
                        this.accountData.forEach((item) => {
                          if (
                            item.key == Object.keys(params.row)[i] &&
                            item.key != 'password'
                          ) {
                            item.value = Object.values(params.row)[i]
                          }
                        })
                      }
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
                        (params.row.isForzen == 0 ? '冻结' : '启用') +
                        '账号吗?'
                    },
                    style: { textAlign: 'left', zIndex: '99' },
                    on: {
                      'on-ok': () => {
                        let data = {
                          id: params.row.id,
                          isForzen: params.row.isForzen == 0 ? 1 : 0
                        }

                        editAccountState(data).then((res) => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.handleSearch()
                              this.$Message.success(res.data.msg)
                            })
                          } else if (res.data.code == 400) {
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
                          type: params.row.isForzen == 0 ? 'error' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.isForzen == 0 ? '冻结' : '启用'
                    )
                  ]
                )
              ]
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
                      title: '您确定要删除账号吗?'
                    },
                    style: {
                      textAlign: 'left',
                      zIndex: '99',
                      marginLeft: '5px'
                    },
                    on: {
                      'on-ok': () => {
                        let data = {
                          id: params.row.id
                        }

                        deleteAccountState(data).then((res) => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.handleSearch()
                              this.$Message.success(res.data.msg)
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
              ]
            }
          ]
        }
      ],
      tableData: [],
      accountData: [
        {
          label: '账号类型',
          key: 'uType',
          value: '',
          icon: 'md-checkmark-circle-outline',
          option: [
            { label: '总控账号', value: 1 },
            { label: '信息账号', value: 2 },
            { label: '代理账号', value: 3 }
          ]
        },
        {
          label: '持有人',
          key: 'uName',
          value: '',
          icon: 'ios-trophy-outline'
        },
        {
          label: '账号',
          key: 'account',
          // disabled: true,
          value: '',
          icon: 'md-person'
        },
        {
          label: '密码',
          key: 'password',
          value: '',
          icon: 'ios-medical-outline'
        },
        {
          label: 'IP地址限制',
          key: 'ipLimit',
          value: '',
          icon: 'ios-link',
          des: '输入IP地址多个IP用逗号隔开'
        },
        {
          label: '代理',
          key: 'agentId',
          disabled: true,
          value: '',
          option: [],
          icon: 'ios-navigate-outline'
        },
        {
          label: '后台域名',
          key: 'realmName',
          disabled: true,
          value: '',
          icon: 'ios-globe-outline'
        }
      ],
      showAgent: false,
      agentList: [],
      modalType: { title: '', type: '', id: '' },
      ruleInline: {
        uType: [
          {
            required: true,
            message: '选择账号类型',
            trigger: 'change'
          }
        ],
        uName: [
          {
            required: true,
            message: '请输入持有人',
            trigger: 'blur'
          }
        ],
        account: [
          {
            required: true,
            message: '请输入账号',
            trigger: 'blur'
          }
        ],
        password: [
          {
            required: true,
            message: '请输入密码',
            trigger: 'blur'
          }
        ]
      },
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
    handleSearch (type = 0) {
      type = this.req[0].value
      let Data = type
        ? { uType: type, page: this.pageData.page }
        : { page: this.pageData.page }

      getAccountData(Data).then((res) => {
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

    showAgentSel (num) {
      this.accountData.forEach((item) => {
        if (item.hasOwnProperty('disabled')) {
          if (num == 3) {
            item.disabled = false
            if (item.key == 'agentId') {
              getSelectAgent().then((res) => {
                let resolvedData = []
                let resolveData = (data) => {
                  if (!resolvedData.find((x) => x.value == data.id)) {
                    resolvedData.push({
                      value: data.id,
                      label: data.name || data.nickName
                    })
                  }
                  if (data.subList) {
                    data.subList.map((d) => resolveData(d))
                  }
                }
                res.data.data.map((d) => resolveData(d))
                item.option = []
                item.option.push(...resolvedData)
              })
            }
          } else {
            item.disabled = true
          }
        }
      })
    },

    addAccount () {
      this.accountModalloading = false
      setTimeout(() => {
        this.accountModalloading = true
      }, 500)

      let Data = {}
      this.accountData.map((item) => {
        if (item.value != null || item.value === 0) {
          Data[item.key] = item.value
        }
      })

      let PW = Data.password

      if (PW && PW.includes(' ')) {
        this.$Message.error('密码不能由空格组成')
        return
      }

      if (this.modalType.type == 1) {
        addAccountData(Data).then((res) => {
          res.data.code == 200
            ? this.$Message.success('成功创建账号')
            : this.$Message.error(res.data.code + '：' + res.data.msg)
          this.$nextTick(() => {
            this.handleSearch()
          })

          this.showAccount = false
        })
      } else if (this.modalType.type == 2) {
        Data.id = this.modalType.id

        editAccountData(Data).then((res) => {
          this.$nextTick(() => {
            res.data.code == 200
              ? this.$Message.success('成功修改账号')
              : this.$Message.error(res.data.msg)
            this.handleSearch()
            this.showAccount = false
          })
        })
      }
    },

    showCreateAccount () {
      this.showAccount = true

      this.modalType.title = '添加账号'
      this.modalType.type = 1
      this.accountData.forEach((item) => {
        item.value = ''
      })

      if (this.req[0].value) {
        this.accountData[0].value = this.req[0].value
      } else {
        this.accountData[0].value = ''
      }
    },

    changePage (index) {
      this.pageData.page = index
      this.handleSearch()
    },
    changePageSize (index) {
      this.pageData.pageSize = index
      this.handleSearch()
    },
    changeAccountType (isShow) {
      if (isShow) {
        this.accountData[0].value = 9
        setTimeout(() => {
          this.accountData[0].value = this.req[0].value
          // 触发change
          this.showAgentSel(this.accountData[0].value)
          setTimeout(() => {
            if (this.modalType.type == 2) {
            }
          }, 100)
        }, 100)
      }
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
  margin-bottom: 20px;
  vertical-align: initial;
}
.ip-input-style.width {
  width: 500px;
}
.ivu-form-item-required .ivu-form-item-label:before {
  content: "*";
  display: inline-block;
  margin-right: 4px;
  line-height: 1;
  font-family: SimSun;
  font-size: 14px;
  color: #ed4014;
}
</style>
