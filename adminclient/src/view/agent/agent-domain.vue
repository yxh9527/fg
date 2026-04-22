<template>
  <div>
    <Card>
      <div>
        <span>游戏客户端地址:</span>
        <span style="margin-left: 20px">{{ gameUrl }}</span>
      </div>
      <div>
        <span>游戏回放地址:</span>
        <span style="margin-left: 20px">{{ replay }}</span>
      </div>
      <div>
        <span style="margin-left: 20px"
          ><Button type="primary" @click="openUpdateGameUrl">修改</Button></span
        >
        <div style="color: red; margin-top: 10px">
          <b>注意:</b>修改游戏客户端地址，修改后要等10s左右才能生效
        </div>
      </div>
      <Modal v-model="showEditGameUrl" title="修改游戏客户端地址">
        <div slot="footer" align="center">
          <Button
            class="btn"
            size="default"
            type="primary"
            @click="gameUrlSaveHandler"
            >确定</Button
          >
        </div>
        <Form>
          <FormItem label="游戏客户端地址" prop="">
            <Input
              type="text"
              v-model="gameUrl"
              placeholder="https://127.0.0.1:1234"
            >
            </Input>
          </FormItem>
          <FormItem label="游戏回放地址" prop="">
            <Input
              type="text"
              v-model="replay"
              placeholder="https://127.0.0.1:1234"
            >
            </Input>
          </FormItem>
        </Form>
      </Modal>
    </Card>
    <Card style="margin-top: 10px">
      <div style="margin: 10px">
        <Button type="primary" @click="openAddModal">添加配置</Button>
      </div>
      <Modal
        v-model="showConfigModal"
        :title="configModalType ? '编辑代理配置' : '添加代理配置'"
      >
        <div slot="footer" align="center">
          <Button
            class="btn"
            size="default"
            type="default"
            @click="handleConfigModal(true)"
            >取消</Button
          >
          <Button
            class="btn"
            size="default"
            type="primary"
            @click="handleConfigModal(false)"
            >确定</Button
          >
        </div>
        <Form>
          <FormItem label="配置名称" prop="">
            <Input type="text" v-model="modalData.name" placeholder="默认配置">
            </Input>
          </FormItem>
          <!-- <FormItem label="Client Api 地址" prop="">
            <Input
              type="text"
              v-model="modalData.client_api_urls"
              placeholder="https://127.0.0.1:1080/client"
            >
            </Input>
          </FormItem> -->
          <FormItem label="大厅地址" prop="">
            <Input
              type="text"
              v-model="modalData.hall_urls"
              placeholder="https://127.0.0.1:1080/hall"
            >
            </Input>
          </FormItem>
          <FormItem label="最大分数" prop="">
            <Input type="number" v-model="modalData.max_score" placeholder="0">
            </Input>
          </FormItem>
          <FormItem label="最小分数" prop="">
            <Input type="number" v-model="modalData.min_score" placeholder="0">
            </Input>
          </FormItem>
        </Form>
      </Modal>
      <tables ref="tables" v-model="domainList" :columns="configColumns" />
      <div style="margin-top: 20px; text-align: center">
        <Page
          :total="total"
          :current.sync="ListPage"
          :page-size="20"
          @on-change="getDomainConfig"
        />
      </div>
    </Card>
  </div>
</template>

<script>
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import Tables from '_c/tables'

