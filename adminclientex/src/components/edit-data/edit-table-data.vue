<template>
  <div>
    <div
      class="listData"
      style="position: relative"
      v-for="(item, num) in listData"
      :key="item.key"
    >
      <div style="width: 340px; overflow: hidden; min-height: 43px">
        <!-- 判断为id，该项只读 -->
        <div v-if="(item.key == 'id') | (item.key == 'account')">
          <i-input
            :maxlength="50"
            v-model="item.value"
            readonly
            :disabled="item.isdisabled"
          >
            <span slot="prepend">{{ item.title }}</span>
          </i-input>
        </div>

        <!-- 判断为状态，该项单选 -->
        <div v-else-if="item.key == 'state'" class="edit-checkbox">
          <label>状态</label>
          <RadioGroup v-model="item.value" @on-change="change">
            <Radio label="1">正常</Radio>
            <Radio label="0">维护</Radio>
          </RadioGroup>
        </div>

        <!-- 判断为游戏列表，该项分类多选 -->
        <div v-else-if="item.key == 'gameIds'" class="edit-checkbox">
          <label>代理游戏</label>
          <div class="gameList_edit">
            <div v-for="(items, index) in gameIds.option" :key="items.id">
              <div>
                <Tabs value="name1">
                  <TabPane :label="items.name" name="name1"></TabPane>
                </Tabs>
                <div
                  style="
                    margin-top: -42px;
                    position: relative;
                    z-index: 2;
                    margin-left: 80px;
                  "
                >
                  <Button
                    @click="handleCheckAll(items.id - 1, num, true)"
                    size="small"
                    type="primary"
                    style="margin-right: 10px"
                    >全选</Button
                  >
                  <Button
                    @click="handleCheckAll(items.id - 1, num)"
                    size="small"
                    type="primary"
                    >取消全选</Button
                  >
                </div>
              </div>
              <CheckboxGroup
                style="margin-top: 10px"
                v-if="gameIds.value[index].id == items.id"
                v-model="gameIds.value[index].val"
                @on-change="boxChange(num)"
              >
                <Checkbox
                  v-for="subitems in items.gameList"
                  :key="subitems.id"
                  :label="subitems.id"
                  >{{ subitems.name }}</Checkbox
                >
              </CheckboxGroup>
              <Divider style="margin: 6px 0" />
            </div>
          </div>
        </div>

        <!-- 判断为上级代理，该项为下拉菜单，并绑定代理id -->
        <div v-else-if="item.key == 'upperLevel'" class="edit-checkbox">
          <label>{{ item.title }}</label>
          <Select v-model="item.value" @on-change="change" style="width: 230px">
            <Option
              v-for="items in superior"
              :value="items.id"
              :key="items.id"
              >{{ items.name }}</Option
            >
          </Select>
        </div>

        <!-- 判断为站点，该项为下拉菜单，并绑定站点id -->
        <!-- <div v-else-if="item.key=='webId'" class="edit-checkbox">
        <label>{{ item.title }}</label>
        <Select disabled v-model="item.value" @on-change="change" style="width:230px;">
          <Option v-for="items in webName" :value="items.id" :key="items.id">{{ items.name }}</Option>
        </Select>
        </div>-->

        <i-input
          v-else-if="item.key == 'webId'"
          v-model="webName.find((x) => x.id == item.value).name"
          disabled
        >
          <span slot="prepend">{{ item.title }}</span>
        </i-input>

        <!-- 判断生效时间，为永久将变为只读 -->
        <div v-else-if="item.key == 'startTime' || item.key == 'endTime'">
          <div
            v-if="row.isPermanent != undefined && false"
            style="top: -7px"
            class="edit-checkbox"
          >
            <label>{{ item.title }}</label>
            <i-input
              v-if="isPermanent == 1"
              value="永久"
              style="width: 230px; margin-top: -1px"
              readonly
            ></i-input>

            <DatePicker
              v-if="item.key == 'startTime' && !isPermanent"
              type="datetime"
              v-model="item.value"
              style="width: 230px"
              @on-change="changeStartTime"
            ></DatePicker>

            <DatePicker
              v-if="item.key == 'endTime' && !isPermanent"
              type="datetime"
              v-model="item.value"
              style="width: 230px"
              @on-change="changeEndTime"
            ></DatePicker>

            <div
              v-if="item.key == 'startTime'"
              style="margin-top: 8px; top: -2px"
              class="edit-checkbox"
            >
              <label>是否永久</label>
              <RadioGroup v-model="isPermanent" @on-change="changeTime">
                <Radio :label="1">永久</Radio>
                <Radio :label="0">非永久</Radio>
              </RadioGroup>
            </div>
          </div>
        </div>

        <!-- 判断为消息类型，该项为下拉菜单 -->
        <div v-else-if="item.key == 'msgType'" class="edit-checkbox">
          <label>{{ item.title }}</label>
          <Select v-model="item.value" @on-change="change" style="width: 230px">
            <Option :value="1">活动消息</Option>
            <Option :value="2">维护公告</Option>
          </Select>
        </div>

        <!-- 判断是否为服务器配置 -->
        <div
          v-else-if="
            item.key == 'mysqlConf' ||
            item.key == 'redisConf' ||
            item.key == 'etcdConf'
          "
          class="edit-checkbox"
          style="margin-top: 10px"
        >
          <label style="margin-bottom: 5px">{{ item.title }}</label>
          <i-input
            type="textarea"
            :rows="5"
            v-model="item.value"
            :placeholder="'请输入' + item.lable"
            @on-blur="change"
          ></i-input>
        </div>

        <!-- 判断为操作与接收点，该项隐藏 -->
        <div
          v-else-if="item.key == 'handle' || item.key == 'receive'"
          style="display: none"
        ></div>

        <!-- 判断常规项，使用输入框 -->

        <i-input
          v-else
          :maxlength="item.title == '序号' ? 20 : 50"
          v-model="item.value"
          :disabled="item.isdisabled"
          @on-blur="change"
        >
          <span slot="prepend">{{ item.title }}</span>
        </i-input>
      </div>

      <div style="width: 10px; position: absolute; right: -20px; top: 0px">
        <Badge v-if="item.showRedPoint">
          <Icon
            type="ios-medical"
            title="必填项"
            style="line-height: 0px"
            slot="count"
            color="#ed4014"
            size="12"
          />
        </Badge>
      </div>
    </div>
  </div>
