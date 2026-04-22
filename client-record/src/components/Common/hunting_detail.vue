<template>
  <el-dialog
    ref="dialog"
    :visible.sync="detail_visible"
    :modal-append-to-body="false"
    :show-close="false"
    :before-close="close"
    width="80%"
  >
    <div class="content-box hunting-detail">
      <el-row type="flex" justify="center">
        <el-col :span="22" type="flex" align="middle" style="margin-top:30px;">
          <p style="font-weight: bold;font-size: 24px;margin:0 0 10px 0;">{{ $t("message.game_details") }}</p>
          <template v-if="game_id === 5007 && type === 75">
            <div class="description">
              <span class="description-label">{{ $t("message.killScenesID") }}</span>
              <span class="description-item">{{ table_data[0].hit_report_id || '&nbsp;' }}</span>
              <template v-if="type_5007_75_step === 2">
                <span class="description-label">{{ $t("message.stepOneScenesID") }}</span>
                <span class="description-item">{{ table_data[0].id_one || '&nbsp;' }}</span>
                <span class="description-label">{{ $t("message.stepOneOrderNumber") }}</span>
                <span class="description-item">{{ table_data[0].elf_chess_id || '&nbsp;' }}</span>
              </template>
            </div>
            <el-table :data="scenes_data" stripe>
              <el-table-column
                prop="elf_chess_id"
                :label="$t('message.playSerial')"
                align="center"
                min-width="50%"
              ></el-table-column>
              <el-table-column
                prop="nickname"
                :label="$t('message.player')"
                align="center"
                min-width="35%"
              ></el-table-column>
              <el-table-column
                prop="room_ante"
                :label="$t('message.roomAnte')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="elf_chess_type"
                :label="type_5007_75_step <= 1 ? $t('message.battleChoice') : $t('message.remorseChoice')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="elf_chess_status"
                :label="$t('message.result')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="start_chips"
                :label="$t('message.pre_balance')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="end_chips"
                :label="$t('message.post_balance')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="fish_chips"
                :label="$t('message.gain')"
                align="center"
                min-width="25%"
              >
                <template slot-scope="scope">
                  {{ scope.row.fish_chips }}
                </template>
              </el-table-column>
              <el-table-column
                prop="time"
                :label="$t('message.settleTime')"
                align="center"
                min-width="45%"
              >
                <template slot-scope="scope">
                  {{ scope.row.time }}
                </template>
              </el-table-column>
            </el-table>
          </template>
          <div v-else-if="game_id === 5007 && type === 80">
            <div class="description">
              <span class="description-label">{{ $t("message.killScenesID") }}</span>
              <span class="description-item">{{ data.hit_report_id || '&nbsp;' }}</span>
            </div>
            <el-table :data="scenes_data" stripe>
              <el-table-column
                prop="play_id"
                :label="$t('message.playSerial')"
                align="center"
                min-width="50%"
              ></el-table-column>
              <el-table-column
                prop="fish_type_url"
                :label="$t('message.type')"
                align="center"
                min-width="50%"
              >
                <template slot-scope="scope">
                  <img :src="scope.row.fish_type_url" align="center" style="width:34px;height:32px;"/>
                </template>
              </el-table-column>
              <el-table-column
                prop="nickname"
                :label="$t('message.player')"
                align="center"
                min-width="35%"
              ></el-table-column>
              <el-table-column
                prop="room_ante"
                :label="$t('message.roomAnte')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="play_name"
                :label="$t('message.playTypeName')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="result"
                :label="$t('message.result')"
                align="center"
                min-width="25%"
              >
                <template slot-scope="scope">
                  {{ scope.row.result }}%
                </template>
              </el-table-column>
              <el-table-column
                prop="start_chips"
                :label="$t('message.pre_balance')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="end_chips"
                :label="$t('message.post_balance')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="bpay"
                :label="$t('message.gain')"
                align="center"
                min-width="25%"
              ></el-table-column>
              <el-table-column
                prop="time"
                :label="$t('message.settleTime')"
                align="center"
                min-width="45%"
              >
                <template slot-scope="scope">
                  {{ scope.row.time }}
                </template>
              </el-table-column>
            </el-table>
          </div>
          <el-table :data="table_data" stripe v-else>
            <el-table-column
              prop="bullet_serial_id"
              :label="$t('message.bet_number')"
              align="center"
              min-width="25%"
            >
              <!--流水号-->
            </el-table-column>
            <el-table-column
              prop="bullet_chips"
              :label="$t('message.bullet_price')"
              align="center"
              min-width="25%"
            >
              <!--子弹价值-->
            </el-table-column>
            <el-table-column
              prop="bpay"
              :label="$t('message.profit')"
              align="center"
              min-width="25%"
            >
              <!--盈利-->
            </el-table-column>
            <el-table-column
              prop="room_ante"
              :label="$t('message.times_scene')"
              align="center"
              min-width="10%"
            >
              <!--场景倍数-->
            </el-table-column>
            <el-table-column
              prop="bullet_consumption"
              :label="$t('message.bulletConsumption')"
              align="center"
              min-width="20%"
              v-if="game_id === 5007 || game_id === 5009 || game_id === 5013 || game_id === 5014"
            ></el-table-column>
            <el-table-column
              prop="enchant_multiple"
              :label="$t('message.attachMagic')"
              align="center"
              min-width="20%"
              v-if="game_id === 5007 || game_id === 5009 || game_id === 5013 || game_id === 5014"
            ></el-table-column>
            <el-table-column
              prop="time"
              :label="$t('message.shoot_time') + timeLabel()"
              align="center"
              min-width="40%"
            >
              <!--射击时间-->
            </el-table-column>
            <el-table-column
              prop="is_item"
              :label="$t('message.isPropRat')"
              align="center"
              min-width="20%"
              v-if="game_id == 5008"
            >
              <!--是否为道具鼠-->
              <template
                slot-scope="scope"
              >{{scope.row.is_item==true?$t('message.is_yes'):$t('message.is_no')}}</template>
            </el-table-column>
            <el-table-column
              prop="time"
              :label="$t('message.DurationProps')"
              align="center"
              min-width="20%"
              v-if="game_id == 5008"
            >
              <!--道具时长-->
              <template slot-scope="scope">{{scope.row.item_time==0?'-':scope.row.item_time}}</template>
            </el-table-column>
            <el-table-column
              prop="fish_type_url"
              :label="$t('message.type')"
              align="center"
              min-width="20%"
            >
              <!--类型-->
              <template slot-scope="scope">
                <img :src="scope.row.fish_type_url" />
              </template>
            </el-table-column>
            <el-table-column
              prop="is_dead_url"
              :label="$t('message.hit')"
              align="center"
              min-width="10%"
            >
              <!--是否命中-->
              <template slot-scope="scope">
                <img :src="scope.row.is_dead_url" />
              </template>
            </el-table-column>
          </el-table>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="24" align="center">
          <div>
            <el-pagination
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
              :current-page="current_page"
              :page-sizes="[10, 20, 50]"
              :pager-count="5"
              :page-size="current_size"
              layout="total, sizes, prev, pager, next, jumper"
              :total="total"
            ></el-pagination>
          </div>
        </el-col>
      </el-row>

      <el-row type="flex" justify="center">
        <el-col :span="22" type="flex" align="middle" style="margin-bottom:30px;">
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
.hunting-detail .description {
  background: #fff;
  margin-bottom: 10px;
  padding: 10px;
  font-size: 0;
  text-align: left;
}
.hunting-detail .description-label,
.hunting-detail .description-item {
  padding: 0 10px;
  display: inline-block;
  border: 1px solid #ebeef5;
  height: 30px;
  line-height: 30px;
  font-size: 14px;
}
.hunting-detail .description-label {
  background: #fafafa;
}
.hunting-detail .description-label:not(:first-child) {
  margin-left: -1px;
}
.hunting-detail .description-item {
  margin-left: -1px;
}
</style>

