<template>
  <el-dialog
    ref="dialog"
    :visible.sync="detail_visible"
    :modal-append-to-body="false"
    :show-close="false"
    :before-close="close"
    width="80%"
  >
    <div class="content-box">
      <el-row type="flex" justify="center">
        <el-col :span="22" type="flex" align="middle" style="margin-top:30px;">
          <el-col :span="24" class="bg-purple-dark">
            <p style="font-weight: bold;">{{ $t("message.game_details") }}</p>
          </el-col>
          <el-table :data="tableData" :show-header="false" :border="true" stripe>
            <el-table-column prop="key" align="center"></el-table-column>
            <el-table-column prop="value" align="center"></el-table-column>
          </el-table>
        </el-col>
      </el-row>

      <el-row type="flex" justify="center" v-if="game_id === 7027">
        <el-col :span="24" align="middle" class="bg-purple-dark">
          <p style="font-weight: bold;">
            {{ $t('message.game_results') }}
            <span style="display:inline-block;margin:0 40px;">
              {{ $t('message.drawMultiple') }}: {{ info.rate }}
            </span>
            {{ $t('message.bonus') }}: {{ info.reward_money }}
          </p>
        </el-col>
      </el-row>
      <el-row type="flex" justify="center" v-else>
        <el-col :span="22" type="flex" align="middle">
          <el-col :span="24" class="bg-purple-dark">
            <p style="font-weight: bold;">
              {{ $t("message.game_results") }}:
              <i
                v-for="(result, index) in reward_result"
                :key="index"
                style="font-style:normal;font-weight: bold;color: blue;"
              >{{result}}</i>
              <i
                v-show="has_bonus"
                style="font-style:normal;font-weight: bold;color: red;"
              >{{jp_bonus}}</i>
            </p>
          </el-col>
          <el-col :span="24" v-show="game_show_type == 0">
            <!-- 十二生肖、飞禽走兽、黄金大转轮、巅峰篮球、水果转盘、赛马、燃烧吧足球、王者足球、幸运5、水果机、好运射击、猴子爬树、森林舞会 -->
            <div align="center" class="show-style">
              <div v-for="(item, index) in result_show_url" :key="index" class="img_style">
                <img :src="item" />
              </div>
            </div>

            <el-table :data="specific_bet_info" style="width: 100%">
              <el-table-column prop="bet_area_url" align="center" :label="$t('message.bet_area')">
                <template slot-scope="scope">
                  <img :src="scope.row.bet_area_url" />
                </template>
              </el-table-column>
              <el-table-column prop="bet_num" align="center" :label="$t('message.bet_amount')"></el-table-column>
              <el-table-column
                prop="reward_rate"
                align="center"
                :label="$t('message.reward_multiples')"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;">{{(scope.row.reward_rate / 100).toFixed(2)}}</i>
                </template>
              </el-table-column>
              <el-table-column prop="reward_money" align="center" :label="$t('message.bonus')"></el-table-column>
            </el-table>
          </el-col>

          <el-col :span="24" v-show="game_show_type == 1">
            <!-- 连环夺宝、水果传奇、萌宠夺宝 -->
            <el-table :data="specific_bet_info" style="width: 100%">
              <el-table-column type="index" :label="$t('message.no')" align="center" width="50"></el-table-column>
              <el-table-column
                prop="bet_area_url"
                align="center"
                :label="$t('message.elimination_of_elements')"
                min-width="20%"
              >
                <template slot-scope="scope">
                  <el-col v-show="scope.row.bet_area_url">
                    <img :src="scope.row.bet_area_url" style="width: 45px;height:45px;" />
                    <i style="font-style:normal;">x{{scope.row.num}}</i>
                  </el-col>
                </template>
              </el-table-column>
              <el-table-column
                prop="grid"
                align="center"
                :label="$t('message.eliminate_the_panel')"
                min-width="35%"
              >
                <template slot-scope="scope">
                  <el-col :span="24" align="center">
                    <div v-for="(item, rowIndex) in scope.row.grid" :key="rowIndex" align="center" style="display:table;">
                      <div v-for="(item_one, itemIndex) in item" :key="itemIndex" style="float:left;">
                        <img :src="item_one" class="img-grids" />
                      </div>
                    </div>
                  </el-col>
                </template>
              </el-table-column>
              <el-table-column
                prop="bet_num"
                align="center"
                :label="$t('message.bet_amount')"
                min-width="15%"
              ></el-table-column>
              <el-table-column
                prop="reward_rate"
                align="center"
                :label="$t('message.reward_multiples')"
                min-width="15%"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;">{{(scope.row.reward_rate / 100).toFixed(2)}}</i>
                </template>
              </el-table-column>
              <el-table-column
                prop="bet_nums"
                align="center"
                :label="$t('message.injection_number')"
                min-width="15%"
              ></el-table-column>
              <el-table-column
                prop="reward_money"
                align="center"
                :label="$t('message.bonus')"
                min-width="15%"
              ></el-table-column>
            </el-table>
          </el-col>

          <el-col :span="24" v-show="game_show_type == 2">
            <!-- 浮岛历险记 -->
            <el-col :span="24" align="center" style="background:#FFFAFA;">
              <div align="center" style="margin-top:20px;margin-bottom:20px;display:table;">
                <div v-for="(item, rowIndex) in result_grid_url" :key="rowIndex" align="center">
                  <div v-for="(item_one, itemIndex) in item" :key="itemIndex" style="float:left;">
                    <img :src="item_one" />
                  </div>
                </div>
              </div>
            </el-col>

            <el-table :data="specific_bet_info" style="width: 100%">
              <el-table-column type="index" :label="$t('message.no')" align="center" width="50"></el-table-column>
              <el-table-column
                prop="num"
                align="center"
                :label="$t('message.steps')"
                min-width="10%"
              ></el-table-column>
              <el-table-column
                prop="grid"
                align="center"
                :label="$t('message.route')"
                min-width="30%"
              >
                <template slot-scope="scope">
                  <el-col :span="24" align="center">
                    <div v-for="(item, rowIndex) in scope.row.grid" :key="rowIndex" align="center" style="display:table;">
                      <div v-for="(item_one, itemIndex) in item" :key="itemIndex" style="float:left;">
                        <img :src="item_one" class="img-grids" />
                      </div>
                    </div>
                  </el-col>
                </template>
              </el-table-column>
              <el-table-column
                prop="bet_num"
                align="center"
                :label="$t('message.bet_amount')"
                min-width="20%"
              ></el-table-column>
              <el-table-column
                prop="reward_rate"
                align="center"
                :label="$t('message.reward_multiples')"
                min-width="20%"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;">{{(scope.row.reward_rate / 100).toFixed(2)}}</i>
                </template>
              </el-table-column>
              <el-table-column
                prop="reward_money"
                align="center"
                :label="$t('message.bonus')"
                min-width="20%"
              ></el-table-column>
            </el-table>
          </el-col>

          <el-col :span="24" v-show="game_show_type == 3">
            <!-- 火箭之旅 -->
            <el-table :data="[info]" style="width: 100%">
              <el-table-column prop="settle_multi" :label="$t('message.settlePoint')" align="center"></el-table-column>
              <el-table-column prop="bet_num" :label="$t('message.betAmount')" align="center"></el-table-column>
              <el-table-column prop="reward_money" :label="$t('message.bonus')" align="center"></el-table-column>
            </el-table>
          </el-col>
        </el-col>
      </el-row>

      <el-row
        v-show="bonus_game_info"
        type="flex"
        justify="center"
        style="margin-top:10px;min-height: 100px;"
      >
        <el-col :span="22" style="background:white;">
          <el-col :span="24" class="bg-purple-dark" style="font-weight: bold;">
            <p class="el-icon-caret-bottom" style="margin-left:5px;width: 20px;"></p>
            {{ $t("message.game_results") }}: {{ $t("message.bonus_game") }} {{ $t("message.bonus") }}：{{xiao_all_bonus}}
          </el-col>

          <div class="grid-content bg-purple-light">
            <el-col :span="24" v-for="(item, index) in bonus_game_info" :key="index">
              <div class="grid-content">{{item.bonus_type}}:{{item.bonus_num}}</div>
            </el-col>
          </div>
        </el-col>
      </el-row>

      <el-row
        v-for="(item, index) in conin_info"
        :key="index"
        v-show="item.bonus != 0"
        type="flex"
        justify="center"
        style="margin-top:10px;min-height: 100px;"
      >
        <el-col :span="22" style="background:white;">
          <div class="grid-content bg-purple-dark" style="border-radius: 0;font-weight: bold;">
            <p class="el-icon-caret-bottom" style="margin-left:5px;width: 20px;"></p>
            {{ $t("message.game_results") }}: {{conin_type_name[item.type - 1]}} {{ $t("message.bonus") }}：{{item.bonus}}
          </div>

          <div class="grid-content bg-purple-light">
            <el-col :span="24">
              <div
                class="grid-content"
                style="margin-left: 10px"
              >{{conin_type_name[item.type - 1]}}：{{item.bonus}}</div>
            </el-col>
          </div>
        </el-col>
      </el-row>

      <el-row type="flex" justify="center">
        <el-col :span="22" type="flex" align="middle" style="margin-top:30px;margin-bottom:30px;">
          <el-button
            @click.native.prevent="close()"
            style="font-weight: bold;"
          >{{ $t("message.close") }}</el-button>
        </el-col>
      </el-row>
    </div>
  </el-dialog>
