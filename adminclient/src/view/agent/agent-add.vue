<template>
  <div v-if="state == 0">
    <Card style="padding: 30px 0">
      <Row>
        <i-col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </i-col>
        <i-col span="8" offset="8">
          <h3 style="margin: 20px 0">基本信息</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <FormItem
              v-for="item of itemCreate"
              :key="item.lable"
              :label="item.lable"
            >
              <i-input
                :maxlength="50"
                v-if="item.type == 'text'"
                v-model="item.value"
                :placeholder="'请输入' + item.lable"
              ></i-input>

              <Select
                v-if="item.type == 'select'"
                v-model="item.value"
                clearable
                style="width: 100%"
              >
                <Option
                  v-for="items in item.option"
                  :value="items.id"
                  :key="items.id"
                  >{{ items.name }}</Option
                >
              </Select>

              <template v-if="item.type == 'Tree' && item.data">
                <Tree
                  :data="item.data"
                  @on-select-change="higherAgentSelect"
                  style="transform: translateY(-7px)"
                ></Tree>
              </template>

              <Badge v-if="item.required" style="transform: translateY(-5px)">
                <Icon
                  type="ios-medical"
                  title="必填项"
                  style="margin-left: 20px"
                  slot="count"
                  color="#ed4014"
                  size="12"
                />
              </Badge>
            </FormItem>
            <!-- <FormItem label="生效时间" v-if="this.isPermanent == false">
              <DatePicker
                type="datetime"
                style="width:190px"
                v-model="startTime"
                placeholder="开始时间"
              ></DatePicker>&nbsp;—
              <DatePicker type="datetime" style="width:190px" v-model="endTime" placeholder="结束时间"></DatePicker>
              <Badge>
                <Icon
                  type="ios-medical"
                  title="必填项"
                  style="margin-left: 20px;"
                  slot="count"
                  color="#ed4014"
                  size="12"
                />
              </Badge>
            </FormItem>-->
            <!-- <FormItem label="性质">
              <Checkbox v-model="isPermanent">永久</Checkbox>
            </FormItem>-->
            <FormItem>
              <Button type="primary" size="large" @click="next">下一步</Button>
              <Button style="margin-left: 8px" size="large" @click="resForm"
                >重置</Button
              >
            </FormItem>
          </Form>
        </i-col>
      </Row>
    </Card>
  </div>
  <div v-else-if="state == 1">
    <Card style="padding: 30px 0">
      <Row>
        <i-col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </i-col>
        <i-col span="8" offset="8">
          <h3 style="margin: 20px 0">配置代理商信息</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <template v-for="item of itemSeting">
              <FormItem
                v-if="
                  !(
                    item.type == 'textarea' &&
                    itemSeting.find((x) => x.key == 'useDefault').value == true
                  )
                "
                :key="item.lable"
                :label="item.lable"
              >
                <Checkbox v-if="item.type == 'checkbox'" v-model="item.value">{{
                  ""
                }}</Checkbox>

                <i-input
                  v-if="
                    item.type == 'textarea' &&
                    itemSeting.find((x) => x.key == 'useDefault').value == false
                  "
                  type="textarea"
                  :rows="5"
                  v-model="item.value"
                  :placeholder="'请输入' + item.lable"
                ></i-input>
                <i-input
                  v-if="item.type == 'text'"
                  v-model="item.value"
                  :maxlength="50"
                  :placeholder="'请输入' + item.lable"
                ></i-input>

                <Badge v-if="item.required">
                  <Icon
                    type="ios-medical"
                    title="必填项"
                    style="margin-left: 20px"
                    slot="count"
                    color="#ed4014"
                    size="12"
                  />
                </Badge>
              </FormItem>
            </template>
            <!-- <FormItem label="代理商拥有游戏">
              <div v-for="items in gameIds.option" :key="items.id">
                <div>
                  <Checkbox
                    :indeterminate="indeterminate"
                    :value="checkAll[items.id - 1]"
                    @click.prevent.native="handleCheckAll(items.id - 1)"
                    disabled
                    >{{ items.name }}</Checkbox
                  >
                </div>
                <CheckboxGroup v-model="gameIds.value[items.id - 1].val">
                  <Checkbox
                    v-for="subitems in items.gameList"
                    :key="subitems.id"
                    :label="subitems.id"
                    disabled
                    >{{ subitems.name }}</Checkbox
                  >
                </CheckboxGroup>
                <Divider style="margin: 6px 0" />
              </div>
            </FormItem> -->
            <FormItem>
              <Button type="primary" size="large" @click="submit"
                >下一步</Button
              >
              <Button style="margin-left: 8px" @click="previous" size="large"
                >上一步</Button
              >
            </FormItem>
          </Form>
        </i-col>
      </Row>
    </Card>
  </div>
  <div v-else>
    <Card style="padding: 30px 0">
      <Row>
        <i-col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </i-col>
        <i-col span="8" offset="8">
          <Card style="padding: 15px">
            <List>
              <h3>基本信息：</h3>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.title"
                :title="item.title"
                >{{ item.lable }}：{{ item.value }}</ListItem
              >
              <br />
              <h3>配置代理商信息：</h3>
              <ListItem
                class="label-style"
                v-for="item of itemSeting"
                :key="item.title"
                :title="item.title"
                >{{ item.lable }}：{{ item.value }}</ListItem
              >
            </List>
          </Card>
          <div style="text-align: center; margin-top: 30px">
            <Button type="success" size="large" @click="toagent">完成</Button>
          </div>
        </i-col>
      </Row>
    </Card>
  </div>