</template>
<script>
import { getSelectClassGames, getSelectAgent } from "@/api/data";
import { getDate, forEach } from "@/libs/tools";
export default {
  name: "game-detial",
  props: {
    row: Object,
    columns: Array,
  },
  data() {
    return {
      listData: [],
      gameSelect: [],
      indeterminate: true,
      checkAll: [],
      gameIds: {
        value: [],
        option: [],
      },
      gameVal: [],
      superior: [],
      webName: [],
      isPermanent: null,
    };
  },
  methods: {
    // 绑定分类多选/全选的菜单和值。
    handleCheckAll(tid, num, iscancel) {
      if (this.indeterminate) {
        this.checkAll[tid] = false;
      } else {
        this.checkAll[tid] = !this.checkAll[tid];
      }
      this.indeterminate = false;
      if (iscancel) {
        this.gameIds.value[tid].val = Object.assign(
          this.gameIds.option[tid].gameList.map((item) => {
            return item.id;
          })
        );
        this.boxChange(num);
      } else {
        this.gameIds.value[tid].val = [];
        this.boxChange(num);
      }
    },

    // 在游戏列表选项产生变化时，向上级页面返回值
    boxChange(num) {
      this.listData[num].value = JSON.stringify(this.gameIds.value);

      const data = this.listData;
      if (this.isPermanent != null) {
        // data.push({ key: "isPermanent", value: this.isPermanent });
      }
      this.$emit("sendEditData", data);
    },

    // 在输入框失去焦点时，向上级页面返回值
    change(event) {
      let isValid = true;
      this.listData.map((item) => {
        if (item.showRedPoint && item.value === "") {
          this.$emit("sendError", item.title + "不能为空");
          isValid = false;
        }
      });

      if (isValid) {
        let data = this.listData;
        this.$emit("sendEditData", data);
      }
    },
    changeTime(event) {
      let data = this.listData;
      if (this.isPermanent != null) {
        // data.push({
        //   key: "isPermanent",
        //   title: "是否永久",
        //   value: event
        // });
      }
      if (event) {
        for (const key in data) {
          if (
            data.hasOwnProperty(key) &&
            (data[key].key == "startTime" || data[key].key == "endTime")
          ) {
            data[key].value = "";
          }
        }
      } else {
        for (const key in data) {
          if (data[key].key == "startTime") {
            data[key].value = getDate(new Date().setHours(0, 0, 0));
          }
          if (data[key].key == "endTime") {
            data[key].value = getDate(
              new Date(new Date().setMonth(new Date().getMonth() + 1)).setHours(
                0,
                0,
                0
              )
            );
          }
        }
      }

      this.$emit("sendEditData", data);
    },
    changeStartTime(event) {
      let data = this.listData.map((item) => {
        if (item.key == "startTime") {
          item.value = getDate(event);
          return item;
        } else {
          return item;
        }
      });

      this.$emit("sendEditData", data);
    },
    changeEndTime(event) {
      let data = this.listData.map((item) => {
        if (item.key == "endTime") {
          item.value = getDate(event);
          return item;
        } else {
          return item;
        }
      });

      this.$emit("sendEditData", data);
    },
  },
  mounted() {
    let data2 = [];
    // 分解定义数据列表
    this.listData = this.columns.map((item) => {
      return {
        key: item.key,
        title: item.title,
        value: "",
        showRedPoint: item.showRedPoint,
        isdisabled: item.isdisabled,
      };
    });
    for (let keys in this.row) {
      data2.push({ key: keys, value: this.row[keys] });
    }
    // 给定义的数据列表进行循环赋值
    for (let j = 0; j < data2.length; j++) {
      data2[j].value =
        data2[j].key.indexOf("Time") != -1
          ? getDate(data2[j].value > 15000 ? data2[j].value * 1000 : "")
          : data2[j].value;

      for (let i = 0; i < this.listData.length; i++) {
        if (data2[j].key == this.listData[i].key) {
          this.listData[i].value = data2[j].value;
        }

        if (this.listData[i].key == "gameList") {
          this.listData[i].key = "gameIds";
          this.gameIds.value = this.row.gameList.map((item) => {
            return { id: item.id, val: item.val };
          });
        }

        if (this.listData[i].key == "superior") {
          this.listData[i].key = "upperLevel";
        }

        if (this.listData[i].key == "webName") {
          this.listData[i].key = "webId";
        }

        if (this.listData[i].key == "state") {
          this.listData[i].value = this.listData[i].value.toString();
        }
      }
    }
    if (this.row.isPermanent != null) {
      this.isPermanent = this.row.isPermanent ? 1 : 0;
    }

    // 获取游戏分级列表内容，并赋值
    // getSelectClassGames().then((res) => {
    //   res.data.data[0].gameList = res.data.data[0].gameList.filter((x) => {
    //     if (
    //       ["炸金花", "二人麻将", "捕鱼大作战", "血流成河", "血战到底"].includes(
    //         x.name
    //       )
    //     ) {
    //       return false;
    //     } else {
    //       return true;
    //     }
    //   });
    //   this.gameIds.option.push(...Object.assign(res.data.data));
    //   this.checkAll.push(false);
    // });
    // 判断上级代理，并对选项进行赋值
    if (Object.keys(this.row).indexOf("superior") != -1) {
      getSelectAgent().then((res) => {
        this.superior = [];
        res.data.data.forEach((item) => {
          if (item.id != this.row.id) {
            this.superior.push(item);
          }
        });
      });
    }
    // 判断站点名称，并对选项进行赋值
    if (sessionStorage.getItem("siteOption")) {
      this.webName.push(
        ...Object.assign(JSON.parse(sessionStorage.getItem("siteOption")))
      );
    }
  },
};
</script>

<style>
.listData {
  display: inline-block;
  width: 340px;
  margin: 5px 10px;
}
.listData .ivu-input-group-prepend > span {
  min-width: 80px;
  display: inline-block;
}

.edit-checkbox > label {
  width: 95px;
  display: inline-block;
  margin-right: 10px;
  text-align: center;
  padding: 4px 7px;
  line-height: 22px;
  color: #515a6e;
  background-color: #f8f8f9;
  border: 1px solid #dcdee2;
  border-radius: 4px;
}

.gameList_edit {
  display: inline-block;
  vertical-align: middle;
  width: 229px;
}
.ivu-table-cell {
  text-overflow: clip;
}
</style>
