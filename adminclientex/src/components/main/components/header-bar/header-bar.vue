<template>
  <div class="header-bar">
    <!-- 报警提示滚动信息 -->
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
import axios from "@/libs/api.request";
import siderTrigger from "./sider-trigger";
import customBreadCrumb from "./custom-bread-crumb";
import "./header-bar.less";
import { getToken } from "@/libs/util";
export default {
  name: "HeaderBar",
  components: {
    siderTrigger,
    customBreadCrumb,
  },
  props: {
    collapsed: Boolean,
  },
  data() {
    return {
      online: 0,
      allOnline: 0,
      modal1: false,
      list: [], // 在线人数
      proxyList: JSON.parse(sessionStorage.getItem("siteOption")) || [],
    };
  },
  computed: {
    breadCrumbList() {
      return this.$store.state.app.breadCrumbList;
    },
  },
  methods: {
    /**
     * 统计在线人数
     */

    handleCollpasedChange(state) {
      this.$emit("on-coll-change", state);
    },
    /**
     * 显示报警信息
     */
    keepFetchWarning() {
      clearInterval(window.scrollWarningTimer);
      // warning/check

      //使内容滚动起来
      let requestReady = 0;
      let scrollContent = () => {
        if (requestReady == 1) {
          return;
        }
        requestReady = 1;
        axios
          .request({
            url: "v1/warning/check",
            method: "get",
            params: {
              token: getToken(),
            },
          })
          .then((data) => {
            let warningContent = "redis实时统计失效，请管理员联系运维处理!";
            let container = document.querySelector("#warningText");
            requestReady = 0;
            if (
              data &&
              data.data &&
              data.data.data &&
              data.data.data.warning == "1"
            ) {
              container.innerHTML = warningContent;
            } else {
              container.innerHTML = "";
            }
          });
      };
      //保存滚动定时器
      // scrollContent();
      // window.scrollWarningTimer = setInterval(scrollContent, 10000);
    },

    getName(id) {
      let name = "";
      this.proxyList &&
        this.proxyList.map((item) => {
          if (item.agentList) {
            return item.agentList.map((agent) => {
              if (agent.id == id) {
                name = agent.name;
              }
            });
          }
        });
      return name;
    },

    // openModal() {
    //   this.proxyList = JSON.parse(sessionStorage.getItem("siteOption")) || [];
    //   axios
    //     .request({
    //       url: "v2/user/online",
    //       method: "get",
    //       params: {
    //         token: getToken(),
    //       },
    //     })
    //     .then(({ data }) => {
    //       this.modal1 = true;
    //       this.list = (data && data.data) || [];
    //       if (data && data.data.length > 0) {
    //         let resual = data.data.reduce(
    //           (total, item) => total + item.onlineUser,
    //           0
    //         );
    //         this.allOnline = resual;
    //       }
    //     });
    // },
  },
  mounted() {
    this.keepFetchWarning();
  },
};
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
