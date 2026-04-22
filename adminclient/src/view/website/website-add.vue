 <template>
  <div v-if="state == 0">
    <Card style="padding: 30px 0">
      <Row>
        <i-col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <h3 style="margin: 20px 0">基本信息</h3>
          <Divider />
          <Form label-position="right" :label-width="120">
            <FormItem
              v-for="item of itemCreate"
              :key="item.lable"
              :label="item.lable"
            >
              <i-input
                v-model="item.value"
                :maxlength="50"
                :placeholder="'请输入' + item.lable"
              ></i-input>
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
            <FormItem>
              <Button type="primary" size="large" @click="submit">提交</Button>
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
    <Card style="padding: 30px 0">
      <Row>
        <i-col span="8" offset="8">
          <Step :state="state" :data="stepList"></Step>
          <Card style="padding: 15px">
            <h3>基本信息：</h3>
            <List>
              <ListItem>站点ID：10001</ListItem>
              <ListItem
                class="label-style"
                v-for="item of itemCreate"
                :key="item.title"
                :title="item.title"
                >{{ item.lable }}：{{ item.value }}</ListItem
              >
            </List>
          </Card>
          <div style="text-align: center; margin-top: 30px">
            <Button type="success" size="large" @click="towebsite">完成</Button>
          </div>
        </i-col>
      </Row>
    </Card>
  </div>
</template>

<script>
import Step from '_c/step'
import { createSiteData } from '@/api/data'
export default {
  name: 'website_add',
  components: {
    Step
  },
  inject: ['reFreshSiteAgentList'],
  data () {
    return {
      state: this.$route.meta.state,
      stepList: [
        { title: '基本信息', content: '填写基本信息' },
        { title: '创建成功', content: '站点创建成功' }
      ],
      itemCreate: [
        { lable: '站点名称', key: 'nickName', value: '', required: true },
        { lable: '站点域名', key: 'realmName', value: '', required: true },
        { lable: '负责人', key: 'contacts', value: '', required: true },
        { lable: '负责人联系方式', key: 'phone', value: '', required: true },
        { lable: '站点邮箱', key: 'email', value: '', required: true },
        { lable: '备注信息', key: 'remarks', value: '' }
      ]
    }
  },
  methods: {
    towebsite () {
      // 刷新站点和代理信息
      this.reFreshSiteAgentList()

      this.$router.push({
        name: 'website'
      })
    },
    resForm () {
      this.itemCreate.forEach((item) => {
        item.value = ''
      })
    },
    next () {
      this.state += 1
    },
    submit () {
      const Data = []
      this.itemCreate.map((item) => {
        Data.push({
          [item.key]: item.value
        })
      })
      createSiteData(Data).then((res) => {
        if (res.data.code === 200) {
          this.state += 1
        }
      })
    }
  },
  mounted () {}
}
</script>
