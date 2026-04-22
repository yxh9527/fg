<template>
  <div>
    <Tabs @on-click="changeTab" value="bslx" :animated="false">
      <TabPane label="比赛类型" name="bslx">
        <Card style="margin: 10px 0">
          <Button
            @click="showAddMatchTypeItem"
            style="margin-bottom: 20px"
            type="primary"
            >添加</Button
          >
          <Table :columns="matchTypeColumns" :data="matchTypeData"> </Table>
        </Card>

        <Modal
          v-model="editMatchType"
          title="比赛类型修改"
          @on-ok="updateMatchTypeItem"
        >
          <div v-if="curMatchTypeItem">
            <div>
              赛事类型名称：<Input
                type="text"
                v-model="curMatchTypeItem.name"
              />
            </div>
          </div>
        </Modal>
      </TabPane>
      <TabPane label="游戏类型" name="yxlx">
        <Card style="margin: 10px 0">
          <Button
            @click="showAddGameTypeItem"
            style="margin-bottom: 20px"
            type="primary"
            >添加</Button
          >
          <Table :columns="gameTypeColumns" :data="gameTypeData"> </Table>
        </Card>

        <Modal
          v-model="editGameType"
          title="游戏类型修改"
          @on-ok="updateGameTypeItem()"
        >
          <div v-if="curGameTypeItem">
            <div>
              游戏类型名称：<Input type="text" v-model="curGameTypeItem.name" />
            </div>
          </div>
        </Modal>
      </TabPane>

      <TabPane label="比赛奖励" name="bsjl">
        <div>
          <Card style="margin-top: 10px; margin-bottom: 200px">
            <div style="margin-bottom: 20px">
              <Button @click="showMatchAwardPage" type="primary">添加 </Button>
            </div>
            <Table :columns="matchAwardColumns" :data="matchAwardData"> </Table>
          </Card>
          <Modal
            v-model="editMatchAward"
            title="比赛奖励修改"
            @on-ok="updateMatchAwardItem"
          >
            <div v-if="curMatchAwardItem" style="width: 500px">
              <div>
                名称：<Input type="text" v-model="curMatchAwardItem.name" />
              </div>
              <div>
                比例(%)：<textarea
                  v-model="curMatchAwardItem.detail"
                  style="
                    width: 500px;
                    border-color: lightgray;
                    border-radius: 5px;
                  "
                  placeholder="请输入内容..."
                ></textarea>
              </div>
            </div>
          </Modal>
        </div>
      </TabPane>

      <TabPane label="盲注列表" name="mzlb">
        <Card style="margin: 10px 0">
          <Button
            @click="showBlinditem"
            style="margin-bottom: 20px"
            type="primary"
            >添加</Button
          >
          <Table :columns="blindsColumns" :data="blindData"> </Table>
        </Card>
        <Modal
          v-model="editBlind"
          title="盲注列表"
          :width="1000"
          :mask-closable="false"
        >
          <div v-if="curBlindItem">
            <div>名称：<Input type="text" v-model="curBlindItem.name" /></div>
            <Table
              class="mzlbshow"
              style="margin-top: 10px"
              :columns="blindsDataColumns"
              :data="blindsDataDetail"
            >
            </Table>
          </div>
          <template slot="footer">
            <Button @click="addBlindsDetailItem" style="margin-bottom: 20px"
              >添加</Button
            >
            <Button @click="updateBlindItem" style="margin-bottom: 20px"
              >保存</Button
            >
          </template>
        </Modal>
      </TabPane>

      <TabPane label="比赛列表" name="bslb">
        <Card>
          <div class="inlineForm">
            <div>代理选择：</div>
            <div>
              <Select
                :placeholder="agentParams.msg"
                filterable
                filter-by-label
                clearable
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
            <div style="margin-left: 20px">比赛状态：</div>
            <div>
              <Select
                :placeholder="matchState.msg"
                filterable
                filter-by-label
                clearable
                v-model="matchState.val"
                style="width: 100px"
              >
                <Option
                  v-for="item in matchState.option"
                  :value="item.id"
                  :key="item.id"
                  :label="item.name"
                >
                  {{ item.name }}</Option
                >
              </Select>
            </div>
            <div style="margin-left: 20px">
              <label style="margin-right: 5px">时间选择：</label>
              <DatePicker
                type="datetime"
                style="width: 175px"
                v-model="matchSearchTimes.startTime"
                placeholder="选择开始时间"
              ></DatePicker
              >—
              <DatePicker
                type="datetime"
                style="width: 175px"
                v-model="matchSearchTimes.endTime"
                placeholder="选择结束时间"
              ></DatePicker>
            </div>

            <div style="padding-left: 20px">
              <Button @click="getMatchList()">查询 </Button>
            </div>
            <div style="padding-left: 20px">
              <Button @click="showAddMatchPage">添加 </Button>
            </div>
          </div>
        </Card>
        <Card style="margin: 10px 0">
          <Tables
            ref="tables"
            v-model="matchData"
            width="100%"
            :columns="matchColumns"
          />
          <br />
          <div class="pageStyle">
            <Page
              :total="matchListPageData.current"
              :current="matchListPageData.page"
              :page-size="matchListPageData.pageSize"
              :page-size-opts="matchListPageData.pageOpts"
              show-elevator
              show-sizer
              show-total
              title="输入页数后，按Enter键跳转页面"
              @on-change="matchListChangePage"
              @on-page-size-change="matchListChangePageSize"
            />
          </div>
        </Card>
        <Modal
          v-model="editMatch"
          :title="this.matchTag"
          @on-ok="updateMatchItem"
        >
          <div v-if="curMatchItem">
            <div>
              <span>赛事名称:</span
              ><span
                ><Input
                  type="text"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.name"
              /></span>
            </div>
            <div>
              <span>赛事类型:</span
              ><span>
                <CheckboxGroup
                  style="margin-left: 60px; width: 100%"
                  @on-change="SelectQuestionType"
                  v-model="selectTypes"
                >
                  <Checkbox
                    v-for="item in matchTypeParams.option"
                    :label="item.id"
                    :key="item.id"
                    ref="checkboxGroup"
                  >
                    <span>{{ item.name }}</span>
                  </Checkbox>
                </CheckboxGroup>
              </span>
            </div>
            <div>
              <span>游戏类型:</span
              ><span>
                <Select
                  :placeholder="gameTypeParams.msg"
                  filterable
                  filter-by-label
                  clearable
                  v-model="gameTypeParams.val"
                  style="width: 200px; margin-left: 10px"
                >
                  <Option
                    v-for="item in gameTypeParams.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </span>
            </div>
            <div>
              <span>中场休息:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.interval"
              /></span>
            </div>
            <div>
              <span>赛事奖励:</span
              ><span>
                <Select
                  :placeholder="matchAwardParmas.msg"
                  filterable
                  filter-by-label
                  clearable
                  v-model="matchAwardParmas.val"
                  style="width: 200px; margin-left: 10px"
                >
                  <Option
                    v-for="item in matchAwardParmas.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </span>
            </div>
            <div>
              <span>初始分数:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.num"
                />
              </span>
            </div>
            <div>
              <span>买入上限:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.max_buy"
                />
              </span>
            </div>
            <div>
              <span>延迟报名:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.delay"
                />
              </span>
            </div>
            <div>
              <span>游戏种类:</span
              ><span>
                <Select
                  filterable
                  v-model="games.value"
                  style="width: 200px; margin-left: 10px"
                >
                  <Option
                    v-for="item in games.option"
                    :value="item.id"
                    :key="item.id"
                    >{{ item.name }}</Option
                  >
                </Select>
              </span>
            </div>
            <div>
              <span>报名费用:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.fee"
                />
              </span>
            </div>
            <div>
              <span>开房房费:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.room_fee"
                />
              </span>
            </div>
            <div>
              <span>盲注列表:</span
              ><span>
                <Select
                  :placeholder="blindsParams.msg"
                  filterable
                  filter-by-label
                  clearable
                  v-model="blindsParams.val"
                  style="width: 200px; margin-left: 10px"
                >
                  <Option
                    v-for="item in blindsParams.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </span>
            </div>
            <div>
              <span>奖励积分:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.score"
                />
              </span>
            </div>
            <div>
              <span>赛事说明:</span
              ><span
                ><Input
                  type="text"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.desc"
                />
              </span>
            </div>
            <div>
              <span>赛事资讯:</span
              ><span
                ><Input
                  type="text"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.news"
                />
              </span>
            </div>
            <div>
              <span>赛事标签:</span
              ><span
                ><Input
                  type="text"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.tag"
              /></span>
            </div>
            <div v-if="this.matchTag == '修改比赛'">
              <span>赛事状态:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.state"
              /></span>
            </div>
            <div>
              <span>基础人数:</span
              ><span
                ><Input
                  type="number"
                  style="width: 200px; margin-left: 10px"
                  v-model="curMatchItem.base"
              /></span>
            </div>
            <div>
              <span>结算类型:</span
              ><span>
                <Select
                  :placeholder="matchSettleTypeParmas.msg"
                  filterable
                  filter-by-label
                  clearable
                  v-model="matchSettleTypeParmas.val"
                  style="width: 200px; margin-left: 10px"
                >
                  <Option
                    v-for="item in matchSettleTypeParmas.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </span>
            </div>
            <div>
              <span>开始时间:</span
              ><span>
                <DatePicker
                  style="width: 200px; margin-left: 10px"
                  type="datetime"
                  :value="this.matchStartTime"
                  format="yyyy-MM-dd HH:mm:ss"
                  @on-change="TimePickerOnChange"
                ></DatePicker
              ></span>
            </div>
          </div>
        </Modal>
      </TabPane>
      <TabPane label="奖励券" name="jlq">
        <div>
          <Card>
            <div class="inlineForm">
              <div>比赛选择：</div>
              <div>
                <Select
                  :placeholder="matchParams.msg"
                  filterable
                  clearable
                  remote
                  @on-query-change="matchListSearchMethod"
                  :loading="matchListLoading"
                  v-model="matchParams.val"
                  style="width: 400px"
                >
                  <Option
                    v-for="item in matchParams.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </div>
              <div style="padding-left: 20px">
                <Button @click="getMatchTikectList">查询 </Button>
              </div>
              <div style="padding-left: 20px">
                <Button @click="showCreateTicketPage">添加 </Button>
              </div>
            </div>
          </Card>
          <Card style="margin-top: 10px; margin-bottom: 200px">
            <Table :columns="ticketColumns" :data="ticketData"> </Table>
            <br />
            <div class="pageStyle">
              <Page
                :total="ticketListPageData.current"
                :current="ticketListPageData.page"
                :page-size="ticketListPageData.pageSize"
                :page-size-opts="ticketListPageData.pageOpts"
                show-elevator
                show-sizer
                show-total
                title="输入页数后，按Enter键跳转页面"
                @on-change="ticketListChangePage"
                @on-page-size-change="ticketListChangePageSize"
              />
            </div>
          </Card>
          <Modal
            v-model="createTticket"
            title="添加奖励券"
            @on-ok="createTicket"
          >
            <div v-if="createTicketItem">
              <div>
                名称：<Input type="text" v-model="createTicketItem.name" />
              </div>
              <div>
                比赛:
                <Select
                  :placeholder="addTicketMatchParams.msg"
                  filterable
                  clearable
                  v-model="addTicketMatchParams.val"
                  @on-query-change="addTicketMatchListSearchMethod"
                  :loading="matchListLoading"
                >
                  <Option
                    v-for="item in addTicketMatchParams.option"
                    :value="item.id"
                    :key="item.id"
                    :label="item.name"
                  >
                    {{ item.name }}</Option
                  >
                </Select>
              </div>
              <div>
                数量:<Input type="number" v-model="createTicketItem.count" />
              </div>
              <div>
                价值:<Input type="number" v-model="createTicketItem.price" />
              </div>
            </div>
          </Modal>

          <Modal v-model="showTicketDetail" title="奖励券详情" width="1600">
            <Table :columns="ticketDetailColumns" :data="ticketDetaildata">
            </Table>
            <br />
            <div class="pageStyle">
              <Page
                :total="ticketDetailListPageData.current"
                :current="ticketDetailListPageData.page"
                :page-size="ticketDetailListPageData.pageSize"
                :page-size-opts="ticketDetailListPageData.pageOpts"
                show-elevator
                show-sizer
                show-total
                title="输入页数后，按Enter键跳转页面"
                @on-change="ticketDetailListChangePage"
                @on-page-size-change="ticketDetailListChangePage"
              />
            </div>
          </Modal>
        </div>
      </TabPane>
    </Tabs>
  </div>
