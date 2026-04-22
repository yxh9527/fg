<template>
  <div>
    <Card>
      <div style="padding:50px 0">
        <div style="padding:0 150px">
          <Steps :current="currentStep">
            <Step title="填写基本信息"></Step>
            <Step title="人机设置"></Step>
            <Step title="创建成功"></Step>
          </Steps>
        </div>

        <div v-if="currentStep == 0" style="width:500px;margin:0 auto; padding:30px 0;">
          <Form ref="baseInfoForm" :model="baseformItem" :label-width="180" :rules="baseformRule">
            <FormItem label="站点" prop="site">
              <Select v-model="baseformItem.site" @on-change="siteChanged">
                <template v-for="site in siteOptions">
                  <Option :value="site.id" :key="site.id">{{
                    site.name
                  }}</Option>
                </template>
              </Select>
            </FormItem>
            <FormItem label="代理" prop="agent">
              <Select filterable v-model="baseformItem.agent">
                <template v-for="agent in getAgentList">
                  <Option :value="agent.id" :key="agent.id">{{
                    agent.name
                  }}</Option>
                </template>
              </Select>
            </FormItem>
            <FormItem label="游戏" prop="game">
              <Select v-model="baseformItem.game">
                <Option v-for='game in gameOptions' :key="game.id" :value="game.id">{{game.name}}</Option>
              </Select>
            </FormItem>
            <FormItem label="名称" prop="groupName">
              <i-input :maxlength="50" v-model="baseformItem.groupName" placeholder="请输入群名称"></i-input>
            </FormItem>
            <FormItem label="群公告" prop="groupNotice">
              <i-input :maxlength="150" :rows="5" type="textarea" v-model="baseformItem.groupNotice"
                placeholder="请输入群公告"></i-input>
            </FormItem>
            <FormItem label="准入金额" prop="admittance">
              <InputNumber :max="999" :min="1" v-model="baseformItem.admittance"></InputNumber>
            </FormItem>
            <FormItem label="玩家人数上限" prop="playersLimit">
              <InputNumber :max="999" :min="1" v-model="baseformItem.playersLimit"></InputNumber>
            </FormItem>
            <FormItem label="位置号" prop="seatNumber">
              <InputNumber :max="999" :min="1" v-model="baseformItem.seatNumber"></InputNumber>
            </FormItem>
            <FormItem label="抽水" prop="sharing">
              <InputNumber :max="100" :min="0" v-model="baseformItem.sharing"></InputNumber>%
            </FormItem>
          </Form>
        </div>

        <div v-if="currentStep == 1" style="width:500px;margin:0 auto; padding:30px 0;">
          <Form ref="computerForm" :model="computerSetting" :label-width="180">
            <FormItem :label="sumTitle" prop="sum">
              <InputNumber :max="computerSetting.sum[1]" :min="sumRange[0]" :precision="0"
                v-model="computerSetting.sum[0]"></InputNumber>
              <div style="display:inline-block;
                padding-right:10px;
                padding-left:10px">
                —
              </div>
              <InputNumber :max="sumRange[1]" :min="computerSetting.sum[0]" :precision="0"
                v-model="computerSetting.sum[1]"></InputNumber>
            </FormItem>
            <FormItem label="随机金额间隔" prop="interval_0">
              <Select v-model="computerSetting.interval_0">
                <Option :value="1">1</Option>
                <Option :value="5">5</Option>
                <Option :value="10">10</Option>
                <Option :value="50">50</Option>
                <Option :value="100">100</Option>
              </Select>
            </FormItem>
            <FormItem :label="redPacketTitle" prop="redPacket">
              <InputNumber :max="computerSetting.redPacket[1]" :min="redPacketRange[0]" :precision="0"
                v-model="computerSetting.redPacket[0]"></InputNumber>
              <div style="display:inline-block;padding-right:10px;padding-left:10px">—</div>
              <InputNumber :max="redPacketRange[1]" :min="computerSetting.redPacket[0]" :precision="0"
                v-model="computerSetting.redPacket[1]"></InputNumber>
            </FormItem>
            <FormItem label="随机金额间隔" prop="interval_1">
              <Select v-model="computerSetting.interval_1">
                <Option :value="1">1</Option>
                <Option :value="5">5</Option>
                <Option :value="10">10</Option>
                <Option :value="50">50</Option>
                <Option :value="100">100</Option>
              </Select>
            </FormItem>
            <FormItem label="雷号" prop="thunderNumber">
              {{ computerSetting.thunderNumber }}
            </FormItem>
          </Form>
        </div>

        <div v-if="currentStep == 2" style="width:500px;margin:0 auto; padding:30px 0;">
          <div style='text-align:center;padding-bottom: 30px;'>
            <Icon style='color:green;font-size:80px' type="md-checkmark-circle" />
            <div style='font-size:20px;'>创建成功</div>
          </div>
          <Card>
            <p slot="title">基本信息:</p>
            <p>
              <table style='width:100%;' cellspacing='10'>
                <tr>
                  <td width='80' align='right'>站点:</td>
                  <td>{{siteOptions.length ? siteOptions.find(site=>site.id == baseformItem.site).name : ''}}</td>
                  <td width='80' align='right'>代理:</td>
                  <td>{{getAgentList.length? getAgentList.find(agent=>agent.id == baseformItem.agent).name : ''}}</td>
                </tr>
                <tr>
                  <td width='80' align='right'>游戏:</td>
                  <td>{{gameOptions.find(game=>game.id == baseformItem.game).name}}</td>
                  <td width='80' align='right'>名称:</td>
                  <td>{{baseformItem.groupName}}</td>
                </tr>
                <tr>
                  <td width='80' align='right'>准入金额:</td>
                  <td>{{baseformItem.admittance}}</td>
                  <td width='80' align='right'>玩家上限:</td>
                  <td>{{baseformItem.playersLimit}}</td>
                </tr>
                <tr>
                  <td width='80' align='right'>群公告:</td>
                  <td colspan='3'>{{baseformItem.groupNotice}}</td>
                </tr>
              </table>
            </p>
          </Card>
        </div>

        <!-- 确认按钮 -->
        <div style="padding:30px 0; text-align:center;">
          <template v-if='currentStep<2'>
            <Button @click="baseFormSubmit" type="primary" style="margin-right:50px">下一步</Button>
            <Button @click="resetBaseForm">重置</Button>
          </template>
          <template v-else>
            <Button type="primary" @click="resetBaseForm">完成</Button>
          </template>
        </div>

      </div>
    </Card>
  </div>
