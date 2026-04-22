<template>
  <div class="header-bar">
    <!-- 鎶ヨ鎻愮ず婊氬姩淇℃伅 -->
    <div class="warningMessage">
      <div class="wrap">
        <div id="warningText"></div>
      </div>
    </div>
    <custom-bread-crumb
      show-icon
      style="margin-left: 30px"
      :list="breadCrumbList"
    ></custom-bread-crumb>
    <div class="custom-content-con">
      <slot></slot>
    </div>
  </div>
</template>
<script>
import siderTrigger from './sider-trigger'
import customBreadCrumb from './custom-bread-crumb'
import './header-bar.less'

export default {
  name: 'HeaderBar',
  components: {
    siderTrigger,
    customBreadCrumb
  },
  props: {
    collapsed: Boolean
  },
  data () {
    return {
      online: 0,
      allOnline: 0,
      modal1: false,
      list: [], // 鍦ㄧ嚎浜烘暟
      proxyList: JSON.parse(sessionStorage.getItem('siteOption')) || []
    }
  },
  computed: {
    breadCrumbList () {
      return this.$store.state.app.breadCrumbList
    }
  },
  methods: {
    handleCollpasedChange (state) {
      this.$emit('on-coll-change', state)
    },
    keepFetchWarning () {
      clearInterval(window.scrollWarningTimer)
    },
    getName (id) {
      let name = ''
      this.proxyList &&
        this.proxyList.map((item) => {
          if (item.agentList) {
            return item.agentList.map((agent) => {
              if (agent.id == id) {
                name = agent.name
              }
              return undefined
            })
          }
          return undefined
        })
      return name
    }
  },
  mounted () {
    this.keepFetchWarning()
  }
}
</script>
<style scoped>
.warningMessage {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  padding-right: 150px;

  padding-left: 150px;
}

.warningMessage .wrap {
  overflow: hidden;
  position: relative;
  left: 0;
  top: 0;
  height: 100%;
}

#warningText {
  position: absolute;
  left: 0;
  top: 0;
  white-space: nowrap;
  color: rgb(255, 94, 0);
}
</style>
