 <template>
  <div v-if="state==0">
    <Card style="padding:30px 0">
      <Row>
        <Col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </Col>
        <Col span="8" offset="8">
          <h3 style="margin:20px 0">基本信息</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <FormItem v-for="item of itemCreate" :key="item.lable" :label="item.lable">
              <Input v-model="item.input" :maxlength='50' :placeholder="'请输入' + item.lable"></Input>
            </FormItem>
            <FormItem label="生效时间">
              <DatePicker type="date" style="width:175px" placeholder="选择开始日期"></DatePicker>&emsp;—&emsp;
              <DatePicker type="date" style="width:175px" placeholder="选择结束日期"></DatePicker>
            </FormItem>
            <FormItem label="性质">
              <RadioGroup>
                <Radio>永久</Radio>
              </RadioGroup>
            </FormItem>
            <FormItem>
              <Button type="primary" size="large" @click="next">下一步</Button>
              <Button style="margin-left: 8px" size="large">重置</Button>
            </FormItem>
          </Form>
        </Col>
      </Row>
    </Card>
  </div>
  <div v-else-if="state==1">
    <Card style="padding:30px 0">
      <Row>
        <Col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </Col>
        <Col span="8" offset="8">
          <h3 style="margin:20px 0">配置代理商信息</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <FormItem v-for="item of itemSeting" :key="item.lable" :label="item.lable">
              <Input v-model="item.input" :maxlength='50' :placeholder="'请输入' + item.lable"></Input>
            </FormItem>
            <FormItem>
              <Button type="primary" size="large" @click="next">下一步</Button>
              <Button style="margin-left: 8px" size="large">重置</Button>
            </FormItem>
          </Form>
        </Col>
      </Row>
    </Card>
  </div>
  <div v-else>
    <Card style="padding:30px 0">
      <Row>
        <Col span="16" offset="4">
          <Step :state="state" :data="stepList"></Step>
        </Col>
        <Col span="8" offset="8">
          <Card style="padding:15px;">
            <List>
              <h3>基本信息：</h3>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.title"
                :title="item.title"
              >{{item.lable}}：{{item.input}}</ListItem>
              <br />
              <h3>配置代理商信息：</h3>
              <ListItem
                class="label-style"
                v-for="item of itemSeting"
                :key="item.title"
                :title="item.title"
              >{{item.lable}}：{{item.input}}</ListItem>
            </List>
          </Card>
          <div style="text-align: center;margin-top:30px;">
            <Button type="success" size="large" to="/agency">完成</Button>
            <Button style="margin-left: 15px" size="large" @click="change">修改</Button>
          </div>
        </Col>
      </Row>
    </Card>
  </div>
</template>

<script>
import Step from '_c/step'
export default {
  name: 'agency-add',
  components: {
    Step
  },
  data () {
    return {
      state: this.$route.meta.state,
      stepList: [
        { title: '填写基本信息', content: '填写基本信息' },
        { title: '配置代理商', content: '配置代理商' },
        { title: '创建成功', content: '创建成功' }
      ],
      itemCreate: [
        { lable: '代理商标识', input: '' },
        { lable: '代理商昵称', input: '' },
        { lable: '代理商邮箱', input: '' },
        { lable: '所属站点', input: '' },
        { lable: '上级代理商', input: '' },
        { lable: '备注信息', input: '' },
        { lable: '负责人', input: '' },
        { lable: '负责人联系方式', input: '' }
      ],
      itemSeting: [
        { lable: '代理商初始点数', input: '' },
        { lable: '上级剩余点数', input: '' },
        { lable: '点数预警值', input: '' },
        { lable: '点数停用值', input: '' },
        { lable: '代理商直管下级数量', input: '' },
        { lable: '代理商拥有游戏', input: '' },
        { lable: '代理前端域名', input: '' }
      ]
    }
  },
  methods: {
    next () {
      this.state += 1
    },
    change () {
      this.state = 0
    }
  },
  mounted () {}
}
</script>

<style>
</style>
