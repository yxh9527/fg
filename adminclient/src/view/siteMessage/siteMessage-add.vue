 <template>
  <div v-if="state==0">
    <Card style="padding:30px 0">
      <Row>
        <Col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <h3 style="margin:20px 0">消息基本内容</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <FormItem v-for="item of itemCreate" :key="item.lable" :label="item.lable">
              <Input v-model="item.input" :maxlength='50' :placeholder="'请输入' + item.lable"></Input>
            </FormItem>
            <FormItem v-for="item of itemTime" :key="item.lable" :label="item.lable">
              <DatePicker type="datetime" :placeholder="'请选择' + item.lable" style="width: 300px"></DatePicker>
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
        <Col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <Card style="padding:15px;">
            <h3>基本信息：</h3>
            <List>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.title"
                :title="item.title"
              >{{item.lable}}：{{item.input}}</ListItem>
            </List>
          </Card>
          <div style="text-align: center;margin-top:30px;">
            <Button type="success" size="large" to="/siteMessage">完成</Button>
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
  name: 'game_add',
  components: {
    Step
  },
  data () {
    return {
      state: this.$route.meta.state,
      stepList: [
        { title: '填写消息基本内容', content: '填写基本信息' },
        { title: '创建成功', content: '创建成功' }
      ],
      itemCreate: [
        { lable: '消息序号', input: '' },
        { lable: '消息标题', input: '' },
        { lable: '消息类型', input: '' },
        { lable: '收件人', input: '' },
        { lable: '消息内容', input: '' },
        { lable: '备注', input: '' }
      ],
      itemTime: [
        { lable: '发布时间', input: '' },
        { lable: '停止时间', input: '' }
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
