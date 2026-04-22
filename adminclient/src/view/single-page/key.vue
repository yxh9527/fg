<template>
  <div>
    <Card>
      <h3>Key页面</h3>
      <br />
      <ul>
        <li class="label-style" style="width:300px">
          <i-input v-model="key" :maxlength='50'>
            <label slot="prepend">Key:</label>
          </i-input>
        </li>
        <br />
        <li class="label-style" style="width:300px">
          <label>Config:</label>
          <i-input type="textarea" :maxlength='150' :autosize="{ minRows: 5, maxRows: 10 }" v-model="config"></i-input>
        </li>
        <li class="label-style">
          <Button type="primary" @click="submit()">提交</Button>
        </li>
      </ul>
    </Card>
  </div>
</template>
<script>
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
export default {
  name: 'key_page',
  data () {
    return { key: '', config: '' }
  },
  methods: {
    submit () {
      axios
        .request({
          url: 'v1/game/config',
          method: 'post',
          data: {
            token: getToken(),
            path: this.key,
            conf: this.config
          }
        })
        .then(res => {
          //
          if (res.data.code == 200) {
            this.$Message.success('修改配置成功')
          }
        })
    }
  }
}
</script>