export default {
  name: 'agency-domain',
  components: {
    Tables
  },
  inject: ['handleLogOut', 'reFreshSiteAgentList'],
  data () {
    return {
      gameUrl: '',
      replay: '',
      showEditGameUrl: false,
      // 添加弹出框
      showConfigModal: false,
      configModalType: 0, // 0添加 1修改
      // 添加弹出框数据
      modalData: {
        name: '',
        client_api_urls: '',
        hall_urls: '',
        max_score: 0,
        min_score: 0
      },
      // 当前页
      ListPage: 1,
      total: 0,
      // 当前页数据
      domainList: [],
      // 数据表格
      configColumns: [
        { title: '序号', type: 'index', align: 'center' },
        { title: '名称', key: 'name', align: 'center' },
        // { title: "Client Api 地址", key: "client_api_urls", align: "center" },
        { title: '大厅地址', key: 'hall_urls', align: 'center' },
        { title: '最大分数', key: 'max_score', align: 'center' },
        { title: '最小分数', key: 'min_score', align: 'center' },
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
                      this.modalData = params.row || {}
                      this.showConfigModal = true
                      this.configModalType = 1
                    }
                  }
                },
                '编辑'
              )
            },
            (h, params) => {
              return [
                h(
                  'Button',
                  {
                    props: {
                      type: 'error',
                      size: 'small'
                    },
                    on: {
                      click: () => {
                        this.deleteConfig(params.row.id)
                      }
                    }
                  },
                  '删除'
                )
              ]
            }
          ]
        }
      ]
    }
  },
  methods: {
    // 获取默认api和域名设置
    async getDomainConfig (page) {
      let params = {
        token: getToken(),
        pageSize: 20,
        page: page || this.ListPage
      }
      let data = await axios.request({
        url: 'v2/game/apiConfigList',
        method: 'post',
        params
      })
      // 设置默认配置
      if (data && data.data && data.data.code == 200) {
        this.domainList = data.data.data.data
        this.total = data.data.data.total
      }
    },

    // 获取客户端游戏地址
    async loadGameUrl () {
      let params = {
        token: getToken()
      }
      return axios.request({
        url: 'v2/game/getGameUrl',
        method: 'post',
        params
      })
    },

    async updateGameUrl (d, r) {
      let data = {
        token: getToken(),
        gameUrl: d,
        replay: r
      }
      return axios.request({
        url: 'v2/game/updateGameUrl',
        method: 'post',
        params: data
      })
    },

    // 打开添加弹出框
    openAddModal () {
      this.showConfigModal = true
      this.configModalType = 0
      // 重置弹出框数据
      this.modalData = {
        name: '',
        client_api_urls: '',
        hall_urls: '',
        max_score: 0,
        min_score: 0
      }
    },

    async gameUrlLoadHandler () {
      let data = await this.loadGameUrl()
      console.log(data)
      if (data && data.data && data.data.code == 200) {
        this.gameUrl = data.data.data.game_url.join(',')
        this.replay = data.data.data.replays.join(',')
      }
    },

    // 打开添加弹出框
    openUpdateGameUrl () {
      this.showEditGameUrl = true
    },

    async gameUrlSaveHandler () {
      await this.updateGameUrl(this.gameUrl, this.replay)
    },

    // 添加代理配置
    async handleConfigModal (isCancel) {
      if (isCancel) {
        this.showConfigModal = false
        return
      }
      if (
        this.modalData.name == '' ||
        // this.modalData.client_api_urls == "" ||
        this.modalData.hall_urls == ''
      ) {
        this.$Message.error('不能为空')
        return
      }

      if (this.modalData.max_score == 0 || this.modalData.min_score == 0) {
      }

      ((await this.configModalType) && this.updateConfig()) || this.AddConfig()
      this.modalData = {
        name: '',
        client_api_urls: '',
        hall_urls: '',
        max_score: 0,
        min_score: 0
      }
    },

    // add
    async AddConfig () {
      let params = JSON.parse(JSON.stringify(this.modalData))
      params.token = getToken()
      let data = await axios.request({
        url: 'v2/game/apiConfigAdd',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.$Message.info(data.data.msg)
        this.getDomainConfig(1)
        this.ListPage = 1
        this.showConfigModal = false
      } else {
        this.$Message.error(data.data.msg)
      }
    },

    // update
    async updateConfig () {
      let params = JSON.parse(JSON.stringify(this.modalData))
      params.token = getToken()
      let data = await axios.request({
        url: 'v2/game/apiConfigUpdate',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.getDomainConfig()
        this.showConfigModal = false
      }
    },

    // delete
    async deleteConfig (id) {
      let dofunc = async () => {
        let params = {
          token: getToken(),
          id
        }
        let data = await axios.request({
          url: 'v2/game/apiConfigDel',
          method: 'post',
          params
        })
        if (data && data.data && data.data.code == 200) {
          this.getDomainConfig(1)
        } else {
          this.$Message.error(data.data.msg)
        }
      }
      this.$Modal.confirm({
        title: '请确认',
        content: '确认删除配置吗？',
        onOk: dofunc
      })
    }
  },
  mounted () {
    // 获取初始化列表
    this.getDomainConfig(1)
    this.gameUrlLoadHandler()
  }
}
</script>

<style lang="less">
</style>
