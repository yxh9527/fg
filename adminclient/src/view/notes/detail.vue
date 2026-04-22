<template>
  <div>
    <Card>
      <div class="inlineForm">
        <div style="margin-left: 5px">站点：</div>
        <div>
          <Select
            v-model="params.webId"
            @on-change="siteChanged"
            style="width: 100px"
          >
            <Option
              v-for="item in siteOption"
              :value="item.id"
              :key="item.id"
              >{{ item.name }}</Option
            >
          </Select>
        </div>
        <div style="padding-left: 5px">代理：</div>
        <div>
          <Select filterable v-model="params.agentId" style="width: 100px">
            <Option
              v-for="item in agentOption"
              :value="item.id"
              :key="item.id"
              >{{ item.name }}</Option
            >
          </Select>
        </div>
        <div style="padding-left: 5px">游戏：</div>
        <div>
          <Select
            filterable
            clearable
            v-model="games.value"
            style="width: 250px"
          >
            <Option
              v-for="item in games.option"
              :value="item.number"
              :key="item.number"
              :label="item.label"
              >{{ item.label }}</Option
            >
          </Select>
        </div>
        <div style="padding-left: 5px">账号：</div>
        <Input
          v-model="params.account"
          placeholder="请输入"
          style="width: 100px"
        />
        <div style="padding-left: 5px">昵称：</div>
        <Input
          v-model="params.nickName"
          placeholder="请输入"
          style="width: 100px"
        />
        <div style="padding-left: 5px">局号：</div>
        <Input
          v-model="params.officeNumber"
          placeholder="请输入"
          style="width: 100px"
        />
        <div style="padding-left: 5px">Hash：</div>
        <Input
          v-model="params.hash"
          placeholder="请输入"
          style="width: 150px"
        />
        <div style="padding-left: 5px">时间：</div>
        <DatePicker
          type="datetime"
          v-model="startDate"
          placeholder="开始时间"
          :options="startDateRestrict"
          style="width: 180px"
          :clearable="false"
        ></DatePicker>
        <div style="padding: 0 5px">—</div>
        <DatePicker
          type="datetime"
          v-model="endDate"
          placeholder="结束时间"
          :options="endDateRestrict"
          style="width: 180px"
          :clearable="false"
        ></DatePicker>
        <div style="padding: 0 5px">已完成</div>
        <Checkbox v-model="params.complete"></Checkbox>
        <Button
          @click="
            pageData.page = 1;
            handleSearch();
          "
          style="margin-left: 5px"
          type="primary"
          >搜索</Button
        >

        <Button @click="exportSearch()" style="margin-left: 5px" type="primary"
          >导出</Button
        >
      </div>
    </Card>
    <Card title="对局详情" style="margin-top: 10px">
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
      <Modal v-model="viewSettlementDetail" width="65%" title="游戏详情">
        <div>
          <iframe
            :src="this.viewGameDetailUrl"
            width="100%"
            height="500"
            frameborder="0"
          ></iframe>
        </div>
        <div slot="footer"></div>
      </Modal>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
