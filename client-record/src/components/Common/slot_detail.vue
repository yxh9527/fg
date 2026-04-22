<template>
  <el-dialog
    ref="dialog"
    v-loading="loading"
    :visible.sync="detail_visible"
    :modal-append-to-body="false"
    :show-close="false"
    :before-close="close"
    width="60%"
  >
    <div class="content-box">
      <el-row type="flex" justify="center" class="J_Result">
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

      <!-- type_id=3奖励游戏，如麻将来了 -->
      <el-row type="flex" justify="center">
        <el-col
          :span="22"
          type="flex"
          v-if="data.info && data.info.game_id === 3331 && data.info && data.info.type_id === 3"
        >
          <el-col :span="24" class="bg-purple-dark" style="font-weight: bold;">
            <p class="el-icon-caret-bottom" style="margin-left:5px;width: 20px;"></p>
            {{ $t("message.game_results") }}: {{data.info.type}} {{ $t("message.bonus") }}：{{data.info.sum_bonus ? Number(data.info.sum_bonus).toFixed(2) : 0}}
          </el-col>
          <el-table :data="data.info.little_game_list" style="width: 100%">
            <el-table-column
              prop="index"
              align="center"
              :label="$t('message.index')"
            >
            </el-table-column>
            <el-table-column
              prop="point"
              align="center"
              :label="$t('message.point')"
            >
            </el-table-column>
            <el-table-column
              prop="bonus"
              align="center"
              :label="$t('message.bonus')"
            >
            </el-table-column>
          </el-table>
        </el-col>
        <!--非消除类游戏-->
        <el-col
          :span="22"
          type="flex"
          v-else-if="is_all_lines == 0 || is_all_lines == 1 || is_all_lines == 4"
        >
          <el-col :span="24" class="bg-purple-dark" style="font-weight: bold;">
            <p class="el-icon-caret-bottom" style="margin-left:5px;width: 20px;"></p>
            {{ $t("message.game_results") }}: {{game_result_name}} {{ $t("message.bonus") }}：{{ data.info && data.info.type_id === 10 ? lines_bonus : lines_bonus.toFixed(2)}}
          </el-col>
          <el-col :span="24" style="background:#FFFAFA;">
            <!-- 翻倍小游戏 -->
            <template v-if="data.info && data.info.type_id === 10">
              <div class="result-times-game">
                <el-row type="flex" justify="center">
                  <template v-for="(item, gridIndex) in grid_data">
                    <span v-for="(item_one, itemIndex) in item" :key="`${gridIndex}-${itemIndex}`">
                      <img :src="item_one" style="max-width:40px" />
                    </span>
                  </template>
                </el-row>
              </div>
            </template>
            <!-- 喜气洋洋 / 国王游戏 -->
            <template v-else-if="game_id === 3329">
              <el-col :span="18" class="el-grids-style align-grid" style="display:flex;align-items:center;">
                <el-col v-for="(item, gridIndex) in grid_data" :key="gridIndex" v-bind:style="grids_style" align="center">
                  <div v-for="sub, idx in item" :key="idx" class="el-grids">
                    <img :src="sub.img"/>
                    <span v-if="sub.reward !== '0'" class="grid-reward">{{ sub.reward }}</span>
                  </div>
                </el-col>
              </el-col>
            </template>
            <!-- 鹊桥会 -->
            <template v-else-if="game_id === 2532">
              <div class="result-bg" style="display:flex;align-items:center;justify-content:center;">
                <div v-for="row, idx in data.info.resultData" :key="idx" style="display:flex;align-items:center;justify-content:center;padding:20px;">
                  <div v-for="item, idx in row" :key="idx" type="flex" justify="center" align="center">
                    <div v-for="url, idx in item" :key="idx">
                      <img :src="url" alt="" />
                    </div>
                  </div>
                </div>
              </div>
            </template>
            <template v-else>
              <el-col :span="18" class="el-grids-style">
                <el-col v-for="(item, gridIndex) in grid_data" :key="gridIndex" v-bind:style="grids_style" align="center">
                  <div v-for="(item_one, itemIndex) in item" :key="itemIndex" class="el-grids">
                    <img :src="item_one" />
                  </div>
                </el-col>
              </el-col>
            </template>
          </el-col>
          <div
            v-show="choose_list.length > 0"
            class="choose_body"
            v-bind:style="{ left: left_value }"
          >
            <div class="choose_body_font">
              <div>{{all_bets}}</div>
              <div>{{ $t("message.re_do") }}</div>
            </div>
          </div>
          <!--非满线普通中奖数据表-->
          <el-col v-show="is_all_lines == 0" :span="24">
            <!-- 翻倍小游戏 -->
            <el-table v-if="data.info && data.info.type_id === 10" :data="data.info.m_list_info[0].line_num">
              <el-table-column
                prop="times"
                align="center"
                :label="$t('message.multiplie')"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;" v-show="scope.row.times">{{scope.row.times}}</i>
                </template>
              </el-table-column>
              <el-table-column
                prop="wins"
                align="center"
                :label="$t('message.bonus')"
              ></el-table-column>
            </el-table>
            <el-table v-else :data="filteredLines" style="width: 100%">
              <!-- 喜气洋洋 / 国王游戏 -->
              <template v-if="game_id === 3329">
                <el-table-column
                  prop="line_id"
                  align="center"
                  :label="$t('message.lines')"
                  min-width="25%"
                >
                  <template slot-scope="scope">
                    <el-col v-if="scope.row.ts == 1">
                      <i style="font-style:normal;">{{scope.row.line_id}}</i>
                      <img v-show="scope.row.line_shape_url" :src="scope.row.line_shape_url" />
                    </el-col>
                    <el-col v-else>
                      <i style="font-style:normal;">{{ $t("message.special_style") }}</i>
                    </el-col>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="direction"
                  align="left"
                  :label="$t('message.direction')"
                  min-width="15%"
                ></el-table-column>
                <el-table-column
                  prop="symbol"
                  align="left"
                  :label="$t('message.symbol')"
                  width="235"
                >
                  <template slot-scope="scope">
                    <span v-for="item, i in scope.row.symbol" :key="i" class="symbol-grid">
                      <img :src="item.img" />
                      <span v-if="item.reward !== '0'" class="grid-reward">{{ item.reward }}</span>
                    </span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="bonus"
                  align="left"
                  :label="$t('message.bonus')"
                  min-width="20%"
                >
                  <template slot-scope="scope">{{ scope.row.bonus }}</template>
                </el-table-column>
              </template>
              <template v-else>
                <el-table-column
                  prop="line_id"
                  align="center"
                  :label="$t('message.lines')"
                  min-width="25%"
                >
                  <template slot-scope="scope">
                    <el-col v-if="scope.row.ts == 1">
                      <i style="font-style:normal;">{{scope.row.line_id}}</i>
                      <img v-show="scope.row.line_shape_url" :src="scope.row.line_shape_url" style="vertical-align: middle;" />
                    </el-col>
                    <el-col v-else>
                      <i style="font-style:normal;">{{ $t("message.special_style") }}</i>
                    </el-col>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="direction"
                  align="left"
                  :label="$t('message.direction')"
                  min-width="15%"
                ></el-table-column>
                <el-table-column
                  prop="symbol"
                  align="left"
                  :label="$t('message.symbol')"
                  min-width="25%"
                >
                  <template slot-scope="scope">
                    <img v-for="(item, symbolIndex) in scope.row.symbol" :key="symbolIndex" :src="item" align="left" />
                  </template>
                </el-table-column>
                <el-table-column
                  prop="times"
                  align="left"
                  :label="$t('message.multiplie')"
                  min-width="15%"
                >
                  <template slot-scope="scope">
                    <i style="font-style:normal;" v-show="scope.row.times">X {{scope.row.times}}</i>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="bonus"
                  align="left"
                  :label="$t('message.bonus')"
                  min-width="20%"
                ></el-table-column>
              </template>
            </el-table>
          </el-col>
          <!--满线普通中奖数据表-->
          <el-col v-show="is_all_lines == 1" :span="24">
            <el-table :data="filteredLines" style="width: 100%">
              <el-table-column
                prop="img_url"
                align="left"
                :label="$t('message.symbol')"
                min-width="30%"
              >
                <template slot-scope="scope">
                  <img v-for="(item, imgIndex) in scope.row.img_url" :key="imgIndex" :src="item" align="left" />
                </template>
              </el-table-column>
              <el-table-column
                prop="lin_num"
                align="center"
                :label="$t('message.combination')"
                min-width="20%"
              ></el-table-column>
              <el-table-column
                prop="times"
                align="center"
                :label="$t('message.multiplie')"
                min-width="20%"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;" v-show="scope.row.times">X {{scope.row.times}}</i>
                </template>
              </el-table-column>
              <el-table-column
                prop="bonus"
                align="center"
                :label="$t('message.bonus')"
                min-width="30%"
              ></el-table-column>
            </el-table>
          </el-col>
          <!--满线红利中奖数据表-->
          <el-col v-show="is_all_lines == 4" :span="24">
            <el-table :data="filteredLines" style="width: 100%">
              <el-table-column
                prop="img_url"
                align="left"
                :label="$t('message.symbol')"
                min-width="30%"
              >
                <template slot-scope="scope">
                  <img v-for="(item, imgIndex) in scope.row.img_url" :key="imgIndex" :src="item" align="left" />
                </template>
              </el-table-column>
              <el-table-column
                prop="lin_num"
                align="center"
                :label="$t('message.combination')"
                min-width="20%"
              ></el-table-column>
              <el-table-column
                prop="times"
                align="center"
                :label="$t('message.multiplie')"
                min-width="20%"
              >
                <template slot-scope="scope">
                  <i style="font-style:normal;" v-show="scope.row.times">X {{scope.row.times}}</i>
                </template>
              </el-table-column>
              <el-table-column
                prop="bonus"
                align="center"
                :label="$t('message.bonus')"
                min-width="30%"
              ></el-table-column>
            </el-table>
          </el-col>
        </el-col>
        <!--消除类游戏-->
        <el-col :span="22" type="flex" v-else>
          <el-collapse
            v-model="activeNames"
            @change="handleChange"
            v-show="lines_info.length > 0"
            v-for="(line_data, index) in lines_info"
            :key="index"
          >
            <el-collapse-item style="margin-bottom:20px;" :name="index">
              <template slot="title">
                <div style="background: #e2e6eb;font-weight: bold;">
                  <p class="el-icon-caret-bottom" style="margin-left:5px;width: 20px;"></p>
                  {{ $t("message.game_results") }}{{ index + 1}}: {{game_result_name}} {{ $t("message.bonus") }}：{{ countBonus(index) }}
                </div>
              </template>
              <el-col :span="24" type="flex">
                <el-col :span="24" class="result-bg">
                  <!-- 国王游戏/疯狂黄金城 -->
                  <div v-if="game_id === 3328 || game_id === 3330 || game_id === 2278" style="margin: 10px">
                    <div style="display: flex; justify-content: center;">
                      <img v-for="img, i in line_data.topRow" :key="i" :src="img" style="display: block;" />
                    </div>
                    <div style="display: flex; justify-content: center;">
                      <div                      
                        v-for="(item, gird_index) in line_data.grid_data"
                        :key="gird_index"
                      >
                        <img v-for="img, i in item" :key="i" :src="img" style="display: block;" />
                      </div>
                    </div>
                  </div>
                  <!-- 麻将来了 -->
                  <div v-else-if="game_id === 3331" style="display:flex;justify-content:center;align-items:center;margin: 10px 0;">
                    <el-row type="flex" justify="center">
                      <el-col
                        v-for="(item, gird_index) in line_data.grid_data"
                        :key="gird_index"
                        align="center"
                      >
                        <div
                          v-for="(item_one, item_index) in item"
                          :id="index + '_' + gird_index + '_' + item_index"
                        >
                          <img :src="item_one" />
                        </div>
                      </el-col>
                    </el-row>
                  </div>
                  <!-- 粉红女郎 -->
                  <el-row v-else-if="game_id === 3304" type="flex" justify="center" align="middle" class="el-grids-style">
                    <el-col
                      v-for="(item, gird_index) in line_data.grid_data"
                      :key="gird_index"
                      v-bind:style="grids_style"
                      align="center"
                    >
                      <div
                        v-for="(item_one, item_index) in item"
                        class="el-grids"
                        :id="index + '_' + gird_index + '_' + item_index"
                      >
                        <img :src="item_one" />
                      </div>
                    </el-col>
                  </el-row>
                  <el-col v-else :span="18" class="el-grids-style">
                    <el-col
                      v-for="(item, gird_index) in line_data.grid_data"
                      :key="gird_index"
                      v-bind:style="grids_style"
                      align="center"
                    >
                      <div
                        v-for="(item_one, item_index) in item"
                        class="el-grids"
                        :id="index + '_' + gird_index + '_' + item_index"
                      >
                        <img :src="item_one" />
                      </div>
                    </el-col>
                  </el-col>
                </el-col>
                <!--非满线消除中奖数据表-->
                <el-col v-show="is_all_lines == 2" :span="24">
                  <el-table
                    :data="line_data.line_num"
                    style="width: 100%"
                    @cell-mouse-enter="handleMoueEnter(index)"
                    @cell-mouse-leave="handleMouseLeave"
                  >
                    <el-table-column
                      prop="lineID"
                      align="center"
                      :label="$t('message.lines')"
                      min-width="25%"
                    >
                      <template slot-scope="scope">
                        <el-col v-if="scope.row.ts == 1">
                          <i style="font-style:normal;">{{scope.row.lineID}}</i>
                          <img v-show="scope.row.argument" :src="scope.row.argument" />
                        </el-col>
                        <el-col v-else>
                          <i style="font-style:normal;">{{ $t("message.special_style") }}</i>
                        </el-col>
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="direction"
                      align="left"
                      :label="$t('message.direction')"
                      min-width="15%"
                    ></el-table-column>
                    <el-table-column
                      prop="symbol"
                      align="left"
                      :label="$t('message.symbol')"
                      min-width="25%"
                    >
                      <template slot-scope="scope">
                        <img v-for="(item, symbolIndex) in scope.row.symbol" :key="symbolIndex" :src="item" align="left" />
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="times"
                      align="left"
                      :label="$t('message.multiplie')"
                      min-width="15%"
                    ></el-table-column>
                    <el-table-column
                      prop="wins"
                      align="left"
                      :label="$t('message.bonus')"
                      min-width="20%"
                    ></el-table-column>
                  </el-table>
                </el-col>
                <!--满线消除中奖数据表-->
                <el-col v-show="is_all_lines == 3" :span="24">
                  <el-table v-if="line_data.line_num.length" :data="line_data.line_num" style="width: 100%" class="table">
                    <el-table-column
                      prop="img_url"
                      align="left"
                      :label="$t('message.symbol')"
                      min-width="30%"
                    >
                      <template slot-scope="scope">
                        <img v-for="(item, imgIndex) in scope.row.img_url" :key="imgIndex" :src="item" align="left" />
                      </template>
                    </el-table-column>
                    <el-table-column
                      prop="lin_num"
                      align="center"
                      :label="$t('message.combination')"
                      min-width="20%"
                    ></el-table-column>
                    <el-table-column
                      prop="times"
                      align="center"
                      :label="$t('message.multiplie')"
                      min-width="20%"
                    ></el-table-column>
                    <el-table-column
                      prop="wins"
                      align="center"
                      :label="$t('message.bonus')"
                      min-width="30%"
                    ></el-table-column>
                  </el-table>
                </el-col>
              </el-col>
            </el-collapse-item>
          </el-collapse>
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
.result-bg {
  background: #cf9e9e;
}
.result-bg img {
  max-width: 40px;
  width: 100%;
}
.el-grids-style {
  background: #cf9e9e;
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 10px 0;
}
.el-grids-style .el-col {
  width: auto !important;
}
.el-grids {
  width: 40px;
  height: 40px;
  position: relative;
}
.grid-reward {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  white-space: nowrap;
}
.symbol-grid {
  position: relative;
  height: 40px;
  margin: 4px 15px;
  text-align: center;
  display: inline-block;
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
.result-times-game {
  padding: 10px;
  background: #cf9e9e;
}
.table .cell img {
  max-width: 40px;
}
</style>

<script>
import http from '../../assets/js/http'
export default {
  data() {
    return {
      loading: true,
      detail_visible: false,
      is_all_lines: 0,
      recordData: {},
      grids_style: { width: '', 'margin-bottom': '' },
      tableData: [
        { key: '', value: '' },
        { key: '', value: '' },
        { key: '', value: '' },
        { key: '', value: '' },
        { key: '', value: '' },
        { key: '', value: '' }
      ],
      lines_info: [],
      grid_data: [],
      grid_show: [],
      scatter_info: [],
      bonus_game_info: [],
      conin_info: [],
      choose_list: [],
      conin_type_name: [],
      percent_list: ['16%', '30%', '44%', '58%', '71%'],
      grid_width_list: ['100%', '50%', '33.33%', '25%', '20%', '16.66%'],
      game_result_name: '',
      all_bets: 0,
      lines_bonus: 0,
      xiao_all_bonus: 0,
      index: 1,
      left_value: '',
      grid_width: '',
      activeNames: [],
      game_id: 0,
      data: {}
    }
  },
  computed: {
    filteredLines() {
      return this.lines_info.filter(lines_info => lines_info.bonus > 0)
    }
  },
  methods: {
    handleMoueEnter(index, row) {
      // console.log(index)
    },
    handleMouseLeave(row) {
      this.isOpen = false
      // console.log(row)
    },
    handleChange(val) {
      // console.log(val)
    },
    timeLabel() {
      let utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    countBonus(index) {
      let data = this.lines_info[index].line_num
      let sum = 0
      if (data) {
        for (var item of data) {
          sum += parseFloat(item.wins)
        }
      }
      return sum.toFixed(2)
    },
    handleGrids() {
      if (this.grid_show.length > 0) {
        var grid = []
        for (var item in this.grid_data) {
          var temp1 = []
          var data = this.grid_data[item]
          var len = this.grid_data[item].length
          for (var item1 of this.grid_show[item]) {
            var key = item1 % len
            temp1.push(data[key - 1])
          }
          grid.push(temp1)
        }
        this.grid_data = grid
      }
    },
    open() {
      setTimeout(() => {
        this.loading = false
      }, 5000)
      this.detail_visible = true
      this.conin_type_name = [
        this.$t('message.fight_monster_bonuses'),
        this.$t('message.collect_and_reward'),
        this.$t('message.red_envelope_rewards'),
        this.$t('message.slyder_Adventures'),
        this.$t('message.jackPot')
      ]
      this.recordData = Lockr.get('slot_detail')
      this.game_result_name = this.recordData.game_type
      this.tableData[0].key = this.$t('message.game_time') + this.timeLabel()
      this.tableData[0].value = this.recordData.time
      this.tableData[1].key = this.$t('message.game_number')
      this.tableData[1].value = this.recordData.id
      this.tableData[2].key = this.$t('message.game_name')
      this.tableData[2].value = this.recordData.game_name
      this.tableData[3].key = this.$t('message.bet')
      this.tableData[4].key = this.$t('message.detail_total_bets')
      // this.tableData[4].value = this.recordData.all_bets
      this.tableData[5].key = this.$t('message.total_bonus')
      this.tableData[5].value = this.recordData.all_bonus

      this.apiGet(this.recordData.dt_url).then(res => {
        this.handelResponse(res, data => {
          this.data = data
          this.game_id = data.info.game_id
          this.is_all_lines = data.info.is_all_lines
          this.scatter_info = data.info.scatter_info
          this.tableData[4].value = data.info.all_bets
          if (this.is_all_lines == 0 || this.is_all_lines == 1) {
            this.grid_data = data.info.grids
            this.grid_show = data.info.grid_show
            this.lines_info = this.is_all_lines == 0 ? data.info.lines_info : data.info.m_list_info
            this.handleGrids()
            this.grids_style.width = this.grid_width_list[this.grid_data.length - 1]
            if (data.info.choose_list.length > 0) {
              this.choose_list = data.info.choose_list
              this.left_value = this.percent_list[this.choose_list[0] - 1]
              this.all_bets = data.info.all_bets
              this.grids_style['margin-bottom'] = '60px'
            }
            for (var item of this.lines_info) {
              this.lines_bonus += parseFloat(item.bonus)
            }
            if (this.scatter_info) {
              for (var scatter of this.scatter_info) {
                this.lines_bonus += parseFloat(scatter.bonus)
                var temp = {}
                temp.bets = scatter.bets
                temp.img_num = scatter.num
                temp.times = scatter.times
                temp.bonus = scatter.bonus
                if (this.is_all_lines == 1) {
                  temp.img_url = []
                  for (var i1 = 0; i1 < scatter.num; i1++) {
                    temp.img_url.push(scatter.img_url)
                  }
                  temp.lin_num = ''
                } else if (this.is_all_lines == 0) {
                  temp.line_id = scatter.num + ' X'
                  temp.line_shape_url = ''
                  temp.direction = ''
                  temp.symbol = []
                  for (var i2 = 0; i2 < scatter.num; i2++) {
                    temp.symbol.push(scatter.img_url)
                  }
                }
                this.lines_info.push(temp)
              }
            }
            if (this.data.info.type_id === 10) {
              this.lines_bonus = this.data.info.sum_bonus
            }
          } else if (this.is_all_lines == 2 || this.is_all_lines == 3) {
            this.activeNames = Array.from(Array(20), (v, k) => k)
            this.lines_info = this.is_all_lines == 2 ? data.info.lines_info : data.info.m_list_info
            this.grids_style.width = this.grid_width_list[this.lines_info[0].grid_data.length - 1]
            if (this.scatter_info) {
              for (var scatter1 of this.scatter_info) {
                var temp1 = {}
                temp1.times = 'X ' + scatter1.times
                temp1.wins = scatter1.bonus
                if (this.is_all_lines == 3) {
                  temp1.bets = scatter1.bets
                  temp1.img_url = []
                  for (var i3 = 0; i3 < scatter1.num; i3++) {
                    temp1.img_url.push(scatter1.img_url)
                  }
                  temp1.lin_num = ''
                } else if (this.is_all_lines == 2) {
                  temp1.lineID = scatter1.num + ' X'
                  temp1.argument = ''
                  temp1.direction = ''
                  temp1.is_win = ''
                  temp1.symbol = []
                  for (var i4 = 0; i4 < scatter1.num; i4++) {
                    temp1.symbol.push(scatter1.img_url)
                  }
                }
                this.lines_info[this.lines_info.length - 1].line_num.push(temp1)
              }
            }
          } else if (this.is_all_lines == 4) {
            this.grid_data = data.info.grids
            this.grid_show = data.info.grid_show
            this.lines_info = data.info.m_list_info
            this.handleGrids()
            this.grids_style.width = this.grid_width_list[this.grid_data.length - 1]
            if (data.info.choose_list.length > 0) {
              this.choose_list = data.info.choose_list
              this.left_value = this.percent_list[this.choose_list[0] - 1]
              this.all_bets = data.info.all_bets
              this.grids_style['margin-bottom'] = '60px'
            }
            for (let item of this.lines_info) {
              this.lines_bonus += parseFloat(item.bonus)
            }
          }

          this.bonus_game_info = data.info.bonus_game_info.bonus_info
          this.xiao_all_bonus = data.info.bonus_game_info.xiao_all_bonus
          this.conin_info = data.info.conin_info
          this.tableData[3].value = data.info.line_bets

          window.setTimeout(() => {
            // 喜气洋洋,国王游戏,麻将来了,群星闪耀,希腊传说,幸运水果机，鹊桥会
            if (this.game_id === 3329 || this.game_id === 3328 || this.game_id === 3330 || this.game_id === 3331 || this.game_id === 2278 ||
              this.game_id === 2529 || this.game_id === 2261 || this.game_id === 3301 || this.game_id === 2532 || this.game_id === 3304) {
              const tr = window.document.querySelector('.J_Result tbody tr:nth-child(4)')
              if (tr) {
                tr.style.display = 'none'
              }
            }
          }, 150)

          // 国王游戏
          if (this.game_id === 3328 || this.game_id === 3330) {
            for (let i = 0, len = this.lines_info.length; i < len; i++) {
              this.lines_info[i].topRow = this.lines_info[i].grid_data.pop()
            }
          }

          // 鹊桥会
          if (this.game_id === 2532) {
            const resultList = this.data.info.grids.length ? [this.data.info.grids] : []
            if ((this.data.info.type_id === 19 || this.data.info.type_id === 2) && this.data.info.grids.length) {
              const mid = this.data.info.grids.length / 2
              resultList[0] = this.data.info.grids.slice(0, mid)
              resultList[1] = this.data.info.grids.slice(mid)
            }
            this.data.info.resultData = resultList
          }

          this.loading = false
        })
      })
    },
    close() {
      Lockr.rm('slot_detail')
      this.loading = true
      this.game_id = 0
      this.lines_info.splice(0, this.lines_info.length)
      this.grid_data.splice(0, this.grid_data.length)
      this.conin_info.splice(0, this.conin_info.length)
      this.choose_list.splice(0, this.choose_list.length)
      this.all_bets = 0
      this.left_value = ''
      this.lines_bonus = 0
      this.xiao_all_bonus = 0
      this.detail_visible = false
      this.is_all_lines = 0
      this.grids_style['margin-bottom'] = ''
      this.data = {}
    }
  },
  mixins: [http]
}
</script>


// WEBPACK FOOTER //
// slot_detail.vue?6c989875
