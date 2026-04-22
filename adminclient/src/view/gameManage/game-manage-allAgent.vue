<template>
  <div>
    <Card>
      <ul style="height: 43px;">
        <div style="float: right; padding: 5px 0 0;">
          <Button type="error" @click="stopAllGame" style="margin-right:20px;">冻结所有代理此游戏</Button>
          <Button type="primary" @click="startAllGame" style="margin-right:20px;">解冻所有代理此游戏</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <tables
        ref="tables"
        @on-selection-change="setSelections"
        v-model="tableData"
        :columns="columns"
      />
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
    </Card>
    <Modal
      v-model="modalEdit.state"
      title="修改游戏房间配置"
      width="1200"
      :styles="{ top: '20px' }"
      class="modalEditBox"
      @on-ok="submitConfig"
    >
      <Row
        v-for="(item, index) in configInfo"
        :key="item.id"
        :gutter="16"
        style="padding-right: 50px;"
      >
        <template v-for="items in Object.keys(item)">
          <i-col v-if="mapConfigTitle(items)" :key="items" class="config-item">
            <i-input
              :maxlength="50"
              :type="'text'"
              v-model="item[items]"
              :placeholder="mapConfigTitle(items)"
            >
              <span slot="prepend">{{ mapConfigTitle(items) }}</span>
            </i-input>
          </i-col>
        </template>
        <Button
          v-if="index"
          icon="md-remove"
          style="height: 26px;width: 26px;"
          title="移除配置项"
          shape="circle"
          @click="removeConfigItem(index)"
        ></Button>
      </Row>
      <Button type="info" icon="md-add" style="margin-top:10px;" @click="addConfigItem">添加配置项</Button>
    </Modal>
  </div>
</template>

