 <template>
  <div v-if="state==0">
    <Card>
      <Row>
        <i-col span="10" offset="6">
          <Step :state="state" :data="stepList"></Step>
        </i-col>
        <br />
        <Form label-position="right" :label-width="120">
          <i-col span="8" offset="2">
            <h3 style="margin:20px 0">基本信息</h3>
            <Divider />
            <FormItem v-for="item of itemCreate" :key="item.lable" :label="item.lable">
              <i-input
                v-if="item.type=='text'"
                :maxlength='50'
                v-model="item.value"
                :placeholder="'请输入' + item.lable"
              ></i-input>
              <Select v-if="item.type=='select'" v-model="item.value" style="width: 100%">
                <Option
                  v-for="items in item.option"
                  :value="items.id"
                  :key="items.id"
                >{{ items.name }}</Option>
              </Select>
            </FormItem>
          </i-col>
          <i-col span="12" offset="1">
            <h3 style="margin:20px 0">房间配置信息</h3>
            <Divider />
            <div v-for="item of confParam" :key="item.lable1">
              <FormItem :label="item.lable1" style="display: inline-block">
                <i-input v-model="item.value1" :maxlength='50' :placeholder="'请输入' + item.lable1"></i-input>
              </FormItem>
              <FormItem :label="item.lable2" style="display: inline-block">
                <i-input v-model="item.value2" :maxlength='50' type="text" :placeholder="'请输入' + item.lable2"></i-input>
              </FormItem>
            </div>
          </i-col>
          <div style="text-align: center;clear: both;padding-top: 40px;padding-bottom: 20px;">
            <Button type="primary" size="large" @click="submit">下一步</Button>
            <Button style="margin-left: 8px" size="large" @click="resForm">重置</Button>
          </div>
        </Form>
      </Row>
    </Card>
  </div>
  <div v-else>
    <Card>
      <Row>
        <i-col span="10" offset="7">
          <Step :state="state" :data="stepList"></Step>
          <Card style="padding:15px;">
            <h3>基本信息：</h3>
            <List>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.lable"
                :title="item.lable"
              >{{item.lable}}：{{item.value}}</ListItem>
            </List>
          </Card>
          <div style="text-align: center;margin-top:30px;">
            <Button type="success" size="large" to="/gameManage">完成</Button>
          </div>
        </i-col>
      </Row>
    </Card>
  </div>
</template>

<script>
import Step from '_c/step'
import { createGameData } from '@/api/data'
export default {
  name: 'game_add',
  components: {
    Step
  },
  data () {
    return {
      state: this.$route.meta.state,
      stepList: [
        { title: '填写基本信息', content: '填写基本信息' },
        { title: '创建成功', content: '站点创建成功' }
      ],
      itemCreate: [
        { lable: '游戏序号', type: 'text', key: 'number', value: '' },
        { lable: '游戏名称', type: 'text', key: 'name', value: '' },
        { lable: '游戏Url', type: 'text', key: 'url', value: '' },
        {
          lable: '游戏分类',
          type: 'select',
          key: 'gameType',
          value: '',
          option: []
        },
        {
          lable: '游戏平台',
          type: 'select',
          key: 'gameClass',
          value: '',
          option: []
        },
        { lable: '单局时长设定', type: 'text', key: 'limitTime', value: '' },
        { lable: '备注信息', type: 'text', key: 'remarks', value: '' }
      ],
      confParam: [
        {
          roomType: 1,
          lable1: '初级房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '初级房准入',
          key2: 'getIntoScore',
          value2: ''
        },
        {
          roomType: 2,
          lable1: '中级房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '中级房准入',
          key2: 'getIntoScore',
          value2: ''
        },
        {
          roomType: 3,
          lable1: '高级房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '高级房准入',
          key2: 'getIntoScore',
          value2: ''
        },
        {
          roomType: 4,
          lable1: '王者房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '王者房准入',
          key2: 'getIntoScore',
          value2: ''
        },
        {
          roomType: 5,
          lable1: '至尊房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '至尊房准入',
          key2: 'getIntoScore',
          value2: ''
        },
        {
          roomType: 6,
          lable1: '尊享房底注',
          key1: 'baseScore',
          value1: '',
          lable2: '尊享房准入',
          key2: 'getIntoScore',
          value2: ''
        }
      ]
    }
  },
  methods: {
    resForm () {
      this.itemCreate.forEach(item => {
        item.value = ''
      })
    },
    next () {
      this.state += 1
    },
    change () {
      this.state = 0
    },
    submit () {
      let Data = []
      let Room = []
      this.itemCreate.map(item => {
        Data.push({
          [item.key]: item.value
        })
      })
      this.confParam.map(item => {
        Room.push({
          roomType: item.roomType,
          [item.key1]: item.value1,
          [item.key2]: item.value2
        })
      })
      Data.push({ confParam: [...Room] })
      createGameData(Data).then(res => {
        if (res.data.msg == '操作成功') {
          this.state += 1
        }
      })
    }
  },
  mounted () {
    this.itemCreate.map(item => {
      // if (item.key == "gameClass") {
      //   getClassList(sessionStorage.getItem("agentVal")).then(res => {
      //     item.option.push(...Object.assign(res.data.data));
      //   });
      // }
      // if (item.key == "gameType") {
      //   getTypeList(sessionStorage.getItem("agentVal")).then(res => {
      //     item.option.push(...Object.assign(res.data.data));
      //   });
      // }
    })
  }
}
</script>

<style>
</style>
