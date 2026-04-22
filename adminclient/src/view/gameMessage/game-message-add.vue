<template>
  <div v-if="state == 0">
    <Card style="padding:30px 0">
      <Row>
        <i-col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <h3 style="margin:20px 0">消息基本内容</h3>
          <Divider />

          <Form label-position="right" :label-width="120">
            <FormItem label="游戏">
              <Select v-model="gameIds">
                <Option :value="-1">全部</Option>
                <Option
                  v-for="item in gameList"
                  :key="item.id"
                  :value="item.id"
                  >{{ item.name }}</Option
                >
              </Select>
              <Badge style="transform: translateY(-5px);">
                <Icon
                  type="ios-medical"
                  title="必填项"
                  style="margin-left: 20px;"
                  slot="count"
                  color="#ed4014"
                  size="12"
                />
              </Badge>
            </FormItem>
            <FormItem
              v-for="item of itemCreate"
              :key="item.lable"
              :label="item.lable"
            >
              <i-input
                v-if="item.type == 'text'"
                :maxlength="item.lable == '消息序号' ? 20 : 50"
                type="text"
                v-model="item.value"
                :placeholder="'请输入' + item.lable"
              ></i-input>

              <Select
                v-if="item.type == 'select'"
                v-model="item.value"
                style="width: 100%"
              >
                <Option
                  v-for="items in item.option"
                  :value="items.value"
                  :key="items.label"
                  >{{ items.label }}</Option
                >
              </Select>

              <i-input
                v-if="item.type == 'textarea'"
                type="textarea"
                maxlength="150"
                show-word-limit
                :rows="5"
                v-model="item.value"
                :placeholder="'请输入' + item.lable"
              ></i-input>

              <DatePicker
                v-if="item.type == 'datetime'"
                type="datetime"
                v-model="item.value"
                :placeholder="'请选择' + item.lable"
                style="width: 100%"
              ></DatePicker>

              <Badge v-if="item.required" style="transform: translateY(-5px);">
                <Icon
                  type="ios-medical"
                  title="必填项"
                  style="margin-left: 20px;"
                  slot="count"
                  color="#ed4014"
                  size="12"
                />
              </Badge>
            </FormItem>
            <FormItem>
              <Button type="primary" size="large" @click="submit"
                >创建消息</Button
              >
              <Button style="margin-left: 8px" size="large" @click="resForm"
                >重置</Button
              >
            </FormItem>
          </Form>
        </i-col>
      </Row>
    </Card>
  </div>
  <div v-else>
    <Card style="padding:30px 0">
      <Row>
        <i-col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <Card style="padding:15px;">
            <h3>基本信息：</h3>
            <List>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.title"
                :title="item.title"
                >{{ item.lable }}：{{ item.value }}</ListItem
              >
            </List>
          </Card>
          <div style="text-align: center;margin-top:30px;">
            <Button type="success" size="large" to="/game-message">完成</Button>
          </div>
        </i-col>
      </Row>
    </Card>
  </div>
</template>

<script>
import Step from '_c/step'
import { createGameMsgData, getLinkageList } from '@/api/data'
import { getDate } from '@/libs/tools'
export default {
  name: 'game_add',
  components: {
    Step
  },
  data () {
    return {
      state: this.$route.meta.state,
      stepList: [
        { title: '基本信息', content: '填写基本信息' },
        { title: '创建成功', content: '游戏消息创建成功' }
      ],
      itemCreate: [
        {
          lable: '消息序号',
          type: 'text',
          key: 'number',
          value: '',
          required: true
        },
        {
          lable: '消息标题',
          type: 'text',
          key: 'title',
          value: '',
          required: true
        },
        {
          lable: '消息类型',
          type: 'select',
          key: 'msgType',
          value: '',
          option: [
            { label: '活动消息', value: 1 },
            { label: '维护公告', value: 2 }
          ],
          required: true
        },
        {
          lable: '发布时间',
          type: 'datetime',
          key: 'startTime',
          value: '',
          required: true
        },
        {
          lable: '停止时间',
          type: 'datetime',
          key: 'endTime',
          value: '',
          required: true
        },
        {
          lable: '消息内容',
          type: 'textarea',
          key: 'info',
          value: '',
          required: true
        },
        { lable: '备注', type: 'text', key: 'remarks', value: '' }
      ],
      linkageList: [],
      gameList: [],
      webId: sessionStorage.getItem('siteVal'),
      siteTag: '',
      agentId: '',
      agentTag: '',
      gameIds: -1,
      gameShow: false
    }
  },
  methods: {
    resForm () {
      this.itemCreate.forEach(item => {
        item.value = ''
      })
      this.webId = ''
    },
    next () {
      this.state += 1
    },
    submit () {
      const Data = []
      Data.push(
        { webId: sessionStorage.getItem('siteVal') },
        { agentId: sessionStorage.getItem('agentVal') },
        { gameIds: this.gameIds ? this.gameIds : -1 }
      )
      this.itemCreate.map(item => {
        if (item.type == 'datetime') {
          Data.push({
            [item.key]: getDate(item.value)
          })
        } else {
          Data.push({
            [item.key]: item.value
          })
        }
      })

      /**
       * 验证时间范围合法
       */
      if (Data.find(x => x.startTime || x.endTime)) {
        let startTime = new Date(
          Data.find(x => x.startTime).startTime
        ).getTime()
        let endTime = new Date(Data.find(x => x.endTime).endTime).getTime()
        if (endTime - startTime <= 0) {
          this.$Message.error('开始时间不允许大于结束时间')
          return
        }
      }

      createGameMsgData(Data).then(res => {
        if (res.data.code == 200) {
          this.state += 1
          this.$Message.success('创建游戏消息成功')
        } else this.$Message.error(res.data.msg)
      })
    },
    changeOption (item) {
      this.siteTag = item.tag
      this.agentId = ''
      this.gameId = ''
    },
    changeOption1 (item) {
      this.agentTag = item.tag
      this.gameId = ''
      if (item.value > -1) this.gameShow = true
      else this.gameShow = false
    }
  },
  mounted () {
    getLinkageList().then(res => {
      this.linkageList.push(...res.data.data)
    })
    this.gameList.push(...JSON.parse(sessionStorage.getItem('gameOption')))
  }
}
</script>

<style></style>
