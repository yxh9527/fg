<template>
  <div>
    <Tabs @on-click="changeTab" value="name1" :animated="false">
      <TabPane label="游戏设置" name="name1">
        <Card style="margin: 0 0 10px">
          <div style="display: flex; align-items: center">
            <div style="display: flex; align-items: center">
              玩家数据重置周期：
            </div>
            <div>
              <InputNumber
                :max="31"
                :min="0"
                :step="1"
                :precision="0"
                v-model="userControlTimeRange"
              ></InputNumber>
              <div style="display: inline-block; margin-left: 5px">天</div>
              <Button
                @click="userControlTimeRangeChange"
                type="primary"
                style="margin-left: 10px"
                >保存</Button
              >
            </div>
            <div style="margin-left: 20px; color: red">
              <b>下次重置时间</b>:{{ this.cycleRestTime }}
            </div>
          </div>
        </Card>
        <Card style="margin: 10px 0">
          <div class="inlineForm">
            <div>水池重置周期（天）：</div>
            <RadioGroup v-model="resetParams.interval">
              <Radio
                v-for="item in radioData"
                :label="item.value"
                :key="item.value"
              >
                <span>{{ item.label }}</span>
              </Radio>
            </RadioGroup>
            <div style="padding-left: 20px">
              下次重置时间：<label>{{ this.resetParams.nextRestTime }}</label>
            </div>
            <div style="padding-left: 20px">
              <Button
                @click="poolResetTimeRangeChange"
                style="margin-left: 20px"
                type="primary"
              >
                保存
              </Button>
              <Button
                @click="poolResetNow"
                style="margin-left: 20px"
                type="primary"
              >
                立即重置
              </Button>
              <label style="margin-left: 20px; color: red"
                ><b>说明</b
                >:水池重置,到周期后。系统会保留亏损水池,重置盈利水池！<b
                  >所有修改都会延迟最长2分钟执行！</b
                >
              </label>
            </div>
          </div>
        </Card>
        <Card>
          <div class="inlineForm">
            <div style="width: 500px">
              <span>游戏选择：</span>
              <Select
                v-model="paramgame"
                style="width: 400px"
                clearable
                filterable
                filter-by-label
              >
                <Option
                  v-for="item in gameOptions"
                  :value="item.number"
                  :key="item.number"
                  :label="item.label"
                  >{{ item.label }}
                </Option>
              </Select>
            </div>
            <div style="margin-left: 10px; width: 300px">
              <span>代理选择：</span>
              <Select
                :placeholder="agentParams.msg"
                filterable
                filter-by-label
                clearable
                v-model="agentParams.val"
                style="width: 200px; margin-left: 5px"
              >
                <Option
                  v-for="item in agentParams.option"
                  :value="item.id"
                  :key="item.id"
                  :label="item.name"
                >
                  {{ item.name }}</Option
                >
              </Select>
            </div>
            <Button
              @click="agentSearch"
              style="margin-left: 20px"
              type="primary"
              >搜索</Button
            >

            <Button
              @click="addRoomInfo"
              style="margin-left: 20px"
              type="primary"
              >添加配置</Button
            >
          </div>
        </Card>

        <Card style="margin: 10px 0; min-height: 150px" title="水池配置">
          <Table :columns="agentColumns" :data="agentData"> </Table>
        </Card>

        <Modal
          v-model="editRoomInfo"
          :title="currentEditRoom.tag"
          @on-ok="confirmEditRoomInfo"
        >
          <div v-if="currentEditRoom.data">
            <div>
              游戏名称：{{
                (currentEditRoom.data.nameZH &&
                  currentEditRoom.data.nameZH.length) > 0
                  ? currentEditRoom.data.name +
                    " [" +
                    currentEditRoom.data.nameZH +
                    "]"
                  : currentEditRoom.data.name
              }}
            </div>
            <div>
              <span>正常水位</span>
              <Input
                type="number"
                v-model.number="currentEditRoom.data.normal"
              />
              <span>初始水位比例</span>
              <Input
                type="number"
                v-model.number="currentEditRoom.data.normalRate"
              />
            </div>
            <div>
              <span>最高水位</span>
              <Input type="number" v-model.number="currentEditRoom.data.max" />
              <span>最高水位比例</span>
              <Input
                type="number"
                v-model.number="currentEditRoom.data.maxRate"
              />
            </div>
            <div>
              <span>最低水位</span>
              <Input type="number" v-model.number="currentEditRoom.data.min" />
              <span>最低水位比例</span>
              <Input
                type="number"
                v-model.number="currentEditRoom.data.minRate"
              />
            </div>
            <div>
              <span>税收比例</span>
              <Input
                type="number"
                v-model.number="currentEditRoom.data.revenue"
              />
            </div>
            <div>
              <span>基数</span>
              <Input type="number" v-model.number="currentEditRoom.data.base" />
            </div>
          </div>
          <p v-if="currentEditRoom.agentId >= 0">
            <span style="color: red">*保存为代理配置</span>
          </p>
          <p v-else>
            <span style="color: red">*保存为默认配置</span>
          </p>
        </Modal>

        <Modal
            v-model="showBairenGameSettingData"
            title="游戏算法设置"
            @on-ok="saveBairenGameSettingData(currentBairenGameSettingData.isAdd)"
        >
            <div v-if="currentBairenGameSettingData">
                <div>
                    <span>名字</span>
                    <Input type="text" v-model="currentBairenGameSettingData.name" :disabled="currentBairenGameSettingData.name=='single'||currentBairenGameSettingData.name=='default'"/>
                    <span>盈亏比例Min</span>
                    <Input type="number" v-model="currentBairenGameSettingData.min" />
                    <span>盈亏比例Max</span>
                    <Input type="number" v-model="currentBairenGameSettingData.max" />
                    <span>低水位中奖概率</span>
                    <Input type="number" v-model="currentBairenGameSettingData.low_odds" />
                    <span>低水位中奖倍数</span>
                    <Input type="number" v-model="currentBairenGameSettingData.low_multiple" />
                    <span>正常水位中奖概率</span>
                    <Input type="number" v-model="currentBairenGameSettingData.normal_odds" />
                    <span>正常水位中奖倍数</span>
                    <Input type="number" v-model="currentBairenGameSettingData.normal_multiple" />
                    <span>高水位中奖概率</span>
                    <Input type="number" v-model="currentBairenGameSettingData.high_odds" />
                    <span>高水位中奖倍数</span>
                    <Input type="number" v-model="currentBairenGameSettingData.high_multiple" />
                </div>
            </div>
            <div v-else>
              <span>暂无配置</span>
            </div>
        </Modal>
      </TabPane>
      <TabPane label="分段奖励配置" name="name2">
        <Card style="margin: 10px 0; min-height: 200px" title="分段奖励配置">
          <Table :columns="gameSettingColumn" :data="gameSettingData"></Table>
        </Card>
      </TabPane>
      <TabPane label="库存预警" name="name3">
        <Card>
          <div class="inlineForm">
            <div>站点选择：</div>
            <div>
              <Select
                placeholder="选择站点"
                clearable
                filterable
                filter-by-label
                v-model="siteParams.val"
                @on-change="setSiteSession"
                style="width: 200px"
              >
                <Option
                  v-for="item in siteParams.option"
                  :value="item.id"
                  :key="item.id"
                  :label="item.name"
                >
                  {{ item.name }}</Option
                >
              </Select>
            </div>
            <div style="padding-left: 20px">代理选择：</div>
            <div>
              <Select
                :placeholder="agentParams.msg"
                clearable
                filterable
                filter-by-label
                v-model="agentParams.val"
                style="width: 200px"
              >
                <Option
                  v-for="item in agentParams.option"
                  :value="item.id"
                  :key="item.id"
                  :label="item.name"
                >
                  {{ item.name }}</Option
                >
              </Select>
            </div>
            <div style="margin-left: 20px">游戏选择：</div>
            <div>
              <Select
                v-model="paramgame"
                style="width: 400px"
                clearable
                filterable
                filter-by-labe
              >
                <Option
                  v-for="item in gameOptions"
                  :value="item.number"
                  :key="item.number"
                  :label="item.label"
                  >{{ item.label }}
                </Option>
              </Select>
            </div>
            <Button
              @click="searchStock"
              style="margin-left: 20px"
              type="primary"
              >搜索</Button
            >
            <Button style="margin-left: 20px" @click="renderLineChart"
              >生成曲线图</Button
            >
          </div>
        </Card>

        <Card style="margin: 10px 0">
          <Table :columns="stockColumns" :data="stockData"></Table>
          <div style="margin-top: 20px; text-align: center">
            <Page
              :total="stockTotal"
              :current.sync="stockPage"
              :page-size="20"
              @on-change="searchStock"
            />
          </div>
        </Card>
      </TabPane>

      <TabPane label="汇率配置" name="hlpz">
        <Card style="margin: 10px 0; min-height: 200px;" title="汇率配置">
          <div style="height: 800px; overflow-y: auto;">
            <Button type="primary" style="margin: 5px 0px 5px 0px" @click="exchangeEditType = 2; exchangeEdit = true;">新增汇率</Button>
            <Table :columns="exchangeSettings" :data="exchangeData"></Table>
          </div>

        </Card>

        <Modal v-model="exchangeEdit" title="汇率配置" :width="700">
          <Card>
            <div class="inlineForm">
              货币符号:<Input type="text" style="width:180px"  v-model="currentCurency.currency" />
              <span style="margin-left:15px">汇率:<Input type="number" style="width:180px;"  v-model="currentCurency.exchange"/></span>
              <Button type="primary" style="margin-left: 15px" @click="EditCurrency(); exchangeEdit = false;">保存</Button>
            </div>
          </Card>
          <div slot="footer"></div>
        </Modal>

      </TabPane>

    </Tabs>

    <Modal v-model="renderLineChartModal" title="曲线图" :width="840">
      <Card>
        <div id="renderLineChartDom" style="height: 300px; width: 800px"></div>
      </Card>
      <div slot="footer"></div>
    </Modal>
  </div>
