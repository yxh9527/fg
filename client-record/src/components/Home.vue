<template>
  <div class="home-main">
    <div class="main-header">
      <el-col class="header-info">
        <!-- el-col style="width:180px;margin-left: 12px;"><img src="../assets/images/logo.png" style="margin-top: 4px;"></el-col -->
        <el-col align="center" style="width:160px;margin-left: 50px;margin-top: 6px;">
          <el-button
            type="text"
            v-bind:class="{btn_blue: btn_record, btn_white: btn_account}"
            @click="clickRecord"
          >{{ $t("message.record") }}</el-button>
        </el-col>
        <el-col align="center" style="width:120px;margin-top: 6px;">
          <el-button
            type="text"
            v-bind:class="{btn_blue: btn_account, btn_white: btn_record}"
            @click="clickAccount"
          >{{ $t("message.account") }}</el-button>
        </el-col>
        <el-col v-if="history_url" align="center" style="width:120px;margin-top: 6px;">
          <el-button
            type="text"
            class="btn-history"
            @click="clickHistory"
          >{{ $t("message.history") }}</el-button>
        </el-col>
        <el-col class="user-info-style">
          <!--div>{{ $t("message.user_name") }}：{{user_name}}</div-->
          <div style="font-size: 15px;">{{ $t("message.user_balance") }}：{{chips}}</div>
        </el-col>
        <el-col v-if="is_app" class="close-style">
          <el-button class="close-button-style" @click="closeWindow">x</el-button>
        </el-col>
      </el-col>

      <el-col>
        <el-form :inline="true" :model="query_data" class="header-form">
          <el-form-item
            :label="$t('message.game_type')"
            class="form-item-style"
            style="margin-left: 5px;"
          >
            <el-select
              size="small"
              style="width: 100px;"
              v-model="query_data.game_type"
              :filterable="!is_app"
              @change="changeType($event)"
            >
              <el-option
                v-for="item in game_type_list"
                :key="item.id"
                :label="item.name"
                :value="item.gtid"
              ></el-option>
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('message.game_name')" class="form-item-style">
            <el-select
              size="small"
              style="width: 150px;"
              v-model="query_data.game_name"
              :filterable="!is_app"
              @change="changeName($event)"
            >
              <el-option v-for="item in game_name_list" :key="item.game_id" :label="item.name" :value="item.game_id"></el-option>
            </el-select>
          </el-form-item>

          <el-form-item
            :label="$t('message.game_page')"
            class="form-item-style"
            v-show="query_data.game_type_id != 20 || query_data.query_type == 0"
          >
            <el-select
              size="small"
              style="width: 80px;"
              v-model="query_data.game_size"
              :filterable="!is_app"
              @change="changeSize($event)"
            >
              <el-option v-for="item in game_size_list" :key="item.value" :label="item.label" :value="item.value"></el-option>
            </el-select>
          </el-form-item>

          <el-form-item
            :label="$t('message.bill_type')"
            class="form-item-style"
            v-show="query_data.game_type_id == 20 && query_data.query_type == 1"
          >
            <el-select
              size="small"
              style="width: 100px;"
              v-model="query_data.bill_type"
              :filterable="!is_app"
              @change="changeBillType($event)"
            >
              <el-option v-for="item in type_arr" :key="item.id" :label="item.name" :value="item.type"></el-option>
            </el-select>
          </el-form-item>

          <el-form-item :label="$t('message.query_time') + timeLabel()" class="form-item-style">
            <el-select
              size="small"
              style="width: 120px;"
              v-model="query_data.game_time_label"
              :filterable="!is_app"
              @change="changeDate($event)"
            >
              <el-option v-for="item in game_time_list" :key="item.value" :label="item.label" :value="item.value"></el-option>
            </el-select>
          </el-form-item>

          <el-form-item class="form-item-style">
            <el-button type="primary" @click="clickBtn">{{ $t("message.query") }}</el-button>
          </el-form-item>
        </el-form>
      </el-col>
    </div>

    <div class="main-cnt">
      <el-col :span="24">
        <router-view v-if="is_router_alive"></router-view>
      </el-col>
    </div>
    <slot_detail ref="slot_detail" :s_info="s_info"></slot_detail>
    <chess_detail ref="chess_detail" :s_info="s_info"></chess_detail>
    <fruit_detail ref="fruit_detail" :s_info="s_info"></fruit_detail>
    <hunting_detail ref="hunting_detail" :s_info="s_info"></hunting_detail>
  </div>