</template>

<script>
export default {
  data () {
    return {
      currentStep: 2,
      // 基本信息表单
      baseformItem: {
        site: Number(sessionStorage.getItem('siteVal')),
        agent: Number(sessionStorage.getItem('agentVal')),
        game: 1,
        groupName: '',
        groupNotice: '',
        admittance: 0,
        playersLimit: 50,
        seatNumber: 1,
        sharing: 0
      },
      // 游戏选项
      gameOptions: [
        {
          id: 1,
          name: '扫雷红包'
        },
        {
          id: 2,
          name: '牛牛红包'
        },
        {
          id: 3,
          name: '抢庄牛牛'
        },
        {
          id: 4,
          name: '二八杠'
        }
      ],
      // 站点列表
      siteOptions: [],
      // 基本信息表单数据验证
      baseformRule: {
        site: [
          {
            required: true,
            type: 'number',
            message: '请选择站点',
            trigger: 'blur'
          }
        ],
        agent: [
          {
            required: true,
            type: 'number',
            message: '请选择代理',
            trigger: 'blur'
          }
        ],
        game: [
          {
            required: true,
            type: 'number',
            message: '请选择游戏',
            trigger: 'blur'
          }
        ],
        admittance: [
          {
            required: true,
            type: 'number',
            message: '请输入准入金额',
            trigger: 'blur'
          }
        ],
        seatNumber: [
          {
            required: true,
            type: 'number',
            message: '请输入位置号',
            trigger: 'blur'
          }
        ],
        sharing: [
          {
            required: true,
            type: 'number',
            message: '请输入抽水',
            trigger: 'blur'
          }
        ],
        groupName: [
          {
            required: true,
            validator: (rule, value, callback) => {
              if (value == '') {
                callback(new Error('请输入群名称'))
              } else {
                callback()
              }
            },
            trigger: 'blur'
          }
        ]
      },
      // 人机设置表单
      computerSetting: {
        sum: [10, 100],
        interval_0: 10,
        redPacket: [5, 10],
        interval_1: 10,
        thunderNumber: Math.floor(Math.random() * 10)
      },
      // 人机设置title
      sumTitle: '总金额',
      redPacketTitle: '红包个数',
      // 人机设置范围
      sumRange: [5, 800],
      redPacketRange: [5, 50]
    }
  },
  computed: {
    /**
     * 获取代理列表
     */
    getAgentList () {
      if (this.siteOptions && this.siteOptions.length) {
        return this.siteOptions.find(
          (site) => site.id == this.baseformItem.site
        ).agentList
      } else {
        return []
      }
    }
  },
  methods: {
    /**
     * 切换站点
     */
    siteChanged (value) {
      // 重置代理为当前站点第一个
      let agentList = this.siteOptions.find(
        (site) => site.id == value
      ).agentList
      if (agentList.length > 0) {
        this.baseformItem.agent = agentList[0].id
      } else {
        this.baseformItem.agent = null
      }
    },
    /**
     * 重置基本信息表单
     */
    resetBaseForm () {
      if (this.currentStep == 0) {
        this.$refs['baseInfoForm'].resetFields()
      } else {
        this.$refs['computerForm'].resetFields()
      }
    },
    /**
     * 提交信息表单
     */
    baseFormSubmit () {
      if (this.currentStep == 0) {
        this.$refs['baseInfoForm'].validate((valid) => {
          if (valid) {
            this.currentStep = 1
          } else {
            this.$Message.error('请检查必填项!')
          }
        })
      } else {
        this.currentStep = 2
      }
    }
  },
  mounted () {
    /**
     * 初始化站点列表
     */
    let getSiteTimer = () => {
      if (window.windowData_getLinkageList) {
        this.siteOptions = window.windowData_getLinkageList
      } else {
        setTimeout(getSiteTimer, 100)
      }
    }
    getSiteTimer()
  }
}
</script>