</template>

<script>
// import the component
import {
  getLinkageList
} from '@/api/data'
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import * as dayjs from 'dayjs'
import * as echarts from 'echarts'

export default {
  mounted () {
    // this.resetParams.interval = 7;
    this.initAgentData()
    this.getPoolResetInfo()
    this.getSite() // 获取玩家控制的头部代理参数
    window.test1 = this
    this.getUserControlTimeRange()
    this.getGameSettingList()
    this.getCurrencyData()
    // 监听网页加载完成
    document.onreadystatechange = () => {
      if (document.readyState == 'complete') {
        // let dom = document.querySelector("#renderLineChartDom");
        // this.myChart = echarts.init(dom);
      }
    }
  },
  data () {
    let _this = this
    return {
      value19: '',
      // 新手保护
      freshmanProtectRooms: [1, 2, 3, 4],
      freshmanProtectFactor: 0,
      freshmanProtectTimes: 10,
      freshmanProtectSwitch: false,
      userControlTimeRange: 1,
      temppoolAgent: null,
      temppool: null,
      userAccount: null,
      accSessionRes: null,
      iplatform,
      // 库存合并
      stockMergeDataTotal: 0,
      stockMergeFormModa: 'edit',
      stockMergeGames: [],
      stockMergeAreas: [],
      stockMergeAgentOptions: [
        {
          id: 123,
          label: '站点1',
          children: [
            {
              id: 11,
              label: '代理1'
            },
            {
              id: 122,
              label: '代理2'
            }
          ]
        },
        {
          id: 1235,
          label: '站点2',
          children: [
            {
              id: 331,
              label: '代理22'
            },
            {
              id: 1231,
              label: '代理222'
            }
          ]
        }
      ],

      stockMergeForm: {
        poolId: 1,
        name: null,
        gameId: null,
        areaId: 1,
        normal: 0,
        normalRate: 0,
        agentIds: [],
        // rate: 0,
        max: 0,
        slope: 0,
        maxRate: 0,
        // timeRange: 0,
        min: 0,
        // highFix: 0,
        // minRate: 0,
        // lowFix: 0,
        revenue: 0,
        control: 0
      },
      stockMergeModal: false,
      stockMergeData: [
        {
          a: 1
        }
      ],

      // 库存合并
      renderLineChartModal: false,
      radioData: [
        {
          value: 1,
          label: '1天'
        },
        {
          value: 3,
          label: '3天'
        },
        {
          value: 7,
          label: '7天'
        },
        {
          value: 10,
          label: '10天'
        },
        {
          value: 14,
          label: '14天'
        },
        {
          value: 30,
          label: '30天'
        }
      ],
      resetParams: {
        interval: 7,
        nextRestTime: '',
        resetTime: 0,
        now: false
      },

      siteParams: {
        val: '',
        option: []
      }, // 控制玩家的站点
      agentParams: {
        val: -1,
        option: [],
        msg: '请先选择站点'
      }, // 控制玩家的代理

      isUpdate: false, // 点击修改展示修改框
      spinShow: false, // loading
      agentId: undefined,
      playerData: {
        countMaxBet: 0,
        userControl: {}
      },
      rewardParams: {
        page: 1,
        userId: undefined,
        nickName: undefined,
        agentId: undefined,
        webId: undefined
      },
      editRoomInfo: false,
      canEditName: true,
      currentEditRoom: { tag: '', data: null, t: 1 },
      currentCurency: {currency: '', exchange: 0, id: 0},
      exchangeEditType: 1, // 1 edit 2 add 3 del
      stockGameId: null,
      stockRoomId: null,
      stockPage: 1,
      stockTotal: 1,
      poolStartTime: null,
      poolEndTime: null,
      isAutoRefresh: true,
      cycleRestTime: 0,
      /**
       * 房间
       */
      roomAgentId: undefined,

      stockColumns: [
        {
          title: '当前库存',
          key: 'poolValue',
          align: 'center',
          render (h, params) {
            return h('span', {}, params.row.poolValue.toFixed(2))
          }
        },
        {
          title: '正常水位',
          key: 'normal',
          align: 'center'
        },
        {
          title: '正常水位比例',
          key: 'normalRate',
          align: 'center'
        },
        {
          title: '高水位',
          key: 'max',
          align: 'center'
        },
        {
          title: '高水位比例',
          key: 'maxRate',
          align: 'center'
        },
        {
          title: '低水位',
          key: 'min',
          align: 'center'
        },
        {
          title: '低水位比例',
          key: 'minRate',
          align: 'center'
        },
        {
          title: '税收比例',
          key: 'revenue',
          align: 'center'
        },
        // {
        //   title: "单局盈亏系数",
        //   key: "ctl",
        //   align: "center",
        // },
        {
          title: '时间',
          key: 'ctl',
          align: 'center',
          render (h, params) {
            return h(
              'span',
              {},
              dayjs(params.row.createTime * 1000).format('YYYY-MM-DD HH:mm')
            )
          }
        }
      ],
      stockData: [],
      myChart: null,
      gameSettingData: [],
      exchangeData: [
        // {currency:"USD",exchange:7.1}
      ],
      exchangeEdit: false,
      /**
       * 查询参数
       */
      paramgame: null,
      gameOptions: [],
      paramBairen: null,
      gameBairenGames: [],
      games: JSON.parse(sessionStorage.getItem('games') || '[]'),
      /**
       * 表格配置
       */
      agentColumns: [
        {
          title: 'ID',
          key: 'gameId',
          align: 'center',
          width: 100
        },
        {
          title: '标识',
          key: 'symbol',
          align: 'center',
          width: 150
        },
        {
          title: '游戏名称',
          key: 'name',
          align: 'center',
          render: (h, params) => {
            if (params.row.nameZH && params.row.nameZH.length > 0) {
              return (
                <span>{params.row.name + ' [' + params.row.nameZH + ']'}</span>
              )
            } else {
              return <span>{params.row.name}</span>
            }
          }
        },
        {
          title: '正常水位',
          key: 'normal',
          align: 'center',
          width: 100
        },
        {
          title: '正常比例',
          key: 'normalRate',
          align: 'center',
          width: 120
        },
        {
          title: '高水位',
          key: 'max',
          align: 'center',
          width: 100
        },
        {
          title: '高比例',
          key: 'maxRate',
          align: 'center',
          width: 100
        },
        {
          title: '低水位',
          key: 'min',
          align: 'center',
          width: 100
        },
        {
          title: '低比例',
          key: 'minRate',
          align: 'center',
          width: 100
        },
        {
          title: '税收',
          key: 'revenue',
          align: 'center',
          width: 80
        },
        {
          title: '操作',
          key: 'option',
          align: 'center',
          width: 180,
          render: (h, params) => {
            return [
              h(
                'Button',
                {
                  props: {
                    type: 'warning',
                    size: 'small',
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      this.vieweditRoomInfo(params.row)
                    }
                  }
                },
                '修改'
              ),
              h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    async click () {
                      _this.$Modal.confirm({
                        title: '确定要同步所有默认水池配置吗?',
                        async onOk () {
                          _this.syncAllPool(params.row)
                        },
                        onCancel () {}
                      })
                    }
                  }
                },
                '同步所有'
              )
            ]
          }
        }
      ],

      /**
       * 表格数据
       */
      agentData: [],
      rewardPage: 1,

      gameSettingColumn: [
        {
          title: 'id',
          key: 'id',
          align: 'center',
          width: 50
        },
        {
          title: '配置名称',
          key: 'name',
          align: 'center'
        },
        {
          title: '盈亏比例Min',
          key: 'min',
          align: 'center'
        },
        {
          title: '盈亏比例Max',
          key: 'max',
          align: 'center'
        },
        {
          title: '低概率',
          key: 'low_odds',
          align: 'center'
        },
        {
          title: '低倍数',
          key: 'low_multiple',
          align: 'center'
        },
        {
          title: '正常概率',
          key: 'normal_odds',
          align: 'center'
        },
        {
          title: '正常倍数',
          key: 'normal_multiple',
          align: 'center'
        },
        {
          title: '高概率',
          key: 'high_odds',
          align: 'center'
        },
        {
          title: '高倍数',
          key: 'high_multiple',
          align: 'center'
        },
        {
          title: '操作',
          key: 'gameNumber',
          align: 'center',
          width: 80,
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  style: {
                    marginRight: '10px'
                  },
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.showBairenGameSettingDataPage(params.row, false)
                    }
                  }
                },
                '修改'
              )
            ])
          }
        }
      ],
      // 汇率配置
      exchangeSettings: [
        {
          title: '货币',
          key: 'currency',
          align: 'center',
          width: 300
        },
        {
          title: '汇率(对CNY)',
          key: 'exchange',
          align: 'center'
        },
        {
          title: '操作',
          key: 'opt',
          align: 'center',
          width: 300,
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  style: {
                    marginRight: '10px'
                  },
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.exchangeEditType = 1
                      _this.exchangeEdit = true
                      _this.currentCurency = params.row
                    }
                  }
                },
                '修改'
              ),
              h(
                'Button',
                {
                  style: {
                    marginRight: '10px'
                  },
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.exchangeEditType = 3
                      _this.exchangeData = _this.exchangeData.filter(item => {
                        return item.currency != params.row.currency
                      })
                      _this.EditCurrency()
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      showBairenGameSettingData: false,
      currentBairenGameSettingData: {},
      filter: []
    }
  },
  methods: {
    async saveFreshmanProtect () {
    },
    async poolResetTimeRangeChange () {
      let _this = this
      this.$Modal.confirm({
        title: '提醒',
        render: (h) => {
          return (
            <div>
              <label>
                <b>确定要修改水池重置周期吗？</b>
              </label>
            </div>
          )
        },
        async onOk () {
          let params = {
            token: getToken(),
            interval: _this.resetParams.interval,
            now: false
          }
          let data = await axios.request({
            url: 'v2/govern/poolResetTimeRange',
            method: 'get',
            params
          })
          if (data && data.data && data.data.code == 200) {
            _this.$Message.info('设置成功')
          }
          _this.getPoolResetInfo()
        },
        onCancel () {}
      })
    },

    async poolResetNow () {
      let _this = this
      this.$Modal.confirm({
        title: '提醒',
        render: (h) => {
          return (
            <div>
              <label>
                <b>确定要立即重置水池吗？</b>
              </label>
            </div>
          )
        },
        async onOk () {
          let params = {
            token: getToken()
          }
          let data = await axios.request({
            url: 'v2/govern/poolResetNow',
            method: 'get',
            params
          })
          if (data && data.data && data.data.code == 200) {
            _this.$Message.info('设置成功')
          }
          _this.getPoolResetInfo()
        },
        onCancel () {}
      })
    },

    async getPoolResetInfo () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/govern/poolResetInfo',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        let result = JSON.parse(data.data.data)
        this.resetParams.interval = result.interval
        this.resetParams.nextRestTime = dayjs(result.resetTime * 1000).format(
          'YYYY-MM-DD HH:mm:ss'
        )
      }
    },

    async userControlTimeRangeChange () {
      let params = {
        token: getToken(),
        t: this.userControlTimeRange
      }
      let data = await axios.request({
        url: 'v2/govern/setUserCtl',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.$Message.info('设置成功')
      }
    },

    async getUserControlTimeRange () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/govern/getUserCtl',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.userControlTimeRange = data.data.data.t
        this.cycleRestTime = dayjs(data.data.data.e * 1000).format(
          'YYYY-MM-DD HH:mm:ss'
        )
      }
    },

    async getSessionAcc () {
      let params = {
        token: getToken(),
        acc: this.userAccount
      }
      let data = await axios.request({
        url: 'v2/game/getSession',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.$Message.info(data.data.msg)
        if (data.data.data.length) {
          this.accSessionRes = JSON.parse(data.data.data[0])
        }
      }
    },
    async delSessionAcc () {
      let func = async () => {
        let params = {
          token: getToken(),
          acc: this.userAccount
        }
        let data = await axios.request({
          url: 'v2/game/delSession',
          method: 'post',
          params
        })

        if (data && data.data && data.data.code == 200) {
          this.accSessionRes = null
          this.$Message.info(data.data.msg)
        }
      }
      this.$Modal.confirm({
        title: '提示',
        content: '请确认玩家不在线',
        onOk: func
      })
    },
    async doPoolRefresh () {
      let data = await axios.request({
        url: 'v2/stock/list',
        method: 'get',
        params: {
          token: getToken(),
          areaId: this.stockRoomId,
          gameId: this.paramgame,
          pageSize: 10000,
          page: 1,
          webId: this.siteParams.val ? this.siteParams.val : undefined,
          agentId: this.agentParams.val
            ? this.agentParams.val
            : this.agentParams.val === 0
              ? 0
              : undefined
        }
      })
      if (data && data.data && data.data.code == 200) {
        let scqxData = data.data.data.data
        scqxData.sort(this.sortBy('createTime'))
        // this.myChart.clear();
        // 绘制图表
        this.myChart.setOption({
          title: {
            text: '水池曲线',
            left: '1%'
          },
          tooltip: {
            trigger: 'axis'
          },
          grid: {
            left: '10%',
            right: '5%',
            bottom: '10%'
          },
          xAxis: {
            data: scqxData.map(function (item) {
              return dayjs(item.createTime * 1000).format('MM-DD HH:mm')
            })
          },
          yAxis: {},
          dataZoom: [
            {
              startValue: dayjs(scqxData[0].createTime * 1000).format(
                'MM-DD HH:mm'
              )
            },
            {
              type: 'inside'
            }
          ],
          series: {
            name: '实时水池',
            type: 'line',
            smooth: 100,
            data: scqxData.map(function (item) {
              return item.poolValue
            })
          }
        })

        if (this.isAutoRefresh && this.renderLineChartModal) {
          let _this = this
          setTimeout(function () {
            _this.doPoolRefresh()
          }, 5000)
        }
      }
    },

    async poolDataRefresh () {
      this.doPoolRefresh()
    },
    async renderLineChart () {
      if (this.agentParams.val !== 0 && !this.agentParams.val) {
        this.$Message.error('请先选择一个代理！')
        return
      }
      if (
        this.paramgame === 0 ||
        this.paramgame === null ||
        this.paramgame === undefined
      ) {
        this.$Message.error('请先选择一个游戏！')
        return
      }
      let dom = document.querySelector('#renderLineChartDom')
      if (dom) {
        if (this.myChart == null) {
          this.myChart = echarts.init(dom)
        } else {
          this.myChart.clear()
        }
      }

      this.renderLineChartModal = true
      this.doPoolRefresh()
    },
    setSiteSession (val) {
      this.agentParams.val = ''
      if (val > 0) {
        this.agentParams.option = []
        for (const key in this.siteParams.option) {
          if (this.siteParams.option[key].id == val) {
            this.agentParams.option.push(
              ...this.siteParams.option[key].agentList
            )
            this.agentParams.msg = '选择代理'
          }
        }
        this.agentParams.option.map((item) => {
          item.label = item.name
        })
      }
    },

    changeTab (tab) {
      if (tab == 'name4') {
        window.isctl_temp = 1
        this.$router.push({
          name: 'players'
        })
      }
    },

    async searchStock () {
      if (this.agentParams.val !== 0 && !this.agentParams.val) {
        this.$Message.error('请先选择一个代理！')
        return
      }
      if (
        this.paramgame === 0 ||
        this.paramgame === null ||
        this.paramgame === undefined
      ) {
        this.$Message.error('请先选择一个游戏！')
        return
      }
      let data = await axios.request({
        url: 'v2/stock/list',
        method: 'get',
        params: {
          token: getToken(),
          areaId: this.stockRoomId,
          gameId: this.paramgame,
          pageSize: 20,
          page: this.stockPage,
          webId: this.siteParams.val ? this.siteParams.val : undefined,
          agentId: this.agentParams.val
            ? this.agentParams.val
            : this.agentParams.val === 0
              ? 0
              : undefined
        }
      })
      if (data && data.data && data.data.code == 200) {
        this.stockData = data.data.data.data
        this.stockTotal = data.data.data.total
      } else {
      }
    },

    async searchReward () {
      let params = {
        ...this.rewardParams
      }
      let obj = {}
      for (const key in params) {
        if (params[key]) {
          obj[key] = params[key]
        }
      }
    },

    /**
     * 修改房间调
     */

    async confirmEditRoomInfo () {
      if (
        Number(this.currentEditRoom.data.max) <=
        Number(this.currentEditRoom.data.min)
      ) {
        this.$Message.error('最高水位不能低于或等于最低水位')
        return
      }
      let Data = {
        key: '',
        value: {
          pool: {
            1: {
              min: Number(this.currentEditRoom.data.min),
              max: Number(this.currentEditRoom.data.max),
              normal: Number(this.currentEditRoom.data.normal),
              minRate: Number(this.currentEditRoom.data.minRate),
              maxRate: Number(this.currentEditRoom.data.maxRate),
              normalRate: Number(this.currentEditRoom.data.normalRate),
              revenue: Number(this.currentEditRoom.data.revenue),
              control: Number(this.currentEditRoom.data.control),
              base: Number(this.currentEditRoom.data.base)
            }
          },
          name: this.currentEditRoom.data.name,
          nameZH: this.currentEditRoom.data.nameZH,
          gameId: this.currentEditRoom.data.gameId,
          symbol: this.currentEditRoom.data.symbol
        },
        token: getToken()
      }
      if (this.currentEditRoom.agentId >= 0) {
        // /agent/{agentId}/pool/{symbol}
        Data.key =
          '/agent/' +
          this.currentEditRoom.agentId +
          '/pool/' +
          this.currentEditRoom.data.symbol
      } else {
        // /config/pool/{symbol}
        Data.key = '/config' + '/pool/' + this.currentEditRoom.data.symbol
      }

      let data = await axios.request({
        url: 'v2/govern/edit',
        method: 'post',
        params: Data
      })
      if (data && data.data && data.data.code == 200) {
        this.$Message.info('保存成功')
        this.agentSearch()
      } else {
        this.$Message.error('修改失败')
      }
    },

    // 获取汇率配置数据
    async getCurrencyData () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/exchange',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        let txd = []
        let n = 0
        Object.keys(data.data.data.currency).forEach(k => {
          txd.push({currency: k, exchange: data.data.data.currency[k], id: n++})
        })
        this.exchangeData = txd
      } else {
        this.$Message.error('修改失败')
      }
    },

    // 修改汇率配置
    async EditCurrency () {
      let c = {}
      let n = 0
      this.exchangeData.forEach(item => {
        c[item.currency] = item.exchange
      })

      if (this.exchangeEditType <3) {
        c[this.currentCurency.currency] = this.currentCurency.exchange
      }

      let params = {
        token: getToken(),
        config: JSON.stringify({currency: c})
      }
      console.log(JSON.stringify(params))
      let data = await axios.request({
        url: 'v2/editExchange',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        let txd = []
        Object.keys(data.data.data.currency).forEach(k => {
          txd.push({currency: k, exchange: data.data.data.currency[k], id: n++})
        })
        this.exchangeData = txd
        this.exchangeEdit = false
      } else {
        this.$Message.error('修改失败')
      }
    },

    vieweditRoomInfo (row) {
      this.editRoomInfo = true
      this.currentEditRoom.data = JSON.parse(JSON.stringify(row))
      this.currentEditRoom.add = false
      this.currentEditRoom.tag = '修改房间配置'
      this.currentEditRoom.agentId = this.agentParams.val
        ? this.agentParams.val
        : this.agentParams.val === 0
          ? 0
          : undefined
    },

    addRoomInfo () {
      this.editRoomInfo = true
      this.currentEditRoom.add = true
      this.currentEditRoom.tag = '添加房间配置'
      let game = {}
      let _this = this
      this.gameOptions.map((item) => {
        if (item.number == _this.paramgame) {
          game = item
        }
      })
      this.currentEditRoom.data = {
        gameId: this.paramgame,
        name: game.name,
        nameZH: game.nameZH,
        symbol: game.symbol,
        min: 0,
        max: 0,
        normal: 0,
        normalRate: 0,
        minRate: 0,
        maxRate: 0,
        control: 1,
        revenue: 0,
        base: 0
      }
      this.currentEditRoom.agentId = this.agentParams.val
        ? this.agentParams.val
        : this.agentParams.val === 0
          ? 0
          : undefined
    },

    /**
     * 游戏列表获取
     */
    async initAgentData () {
      let newArray = this.games
      let a = []
      a.push(...newArray)
      a.map((item) => {
        if (item.nameZH.length > 0) {
          item.label = item.name + ' [' + item.nameZH + ']'
        } else {
          item.label = item.name
        }
        return item
      })
      this.gameOptions = a
      this.paramgame = a[0].number
      this.stockGameId = a[0].number
      this.agentSearch()
    },

    /**
     * 代理查询
     */
    async agentSearch () {
      let params = {
        token: getToken(),
        gameId: this.paramgame,
        agentId: this.agentParams.val
          ? this.agentParams.val
          : this.agentParams.val === 0
            ? 0
            : undefined,
        webId: this.siteParams.val ? this.siteParams.val : undefined
      }

      if (this.paramgame == 9999999) {
        params.agentId = null
      }

      if (this.paramgame == '0') {
        params = {
          token: getToken()
        }
      }

      let data = await axios.request({
        url: 'v2/govern/list',
        method: 'get',
        params
      })

      if (data && data.data && data.data.code == 200) {
        let newarr = []
        data.data.data.map((item) => {
          if (!item.value) {
            return
          }
          for (const key in item.value.pool) {
            if (item.value.pool.hasOwnProperty(key)) {
              const element = item.value.pool[key]
              element.name = item.name
              element.symbol = item.value.symbol
              element.nameZH = item.value.nameZH
              element.key = item.key
              element.gameId = item.value.gameId
              newarr.push(JSON.parse(JSON.stringify(element)))
            }
          }
        })
        this.agentData = [...newarr]
      }
    },
    getSite () {
      // 获取玩家控制的站点
      getLinkageList().then((res) => {
        this.siteParams.option = []
        this.siteParams.option.push(...Object.assign(res.data.data))
        this.siteParams.option.map((item) => {
          item.label = item.name
          item.value = item.id
        })
        this.siteParams.val = this.siteParams.option[0].value
        this.setSiteSession(this.siteParams.val)
      })
    },

    // 房间
    async selectRoom (e) {
      this.agentSearch()
    },
    // 提示
    async successTip () {
      this.$Message.info('修改成功')
    },
    async selectRewordByPageOne () {
      // this.rewardParams.page = 1;
      // this.getRewards();
    },

    async getStockMergeAgentList (combo = []) {
      let params = {
        token: getToken(),
        gameId: this.stockMergeForm.gameId,
        areaId: this.stockMergeForm.areaId
      }
      let data = await axios.request({
        url: 'v2/pool/agentList',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        let originA = []
        if (combo.length) {
          combo.map((com) => {
            data.data.data[com.id] = com
          })
        }
        for (const key in data.data.data) {
          const element = data.data.data[key]
          originA.push(element)
        }
        let groupKey = 'webId'
        let result = originA.reduce(function (r, a) {
          r[a[groupKey]] = r[a[groupKey]] || []
          r[a[groupKey]].push(a)
          return r
        }, Object.create(null))
        let tempAgents = []
        for (const key in result) {
          const element = result[key]
          tempAgents.push({
            id: key,
            label: '站点' + key,
            children: element.map((x) => {
              return {
                id: x.id,
                label: x.nickName
              }
            })
          })
        }
        this.stockMergeAgentOptions = tempAgents
      }
    },

    async getStockMergeList (page) {
      let params = {
        token: getToken(),
        page
      }
      let data = await axios.request({
        url: 'v2/pool/list',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.stockMergeData = data.data.data.data
        this.stockMergeDataTotal = data.data.data.total
      }
    },

    async getStockMergeBindAgentList (pool) {
      this.stockMergeForm.agentIds = []
      this.temppool = pool
      let poolId = pool.id
      let params = {
        token: getToken(),
        poolId
      }
      let data = await axios.request({
        url: 'v2/pool/bindAgentList',
        method: 'post',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.temppoolAgent = data.data.data
        let agentIds = []
        this.stockMergeForm.gameId = data.data.data[0].game_id
        this.stockMergeForm.areaId = data.data.data[0].area_id
        this.getStockMergeAgentList(
          data.data.data.map((x) => {
            agentIds.push(x.agent_id)
            return {
              id: x.agent_id,
              nickName: x.agent_name,
              webId: x.web_id
            }
          })
        )

        await Object.assign(this.stockMergeForm, pool)
        this.stockMergeForm.poolId = pool.id

        this.stockMergeAreas = this.stockMergeGames.find(
          (x) => x.number == this.stockMergeForm.gameId
        ).areas

        this.stockMergeForm.agentIds = agentIds
      }
    },

    sortBy (name) {
      return (o, p) => {
        let a, b
        if (typeof o === 'object' && typeof p === 'object' && o && p) {
          a = o[name]
          b = p[name]
          if (a === b) {
            return 0
          }
          if (typeof a === typeof b) {
            return a < b ? -1 : 1
          }
          return typeof a < typeof b ? -1 : 1
        } else {
        }
      }
    },
    showBairenGameSettingDataPage (row, isAdd) {
      console.log(row)
      if (!isAdd) {
        this.currentBairenGameSettingData.id = Number(row.id)
        this.currentBairenGameSettingData.name = row.name
        this.currentBairenGameSettingData.min = Number(row.min)
        this.currentBairenGameSettingData.max = Number(row.max)
        this.currentBairenGameSettingData.low_odds = Number(row.low_odds)
        this.currentBairenGameSettingData.low_multiple = Number(row.low_multiple)
        this.currentBairenGameSettingData.low_rate = Number(row.low_rate)
        this.currentBairenGameSettingData.normal_odds = Number(row.normal_odds)
        this.currentBairenGameSettingData.normal_multiple = Number(row.normal_multiple)
        this.currentBairenGameSettingData.normal_rate = Number(row.normal_rate)
        this.currentBairenGameSettingData.high_odds = Number(row.high_odds)
        this.currentBairenGameSettingData.high_multiple = Number(row.high_multiple)
        this.currentBairenGameSettingData.high_rate = Number(row.high_rate)
      } else {
        this.currentBairenGameSettingData.id = (new Date()).valueOf()
        this.currentBairenGameSettingData.name = ''
        this.currentBairenGameSettingData.min = 0
        this.currentBairenGameSettingData.max = 0
        this.currentBairenGameSettingData.low_odds = 0
        this.currentBairenGameSettingData.low_multiple = 0
        this.currentBairenGameSettingData.low_rate = 0
        this.currentBairenGameSettingData.normal_odds = 0
        this.currentBairenGameSettingData.normal_multiple = 0
        this.currentBairenGameSettingData.normal_rate = 0
        this.currentBairenGameSettingData.high_odds = 0
        this.currentBairenGameSettingData.high_multiple = 0
        this.currentBairenGameSettingData.high_rate = 0
      }
      this.currentBairenGameSettingData.isAdd = isAdd
      this.showBairenGameSettingData = true
    },
    async saveBairenGameSettingData (isAdd) {
      if (Number(this.currentBairenGameSettingData.max) <= Number(this.currentBairenGameSettingData.min) && this.currentBairenGameSettingData.name != 'single') {
        this.$Message.error('高盈亏比不能小于低盈亏比')
        return
      }
      let result = this.gameSettingData.concat()
      if (isAdd) {
        // 合并数据
        result.push(this.currentBairenGameSettingData)
      }
      let tmp = []
      result.forEach(element => {
        if (element.id == this.currentBairenGameSettingData.id && !isAdd) {
          element = this.currentBairenGameSettingData
        }
        tmp.push({
          id: Number(element.id),
          name: element.name,
          min: Number(element.min),
          max: Number(element.max),
          pool_odds: [
            {
              odds: Number(element.low_odds),
              multiple: Number(element.low_multiple),
              rate: Number(element.low_rate)
            },
            {
              odds: Number(element.normal_odds),
              multiple: Number(element.normal_multiple),
              rate: Number(element.normal_rate)
            },
            {
              odds: Number(element.high_odds),
              multiple: Number(element.high_multiple),
              rate: Number(element.high_rate)
            }
          ]
        })
      })
      let Data = {
        token: getToken(),
        gameId: this.paramBairen,
        config: JSON.stringify(tmp)
      }
      // 提交保存
      let data = await axios.request({
        url: 'v2/game/saveGameSettingData',
        method: 'post',
        params: Data
      })
      console.log(data)
      if (data && data.data.code == 200) {
        this.$Message.info('保存成功')
        this.getGameSettingList()
      } else {
        this.$Message.error('保存失败')
      }
    },
    async syncAllPool (params) {
      let p = {
        pool: {
          1: {
            min: Number(params.min),
            max: Number(params.max),
            normal: Number(params.normal),
            minRate: Number(params.minRate),
            maxRate: Number(params.maxRate),
            normalRate: Number(params.normalRate),
            revenue: Number(params.revenue),
            control: Number(params.control),
            base: Number(params.base)
          }
        },
        name: params.name,
        nameZH: params.nameZH,
        gameId: params.gameId,
        symbol: params.symbol
      }
      let Data = {
        token: getToken(),
        config: JSON.stringify(p)
      }
      // 提交保存
      let data = await axios.request({
        url: 'v2/game/syncAllPool',
        method: 'post',
        params: Data
      })
      if (data && data.data.code == 200) {
        this.$Message.info('同步成功')
        this.getGameSettingList()
      } else {
        this.$Message.error('同步失败')
      }
    },
    async getGameSettingList () {
      let data = await axios.request({
        url: 'v2/game/getGameSettingData',
        method: 'post',
        params: {
          token: getToken()
        }
      }).catch(() => {})
      if (data && data.data && data.data.code == 200) {
        let arr = []
        let tmp = JSON.parse(data.data.data)
        tmp.award_config.forEach((item) => {
          arr.push({
            id: item.id,
            name: item.name,
            min: item.min,
            max: item.max,
            low_odds: item.pool_odds[0].odds,
            low_multiple: item.pool_odds[0].multiple,
            low_rate: item.pool_odds[0].rate,
            normal_odds: item.pool_odds[1].odds,
            normal_multiple: item.pool_odds[1].multiple,
            normal_rate: item.pool_odds[1].rate,
            high_odds: item.pool_odds[2].odds,
            high_multiple: item.pool_odds[2].multiple,
            high_rate: item.pool_odds[2].rate
          })
        })
        this.gameSettingData = arr
      }
    }
  }
}
</script>

<style scoped lang="less">
.inlineForm {
  display: flex;
  align-items: center;
  padding: 10px 0;
}
</style>
