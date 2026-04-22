<style lang="less">
@import "./login.less";
</style>

<template>
  <div class="login">
    <div class="login-con">
      <Card icon="log-in" title="欢迎登录" :bordered="false">
        <div class="form-con">
          <login-form @on-success-valid="handleSubmit"></login-form>
          <!-- <p class="login-tip"></p> -->
        </div>
      </Card>
    </div>
    <Modal v-model="modal1" title="选择游戏地址" @on-ok="ok">
      <RadioGroup v-model="url" vertical>
        <Radio :label="item.url" :key="item.url" v-for="item in urlList">
          <Icon type="social-apple"></Icon>
          <span>{{ item.desc }}</span>
        </Radio>
      </RadioGroup>
    </Modal>
  </div>
</template>

<script>
import LoginForm from '_c/login-form'
import { mapActions } from 'vuex'
import { getGameData2 } from '@/api/data'
export default {
  components: {
    LoginForm
  },
  data () {
    return {
      modal1: false,
      urlList: [],
      url: ''
    }
  },
  methods: {
    ...mapActions(['handleLogin', 'getUserInfo']),
    ok () {
      if (this.url) {
        sessionStorage.setItem('node_url', this.url)
        getGameData2().then(({ data }) => {
          if (data.code === 200) {
            sessionStorage.setItem('games', JSON.stringify(data.data))
          } else {
            this.$Message.error(data.code + ' ：&nbsp;' + data.msg)
          }
        })
        this.modal1 = false
        this.$router.push({
          name: 'newHome'
        })
      } else {
        this.$Message.error('选择一个地址')
      }
    },
    handleSubmit ({ name, password }) {
      this.handleLogin({ name, password }).then((res) => {
        if (res.status === 200) {
          this.urlList = res.data.data.web_node || []
          getGameData2().then(({ data }) => {
            if (data.code === 200) {
              sessionStorage.setItem('games', JSON.stringify(data.data))
            } else {
              this.$Message.error(data.code + ' ：&nbsp;' + data.msg)
            }
          })
          switch (res.data.data.user.userType) {
            case 1:
              this.$router.push({
                name: 'newHome'
              })
              break
            case 2:
              this.$router.push({
                name: 'players'
              })
              break
            case 3:
              this.$router.push({
                name: 'agent'
              })
              break
          }

          sessionStorage.setItem('sign', res.data.data.sign)
        }
      })
    }
  }
}
</script>

<style>
</style>