</template>

<style>
.el-dialog__header {
  padding: 0;
}
.el-dialog__body {
  padding: 0;
}
.show-style {
  /* transform:scale(1.6);
  -ms-transform:scale(1.6);
  -webkit-transform:scale(1.6);
  -o-transform:scale(1.6);
  -moz-transform:scale(1.6); */
  margin-top: 20px;
  margin-bottom: 20px;
  display: table;
}
.img_style {
  display: inline-block;
  float: left;
}
.img-grids {
  width: 30px;
  height: 30px;
}
.choose_body {
  width: 12%;
  height: 230px;
  position: absolute;
  z-index: 100;
  top: 65px;
  left: 11%;
  border: 2px solid rgb(255, 0, 0);
  color: red;
}
.choose_body_font {
  bottom: 5px;
  position: absolute;
  text-align: center;
  width: 100%;
}
.bg-purple-dark {
  background: #e2e6eb;
}
.content-box {
  width: 100%;
  margin: auto;
  background: rgb(240, 240, 240);
  position: relative;
}
</style>

<script>
import http from '../../assets/js/http'
export default {
  data() {
    return {
      detail_visible: false,
      has_bonus: false,
      jp_bonus: 0,
      game_show_type: 0,
      recordData: {},
      tableData: [{ key: '', value: '' }, { key: '', value: '' }, { key: '', value: '' }, { key: '', value: '' }, { key: '', value: '' }],
      game_id_list1: [7004, 7005, 7001, 7008, 7007, 7009, 7010, 7006, 7019, 7012, 7000, 7014, 7015, 7003, 7002, 7020],
      game_id_list2: [7011, 7017, 7018, 7021, 7023],
      game_id_list3: [7016],
      game_id_list4: [7028],
      reward_result: [],
      result_show_url: [],
      specific_bet_info: [],
      result_grid_url: [],
      total_bet: {},
      game_id: '',
      info: {}
    }
  },
  computed: {},
  methods: {
    timeLabel() {
      let utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    showGame(game_id) {
      if (this.game_id_list1.indexOf(game_id) != -1) {
        this.game_show_type = 0
      } else if (this.game_id_list2.indexOf(game_id) != -1) {
        this.game_show_type = 1
      } else if (this.game_id_list3.indexOf(game_id) != -1) {
        this.game_show_type = 2
      } else if (this.game_id_list4.indexOf(game_id) != -1) {
        this.game_show_type = 3
      }
    },
    open() {
      let loading = this.$loading({})
      setTimeout(() => {
        loading.close()
      }, 3000)
      this.detail_visible = true
      this.recordData = Lockr.get('fruit_detail')
      this.tableData[0].key = this.$t('message.game_time') + this.timeLabel()
      this.tableData[0].value = this.recordData.time
      this.tableData[1].key = this.$t('message.game_number')
      this.tableData[1].value = this.recordData.id
      this.tableData[2].key = this.$t('message.game_name')
      this.tableData[2].value = this.recordData.game_name
      this.tableData[3].key = this.$t('message.detail_total_bets')
      this.tableData[3].value = this.recordData.all_bets
      this.tableData[4].key = this.$t('message.total_bonus')
      this.tableData[4].value = this.recordData.all_bonus

      this.apiGet(this.recordData.dt_url).then(res => {
        this.handelResponse(res, data => {
          this.game_id = data.info.game_id
          if (data.info.hasOwnProperty('jp_bonus')) {
            this.has_bonus = true
            this.jp_bonus = data.info.jp_bonus
          }
          this.showGame(data.info.game_id)
          this.result_grid_url = data.info.result_grid_url
          this.reward_result = data.info.reward_result
          this.result_show_url = data.info.result_show_url
          this.specific_bet_info = data.info.specific_bet_info
          this.total_bet = data.info.total_bet
          this.info = data.info
          if (data.info.game_id === 7027) {
            this.tableData[3].key = this.$t('message.bet1')
            this.tableData[3].value = data.info.bet_num
            this.tableData[4].key = this.$t('message.risk')
            this.tableData[4].value = data.info.risk
            if (this.tableData[5]) {
              this.tableData[5].key = this.$t('message.level')
              this.tableData[5].value = data.info.num
            } else {
              this.tableData.push({ key: this.$t('message.level'), value: data.info.num })
            }
          } else if (data.info.game_id === 7028) {
            this.tableData[4].key = this.$t('message.drawInfo')
            this.tableData[4].value = data.info.bomb_multi
          }
          loading.close()
        })
      })
    },
    close() {
      Lockr.rm('fruit_detail')
      this.detail_visible = false
      this.has_bonus = false
      this.jp_bonus = 0
      this.reward_result.splice(0, this.reward_result.length)
      this.result_show_url.splice(0, this.result_show_url.length)
      this.specific_bet_info.splice(0, this.specific_bet_info.length)
      this.result_grid_url.splice(0, this.result_grid_url.length)
      this.total_bet = {}
      this.game_id = ''
    }
  },
  mixins: [http]
}
</script>


// WEBPACK FOOTER //
// fruit_detail.vue?5b46f494
