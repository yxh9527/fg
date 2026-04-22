<template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{ item.label }}</label>
          <Input
            v-model="item.value"
            v-if="item.type == 'text'"
            placeholder="请输入"
            style="width: 180px"
            :maxlength="50"
          />
          <!-- <DatePicker
            type="datetime"
            v-if="item.type=='datetime'"
            v-model="item.value"
            :placeholder="'请选择'+item.label"
          ></DatePicker>-->
        </li>

        <div class="search">
          <Button type="primary" @click="handleSearch">搜索</Button>
          <Button @click="handleAllSearch">重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <Button
        type="primary"
        v-if="viewAccess"
        style="margin-bottom: 10px"
        to="/website-add"
        >创建站点</Button
      >
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
          @on-change="changePage"
          @on-page-size-change="changePageSize"
        />
      </div>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getSiteData, editSiteData } from '@/api/data'
import Edits from '_c/edit-data'
import { getDate } from '@/libs/tools'
import { setting } from '@/config'
export default {
  name: 'website',
  components: {
    Tables,
    Edits
  },
  inject: ['viewAccess', 'handleLogOut', 'reFreshSiteAgentList'],
  data () {
    return {
      req: [{ label: '站点昵称', type: 'text', key: 'name', value: '' }],
      createTime: '',
      columns: [
        {
          title: '序号',
          key: 'id',
          width: 80,
          align: 'center',
          isdisabled: true
        },
        {
          title: '站点名称',
          key: 'nickName',
          align: 'center',
          minWidth: 100,
          showRedPoint: true
        },
        // { title: "域名", key: "realmName", align: "center", minWidth: 100 ,showRedPoint:true},
        {
          title: '负责人',
          key: 'contacts',
          align: 'center',
          minWidth: 100,
          showRedPoint: true
        },
        {
          title: '负责人联系方式',
          key: 'phone',
          align: 'center',
          minWidth: 100,
          showRedPoint: true
        },
        {
          title: '邮箱',
          key: 'email',
          align: 'center',
          width: 180,
          showRedPoint: true
        },
        { title: '备注', key: 'remarks', align: 'center', minWidth: 100 },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 200,
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
                      this.$Modal.confirm({
                        title: '修改数据',
                        closable: true,
                        width: 800,
                        onOk: () => {
                          if (Object.keys(this.editData).length > 0) {
                            editSiteData(this.editData).then((res) => {
                              this.editData = []
                              this.$nextTick(() => {
                                this.$Message.success('编辑站点成功')

                                this.reFreshSiteAgentList()

                                this.handleSearch()
                              })
                            })
                          }
                        },
                        render: (h) => {
                          return [
                            h(Edits, {
                              props: {
                                row: params.row,
                                columns: this.columns
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
                                  this.editData = Object.assign(...data)
                                }
                              }
                            })
                          ]
                        }
                      })
                    }
                  }
                },
                '编辑'
              )
            }
          ]
        }
      ],
      tableData: [],
      editData: [],
      isPermanent: null,
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
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    },
    handleSearch () {
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
      getSiteData(Data).then((res) => {
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
      getSiteData(Data).then((res) => {
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
    }
  },
  created () {
    this.columns.forEach((item, index) => {
      if (item.key == 'handle') {
        if (!this.viewAccess) {
          this.columns.splice(index, 1)
        }
      }
    })
  },
  mounted () {
    this.handleSearch()
  }
}
</script>

<style lang="less">
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