import Detial from '../players/players-game-detial.vue'
import JSZip from 'jszip'
import FileSaver from 'file-saver'
import { exportExcel } from '@/libs/excel'
import * as dayjs from 'dayjs'
// 在点击事件的方法中调用
import Clipboard from 'clipboard'
import {
  getSettlement,
  getExportSettlements,
  getExportSettlementCount,
  getGameServers
} from '@/api/data'
export default {
  name: 'detailManage',
  components: {
    Tables,
    Detial
  },
  data () {
    let _this = this
    return {
      spinShow: false,
      startDate: null,
      endDate: null,
      gameUrls: [],
      replays: [],
      viewSettlementDetail: false,
      viewGameDetailUrl: '',
      startDateRestrict: {
        disabledDate (date) {
          if (date.getTime() > dayjs(_this.endDate).unix() * 1000) {
            return true
          }
          if (date.getTime() < Date.now() - 1000 * 60 * 60 * 24 * 30 * 6) {
            return true
          }
        }
      },
      endDateRestrict: {
        disabledDate (date) {
          if (date.getTime() < dayjs(_this.startDate).unix() * 1000) {
            return true
          }
          if (date.getTime() > Date.now()) {
            return true
          }
        }
      },
      // 表列
      columns: [
        {
          title: '代理',
          key: 'agentId',
          width: 80,
          render (h, params) {
            return (
              _this.siteOption &&
              _this.siteOption.map((item) => {
                if (item.agentList) {
                  return item.agentList.map((agent) => {
                    if (agent.id == params.row.agentId) {
                      return h('span', {}, agent.name)
                    }
                  })
                }
              })
            )
          }
        },
        {
          title: '游戏名称',
          key: 'gameName',
          minWidth: 250,
          render (h, params) {
            return h('span', {}, params.row.gameName)
          }
        },
        {
          title: '局号',
          key: 'roundID',
          width: 250,
          render (h, params) {
            return <div>{params.row.roundID}</div>
          }
        },
        { title: '用户ID', key: 'userId', width: 80 },
        { title: '账号', key: 'account', minWidth: 80 },
        { title: '昵称', key: 'nickName', width: 150 },
        {
          title: '试玩',
          key: 'isTourist',
          width: 80,
          align: 'center',
          render (h, params) {
            return params.row.isTourist > 0 ? (
              <span style="color:red;">是</span>
            ) : (
              <span style="color:green;">否</span>
            )
          }
        },
        { title: 'Symbol', key: 'symbol', width: 150 },
        {
          title: '状态',
          key: 'complete',
          width: 100,
          render: (h, params) => {
            return params.row.complete ? (
              <span style="color:green">完成</span>
            ) : (
              <span style="color:red">未完成</span>
            )
          }
        },
        {
          title: '流水',
          align: 'center',
          width: 80,
          render: (h, params) => {
            return h(
              'a',
              {
                attrs: {
                  href:
                    '/players-record-' +
                    params.row.userId +
                    '?agent=' +
                    params.row.agentId +
                    '&on=' +
                    params.row.roundID
                }
              },
              '查询'
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 80,
          render: (h, params) => {
            return h(
              'a',
              {
                class: 'copy-btn',
                on: {
                  click: () => {
                    _this.handleCopy(params.row.log)
                  }
                }
              },
              '复制'
            )
          }
        },
        { title: '有效下注', key: 'bet', width: 100 },
        {
          title: '返奖',
          key: 'win',
          width: 100,
          render: (h, params) => {
            return params.row.win > 0 ? (
              <span style="color:green">
                {Number(params.row.win).toFixed(2)}
              </span>
            ) : (
              <span style="color:red">{Number(params.row.win).toFixed(2)}</span>
            )
          }
        },
        {
          title: '货币',
          key: 'currency',
          width: 80
        },
        {
          title: '索引',
          key: 'rowVersion',
          width: 220
        },
        {
          title: '对局时间',
          key: 'playedDate',
          width: 180,
          render (h, params) {
            return <span>{getDate(params.row.playedDate)}</span>
          }
        }
      ],
      tableData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },

      siteOption: [],
      agentOption: [],
      games: {
        value: '',
        option: []
      },
      /**
       * 补充的查询参数
       */
      params: {
        webId: null,
        agentId: null,
        officeNumber: null,
        startTime: null,
        endTime: null,
        page: 1,
        pageSize: 10,
        gameId: 0,
        hash: null,
        complete: true,
        account: null,
        nickName: null
      }
    }
  },
  methods: {
    // 在methods中
    handleCopy (text) {
      const clipboard = new Clipboard('.copy-btn', {
        text: () => text // 设置要复制的文本
      })
      clipboard.on('success', () => {
        this.$Message.success('复制成功')
        clipboard.destroy() // 销毁实例
      })
      clipboard.on('error', () => {
        this.$Message.error('复制失败')
        clipboard.destroy()
      })
    },

    /**
     * 切换站点
     */
    siteChanged (siteId) {
      this.agent = 9999999
      sessionStorage.setItem('siteVal', siteId)
      this.agentOption = this.siteOption.find(
        (site) => site.id == siteId
      ).agentList
      if (!this.agentOption.find((x) => x.id == 9999999)) {
        this.agentOption.unshift({
          name: '全部',
          id: 9999999
        })
      }
    },
    handleSearch () {
      this.spinShow = true
      this.params.gameId =
        this.games.value === '' || this.games.value === 0
          ? null
          : this.games.value
      this.params.agentId =
        this.params.agentId === 9999999 ? undefined : this.params.agentId
      if (this.startDate && this.endDate) {
        this.params.startTime = dayjs(
          dayjs(this.startDate.getTime()).format('YYYY-MM-DD HH:mm:ss')
        ).unix()
        this.params.endTime = dayjs(
          dayjs(this.endDate.getTime()).format('YYYY-MM-DD HH:mm:ss')
        ).unix()
      } else {
        this.params.startTime = null
        this.params.endTime = null
      }
      let params = { ...this.params, ...this.pageData }
      delete params.pageOpts
      delete params.current
      getSettlement(params).then(({ data }) => {
        let result = data.data
        if (data.code === 200) {
          this.tableData = []
          this.pageData.page = result.page
          this.pageData.current = result.total
          this.tableData = (result && result.data) || []
        } else {
          this.$Message.error('接口报错')
        }
        this.spinShow = false
      })
    },

    exportSearch () {
      this.spinShow = true
      this.params.gameId =
        this.games.value === '' || this.games.value === 0
          ? null
          : this.games.value
      this.params.agentId =
        this.params.agentId === 9999999 ? undefined : this.params.agentId
      if (this.startDate && this.endDate) {
        this.params.startTime = dayjs(
          dayjs(this.startDate.getTime()).format('YYYY-MM-DD HH:mm:ss')
        ).unix()
        this.params.endTime = dayjs(
          dayjs(this.endDate.getTime()).format('YYYY-MM-DD HH:mm:ss')
        ).unix()
      } else {
        this.params.startTime = null
        this.params.endTime = null
      }
      let params = { ...this.params, ...this.pageData }
      delete params.pageOpts
      delete params.current
      getExportSettlementCount(params).then(({ data }) => {
        let result = data.data
        if (result >= 10000) {
          this.spinShow = false
          this.$Message.error('导出数据量大于1w条，请修改查询条件再导出！')
        } else {
          getExportSettlements(params).then(({ data }) => {
            let result = []
            data.data.data.forEach((item) => {
              item.playedDate = dayjs(item.playedDate).format(
                'YYYY-MM-DD HH:mm:ss'
              )
              result.push(item)
            })
            const columns = [
              { title: '玩家id', key: 'userId', width: 100 },
              { title: '账号', key: 'account', width: 100 },
              { title: 'Symbol', key: 'symbol', width: 100 },
              { title: '游戏名称', key: 'gameName', width: 100 },
              { title: '昵称', key: 'nickName', width: 100 },
              { title: '局号', key: 'roundID', width: 100 },
              { title: '游戏时间', key: 'playedDate', width: 100 },
              { title: '代理id', key: 'agentId', width: 80 },
              { title: '余额', key: 'balance', width: 80 },
              { title: '下注', key: 'bet', width: 80 },
              { title: '返奖', key: 'win', width: 80 },
              { title: '是否完成', key: 'complete', width: 80 },
              { title: '税收', key: 'revenue', width: 80 },
              { title: '货币', key: 'currency', width: 80 }
            ]
            const zip = new JSZip()
            let ec = exportExcel(columns, result, 0)
            zip.file(`注单.xlsx`, ec, { binary: true })
            // 生成zip文件并下载
            zip.generateAsync({ type: 'blob' }).then((content) => {
              FileSaver.saveAs(content, `注单${new Date().getTime()}.zip`)
            })
            this.spinShow = false
          })
        }
      })
    },

    imgClick (row) {
      // let path = window.location.protocol + "//" + url;
      // window.location.href = path;
      this.viewGameDetailUrl = this.replays[0] + '/share/' + row.hash
      this.viewSettlementDetail = true
    },
    handleAllSearch () {},
    changePage (index) {
      this.pageData.page = index
      this.handleSearch()
    },
    changePageSize (index) {
      this.pageData.pageSize = index
      this.handleSearch()
    },
    /**
     * set sit
     */
    setSite (val) {
      this.webId = val
      if (val > 0) {
        sessionStorage.setItem('siteVal', val)
        this.agentOption = this.siteOption.find(
          (site) => site.id == val
        ).agentList
        if (!this.agentOption.find((x) => x.id == 9999999)) {
          this.agentOption.unshift({
            name: '全部',
            id: 9999999
          })
        }
      }
    },

    loadGames () {
      let newArray = JSON.parse(sessionStorage.getItem('games'))
      this.games.option.push(
        { id: 0, number: 0, name: '全部', nameZH: '' },
        ...newArray
      )
      this.games.option.map((item) => {
        if (item.nameZH != '') {
          item.label = item.name + ' [' + item.nameZH + ']'
        } else {
          item.label = item.name
        }
        return item
      })
    }
  },

  mounted () {
    /**
     * 填充页面的站点参数
     */
    getGameServers().then((data) => {
      if (data.data.data) this.replays = data.data.data.data.replays
    })
    let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
    let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
    void sid
    this.siteOption = siteOption
    this.loadGames()
    this.handleSearch()
  }
}
</script>

<style lang="less">
.label-style {
  font-size: 16px;
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
.inlineForm {
  display: flex;
  align-items: center;
  padding: 10px 0;
}
</style>
