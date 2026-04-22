<template>
  <Layout style="height: 100%" class="main">
    <Sider
      hide-trigger
      collapsible
      :width="220"
      :collapsed-width="64"
      v-model="collapsed"
      class="left-sider"
      :style="{ overflow: 'hidden' }"
    >
      <side-menu
        style="width: 220px"
        accordion
        ref="sideMenu"
        :active-name="$route.name"
        :collapsed="collapsed"
        @on-select="turnToPage"
        :menu-list="menuList"
      >
        <!-- 需要放在菜单上面的内容，如Logo，写在side-menu标签内部 -->
        <div class="logo-con">
          <img v-show="!collapsed" :src="maxLogo" key="max-logo" />
          <img v-show="collapsed" :src="minLogo" key="min-logo" />
        </div>
        <!-- <div style="width:85%;margin:10px auto">
          <Select placeholder="选择站点" clearable v-model="site.val" @on-change="setSiteSession">
            <Option v-for="item in site.option" :value="item.id" :key="item.id">{{ item.name }}</Option>
          </Select>
          <Select
            :placeholder="agent.msg"
            clearable
            filterable
            style="margin-top:10px"
            v-model="agent.val"
            @on-change="setAgentSession"
          >
            <Option v-for="item in agent.option" :value="item.id" :key="item.id">{{ item.name }}</Option>
          </Select>
        </div> -->
      </side-menu>
    </Sider>
    <Layout>
      <Header class="header-con">
        <marquee
          ref="marqueeMsg"
          v-if="unreadCount && 0"
          class="marqueeStyle"
        ></marquee>
        <header-bar
          :collapsed="collapsed"
          @on-coll-change="handleCollapsedChange"
        >
          <user :user-avatar="userAvatar" />
          <!-- //:message-unread-count="unreadCount" -->
          <language
            v-if="$config.useI18n"
            @on-lang-change="setLocal"
            style="margin-right: 10px"
            :lang="local"
          />
          <error-store
            v-if="
              $config.plugin['error-store'] &&
              $config.plugin['error-store'].showInHeader
            "
            :has-read="hasReadErrorPage"
            :count="errorCount"
          ></error-store>
          <fullscreen v-model="isFullscreen" style="margin-right: 10px" />
        </header-bar>
      </Header>
      <Content class="main-content-con">
        <Layout class="main-layout-con" style="overflow-x: auto">
          <!-- 面包屑菜单 -->
          <!-- <div class="tag-nav-wrapper">
            <tags-nav
              :value="$route"
              @input="handleClick"
              :list="tagNavList"
              @on-close="handleCloseTag"
            />
          </div>-->
          <Content
            v-if="refresh"
            class="content-wrapper"
            style="min-width: 1480px"
          >
            <template v-if="$route.meta.notKeep">
              <router-view />
            </template>
            <keep-alive v-else>
              <router-view />
            </keep-alive>
            <ABackTop
              :height="100"
              :bottom="80"
              :right="50"
              container=".content-wrapper"
            ></ABackTop>
          </Content>
        </Layout>
      </Content>
    </Layout>
  </Layout>