<script>
import http from '../../assets/js/http'
export default {
  data() {
    return {
      detail_visible: false,
      recordData: {},
      table_data: [],
      total: 0,
      current_size: 10,
      current_page: 1,
      game_id: '',
      type: -1, // 场景类型
      type_5007_75_step: -1,
      scenes_data: [],
      data: {}
    }
  },
  computed: {},
  methods: {
    timeLabel() {
      let utc = Lockr.get('timezone')
      return utc == -4 ? this.$t('message.usa_time') : ''
    },
    getFishDetail(page, size) {
      const form_data = {
        params: {
          page: page,
          size: size
        }
      }
      this.game_id = this.recordData.game_id
      this.apiGet(this.recordData.dt_url, form_data).then(res => {
        this.handelResponse(res, data => {
          this.data = data
          this.total = data.info.total
          this.table_data = data.info.data
          if (this.recordData.game_id === 5007 && data.info.data[0].type === 75) {
            const row = data.info.data[0]
            this.type = row.type
            this.type_5007_75_step = row.elf_chess_step
            this.scenes_data = JSON.parse(JSON.stringify(data.info.data))
          } else if (this.recordData.game_id === 5007 && this.recordData.type === 80) {
            this.type = this.recordData.type
            this.scenes_data = JSON.parse(JSON.stringify(data.info.data))
          }
          if (this.table_data.length) {
            this.table_data.push({
              bullet_serial_id: this.$t('message.sub_total') + '：',
              bullet_chips: data.info.s_data.x_bullet_chips,
              bpay: data.info.s_data.x_all_bpay
            })
            this.table_data.push({
              bullet_serial_id: this.$t('message.all_total') + '：',
              bullet_chips: data.info.s_data.bullet_chips,
              bpay: data.info.s_data.all_bpay
            })
          }
        })
      })
    },
    handleSizeChange(val) {
      this.current_size = val
      this.getFishDetail(this.current_page, this.current_size)
    },
    handleCurrentChange(val) {
      this.current_page = val
      this.getFishDetail(this.current_page, this.current_size)
    },
    open() {
      let loading = this.$loading({})
      setTimeout(() => {
        loading.close()
      }, 3000)
      this.detail_visible = true
      this.recordData = Lockr.get('hunting_detail')
      this.getFishDetail(this.current_page, this.current_size)
      loading.close()
    },
    close() {
      this.detail_visible = false
      this.table_data.splice(0, this.table_data.length)
      this.total = 0
      this.current_size = 10
      this.current_page = 1
      this.type = -1
      this.type_5007_75_step = -1
      this.scenes_data = []
    }
  },
  mixins: [http]
}
</script>


// WEBPACK FOOTER //
// hunting_detail.vue?0d0479d5