</template>

<script>
import Step from '_c/step'
import {
  createAgentData,
  getLinkageList,
  getSelectAgent
  // getSelectClassGames,
} from '@/api/data'
export default {
  name: 'agency-add',
  components: {
    Step
  },
  inject: ['handleLogOut', 'reFreshSiteAgentList'],
  data () {
    return {
      originalHigherAgentData: null,
      state: this.$route.meta.state,
      indeterminate: false,
      checkAll: [],
      gameIds: {
        value: [],
        option: []
      },
      stepList: [
        { title: '填写基本信息', content: '填写基本信息' },
        { title: '配置代理商', content: '配置代理商' },
        { title: '创建成功', content: '创建成功' }
      ],
      itemCreate: [
        {
          lable: '代理名',
          type: 'text',
          key: 'nickName',
          value: '',
          required: true
        },
        {
          lable: '持有人',
          type: 'text',
          key: 'uName',
          value: '',
          required: true
        },
        {
          lable: '账号',
          type: 'text',
          key: 'account',
          value: '',
          required: true
        },
        {
          lable: '密码',
          type: 'text',
          key: 'password',
          value: '',
          required: true
        },
        {
          lable: '所属站点',
          type: 'select',
          key: 'webId',
          value: '',
          option: [],
          required: true
        },
        // {
        //   lable: "上级代理",
        //   type: "Tree",
        //   key: "upperLevel",
        //   value: "",
        //   data:null,
        //   option: []
        // },
        {
          lable: '负责人',
          type: 'text',
          key: 'contacts',
          value: '1',
          required: true
        },
        {
          lable: '负责人联系方式',
          type: 'text',
          key: 'phone',
          value: '1',
          required: true
        },
        // { lable: "代理邮箱", type: "text", key: "email", value: "",required: true },
        { lable: '备注信息', type: 'text', key: 'remarks', value: '' }
      ],
      isPermanent: false,
      startTime: '',
      endTime: '',

      itemSeting: [
        // {
        //   lable: "代理商初始点数",
        //   type: "text",
        //   key: "point",
        //   value: 1000000000,
        //   required: true,
        // },
        // {
        //   lable: "点数预警值",
        //   type: "text",
        //   key: "warningPoint",
        //   value: 1123,
        //   required: true,
        // },
        // {
        //   lable: "点数停用值",
        //   type: "text",
        //   key: "stopPoint",
        //   value: 123,
        //   required: true,
        // },
        { lable: '代理前端域名', type: 'text', key: 'realmName', value: '' }
        // {
        //   lable: "使用默认配置",
        //   type: "checkbox",
        //   key: "useDefault",
        //   value: true,
        // },
        // {
        //   lable: "MYSQL配置",
        //   type: "textarea",
        //   key: "mysqlConf",
        //   value: "",
        //   // required: true
        // },
        // {
        //   lable: "Redis配置",
        //   type: "textarea",
        //   key: "redisConf",
        //   value: "",
        //   // required: true
        // },
        // {
        //   lable: "ETCD配置",
        //   type: "textarea",
        //   key: "etcdConf",
        //   value: "",
        //   // required: true
        // },
      ]
    }
  },
  methods: {
    toagent () {
      // 刷新站点和代理信息
      this.reFreshSiteAgentList()

      this.$router.push({
        name: 'agent'
      })
    },
    higherAgentSelect (e) {
      if (e.length) {
        this.itemCreate.map((item) => {
          if (item.key == 'upperLevel') {
            item.value = e[0].id
          }
        })
      } else {
        // 选中空
        this.itemCreate.map((item) => {
          if (item.key == 'upperLevel') {
            item.data = JSON.parse(
              JSON.stringify(this.originalHigherAgentData)
            )
            item.value = ''
          }
        })
      }
    },
    resForm () {
      this.itemCreate.forEach((item) => {
        item.value = ''
      })
      this.isPermanent = false
    },
    handleCheckAll (tid) {
      if (this.indeterminate) {
        this.checkAll[tid] = false
      } else {
        this.checkAll[tid] = !this.checkAll[tid]
      }
      this.indeterminate = false
      if (this.checkAll[tid]) {
        this.gameIds.value[tid].val = Object.assign(
          this.gameIds.option[tid].gameList.map((item) => {
            return item.id
          })
        )
      } else {
        this.gameIds.value[tid].val = []
      }
    },
    next () {
      this.state += 1
    },
    previous () {
      this.state -= 1
    },
    async submit () {
      let Data = []
      this.itemCreate.map((item) => {
        Data.push({
          [item.key]: item.value
        })
      })
      this.itemSeting.map((item) => {
        Data.push({
          [item.key]: item.value
        })
      })
      Data.push(
        {
          isPermanent: 1
        },
        { gameIds: JSON.stringify([...Object.assign(this.gameIds.value)]) }
      )
      createAgentData(Data).then((res) => {
        if (res.data.msg == '操作成功') {
          this.state += 1
        } else {
          this.$Message.error(res.data.msg)
        }
      })
    }
  },
  mounted () {
    for (let i in this.itemCreate) {
      if (this.itemCreate[i].key == 'webId') {
        getLinkageList().then((res) => {
          if (res.data.code == 200) {
            for (let i in res.data.data) {
              if (res.data.data[i].agentList.length > 0) {
                for (let j in res.data.data[i].agentList) {
                  res.data.data[i].agentList[j].gameIds = undefined
                  res.data.data[i].agentList[j].gameList = undefined
                }
              }
            }
            this.itemCreate[i].option.push(...res.data.data)
          }
        })
        if (sessionStorage.getItem('siteVal')) {
          this.itemCreate[i].value = Number(sessionStorage.getItem('siteVal'))
        }
      }
    }

    // 判断上级代理，并对选项进行赋值
    getSelectAgent().then((res) => {
      this.itemCreate.forEach((item) => {
        if (item.key == 'upperLevel') {
          // 代理层级数据转换成树桩结构

          let convert = (data) => {
            if (data.subList) {
              return {
                expand: true,
                id: data.id,
                title: data.nickName || data.name,
                children: data.subList.map((d) => convert(d))
              }
            } else {
              return {
                // expand: true,
                id: data.id,
                title: data.nickName || data.name
                // children:[]
              }
            }
          }

          item.data = res.data.data.map((d) => convert(d))
          item.data.unshift({
            title: '无',
            selected: true,
            id: -1
          })
          this.originalHigherAgentData = JSON.parse(JSON.stringify(item.data))
        }
      })
    })
  }
}
</script>

<style></style>