</template>
<script>
import SideMenu from "./components/side-menu";
import HeaderBar from "./components/header-bar";
import TagsNav from "./components/tags-nav";
import User from "./components/user";
import ABackTop from "./components/a-back-top";
import Fullscreen from "./components/fullscreen";
import Language from "./components/language";
import ErrorStore from "./components/error-store";
import { mapMutations, mapActions, mapGetters } from "vuex";
import { getNewTagList, routeEqual } from "@/libs/util";
import routers from "@/router/routers";
import minLogo from "@/assets/images/logo-min.jpg";
import maxLogo from "@/assets/images/logo.jpg";
import "./main.less";
import {
  //getClassList,
  getLinkageList,
  // getTypeList,
  getSelectGames,
} from "@/api/data";
import axios from "axios";
export default {
  name: "Main",
  components: {
    SideMenu,
    HeaderBar,
    Language,
    TagsNav,
    Fullscreen,
    ErrorStore,
    User,
    ABackTop,
  },
  provide() {
    return {
      reFreshSiteAgentList: this.reFreshSiteAgentList,
      viewAccess: this.viewAccessSuper,
      handleLogOut: this.handleLogOut,
    };
  },
  data() {
    return {
      refresh: true,
      collapsed: false,
      minLogo,
      maxLogo,
      isFullscreen: false,
      site: { val: "", option: [] },
      agent: { val: "", option: [], msg: "请先选择站点" },
    };
  },
  computed: {
    ...mapGetters(["errorCount"]),
    tagNavList() {
      return this.$store.state.app.tagNavList;
    },
    tagRouter() {
      return this.$store.state.app.tagRouter;
    },
    userAvatar() {
      return this.$store.state.user.avatarImgPath;
    },
    cacheList() {
      const list = [
        "ParentView",
        ...(this.tagNavList.length
          ? this.tagNavList
              .filter((item) => !(item.meta && item.meta.notCache))
              .map((item) => item.name)
          : []),
      ];
      return list;
    },
    menuList() {
      return this.$store.getters.menuList;
    },
    local() {
      return this.$store.state.app.local;
    },
    hasReadErrorPage() {
      return this.$store.state.app.hasReadErrorPage;
    },
    unreadCount() {
      return this.$store.state.user.unreadCount;
    },
    viewAccessSuper() {
      return this.$store.state.user.access.indexOf("administrator") > -1;
    },
  },
  methods: {
    ...mapMutations([
      "setBreadCrumb",
      "setTagNavList",
      "addTag",
      "setLocal",
      "setHomeRoute",
      "closeTag",
    ]),
    ...mapActions(["handleLogin", "getUnreadMessageCount", "handleLogOut"]),
    turnToPage(route) {
      let { name, params, query } = {};
      if (typeof route === "string") name = route;
      else {
        name = route.name;
        params = route.params;
        query = route.query;
      }
      if (name.indexOf("isTurnByHref_") > -1) {
        window.open(name.split("_")[1]);
        return;
      }
      this.$router.push({
        name,
        params,
        query,
      });
    },
    handleCollapsedChange(state) {
      this.collapsed = state;
    },
    handleCloseTag(res, type, route) {
      if (type !== "others") {
        if (type === "all") {
          this.turnToPage(this.$config.homeName);
        } else {
          if (routeEqual(this.$route, route)) {
            this.closeTag(route);
          }
        }
      }
      this.setTagNavList(res);
    },
    handleClick(item) {
      this.turnToPage(item);
    },

    setSiteSession(val) {
      this.agent.val = "";
      if (val > 0) {
        sessionStorage.setItem("siteVal", val);
        this.agent.option = [];
        for (const key in this.site.option) {
          if (this.site.option[key].id == val) {
            this.agent.option.push(...this.site.option[key].agentList);
            this.agent.msg = "选择代理";
          }
        }
      }
    },

    setAgentSession(val) {
      if (val != undefined) {
        sessionStorage.setItem("agentVal", val);

        //刷新右边页面
        this.refresh = false;
        setTimeout(() => {
          this.refresh = true;
        }, 1);
      }
    },

    /**
     * 刷新站点代理信息
     */
    reFreshSiteAgentList() {
      getLinkageList().then((res) => {
        window.windowData_getLinkageList = res.data.data;

        if (
          !window.windowData_getLinkageList ||
          window.windowData_getLinkageList.length == 0
        ) {
          return;
        }
        this.site.option = [];

        this.site.option.push(...Object.assign(res.data.data));

        for (let i in res.data.data) {
          if (
            res.data.data[i].agentList &&
            res.data.data[i].agentList.length > 0
          ) {
            for (let j in res.data.data[i].agentList) {
              res.data.data[i].agentList[j].gameIds = undefined;
              res.data.data[i].agentList[j].gameList = undefined;
            }
          }
        }
        sessionStorage.setItem("siteOption", JSON.stringify(res.data.data));

        if (sessionStorage.getItem("siteVal")) {
          this.site.val = Number(sessionStorage.getItem("siteVal"));
          if (sessionStorage.getItem("agentVal")) {
            this.$nextTick(() => {
              this.agent.val = Number(sessionStorage.getItem("agentVal"));
            });
          }
          this.setSiteSession(this.site.val);
        } else {
          this.agent.option = [];
          this.site.val = res.data.data[0].id;
          this.agent.option.push(...res.data.data[0].agentList);
          this.agent.val = res.data.data[0].agentList[0].id;
          sessionStorage.setItem("siteVal", res.data.data[0].id);
          sessionStorage.setItem("agentVal", res.data.data[0].agentList[0].id);
        }

        // getClassList(sessionStorage.getItem("agentVal")).then((res) => {
        //   sessionStorage.setItem("classOption", JSON.stringify(res.data.data));
        // });
        // getTypeList(sessionStorage.getItem("agentVal")).then((res) => {
        //   sessionStorage.setItem("typeOption", JSON.stringify(res.data.data));
        // });
        // getSelectGames(sessionStorage.getItem("agentVal")).then((res) => {
        //   sessionStorage.setItem("gameOption", JSON.stringify(res.data.data));
        // });
      });
    },
  },
  watch: {
    $route(newRoute) {
      const { name, query, params, meta } = newRoute;
      this.addTag({
        route: { name, query, params, meta },
        type: "push",
      });
      this.setBreadCrumb(newRoute);
      this.setTagNavList(getNewTagList(this.tagNavList, newRoute));
      this.$refs.sideMenu.updateOpenName(newRoute.name);
    },
  },
  mounted() {
    /**
     *
     * 动态修改marquee的宽度
     *
     */

    window.test = this;

    let marqueeReference = document.querySelector(".ivu-layout-header");
    let that = this;
    let reloadMarqueeWidth = (entries) => {
      if (that.$refs.marqueeMsg) {
        that.$refs.marqueeMsg.style = `width:${
          document.querySelector(".ivu-layout-header").clientWidth -
          document.querySelector(".custom-bread-crumb").clientWidth -
          370
        }px`;
      }
    };
    const resizeObserver = new ResizeObserver(reloadMarqueeWidth);

    resizeObserver.observe(marqueeReference);

    if (!window.isaddEventreloadMarqueeWidth) {
      window.isaddEventreloadMarqueeWidth = true;
      window.addEventListener("resize", reloadMarqueeWidth);
    }

    getLinkageList().then((res) => {
      window.windowData_getLinkageList = res.data.data;

      if (
        !window.windowData_getLinkageList ||
        window.windowData_getLinkageList.length == 0
      ) {
        return;
      }
      //站点赋值
      this.site.option.push(...Object.assign(res.data.data));
      sessionStorage.setItem("siteOption", JSON.stringify(res.data.data));

      for (let i in res.data.data) {
        if (
          res.data.data[i].agentList &&
          res.data.data[i].agentList.length > 0
        ) {
          for (let j in res.data.data[i].agentList) {
            res.data.data[i].agentList[j].gameIds = undefined;
            res.data.data[i].agentList[j].gameList = undefined;
          }
        }
      }

      if (
        sessionStorage.getItem("siteVal") &&
        this.site.option.find(
          (siteitem) => siteitem.id == Number(sessionStorage.getItem("siteVal"))
        )
      ) {
        this.site.val = Number(sessionStorage.getItem("siteVal"));
        if (sessionStorage.getItem("agentVal")) {
          this.$nextTick(() => {
            this.agent.val = Number(sessionStorage.getItem("agentVal"));
          });
        }
        this.setSiteSession(this.site.val);
      } else {
        this.site.val = res.data.data[0].id;
        this.agent.option.push(...res.data.data[0].agentList);
        this.agent.val = res.data.data[0].agentList[0].id;
        sessionStorage.setItem("siteVal", res.data.data[0].id);
        sessionStorage.setItem("agentVal", res.data.data[0].agentList[0].id);
      }
      // getClassList(sessionStorage.getItem("agentVal")).then((res) => {
      //   sessionStorage.setItem("classOption", JSON.stringify(res.data.data));
      // });
      // getTypeList(sessionStorage.getItem("agentVal")).then((res) => {
      //   sessionStorage.setItem("typeOption", JSON.stringify(res.data.data));
      // });
      // getSelectGames(sessionStorage.getItem("agentVal")).then((res) => {
      //   sessionStorage.setItem("gameOption", JSON.stringify(res.data.data));
      // });

      /**
       * @description 初始化设置面包屑导航和标签导航
       */
      this.setTagNavList();
      this.setHomeRoute(routers);
      const { name, params, query, meta } = this.$route;
      this.addTag({
        route: { name, params, query, meta },
      });
      this.setBreadCrumb(this.$route);

      // 如果当前打开页面不在标签栏中，跳到homeName页
      if (!this.tagNavList.find((item) => item.name === this.$route.name)) {
        this.$router.push({
          name: this.$config.homeName,
        });
      }
    });
  },
};
</script>
<style>
.ivu-radio-group-button.ivu-radio-group-small .ivu-radio-wrapper {
  overflow: hidden;
  border-left: 1px solid #dcdee2;
}
div.ivu-select-dropdown {
  z-index: 901;
}
.marqueeStyle {
  max-width: 800px;
  position: fixed;
  right: 250px;
  height: 42px;
  margin-right: 10px;
  font-size: 16px;
  color: red;
}
body .ivu-badge-count {
  background: none;
}
</style>