<script>
import Tables from '_c/tables'
import axios from '@/libs/api.request'
import { setting } from '@/config'
import {
  getGameData,
  getGameData2,
  editGameData,
  generateHallCache
} from '@/api/data'
import { getToken } from '@/libs/util'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  data () {
    return {
      currentSelectons: [],
      req: [
        {
          label: '游戏名称',
          key: 'gameId',
          value: '',
          option: []
        }
        // {
        //   label: "游戏平台",
        //   key: "platformId",
        //   value: "",
        //   option: [],
        // },
      ],
      configInfo: [],
      modalEdit: {
        id: '',
        state: false,
        agentId: '',
        gameId: '',
        gameServerId: ''
      },
      columns: [
        {
          type: 'selection',
          width: 60,
          align: 'center'
        },
        { title: '游戏ID', key: 'id', align: 'center', minWidth: 95 },
        { title: '游戏名称', key: 'name', align: 'center', minWidth: 95 },

        {
          title: '游戏状态',
          key: 'state',
          align: 'center',
          minWidth: 95,
          render (h, params) {
            return params.row.status == 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:orange">维护</span>
            )
          }
        }
        // {
        //   title: "操作",
        //   key: "handle",
        //   width: 100,
        //   align: "center",
        //   button: [
        //     // (h, params) => {
        //     //   return h(
        //     //     "Button",
        //     //     {
        //     //       props: {
        //     //         type: "primary",
        //     //         size: "small"
        //     //       },
        //     //       style: {
        //     //         marginRight: "5px"
        //     //       },
        //     //       on: {
        //     //         click: () => {
        //     //           this.showGameConfig(params);
        //     //         }
        //     //       }
        //     //     },
        //     //     "房间配置"
        //     //   );
        //     // },
        //     // (h, params) => {
        //     //   return h(
        //     //     "Button",
        //     //     {
        //     //       props: {
        //     //         type: "info",
        //     //         size: "small"
        //     //       },
        //     //       style: {
        //     //         marginRight: "5px"
        //     //       },
        //     //       on: {
        //     //         click: () => {
        //     //           this.$Modal.confirm({
        //     //             title: "自定义配置",
        //     //             align: "center",
        //     //             width: 600,
        //     //             render: h => {
        //     //               return h("Input", {
        //     //                 props: {
        //     //                   type: "textarea",
        //     //                   rows: 6,
        //     //                   value: params.row.extra
        //     //                 },
        //     //                 on: {
        //     //                   input: val => {
        //     //                     params.row.extra = val;
        //     //                   }
        //     //                 }
        //     //               });
        //     //             },
        //     //             onOk: () => {
        //     //               let Data = {
        //     //                 id: params.row.id,
        //     //                 agentId: params.row.agentId,
        //     //                 extra: params.row.extra,
        //     //                 gameId:params.row.gameId
        //     //               };
        //     //               configGameData(Data).then(res => {
        //     //                 if (res.data.code == 200) {
        //     //                   this.$Message.success("自定义配置成功");
        //     //                 }
        //     //               });
        //     //             }
        //     //           });
        //     //         }
        //     //       }
        //     //     },
        //     //     "自定义配置"
        //     //   );
        //     // },
        //     (h, params) => {
        //       return [
        //         h(
        //           "Poptip",
        //           {
        //             props: {
        //               transfer: true,
        //               confirm: true,
        //               placement: "left",
        //               title:
        //                 "您确定要" +
        //                 (params.row.isFrozen == 0 ? "冻结" : "启用") +
        //                 "游戏吗?",
        //             },
        //             style: { textAlign: "left", zIndex: "99" },
        //             on: {
        //               "on-ok": () => {
        //                 let data = {
        //                   agentId: params.row.agentId,
        //                   id: params.row.id,
        //                   isFrozen: params.row.isFrozen == 0 ? 1 : 0,
        //                 };
        //                 editGameState(data).then((res) => {
        //                   if (res.data.code == 200) {
        //                     this.$nextTick(() => {
        //                       this.handleSearch();
        //                       this.$Message.success(res.data.msg);
        //                     });
        //                   } else if (res.data.code == 400) {
        //                     this.$Message.error(res.data.msg);
        //                   }
        //                 });
        //               },
        //             },
        //           },
        //           [
        //             h(
        //               "Button",
        //               {
        //                 props: {
        //                   type: params.row.isFrozen == 0 ? "error" : "success",
        //                   size: "small",
        //                 },
        //               },
        //               params.row.isFrozen == 0 ? "冻结" : "启用"
        //             ),
        //           ]
        //         ),
        //       ];
        //     },
        //   ],
        // },
      ],
      tableData: [],
      editData: [],
      mainStartTime: '',
      mainEndTime: '',
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize
      },
      spinShow: false
    }
  },
  methods: {
    /**
     * 批量冻结解冻
     */
    setSelections (e) {
      this.currentSelectons = e
    },
    async stopAllGame () {
      if (this.currentSelectons.length == 0) {
        this.$Message.error('请选择要冻结的游戏')
        return
      }
      let gameId = this.currentSelectons.map((x) => x.number)
      gameId = gameId.join(',')
      let data = await axios.request({
        url: 'v2/game/changeStatus',
        method: 'post',
        params: {
          token: getToken(),
          status: 0,
          gameId
        }
      })
      let result = data.data
      if (result && result.code == 200) {
        this.$Message.success('操作成功')
        this.handleSearch()
      }
    },
    async startAllGame () {
      if (this.currentSelectons.length == 0) {
        this.$Message.error('请选择要解冻的游戏')
        return
      }
      let gameId = this.currentSelectons.map((x) => x.number)
      gameId = gameId.join(',')
      let data = await axios.request({
        url: 'v2/game/changeStatus',
        method: 'post',
        params: {
          token: getToken(),
          status: 1,
          gameId
        }
      })
      let result = data.data

      if (result && result.code == 200) {
        this.$Message.success('操作成功')
        this.handleSearch()
      }
    },
    mapConfigTitle (items) {
      return this.currentConfigTitle[items]
    },
    showGameConfig (params) {
      /**
       * 按照配置转换游戏配置
       */
      let gameConfig =
        setting.configName.find((config) => {
          return config.id.includes(params.row.gameServerId)
        }) || setting.defaultConfigName

      this.currentConfigTitle = gameConfig.list

      this.configInfo = JSON.parse(JSON.stringify(params.row.config))
      this.modalEdit.id = params.row.id
      this.modalEdit.state = true
      this.modalEdit.agentId = params.row.agentId
      this.modalEdit.gameId = params.row.gameId
      this.modalEdit.gameServerId = params.row.gameServerId
    },
    generateHallCachefun () {
      let Data = { agentId: sessionStorage.getItem('agentVal') }
      generateHallCache(Data).then((res) => {
        this.$Message.success(res.data.msg)
        setTimeout(() => {
          this.handleSearch()
        }, 500)
      })
    },
    setLinkage (i, key) {
      if (key == 'webId') {
        let [k, m] = [null, null]
        for (const j in this.req) {
          if (this.req[j].key == 'webId') k = j
          if (this.req[j].key == 'agentId') m = j
        }
        this.req[m].option = []
        this.req[m].value =
          sessionStorage.getItem('agentVal') != null
            ? Number(sessionStorage.getItem('agentVal'))
            : ''
        this.req[k].option.forEach((item) => {
          if (item.agentList.length > 0 && item.id == i) {
            this.req[m].option.push(...item.agentList)
          }
        })
      }
    },
    handleSearch () {
      this.spinShow = true
      let Data = []
      this.req.map((item) => {
        if (item.value) {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      getGameData2(Data).then((res) => {
        this.spinShow = false
        if (res.data.code == 200) {
          this.tableData = res.data.data
        } else {
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
      getGameData(Data).then((res) => {
        this.spinShow = false
        if (res.data.code == 200) {
          this.tableData = []
          this.pageData.page = 1
          this.tableData = res.data.data.data
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
    submitConfig () {
      let Data = []
      if (
        Object.values(this.configInfo[this.configInfo.length - 1]).indexOf('') >
        -1
      ) {
        this.$Message.warning('配置项不能为空')
      } else {
        const configData = this.configInfo.map((item, index) => {
          item.id = index + 1
          return { [index + 1]: item }
        })
        if (this.configInfo.length > 0) {
          Data = {
            id: this.modalEdit.id,
            agentId: this.modalEdit.agentId,
            gameId: this.modalEdit.gameId,
            config: JSON.stringify(Object.assign(...configData))
          }
        }
        editGameData(Data).then((res) => {
          if (res.data.code == 200) {
            this.$Message.success('修改配置成功')
            this.handleSearch()
          } else {
            this.$Message.error('修改配置失败：' + res.data.msg)
          }
          this.configInfo = []
        })
      }
    },
    addConfigItem () {
      let newConfig = {
        id: this.configInfo.length + 1
      }
      Object.keys(this.currentConfigTitle).map((key) => {
        newConfig[key] = ''
      })
      this.configInfo.push(newConfig)
    },
    removeConfigItem (index) {
      if (index > 0) {
        this.configInfo.splice(index, 1)
      }
    },
    changeConfigName (name, gid = 1) {
      let names = null
      setting.configName.forEach((item, index) => {
        if (item.id.indexOf(gid) != -1) {
          names = { ...item.list }[name] || name
        }
      })
      if (names) {
        return names
      } else return { ...setting.configNameDefault }[name] || name
    }
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
.config-item {
  display: inline;
  line-height: 40px;
  > .ivu-input-wrapper {
    width: 200px;
    display: inline-table;
  }
  > label {
    display: inline-block;
    margin-left: 10px;
    margin-right: 5px;
  }
  .modalEditBox div.ivu-modal-content div.ivu-modal-body {
    min-height: 75px !important;
  }
}
</style>