</template>

<script>
// import the styles
import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import {
  getLinkageList
} from '@/api/data'
import Tables from '_c/tables'
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import { setting } from '@/config'
import * as dayjs from 'dayjs'

export default {
  provide () {
    return {
      getGameSettingList: this.getGameSettingList
    }
  },
  components: {
    Tables
  },
  data () {
    let _this = this
    return {
      matchTag: '',
      siteParams: {
        val: '',
        option: []
      }, // 控制玩家的站点
      agentParams: {
        val: '',
        option: [],
        msg: '请先选择站点'
      },
      matchListLoading: false,
      matchState: {
        val: '',
        option: [
          { id: 0, name: '未发布' },
          { id: 1, name: '已发布' },
          { id: 2, name: '比赛中' },
          { id: 3, name: '已结束' }
        ],
        msg: '比赛状态'
      },
      matchParams: {
        val: '',
        option: [],
        msg: '输入比赛名称'
      },
      // 赛事开始时间范围
      matchSearchTimes: {
        startTime: '',
        endTime: ''
      },
      addTicketMatchParams: {
        val: '',
        option: [],
        msg: '输入比赛名称'
      },

      matchTypeParams: {
        val: '',
        option: [],
        msg: '输入比赛名称'
      },
      gameTypeParams: {
        val: [],
        option: [],
        msg: '输入比赛名称'
      },
      isAllCheck: false,
      selectTypes: [],
      matchAwardParmas: {
        val: '',
        option: [],
        msg: '请先选择赛事奖励'
      },
      matchSettleTypeParmas: {
        val: '',
        option: [
          {
            id: 0, // 普通模式
            name: '普通模式'
          },
          {
            id: 1, // 猎人模式
            name: '猎人模式'
          }
        ],
        msg: '请先选择结算类型'
      },
      games: {
        value: '',
        option: []
      },
      blindsParams: {
        val: '',
        option: [],
        msg: '请先选择盲注表'
      },
      matchStartTime: '',
      agentId: 0,
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      matchListPageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      ticketListPageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      ticketDetailListPageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      editMatchType: false,
      editGameType: false,
      editMatchAward: false,
      editBlind: false,
      editTicket: false,
      editMatch: false,
      createTticket: false,
      showTicketDetail: false,
      curMatchTypeItem: null,
      curGameTypeItem: null,
      curMatchAwardItem: null,
      curBlindItem: null,
      curTicketItem: null,
      curMatchItem: null,
      createTicketItem: null,
      paramgame: null,
      matchTypeColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '名称',
          key: 'name',
          align: 'center'
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 140,
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
                      _this.curMatchTypeItem = params.row
                      _this.editMatchType = true
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
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/del/matchTypeItem',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.matchTypeList()
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      matchAwardColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '名称',
          key: 'name',
          align: 'center'
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '配置详情',
          key: 'detail',
          align: 'center',
          render (h, params) {
            return <span>{_this.genAwardParam(params.row.detail)}</span>
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 140,
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
                      _this.curMatchAwardItem = params.row
                      _this.editMatchAward = true
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
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/matchAward/del',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.getMatchAwardList()
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      blindsColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '名称',
          key: 'name',
          align: 'center'
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 200,
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
                      _this.curBlindItem = params.row
                      _this.curBlindItem.opt = 2
                      _this.editBlind = true
                      let tmp = JSON.parse(params.row.data)
                      _this.blindsDataDetail = tmp.items
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
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/matchBlinds/del',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.getBlindData()
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      blindsDataColumns: [
        {
          title: '小盲注',
          key: 'min',
          align: 'center',
          render (h, params) {
            return h('Input', {
              props: {
                type: 'number',
                value: params.row.min
              },
              on: {
                input: (val) => {
                  _this.blindsDataDetail[params.index].min = val
                }
              }
            })
          }
        },
        {
          title: '大盲注',
          key: 'max',
          align: 'center',
          render (h, params) {
            return h('Input', {
              props: {
                type: 'number',
                value: params.row.max
              },
              on: {
                input: (val) => {
                  _this.blindsDataDetail[params.index].max = val
                }
              }
            })
          }
        },
        {
          title: '延迟报名',
          key: 'delay',
          align: 'center',
          render (h, params) {
            return h('Input', {
              props: {
                type: 'number',
                value: params.row.delay
              },
              on: {
                input: (val) => {
                  _this.blindsDataDetail[params.index].delay = parseInt(val)
                }
              }
            })
          }
        },
        {
          title: '前注',
          key: 'bet',
          align: 'center',
          render (h, params) {
            return h('Input', {
              props: {
                type: 'number',
                value: params.row.bet
              },
              on: {
                input: (val) => {
                  _this.blindsDataDetail[params.index].bet = val
                }
              }
            })
          }
        },
        {
          title: '涨盲时间',
          key: 'interval',
          align: 'center',
          render (h, params) {
            return h('Input', {
              props: {
                type: 'number',
                value: params.row.interval
              },
              on: {
                input: (val) => {
                  _this.blindsDataDetail[params.index].interval = parseInt(val)
                }
              }
            })
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 80,
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  props: {
                    size: 'small'
                  },
                  on: {
                    async click () {
                      for (let i = 0; i < _this.blindsDataDetail.length; i++) {
                        if (
                          _this.blindsDataDetail[i].level == params.row.level
                        ) {
                          _this.blindsDataDetail.splice(i, 1)
                        }
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      ticketColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center',
          width: 180
        },
        {
          title: '比赛',
          key: 'match_name',
          align: 'center'
        },
        {
          title: '券名',
          key: 'name',
          align: 'center'
        },
        {
          title: '数量',
          key: 'create_time',
          align: 'center',
          width: 180
        },
        {
          title: '剩余数量',
          key: 'create_time',
          align: 'center',
          width: 180
        },
        {
          title: '价值',
          key: 'price',
          align: 'center',
          width: 180
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          width: 180,
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 180,
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.showTicketDetail = true
                      let iparams = {
                        token: getToken(),
                        id: params.row.id,
                        page: _this.ticketDetailListPageData.page,
                        pageSize: _this.ticketDetailListPageData.pageSize
                      }
                      let data = await axios.request({
                        url: 'v2/ticket/detail/list',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.ticketDetaildata = data.data.data.list
                        _this.ticketDetailListPageData.current =
                          data.data.data.total
                      }
                    }
                  }
                },
                '查看'
              ),
              h(
                'Button',
                {
                  style: {
                    marginLeft: '5px'
                  },
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/ticket/del',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.getMatchTikectList()
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      ticketDetailColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '比赛',
          key: 'match_name',
          align: 'center',
          width: 300
        },
        {
          title: '券名',
          key: 'name',
          align: 'center',
          width: 300
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          width: 180,
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '所有人',
          key: 'owner',
          align: 'center'
        },
        {
          title: '是否使用',
          key: 'is_use',
          align: 'center'
        },
        {
          title: '使用时间',
          key: 'use_time',
          align: 'center'
        }
      ],

      gameTypeColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center'
        },
        {
          title: '名称',
          key: 'name',
          align: 'center'
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 140,
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
                      _this.curGameTypeItem = params.row
                      _this.editGameType = true
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
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/del/gameTypeItem',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.gameTypeList()
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],
      matchColumns: [
        {
          title: 'ID',
          key: 'id',
          align: 'center',
          width: 60
        },
        {
          title: '名称',
          key: 'name',
          align: 'center',
          width: 200,
          render (h, params) {
            let tmp = params.row.name
            if (params.row.name.length > 12) {
              tmp = params.row.name.slice(0, 12) + '...'
            }
            return h('div', [
              h(
                'Tooltip',
                {
                  props: {
                    placement: 'top',
                    transfer: true
                  }
                },
                [
                  tmp,
                  h(
                    'span',
                    {
                      slot: 'content',
                      style: {
                        whiteSpace: 'normal'
                      }
                    },
                    params.row.name
                  )
                ]
              )
            ])
          }
        },
        {
          title: '开始时间',
          key: 'start_time',
          align: 'center',
          width: 200,
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.start_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '比赛状态',
          key: 'state',
          align: 'center',
          width: 140,
          render (h, params) {
            let str = ''
            switch (params.row.state) {
              case 0:
                str = '代发布'
                break
              case 1:
                str = '发布'
                break
              case 2:
                str = '比赛中'
                break
              case 3:
                str = '结束'
                break
              default:
                break
            }
            return <span>{str}</span>
          }
        },
        {
          title: '代理',
          key: 'agent_id',
          align: 'center',
          width: 100,
          render (h, params) {
            let agentName = ''
            let tmp = ''
            _this.agentParams.option.map((item) => {
              if (item.id == params.row.agent_id) {
                agentName = item.name
              }
            })
            if (agentName.length > 5) {
              tmp = agentName.slice(0, 5) + '...'
            } else {
              tmp = agentName
            }
            return h('div', [
              h(
                'Tooltip',
                {
                  props: {
                    placement: 'top',
                    transfer: true
                  }
                },
                [
                  tmp,
                  h(
                    'span',
                    {
                      slot: 'content',
                      style: {
                        whiteSpace: 'normal'
                      }
                    },
                    agentName
                  )
                ]
              )
            ])
          }
        },
        {
          title: '游戏类型',
          key: 'game_id',
          align: 'center',
          width: 140,
          render (h, params) {
            let games = JSON.parse(sessionStorage.getItem('games'))
            let str = ''
            games.forEach((m) => {
              if (m.number == params.row.game_id) {
                str = m.name
              }
            })
            return <span>{str}</span>
          }
        },
        {
          title: '盲注表',
          key: 'blinds_id',
          align: 'center',
          width: 120,
          render (h, params) {
            let blindList = JSON.parse(sessionStorage.getItem('blindList'))
            let str = ''
            let tmp = ''
            blindList.forEach((m) => {
              if (m.id == params.row.blinds_id) {
                str = m.name
              }
            })
            if (str.length > 5) {
              tmp = str.slice(0, 5) + '...'
            } else {
              tmp = str
            }
            return h('div', [
              h(
                'Tooltip',
                {
                  props: {
                    placement: 'top',
                    transfer: true
                  }
                },
                [
                  tmp,
                  h(
                    'span',
                    {
                      slot: 'content',
                      style: {
                        whiteSpace: 'normal'
                      }
                    },
                    str
                  )
                ]
              )
            ])
          }
        },
        {
          title: '赛事类型',
          key: 'type',
          align: 'center',
          width: 120,
          render (h, params) {
            let arr = params.row.type.split(',')
            let matchType = JSON.parse(sessionStorage.getItem('matchType'))
            let str = []
            let tmp = ''
            arr.forEach((item) => {
              matchType.forEach((m) => {
                if (m.id == parseInt(item)) {
                  str.push(m.name)
                }
              })
            })
            if (str.join(',').length > 5) {
              tmp = str.join(',').slice(0, 5) + '...'
            } else {
              tmp = str.join(',')
            }
            return h('div', [
              h(
                'Tooltip',
                {
                  props: {
                    placement: 'top',
                    transfer: true
                  }
                },
                [
                  tmp,
                  h(
                    'span',
                    {
                      slot: 'content',
                      style: {
                        whiteSpace: 'normal'
                      }
                    },
                    str.join(',')
                  )
                ]
              )
            ])
          }
        },
        {
          title: '奖励',
          key: 'award',
          align: 'center',
          width: 120,
          render (h, params) {
            let matchAward = JSON.parse(sessionStorage.getItem('matchAward'))
            let str = ''
            let tmp = ''
            matchAward.forEach((m) => {
              if (m.id == params.row.award) {
                str = m.name
              }
            })
            if (str.length > 5) {
              tmp = str.slice(0, 5) + '...'
            } else {
              tmp = str
            }
            return h('div', [
              h(
                'Tooltip',
                {
                  props: {
                    placement: 'top',
                    transfer: true
                  }
                },
                [
                  tmp,
                  h(
                    'span',
                    {
                      slot: 'content',
                      style: {
                        whiteSpace: 'normal'
                      }
                    },
                    str
                  )
                ]
              )
            ])
          }
        },
        {
          title: '级别',
          key: 'level',
          align: 'center',
          width: 100
        },
        {
          title: '休息(s)',
          key: 'interval',
          align: 'center',
          width: 100
        },
        {
          title: '报名人数',
          key: 'sign',
          align: 'center',
          width: 120
        },
        {
          title: '彩池',
          key: 'pot',
          align: 'center',
          width: 100
        },
        {
          title: '初始积分',
          key: 'num',
          align: 'center',
          width: 140
        },
        {
          title: '买入上限',
          key: 'max_buy',
          align: 'center',
          width: 140
        },
        {
          title: '报名费',
          key: 'fee',
          align: 'center',
          width: 140
        },
        {
          title: '房费',
          key: 'room_fee',
          align: 'center',
          width: 100
        },
        {
          title: '积分',
          key: 'score',
          align: 'center',
          width: 100
        },
        {
          title: '结算方式',
          key: 'settlement_type',
          align: 'center',
          width: 140
        },
        {
          title: '延迟报名',
          key: 'delay',
          align: 'center',
          width: 140
        },
        {
          title: '创建时间',
          key: 'create_time',
          align: 'center',
          width: 200,
          render (h, params) {
            return (
              <span>
                {dayjs(params.row.create_time * 1000).format(
                  'YYYY-MM-DD HH:mm:ss'
                )}
              </span>
            )
          }
        },
        {
          title: '操作',
          align: 'center',
          width: 260,
          fixed: 'right',
          render (h, params) {
            return h('div', {}, [
              h(
                'Button',
                {
                  style: {
                    marginRight: '5px'
                  },
                  props: {
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/match/publish',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.getMatchList(-1)
                      }
                    }
                  }
                },
                params.row.state == 0 ? '发布' : '取消发布'
              ),
              h(
                'Button',
                {
                  style: {
                    marginRight: '5px'
                  },
                  props: {
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      await axios.request({
                        url: 'v2/match/close',
                        method: 'get',
                        params: iparams
                      })
                    }
                  }
                },
                '关闭'
              ),
              h(
                'Button',
                {
                  style: {
                    marginRight: '5px'
                  },
                  props: {
                    size: 'small'
                  },
                  on: {
                    async click () {
                      _this.curMatchItem = params.row
                      _this.showEditMatchPage()
                    }
                  }
                },
                '修改'
              ),
              h(
                'Button',
                {
                  props: {
                    size: 'small'
                  },
                  on: {
                    async click () {
                      let iparams = {
                        token: getToken(),
                        id: params.row.id
                      }
                      let data = await axios.request({
                        url: 'v2/match/del',
                        method: 'get',
                        params: iparams
                      })
                      if (data && data.data && data.data.code == 200) {
                        _this.getMatchList(-1)
                      }
                    }
                  }
                },
                '删除'
              )
            ])
          }
        }
      ],

      gameOptions: [],
      matchTypeData: [],
      gameTypeData: [],
      matchAwardData: [],
      blindData: [],
      blindsDataDetail: [],
      ticketData: [],
      ticketDetaildata: [],
      matchData: []
    }
  },
  methods: {
    TimePickerOnChange (event) {
      this.matchStartTime = event
    },
    SelectQuestionType (data) {
      this.selectTypes = data
    },
    async matchListSearchMethod (query) {
      if (query !== '') {
        this.matchListLoading = true
        this.matchParams.option = await this.getMatchListByName(query)
      } else {
        this.matchParams.option = []
        this.matchListLoading = false
      }
    },
    async addTicketMatchListSearchMethod (query) {
      if (query !== '') {
        this.matchListLoading = true
        this.addTicketMatchParams.option = await this.getMatchListByName(query)
        this.matchListLoading = false
      } else {
        this.addTicketMatchParams.option = []
        this.matchListLoading = false
      }
    },

    ArrayContain (array, id) {
      var result = false
      for (var i = 0; i < array.length; i++) {
        if (array[i] == id) {
          result = true
          break
        }
      }
      return result
    },
    clearTimeDatess () {
      // 清除快捷日期
      this.matchStartTime = ''
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

    async getMatchAwardList () {
      let params = {
        token: getToken(),
        page: this.pageData.page,
        pageSize: this.pageData.pageSize,
        agentId: this.agentParams.val
          ? this.agentParams.val
          : this.agentParams.val === 0
            ? 0
            : undefined
      }
      let data = await axios.request({
        url: 'v2/matchAward/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.matchAwardData = data.data.data
        sessionStorage.setItem('matchAward', JSON.stringify(data.data.data))
      }
    },

    async createTicket () {
      let params = {
        token: getToken(),
        agentId: this.agentParams.val
          ? this.agentParams.val
          : this.agentParams.val === 0
            ? 0
            : undefined,
        matchId: this.addTicketMatchParams.val
          ? this.addTicketMatchParams.val
          : this.addTicketMatchParams.val === 0
            ? 0
            : undefined,
        name: this.createTicketItem.name,
        price: this.createTicketItem.price,
        count: this.createTicketItem.count
      }
      await axios.request({
        url: 'v2/ticket/add',
        method: 'get',
        params
      })
      this.getMatchTikectList()
    },

    async getMatchTikectList () {
      let params = {
        token: getToken(),
        page: this.ticketListPageData.page,
        pageSize: this.ticketListPageData.pageSize,
        matchId: this.matchParams.val
      }
      let data = await axios.request({
        url: 'v2/ticket/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.ticketListPageData.current = data.data.data.total
        this.ticketData = data.data.data.list
      }
    },

    async updateMatchAwardItem () {
      let params = {
        token: getToken(),
        id: this.curMatchAwardItem.id,
        name: this.curMatchAwardItem.name,
        detail: this.curMatchAwardItem.detail
      }
      if (this.curMatchAwardItem.isAdd == undefined) {
        await axios.request({
          url: 'v2/matchAward/update',
          method: 'get',
          params
        })
      } else {
        await axios.request({
          url: 'v2/matchAward/add',
          method: 'get',
          params
        })
      }
      this.getMatchAwardList()
    },

    async updateTicket () {
      let params = {
        token: getToken(),
        id: this.curMatchAwardItem.id,
        name: this.curMatchAwardItem.name,
        detail: this.curMatchAwardItem.detail
      }
      if (this.curMatchAwardItem.isAdd == undefined) {
        await axios.request({
          url: 'v2/matchAward/update',
          method: 'get',
          params
        })
      } else {
        await axios.request({
          url: 'v2/matchAward/add',
          method: 'get',
          params
        })
      }
      this.getMatchAwardList()
    },

    changePage (index) {
      this.pageData.page = index
      this.getMatchAwardList()
    },
    changePageSize (index) {
      this.pageData.pageSize = index
      this.getMatchAwardList()
    },
    matchListChangePage (index) {
      this.matchListPageData.page = index
      this.getMatchList(-1)
    },
    matchListChangePageSize (index) {
      this.matchListPageData.pageSize = index
      this.getMatchList(-1)
    },
    ticketListChangePage (index) {
      this.ticketListPageData.page = index
      this.getMatchTikectList()
    },
    ticketListChangePageSize (index) {
      this.ticketListPageData.pageSize = index
      this.getMatchTikectList()
    },
    ticketDetailListChangePage (index) {
      this.ticketDetailListPageData.page = index
      this.getMatchTikectList()
    },
    ticketDetailListChangePageSize (index) {
      this.ticketDetailListPageData.pageSize = index
      this.getMatchTikectList()
    },
    //
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
      } else if (tab == 'name5') {
        if (this.agentId != undefined) {
          this.getData({
            agentId: this.agentId
          })
        } else {
          this.getData()
        }
      }
    },

    async matchTypeList () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/match/typeList',
        method: 'get',
        params
      })

      if (data && data.data && data.data.code == 200) {
        this.matchTypeData = data.data.data
        sessionStorage.setItem('matchType', JSON.stringify(data.data.data))
      }
    },

    async gameTypeList () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/game/typeList',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        sessionStorage.setItem('gameType', JSON.stringify(data.data.data))
        this.gameTypeData = data.data.data
      }
    },

    async updateMatchTypeItem () {
      let params = {
        token: getToken(),
        id: this.curMatchTypeItem.id,
        name: this.curMatchTypeItem.name
      }
      if (this.curMatchTypeItem.isAdd == undefined) {
        await axios.request({
          url: 'v2/update/matchType',
          method: 'get',
          params
        })
      } else {
        await axios.request({
          url: 'v2/add/matchType',
          method: 'get',
          params
        })
      }
      this.matchTypeList()
    },

    async updateMatchItem () {
      let params = {
        token: getToken(),
        id: this.curMatchItem.id,
        name: this.curMatchItem.name,
        type: this.selectTypes.join(','),
        game_type: this.gameTypeParams.val,
        interval: this.curMatchItem.interval,
        award: this.matchAwardParmas.val,
        max_buy: this.curMatchItem.max_buy,
        game_id: this.games.value,
        fee: this.curMatchItem.fee,
        room_fee: this.curMatchItem.room_fee,
        blinds_id: this.blindsParams.val,
        score: this.curMatchItem.score,
        desc: this.curMatchItem.desc,
        news: this.curMatchItem.news,
        tag: this.curMatchItem.tag,
        settlement_type: this.matchSettleTypeParmas.val,
        start_time: this.matchStartTime,
        delay: this.curMatchItem.delay,
        num: this.curMatchItem.num,
        state: this.curMatchItem.state,
        sign_count: this.curMatchItem.sign,
        base: this.curMatchItem.base,
        create_time: this.curMatchItem.create_time,
        end_time: this.curMatchItem.end_time,
        pot: this.curMatchItem.pot,
        price: this.curMatchItem.price,
        agent_id: this.agentParams.val
          ? this.agentParams.val
          : this.agentParams.val === 0
            ? 0
            : undefined,
        level: this.curMatchItem.level
      }
      if (this.curMatchItem.isAdd == undefined) {
        await axios.request({
          url: 'v2/match/update',
          method: 'post',
          params
        })
      } else {
        await axios.request({
          url: 'v2/match/add',
          method: 'post',
          params
        })
      }
      this.getMatchList(-1)
    },

    async updateGameTypeItem () {
      let params = {
        token: getToken(),
        id: this.curGameTypeItem.id,
        name: this.curGameTypeItem.name
      }
      if (this.curGameTypeItem.isAdd == undefined) {
        await axios.request({
          url: 'v2/update/gameType',
          method: 'get',
          params
        })
      } else {
        await axios.request({
          url: 'v2/add/gameType',
          method: 'get',
          params
        })
      }
      this.gameTypeList()
    },

    async getBlindData () {
      let params = {
        token: getToken()
      }
      let data = await axios.request({
        url: 'v2/matchBlinds/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.blindData = data.data.data
        sessionStorage.setItem('blindList', JSON.stringify(data.data.data))
      }
    },

    async addBlindsDetailItem () {
      let max = 0
      this.blindsDataDetail.forEach(function (i) {
        if (i) {
          if (i.level >= max) {
            max = i.level
          }
        }
      })
      this.blindsDataDetail.push({
        level: ++max,
        min: '0.0',
        max: '0.0',
        delay: 0,
        bet: '0.0',
        interval: 0
      })
    },

    async updateBlindItem () {
      let d = {}
      d.items = this.blindsDataDetail
      let params = {
        token: getToken(),
        id: this.curBlindItem.id,
        name: this.curBlindItem.name,
        data: JSON.stringify(d)
      }
      if (this.curBlindItem.opt == 1) {
        await axios.request({
          url: 'v2/matchBlinds/add',
          method: 'get',
          params
        })
      } else {
        await axios.request({
          url: 'v2/matchBlinds/update',
          method: 'get',
          params
        })
      }
      this.getBlindData()
      this.editBlind = false
    },

    async showAddMatchTypeItem () {
      this.editMatchType = true
      this.curMatchTypeItem = {}
      this.curMatchTypeItem.name = ''
      this.curMatchTypeItem.isAdd = true
    },

    async showAddGameTypeItem () {
      this.editGameType = true
      this.curGameTypeItem = {}
      this.curGameTypeItem.name = ''
      this.curGameTypeItem.isAdd = true
    },

    async showMatchAwardPage () {
      this.editMatchAward = true
      this.curMatchAwardItem = {}
      this.curMatchAwardItem.name = ''
      this.curMatchAwardItem.detail = ''
      this.curMatchAwardItem.isAdd = true
    },

    async showAddMatchPage () {
      this.editMatch = true
      this.matchTag = '添加比赛'
      this.curMatchItem = {}
      this.curMatchItem.name = ''
      this.curMatchItem.type = 0
      this.curMatchItem.level = 0
      this.curMatchItem.interval = 0
      this.curMatchItem.award = 0
      this.curMatchItem.max_buy = 0
      this.curMatchItem.game_id = 0
      this.curMatchItem.state = 0
      this.curMatchItem.fee = '0'
      this.curMatchItem.room_fee = '0'
      this.curMatchItem.blinds_id = 0
      this.curMatchItem.score = '0'
      this.curMatchItem.desc = ''
      this.curMatchItem.news = ''
      this.curMatchItem.tag = ''
      this.curMatchItem.settlement_type = ''
      this.curMatchItem.delay = 0
      this.selectTypes = []
      this.curMatchItem.isAdd = true
      this.blindsParams.option = this.blindData
      this.matchTypeParams.option = this.matchTypeData
      this.matchAwardParmas.option = this.matchAwardData
    },

    async showEditMatchPage () {
      this.editMatch = true
      this.matchTag = '修改比赛'
      this.gameTypeParams.option = this.gameTypeData
      this.blindsParams.option = this.blindData
      this.matchTypeParams.option = this.matchTypeData
      this.matchAwardParmas.option = this.matchAwardData
      this.blindsParams.val = this.curMatchItem.blinds_id
      this.matchAwardParmas.val = this.curMatchItem.award
      this.matchStartTime = dayjs(this.curMatchItem.start_time * 1000).format(
        'YYYY-MM-DD HH:mm:ss'
      )
      this.games.value = this.curMatchItem.game_id
      this.matchSettleTypeParmas.val = this.curMatchItem.settlement_type
      this.selectTypes = []
      this.curMatchItem.type.split(',').forEach((item) => {
        this.selectTypes.push(Number(item))
      })
      this.gameTypeParams.val = this.curMatchItem.game_type
    },

    async showCreateTicketPage () {
      this.createTticket = true
      this.createTicketItem = {}
      this.createTicketItem.name = ''
      this.createTicketItem.count = 0
      this.createTicketItem.icon = ''
      this.createTicketItem.price = 0
      // this.addTicketMatchParams.option = await this.getMatchList(-1)
    },

    async showTicketDetailPage () {
      this.showTicketDetail = true
    },

    async showBlinditem () {
      this.editBlind = true
      this.curBlindItem = {}
      this.curBlindItem.name = ''
      this.curBlindItem.data = ''
      this.blindsDataDetail = []
      this.curBlindItem.opt = 1 // 插入
    },

    async showBlinditemPage () {
      this.editBlind = true
      this.curBlindItem = {}
      this.curBlindItem.name = ''
      this.curBlindItem.data = []
      this.curBlindItem.opt = 2 // 查看
    },

    async getMatchList (state) {
      let params = {
        token: getToken(),
        agentId: this.agentParams.val
          ? this.agentParams.val
          : this.agentParams.val === 0
            ? 0
            : undefined,
        state: state == -1 ? -1 : this.matchState.val,
        page: this.matchListPageData.page,
        pageSize: this.matchListPageData.pageSize
      }
      if (
        this.matchSearchTimes.startTime != '' &&
        this.matchSearchTimes.endTime != ''
      ) {
        params.startTime = dayjs(this.matchSearchTimes.startTime).unix()
        params.endTime = dayjs(this.matchSearchTimes.endTime).unix()
      }
      let data = await axios.request({
        url: 'v2/match/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.matchData = data.data.data.list
        this.matchListPageData.current = data.data.data.total
        return this.matchData
      }
    },

    async getMatchListByName (query) {
      let params = {
        token: getToken(),
        name: query
      }
      let data = await axios.request({
        url: 'v2/match/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        // this.matchParams.option = data.data.data.list;
        // this.matchListLoading=false;
        return data.data.data.list
      }
    },

    async loadGames () {
      let newArray = JSON.parse(sessionStorage.getItem('games'))
      this.games.option = []
      newArray.forEach((item) => {
        this.games.option.push({ id: item.number, name: item.name })
      })
      console.log(this.games.option)
    },
    genAwardParam (params) {
      let arr = []
      if (params.length > 0) {
        params.split(',').map((item) => {
          item = item + '%'
          arr.push(item)
        })
        return arr.join(',')
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

    async allSet () {
      this.formatParams()
    },

    // 房间
    async selectRoom (e) {
      this.agentSearch()
    },
    // 提示
    async successTip () {
      this.$Message.info('修改成功')
    },
    async selectRewordByPageOne () {}
  },

  async mounted () {
    this.getSite() // 获取玩家控制的头部代理参数
    await this.matchTypeList()
    await this.gameTypeList()
    await this.getMatchAwardList()
    await this.getBlindData()
    await this.getMatchTikectList()
    await this.loadGames()
    // await this.getMatchList(-1);
    this.matchParams.option = this.matchData
    this.blindsParams.option = this.blindData
    this.gameTypeParams.option = this.gameTypeData
    window.test1 = this
  }
}
</script>

<style scoped lang="less">
.inlineForm {
  display: flex;
  align-items: center;
}

.mzlbshow {
  width: 100%; /* 表格容器宽度 */
  overflow-y: auto; /* 添加垂直滚动条 */
  max-height: 500px; /* 设置最大高度 */
}
</style>