</template>
<style>
.home-main {
  width: 100%;
  height: 100%;
  position: absolute;
  min-width: 950px;
}
.main-header {
  position: absolute;
  width: 100%;
  height: 160px;
  left: 0;
  top: 0;
  z-index: 1500;
  padding: 0;
}
.header-info {
  background: #182234;
  height: 60px;
  min-width: 920px;
}
.header-form {
  background: #3a465a;
  width: 100%;
  height: 60px;
  position: absolute;
  min-width: 920px;
}
.user-info-style {
  float: right;
  font-size: 14px;
  margin-top: 25px;
  font-style: normal;
  color: #ffffff;
  width: 200px;
  margin-right: 10px;
}
.close-style {
  position: fixed;
  width: 40px;
  height: 21px;
  right: 5px;
  top: 10px;
  z-index: 1500;
  padding: 0;
}
.close-button-style {
  width: 40px;
  height: 21px;
  padding: 0;
  background: #1f2d3d;
  color: #ffffff;
  font-weight: bold;
  font-size: 14px;
  border-style: none;
}
.btn_white {
  font-size: 20px;
  font-weight: bold;
  color: #ffffff;
}
.btn_blue {
  font-size: 20px;
  font-weight: bold;
  color: #5daafa;
}
.btn-history {
  font-size: 20px;
  font-weight: bold;
  color: #ffffff;
}
.form-item-style {
  /* margin-left:10px; */
  margin-top: 10px;
}
.el-form-item__label {
  color: #ffffff;
}
.main-cnt {
  padding: 0;
  top: 120px;
  width: 100%;
  height: calc(100% - 120px);
  position: absolute;
  overflow: scroll;
  overflow-x: hidden;
  min-width: 920px;
}
.main-footer {
  text-align: center;
  width: 100%;
  height: 60px;
  bottom: 0;
  position: fixed;
}
</style>
<script>
import slot_detail from './Common/slot_detail.vue'
import chess_detail from './Common/chess_detail.vue'
import fruit_detail from './Common/fruit_detail.vue'
import hunting_detail from './Common/hunting_detail.vue'
import http from '../assets/js/http'
export default {
  data() {
    return {
      num: 1,
      is_app: true,
      btn_record: true,
      btn_account: false,
      is_router_alive: true,
      is_bill_type: false,
      user_name: '',
      chips: '',
      game_data: {},
      query_data: {
        query_type: 0,
        game_type: '',
        game_type_id: 0,
        game_name: '',
        game_id: 0,
        game_size: 20,
        bill_type: '',
        bill_type_id: 0,
        game_time_id: 1,
        game_time_label: '',
        game_time: []
      },
      game_type_list: [],
      game_name_list: [],
      game_size_list: [{ value: 20, label: '20' }, { value: 50, label: '50' }, { value: 100, label: '100' }],
      game_time_list: [
        { value: 1, label: '' },
        { value: 2, label: '' },
        { value: 3, label: '' },
        { value: 4, label: '' },
        { value: 5, label: '' },
        { value: 6, label: '' }
      ],
      type_arr: [],
      accouct_chess: [],
      game_id_list: [
        6506,
        6667,
        6006,
        6501,
        6003,
        6008,
        6010,
        6011,
        6013,
        6507,
        6504,
        6014,
        6012,
        6502,
        6304,
        6302,
        6702,
        6703,
        6701,
        6508,
        6509,
        6015,
        6016,
        6017,
        6704
      ],
      game_log_url: '',
      history_url: ''
    }
  },
  components: {
    slot_detail,
    chess_detail,
    fruit_detail,
    hunting_detail
  },
  methods: {
    timeLabel() {
      let utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    initQueryTime() {
      let utc = Lockr.get('timezone')
      if (utc == -4) {
        this.query_data.game_time = [
          moment()
            .utcOffset(-4)
            .startOf('day')
            .format('YYYY-MM-DD HH:mm:ss'),
          moment()
            .utcOffset(-4)
            .format('YYYY-MM-DD HH:mm:ss')
        ]
      } else if (utc == 8) {
        this.query_data.game_time = [
          moment()
            .startOf('day')
            .format('YYYY-MM-DD HH:mm:ss'),
          moment().format('YYYY-MM-DD HH:mm:ss')
        ]
      }
    },
    checkDevice() {
      if (document.cookie && document.cookie.toLowerCase().indexOf('hide_exit_btn=y') >= 0) {
        // 隱藏退出按鈕
        this.is_app = false
      } else {
        var user_agent = navigator.userAgent.toLowerCase()
        // if (user_agent.indexOf('browser_type/android_app') == -1) {
        //   this.is_app = false
        // }
        let isMobile = /Android|webOS|iPhone|iPod|BlackBerry/i.test(navigator.userAgent)
        if (isMobile) {
          this.is_app = true
        } else {
          this.is_app = false
        }
      }
    },
    reload() {
      this.is_router_alive = false
      this.$nextTick(() => (this.is_router_alive = true))
    },
    clickRecord() {
      this.num = 1
      if (this.query_data.game_type_id == 30) {
        // this.game_name_list = this.game_data.list[30]
        // this.query_data.game_name = this.game_name_list[0].name
      }
      this.btn_record = true
      this.btn_account = false
      this.query_data.query_type = 0
      Lockr.set('query_data', this.query_data)
      this.changeRouter(this.query_data.game_type_id)
    },
    clickAccount() {
      this.num = 2
      if (this.query_data.game_type_id == 30) {
        this.game_name_list = this.accouct_chess
        // this.query_data.game_name = this.game_name_list[0].name
        // this.query_data.game_id = this.game_name_list[0].game_id
      }
      this.btn_record = false
      this.btn_account = true
      this.query_data.query_type = 1
      Lockr.set('query_data', this.query_data)
      this.changeRouter(this.query_data.game_type_id)
    },
    clickHistory() {
      window.open(this.history_url, '_blank')
    },
    closeWindow() {
      Lockr.rm('token')
      Lockr.rm('game_id')
      Lockr.rm('language')
      Lockr.rm('query_data')
      // window.opener = null
      window.location.href = 'about:blank'
      window.close()
    },
    changeType(key) {
      this.query_data.game_type_id = key

      if (this.num == 1 && key == 30) {
        // this.game_name_list = this.game_data.list[key]
        // this.query_data.game_name = this.game_name_list[0].name
        // this.query_data.game_id = this.game_name_list[0].game_id
      } else if (this.num == 2 && key == 30) {
        // this.game_name_list = this.accouct_chess
        // this.query_data.game_name = this.game_name_list[0].name
        // this.query_data.game_id = this.game_name_list[0].game_id
      } else {
      }
      this.game_name_list = this.game_data.list[key]
      this.query_data.game_name = this.game_name_list[0].name
      this.query_data.game_id = this.game_name_list[0].game_id
    },
    changeName(id) {
      this.query_data.game_id = id
      // 捕猎需要更新账单类型
      if (this.query_data.game_type_id == 20 && this.query_data.query_type == 1) {
        this.is_bill_type = true
        this.query_data.bill_type_id = 0
        Lockr.set('query_data', this.query_data)
        this.reload()
      }
    },
    changeSize(size) {
      this.query_data.game_size = size
    },
    changeBillType(type) {
      this.query_data.bill_type_id = type
      for (var item of this.type_arr) {
        if (type == item.id) {
          this.query_data.bill_type = item.name
        }
      }
    },
    changeDate(date) {
      let utc = Lockr.get('timezone')
      this.query_data.game_time_id = date
      switch (date) {
        case 1:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .startOf('day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .startOf('day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        case 2:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .subtract(1, 'day')
                .startOf('day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .subtract(1, 'day')
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .subtract(1, 'day')
                .startOf('day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .subtract(1, 'day')
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        case 3:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .subtract(3, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .subtract(3, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        case 4:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .subtract(6, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .subtract(6, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        case 5:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .subtract(29, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .subtract(29, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        case 6:
          if (utc == -4) {
            this.query_data.game_time = [
              moment()
                .utcOffset(-4)
                .subtract(45, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .utcOffset(-4)
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          } else if (utc == 8) {
            this.query_data.game_time = [
              moment()
                .subtract(45, 'day')
                .format('YYYY-MM-DD HH:mm:ss'),
              moment()
                .endOf('day')
                .format('YYYY-MM-DD HH:mm:ss')
            ]
          }
          break
        default:
          console.log('error!')
      }
    },
    clickBtn() {
      this.is_bill_type = false
      var old_data = Lockr.get('query_data')
      this.changeDate(this.query_data.game_time_id)
      Lockr.set('query_data', this.query_data)
      if (old_data.game_type_id == this.query_data.game_type_id) {
        this.reload()
      } else {
        this.changeRouter(this.query_data.game_type_id)
      }
    },
    changeRouter(type) {
      if (this.query_data.query_type == 0) {
        switch (type) {
          case '10':
            this.$router.replace('/slot_record')
            break
          case '20':
            this.$router.replace('/hunting_record')
            break
          case '30':
            this.$router.replace('/chess_record')
            break
          case '40':
            this.$router.replace('/fruit_record')
            break
          default:
            console.log('error!')
        }
      } else {
        switch (type) {
          case '10':
            this.$router.replace('/slot_account')
            break
          case '20':
            this.$router.replace('/hunting_account')
            break
          case '30':
            this.$router.replace('/chess_account')
            break
          case '40':
            this.$router.replace('/fruit_account')
            break
          default:
            console.log('error!')
        }
      }
    },
    initI18n() {
      this.query_data.game_time_label = this.$t('message.today')
      this.game_time_list[0].label = this.$t('message.today')
      this.game_time_list[1].label = this.$t('message.yesterday')
      this.game_time_list[2].label = this.$t('message.lately_3_day')
      this.game_time_list[3].label = this.$t('message.lately_a_week')
      this.game_time_list[4].label = this.$t('message.lately_a_month')
      this.game_time_list[5].label = this.$t('message.lately_45_day')
    },
    initGameData() {
      if (this.game_data) {
        this.game_type_list = this.game_data.type
        if (this.game_data.current_gtid == 0) {
          this.query_data.game_type = this.game_type_list[0].name
          this.query_data.game_type_id = this.game_type_list[0].gtid
        } else {
          this.query_data.game_type_id = this.game_data.current_gtid
          for (var item of this.game_type_list) {
            if (this.query_data.game_type_id == item.gtid) {
              this.query_data.game_type = item.name
            }
          }
        }

        this.query_data.game_id = Lockr.get('game_id')
        this.game_name_list = this.game_data.list[this.query_data.game_type_id]
        for (var itemList of this.game_name_list) {
          if (itemList.game_id == this.query_data.game_id) {
            this.query_data.game_name = itemList.name
          }
        }
      }
      Lockr.set('query_data', this.query_data)
      this.changeRouter(this.query_data.game_type_id)
    }
  },
  created() {
    this.checkDevice()
    this.initQueryTime()
    this.$i18n.locale = Lockr.get('language')
    document.title = this.$t('message.game_details')
    this.initI18n()
    bus.$on('game_size', size => {
      this.query_data.game_size = size
    })
    bus.$on('slot_detail', val => {
      this.$refs.slot_detail.open()
    })
    bus.$on('chess_detail', val => {
      this.$refs.chess_detail.open()
    })
    bus.$on('fruit_detail', val => {
      this.$refs.fruit_detail.open()
    })
    bus.$on('hunting_detail', val => {
      this.$refs.hunting_detail.open()
    })
    bus.$on('type_arr', val => {
      this.type_arr = val
      // 如果已经存在则不需要初始化
      if (!this.query_data.bill_type || this.is_bill_type) {
        this.query_data.bill_type = val[0].name
        this.query_data.bill_type_id = val[0].id
      }
    })
    this.apiGet(_g.apiUrl() + 'player_acc' + '?language=' + Lockr.get('language')).then(res => {
      this.handelResponse(res, data => {
        this.user_name = data.user_name
        this.chips = data.chips
        Lockr.set('user_id', data.user_id)
        this.history_url = data.history_url
      })
    })
    this.game_log_url = _g.apiUrl() + 'gamelog/gameTypeAndList'
    this.query_data.game_id = Lockr.get('game_id')
    Lockr.set('query_data', this.query_data)
    if (this.query_data.game_id) {
      this.game_log_url += '?game_id=' + this.query_data.game_id + '&language=' + Lockr.get('language')
    } else {
      this.game_log_url += '?language=' + Lockr.get('language')
    }
    this.apiGet(this.game_log_url).then(res => {
      this.handelResponse(res, data => {
        this.game_data = data.data
        this.accouct_chess = JSON.parse(JSON.stringify(this.game_data['list']['30']))
        // 屏蔽未上线棋牌端口
        // for (var i = this.game_data['list']['30'].length - 1; i >= 0; i--) {
        //   var tmp = this.game_data['list']['30'][i]
        //   if (this.game_id_list.indexOf(tmp.game_id) == -1) {
        //     this.game_data['list']['30'].splice(i, 1)
        //   }
        // }
        this.initGameData()
      })
    })
    var oMeta = document.createElement('meta')
    oMeta.content = 'width=device-width, initial-scale=0.6, user-scalable=no'
    oMeta.name = 'viewport'
    document.getElementsByTagName('head')[0].appendChild(oMeta)
  },
  watch: {
    sizeChange: function(val) {
      console.log('ok')
    }
  },
  mixins: [http]
}
</script>



// WEBPACK FOOTER //
// Home.vue?47972a88
