import { forEach } from '@/libs/tools';
<template>
  <div class="gogogo">
    <template v-if="fwType == 'Ttby' || fwType == 'Mrby'">
      <RadioGroup v-model="statementType" style="padding-bottom: 20px">
        <Radio border label="all">全部</Radio>
        <Radio border label="hit">命中</Radio>
        <Radio border label="miss">未命中</Radio>
      </RadioGroup>
    </template>

    <!-- 按实际情况判断是否展示庄家信息 -->
    <Table
      v-if="bankerShow"
      :columns="columnsBank"
      :data="bankerData"
      border
    ></Table>

    <!-- 展示玩家详情 -->
    <Table
      size="large"
      :columns="columns"
      :data="detialDataFilter"
      border
    ></Table>
  </div>
</template>
<script>
export default {
  name: 'game-detial',
  props: {
    rowInfo: Array,
    rowId: Number
  },
  data () {
    return {
      // 玩家配置信息
      columns: [],
      statementType: 'all',
      detialData: [],
      fwType: '',
      // 庄家配置信息
      bankerShow: false,
      columnsBank: [],
      bankerData: [],
      slwhmode: 0
    }
  },
  computed: {
    detialDataFilter () {
      let newData = []
      switch (this.statementType) {
        case 'all':
          newData = this.detialData
          break
        case 'hit':
          if (this.fwType == 'Jcby') {
            newData = this.detialData.filter((x) => Number(x.key2) > 0)
          } else {
            newData = this.detialData.filter((x) => x.key8 == 'true')
          }
          break
        case 'miss':
          if (this.fwType == 'Jcby') {
            newData = this.detialData.filter((x) => Number(x.key2) <= 0)
          } else {
            newData = this.detialData.filter((x) => x.key8 != 'true')
          }
          break
      }
      return newData
    }
  },
  methods: {
    pokerCards (item, isReplaceCard) {
      item = JSON.parse(JSON.stringify(item))
      if (item.number > 10 || item.number == 1) {
        item.number =
          { 1: 'A', 11: 'J', 12: 'Q', 13: 'K', 14: '小王', 15: '大王' }[
            item.number
          ] || item.number
      }
      let cards = {
        1: (
          <span class="color-style-black">
            {isReplaceCard ? '*' : ''}
            <img src="/color/hei.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        ),
        2: (
          <span class="color-style-red">
            {isReplaceCard ? '*' : ''}
            <Icon type="md-heart" size="23" />
            {item.number}
            &emsp;
          </span>
        ),
        3: (
          <span class="color-style-black">
            {isReplaceCard ? '*' : ''}
            <img src="/color/hua.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        ),
        4: (
          <span class="color-style-red">
            {isReplaceCard ? '*' : ''}
            <img src="/color/fang.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        ),
        5: (
          <span class="color-style-red">
            {isReplaceCard ? '*' : ''}
            &nbsp;
            {item.number}
            &emsp;
          </span>
        )
      }[item.color]
      return cards
    },
    isPaohuoliantian (red, black) {
      let [redCount, blackCount] = [0, 0]
      red.cards.map((item) => {
        let poker = JSON.parse(JSON.stringify(item.Card.Poker))
        if (poker.number == 7) {
          redCount++
        }
      })
      black.cards.map((item) => {
        let poker = JSON.parse(JSON.stringify(item.Card.Poker))
        if (poker.number == 7) {
          blackCount++
        }
      })
      console.log(redCount, blackCount)
      if (redCount == 2 || blackCount == 2) {
        return true
      }
      return false
    }
  },

  mounted () {
    if (this.rowInfo) {
      console.log(this.rowInfo[0].SettlementDetail)
      const fwType = Object.keys(this.rowInfo[0].SettlementDetail)[0]
      this.fwType = fwType
      this.detialData = []
      this.banker = []
      this.columns = []
      switch (fwType) {
        case 'Xyzp':
          this.columns = [
            {
              title: '下注',
              key: 'effectiveBets',
              render: (h, params) => {
                return <span>{params.row.effectiveBets}</span>
              }
            },
            {
              title: '中奖',
              render: (h, params) => {
                if (params.row.profitLoss + params.row.effectiveBets>0) {
                  return <span>{params.row.profitLoss + params.row.effectiveBets}</span>
                } else {
                  return <span>谢谢参与</span>
                }
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Xyzp)
          })
          break
        case 'Byb':
          this.columns = [
            {
              title: '玩家ID',
              key: 'id'
            },
            {
              title: '中将倍数',
              key: 'multiple',
              render: (h, params) => {
                return <span>{params.row.multiple}</span>
              }
            },
            {
              title: '备注',
              key: 'desc',
              render: (h, params) => {
                return <div style="color:red">该局为搏一搏注单</div>
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Byb)
          })
          break
        case 'Ddz':
          this.columns = [
            {
              title: '玩家ID',
              align: 'center',
              key: 'player_id',
              width: 180
            },
            {
              title: '玩家座位',
              align: 'center',
              width: 100,
              key: 'seat_id'
            },
            {
              title: '身份',
              align: 'center',
              width: 70,
              render (h, params) {
                if (params.row.is_boss) {
                  return <div>地主</div>
                } else {
                  return <div>农民</div>
                }
              }
            },
            {
              title: '初始牌',
              align: 'center',
              key: 'init_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.init_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '倍数',
              align: 'center',
              width: 70,
              key: 'current_multiple'
            },
            {
              title: '炸弹',
              align: 'center',
              width: 70,
              render (h, params) {
                if (params.row.bomb_count) {
                  return <div>{params.row.bomb_count}</div>
                } else {
                  return <div>0</div>
                }
              }
            },
            {
              title: '春天/反春',
              align: 'center',
              width: 110,
              render (h, params) {
                if (params.row.is_spring || params.row.is_no_spring) {
                  return <div>有</div>
                } else {
                  return <div>无</div>
                }
              }
            },
            {
              title: '总倍数',
              align: 'center',
              width: 80,
              key: 'all_current_multiple'
            },
            {
              title: '底分',
              align: 'center',
              width: 100,
              key: 'bet_score'
            },
            {
              title: '分数',
              align: 'center',
              width: 100,
              key: 'change_game_currency'
            }
          ]
          let all_current_multiple =
            this.rowInfo[0].SettlementDetail.Ddz.current_multiple *
            this.rowInfo[1].SettlementDetail.Ddz.current_multiple *
            this.rowInfo[2].SettlementDetail.Ddz.current_multiple
          this.rowInfo.map((item) => {
            let detail = item.SettlementDetail.Ddz
            detail.init_cards.sort((card1, card0) => {
              let card1Point = card1.Card.Poker.number,
                card0Point = card0.Card.Poker.number
              if (card1.Card.Poker.number == 2) {
                card1Point = 0
              }
              if (card0.Card.Poker.number == 2) {
                card0Point = 0
              }
              if (card1.Card.Poker.number >= 14) {
                card1Point = -1
              }
              if (card0.Card.Poker.number >= 14) {
                card0Point = -1
              }
              return card1Point - card0Point
            })
            detail['all_current_multiple'] = all_current_multiple
            this.detialData.push(detail)
          })
          break
        case 'Jszjh':
          this.columns = [
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: 'id',
              key: 'player_id'
            },
            {
              title: '牌型',
              align: 'center',
              key: 'poker_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            { title: '房间底注', key: 'room_bet' },
            {
              title: '结果',
              render: (h, params) => {
                let jsx
                if (params.row.is_win) {
                  jsx = <span>赢家</span>
                } else {
                  jsx = <span></span>
                }
                return jsx
              }
            },
            { title: '牌名', key: 'card_type' },
            { title: '倍数', key: 'card_type_multiple' }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Jszjh)
          })
          break
        case 'Thwj':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },

            {
              title: '牌面',
              align: 'center',
              key: 'poker_cards',
              width: '500',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            { title: '倍数', key: 'card_type_multiple' },
            {
              title: '牌型',
              key: 'card_type'
            },
            {
              title: '分数',
              key: 'change_game_currency',
              render (h, params) {
                return params.row.change_game_currency >= 0 ? (
                  <span style="color:green">
                    {params.row.change_game_currency}
                  </span>
                ) : (
                  <span style="color:red">
                    {params.row.change_game_currency}
                  </span>
                )
              }
            },

            {
              title: '结果',
              key: 'is_win',
              align: 'center',
              width: 120,
              render: (h, params) => {
                return params.row.is_win == true ? (
                  <span style="color:red">赢家</span>
                ) : (
                  ''
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Thwj)
          })
          break
        case 'Xqzjh':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id',
              width: 150
            },
            {
              title: '座位',
              key: 'seat_id',
              width: 150,
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '牌型',
              align: 'center',
              key: 'poker_cards',
              width: 250,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '总有效下注',
              key: 'player_bet_multiple',
              align: 'center',
              width: 150
            },
            {
              title: '喜钱',
              key: 'player_bet_multiple',
              align: 'center',
              width: 70,
              render (h, params) {
                return (
                  <div>{params.row.xi_qian ? +params.row.xi_qian : '/'}</div>
                )
              }
            },
            {
              title: '结果',
              key: 'is_win',
              align: 'center',
              width: 120,
              render: (h, params) => {
                return params.row.is_win == true ? (
                  <span style="color:red">赢家</span>
                ) : (
                  ''
                )
              }
            },
            {
              title: '下注轮次',
              key: 'player_details',
              align: 'center',
              render: (h, params) => {
                let array = []
                params.row.player_details.forEach((item) => {
                  if (item.is_start_contrast_card === true) {
                    array.push(
                      <tr>
                        <td>{item.round_no}</td>
                        <td>{item.bet}</td>
                        <td>{item.is_show_card ? '是' : ''}</td>
                        <td>{item.contrast_card_pos}</td>
                        <td>{item.contrast_card_result ? '赢' : '输'}</td>
                        <td>{item.is_dis_card ? '是' : ''}</td>
                        <td>{item.is_start_contrast_card ? '是' : ''}</td>
                      </tr>
                    )
                  } else {
                    array.push(
                      <tr>
                        <td>{item.round_no}</td>
                        <td>{item.bet}</td>
                        <td>{item.is_show_card ? '是' : ''}</td>
                        <td>{item.contrast_card_pos}</td>
                        <td>{''}</td>
                        <td>{item.is_dis_card ? '是' : ''}</td>
                        <td>{item.is_start_contrast_card ? '是' : ''}</td>
                      </tr>
                    )
                  }
                })
                return (
                  <table class="zjh-table">
                    <tr>
                      <th>轮数</th>
                      <th>该轮下注</th>
                      <th>是否明牌</th>
                      <th>比牌对象</th>
                      <th>比牌结果</th>
                      <th>是否弃牌</th>
                      <th>主动比牌</th>
                    </tr>
                    {array}
                  </table>
                )
              }
            }
          ]

          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Xqzjh)
          })
          break
        case 'Hlzjh':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id',
              align: 'center',
              width: 150
            },
            {
              title: '座位',
              key: 'seat_id',
              align: 'center',
              width: 120,
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '牌型',
              align: 'center',
              width: 250,
              key: 'poker_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '玩家下注',
              key: 'player_bet_multiple',
              align: 'center',
              width: 120
            },
            // { title: "底注", key: "stakes", align: "center", width: 120 },
            {
              title: '结果',
              key: 'is_win',
              align: 'center',
              width: 120,
              render: (h, params) => {
                return params.row.is_win == true ? (
                  <span style="color:red">赢家</span>
                ) : (
                  ''
                )
              }
            },
            {
              title: '下注轮次',
              key: 'player_details',
              align: 'center',
              render: (h, params) => {
                if (params.row.player_details) {
                  let array = []
                  params.row.player_details.forEach((item) => {
                    if (item.is_start_contrast_card === true) {
                      array.push(
                        <tr>
                          <td>{item.round_no}</td>
                          <td>{item.bet}</td>
                          <td>{item.is_show_card ? '是' : ''}</td>
                          <td>{item.contrast_card_pos}</td>
                          <td>{item.contrast_card_result ? '赢' : '输'}</td>
                          <td>{item.is_dis_card ? '是' : ''}</td>
                          <td>{item.is_start_contrast_card ? '是' : ''}</td>
                        </tr>
                      )
                    } else {
                      array.push(
                        <tr>
                          <td>{item.round_no}</td>
                          <td>{item.bet}</td>
                          <td>{item.is_show_card ? '是' : ''}</td>
                          <td>{item.contrast_card_pos}</td>
                          <td>{''}</td>
                          <td>{item.is_dis_card ? '是' : ''}</td>
                          <td>{item.is_start_contrast_card ? '是' : ''}</td>
                        </tr>
                      )
                    }
                  })
                  return (
                    <table class="zjh-table">
                      <tr>
                        <th>轮数</th>
                        <th>该轮下注</th>
                        <th>是否明牌</th>
                        <th>比牌对象</th>
                        <th>比牌结果</th>
                        <th>是否弃牌</th>
                        <th>发起比牌</th>
                      </tr>
                      {array}
                    </table>
                  )
                }
              }
            }
          ]

          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Hlzjh)
          })
          break
        case 'Sg':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '牌型',
              align: 'center',
              width: 250,
              key: 'poker_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '牌类',
              key: 'card_type_name'
            },
            { title: '下注倍数', key: 'bet_multiple' },
            { title: '牌型倍数', key: 'card_type_multiple' },
            { title: '游戏分数', key: 'game_currency' }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Sg)
          })
          break
        case 'Brnn':
          this.bankerShow = true
          this.columnsBank = [
            {
              title: '庄家牌倍数',
              key: 'banker_card_type_multiple'
            },
            {
              title: '庄家牌型',
              key: 'banker_poker_cards',
              align: 'center',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.banker_poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            }
          ]
          this.columns = [
            {
              title: '玩家牌型倍数',
              key: 'card_type_multiple'
            },
            {
              title: '下注分数',
              key: 'bet'
            },
            {
              title: '牌名',
              key: 'name'
            },
            {
              title: '牌型',
              align: 'center',
              key: 'poker_cards',
              width: 440,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '结果',
              key: 'is_win',
              width: 150,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.bankerData.push(item.SettlementDetail.Brnn)
            this.detialData.push(...item.SettlementDetail.Brnn.details)
          })
          break
        case 'Tbnn':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位号',
              key: 'seat_id'
            },
            {
              title: '牌型倍数',
              key: 'card_type_multiple'
            },
            { title: '游戏分数', key: 'game_currency' },
            {
              title: '玩家下注倍数',
              key: 'player_bet_multiple'
            },
            {
              title: '牌名',
              key: 'card_type_name'
            },
            {
              title: '牌型',
              key: 'poker_cards',
              align: 'center',
              width: 330,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '结果',
              key: 'is_win',
              width: 150,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Tbnn)
          })
          break
        case 'Dzpk':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id',
              align: 'center',
              width: 100
            },
            {
              title: '座位号',
              key: 'seat_id',
              align: 'center'
            },
            {
              title: '公共牌',
              key: 'common_cards',
              align: 'center',
              width: 330,
              render: (h, params) => {
                let conts = []
                if (params.row.common_cards) {
                  conts.push(
                    params.row.common_cards.map((item) => {
                      return this.pokerCards(item.Card.Poker)
                    })
                  )
                  return <div>{conts}</div>
                }
                if (!params.row.common_cards && params.row.is_win) {
                  return <div>比牌前赢</div>
                }
              }
            },
            {
              title: '玩家手牌',
              key: 'hand_cards',
              align: 'center',
              width: 180,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.hand_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '帐变金额',
              key: 'delta_score',
              align: 'center'
            },
            {
              title: '第一轮',
              key: 'round1_record',
              align: 'center'
            },
            {
              title: '第二轮',
              key: 'round2_record',
              align: 'center'
            },
            {
              title: '第三轮',
              key: 'round3_record',
              align: 'center'
            },
            {
              title: '第四轮',
              key: 'round4_record',
              align: 'center'
            },
            {
              title: '小盲注',
              key: 'small_blinds',
              align: 'center',
              render: (h, params) => {
                return params.row.small_blinds == true ? <span>是</span> : ''
              }
            },
            {
              title: '大盲注',
              key: 'big_blinds',
              align: 'center',
              render: (h, params) => {
                return params.row.big_blinds == true ? <span>是</span> : ''
              }
            },
            {
              title: '总下注',
              key: 'total_record'
            },
            {
              title: '胜负',
              key: 'is_win',
              align: 'center',
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Dzpk)
          })
          break
        case 'Qznn':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id',
              align: 'center'
            },
            {
              title: '座位号',
              key: 'seat_id',
              align: 'center'
            },
            {
              title: '庄家下注倍数',
              key: 'banker_bet_multiple',
              align: 'center',
              render (h, params) {
                if (!params.row.player_bet_multiple) {
                  return <span>{params.row.banker_bet_multiple}</span>
                }
              }
            },
            {
              title: '牌型倍数',
              key: 'card_type_multiple',
              align: 'center'
            },
            { title: '游戏分数', key: 'game_currency', align: 'center' },
            {
              title: '玩家下注倍数',
              key: 'player_bet_multiple',
              align: 'center'
            },
            {
              title: '牌名',
              key: 'card_type_name',
              align: 'center'
            },
            {
              title: '牌型',
              key: 'poker_cards',
              align: 'center',
              width: 350,
              render: (h, params) => {
                let conts = []
                let ifisReplaceCard
                conts.push(
                  params.row.poker_cards.map((item) => {
                    let isReplaceCard
                    if (params.row.replace_card && !ifisReplaceCard) {
                      if (
                        item.Card.Poker.color ==
                          params.row.replace_card.Card.Poker.color &&
                        item.Card.Poker.number ==
                          params.row.replace_card.Card.Poker.number
                      ) {
                        isReplaceCard = true
                        ifisReplaceCard = true
                      }
                    }
                    return this.pokerCards(item.Card.Poker, isReplaceCard)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '结果',
              key: 'is_win',
              align: 'center',
              width: 150,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            },
            { title: '本局输赢金额', key: 'settle_score', align: 'center' }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qznn)
          })
          break
        case 'Qz21':
          this.bankerShow = true
          this.columnsBank = [
            {
              title: '庄家牌型',
              width: 300,
              align: 'center',
              key: 'banker_card_type'
            },
            {
              title: '庄家手牌',
              key: 'poker_cards',
              align: 'center',
              width: 500,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.poker_cards.map((item) => {
                    return this.pokerCards(item)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '庄家点数',
              align: 'center',
              key: 'banker_point'
            }
          ]
          this.columns = [
            {
              title: '座位',
              key: 'seat_id',
              align: 'center',
              width: 150,
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '牌型',
              align: 'center',
              width: 150,
              key: 'banker_card_type',
              render: (h, params) => {
                return (
                  <span>
                    {params.row.player_card_group_infos[0].banker_card_type}
                    {params.row.player_card_group_infos[1] ? (
                      <p>
                        {params.row.player_card_group_infos[1].banker_card_type}
                      </p>
                    ) : (
                      ''
                    )}
                  </span>
                )
              }
            },
            {
              title: '手牌',
              align: 'center',
              width: 500,
              key: 'player_card_group_infos',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.player_card_group_infos[0].poker_cards.map(
                    (item) => {
                      return this.pokerCards(item)
                    }
                  ),
                  params.row.player_card_group_infos[1] ? (
                    <div>
                      {params.row.player_card_group_infos[1].poker_cards.map(
                        (item) => {
                          return this.pokerCards(item)
                        }
                      )}
                    </div>
                  ) : (
                    ''
                  )
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '玩家点数',
              key: 'banker_point',
              align: 'center',
              render: (h, params) => {
                return (
                  <span>
                    {params.row.player_card_group_infos[0].banker_point}
                    {params.row.player_card_group_infos[1] ? (
                      <p>
                        {params.row.player_card_group_infos[1].banker_point}
                      </p>
                    ) : (
                      ''
                    )}
                  </span>
                )
              }
            },
            { title: '底注', align: 'center', key: 'original_bet' },
            { title: '保险分', align: 'center', key: 'insurance' }
          ]
          this.bankerData.push(
            this.rowInfo[0].SettlementDetail.Qz21.banker_card_group_info
          )
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qz21.player_details)
          })
          break
        case 'Qzpj':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '角色',
              key: 'is_banker',
              render (h, params) {
                return params.row.is_banker ? <span>庄家</span> : ''
              }
            },
            {
              title: '游戏币',
              key: 'game_currency'
            },
            {
              title: '牌型',
              align: 'center',
              width: 300,
              key: 'pai_jiu_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.pai_jiu_cards.map((item) => {
                    return (
                      <button class="paijiu">
                        上&nbsp;{item.Card.PaiJiu.top_number} 下&nbsp;
                      {item.Card.PaiJiu.bottom_number}
                      </button>
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            { title: '点数', key: 'card_type' },
            { title: '倍数', key: 'bet_multiple' },
            { title: '输赢', key: 'change_game_currency' },
            {
              title: '结果',
              key: 'is_win',
              width: 150,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qzpj)
          })
          break
        case 'Slwh':
          this.columns = [
            {
              title: '开奖倍数',
              key: 'lottery_point_multiple'
            },
            {
              title: '开奖名',
              key: 'lottery_point_name'
            },
            {
              title: '玩家下注',
              key: 'total_bet'
            },
            {
              title: '结果',
              key: 'winning_point_names',
              width: 200,
              render: (h, params) => {
                if (
                  this.bankerData[0].includes(params.row.lottery_point_name)
                ) {
                  return (
                    <span>
                      {params.row.lottery_point_name}{' '}
                      {
                        ['', '大三元', '大四喜', '送灯', '霹雳闪电'][
                          this.slwhmode
                        ]
                      }
                    </span>
                  )
                } else {
                  return ''
                }
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.bankerData.push(
              item.SettlementDetail.Slwh.winning_point_names
            )
            this.detialData.push(...item.SettlementDetail.Slwh.bet_details)
            if (item.SettlementDetail.Slwh.mode) {
              this.slwhmode = item.SettlementDetail.Slwh.mode
            } else {
              this.slwhmode = 0
            }
          })
          break
        case 'Jsys':
          this.columns = [
            {
              title: '开奖名',
              key: 'lottery_point_name'
            },
            {
              title: '开奖倍数',
              key: 'lottery_point_multiple'
            },
            {
              title: '玩家下注',
              key: 'total_bet'
            },
            {
              title: '结果',
              key: 'winning_point_names',
              render: (h, params) => {
                return params.row.lottery_point_name == this.bankerData[0] ? (
                  <span>{params.row.lottery_point_name}</span>
                ) : (
                  ''
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.bankerData.push(
              item.SettlementDetail.Jsys.winning_point_names
            )
            this.detialData.push(...item.SettlementDetail.Jsys.bet_details)
          })
          break
        case 'Hcpy':
          this.columns = [
            { title: '奔驰', key: 'benz_bet', align: 'center' },
            { title: '宝马', key: 'bmw_bet', align: 'center' },
            { title: '法拉利', key: 'ferrari_bet', align: 'center' },
            { title: '捷豹', key: 'jaguar_bet', align: 'center' },
            { title: '兰博基尼', key: 'lamborghini_bet', align: 'center' },
            { title: '路虎', key: 'landRover_bet', align: 'center' },
            { title: '玛莎拉蒂', key: 'maserati_bet', align: 'center' },
            { title: '保时捷', key: 'porsche_bet', align: 'center' },
            {
              title: '开奖结果',
              key: 'lottery_results',
              align: 'center',
              width: 280,
              render: (h, params) => {
                if (
                  !params.row.lottery_results ||
                  params.row.lottery_results.length === 0
                ) {
                  return <div>吃灯</div>
                } else {
                  return <div>{params.row.lottery_results.join(',')}</div>
                }
              }
            }
          ]

          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Hcpy)
          })
          break
        case 'Qhb':
          this.columns = [
            {
              title: '玩家ID',
              key: 'id',
              align: 'center',
              render: (h, params) => {
                return params.row.id == this.rowId ? (
                  <span>{params.row.id}（玩家）</span>
                ) : (
                  <span>{params.row.id}</span>
                )
              }
            },
            { title: '抢包金额', key: 'score', align: 'center' },
            { title: '抽水', key: 'pump', align: 'center' },
            {
              title: '本轮发包',
              key: 'dist_score',
              align: 'center',
              render: (h, params) => {
                return Number(params.row.dist_score) ? (
                  <span>{params.row.dist_score}</span>
                ) : null
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qhb)
          })
          break
        case 'Hhdz':
          this.columns = [
            {
              title: '黑方下注详情',
              key: 'black_detail',
              align: 'left',
              render: (h, params) => {
                let data = params.row.black_detail
                return (
                  <dl>
                    <dt>牌型倍数：{data.card_type_multiple}</dt>
                    <dt>
                      牌型：
                    {data.cards.map((item) => {
                      return this.pokerCards(item.Card.Poker)
                    })}
                    </dt>
                    <dt>玩家下注：{data.player_bet}</dt>
                  </dl>
                )
              }
            },
            {
              title: '红方下注详情',
              key: 'red_detail',
              align: 'left',
              render: (h, params) => {
                let data = params.row.red_detail
                return (
                  <dl>
                    <dt>牌型倍数：{data.card_type_multiple}</dt>
                    <dt>
                      牌型：
                    {data.cards.map((item) => {
                      return this.pokerCards(item.Card.Poker)
                    })}
                    </dt>
                    <dt>玩家下注：{data.player_bet}</dt>
                  </dl>
                )
              }
            },
            {
              title: '幸运一击下注',
              key: 'player_lucky_hit_bet',
              align: 'center'
            },
            {
              title: '结果',
              key: 'is_red_win',
              align: 'center',
              render: (h, params) => {
                let [black, red] = [params.row.black_detail, params.row.red_detail]
                if (this.isPaohuoliantian(red, black)) {
                  return (<span style="color:red">炮火连天,<b>庄家通杀</b>,返还玩家<b>20%</b>下注</span>)
                } else {
                  return params.row.is_red_win ? (
                    <span>红方胜</span>
                  ) : (
                    <span>黑方胜</span>
                  )
                }
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Hhdz)
          })
          break
        case 'Ebg':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '身份',
              key: 'is_banker',
              width: 120,
              render: (h, params) => {
                return params.row.is_banker == true ? <span>庄家</span> : ''
              }
            },
            {
              title: '游戏币',
              key: 'game_currency'
            },

            {
              title: '牌型',
              align: 'center',
              key: 'ma_jiang_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.ma_jiang_cards.map((item) => {
                    return (
                      {
                        2: (
                          <button class="majiang" style="letter-spacing: 3px;">
                            {item.Card.MaJiang.number}筒
                          </button>
                        ),
                        10: <button class="majiang">白板</button>
                      }[item.Card.MaJiang.cover] || ''
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '点数',
              key: 'card_type'
            },
            {
              title: '房间底注',
              key: 'stakes'
            },
            { title: '倍数', key: 'multiple' },
            {
              title: '结果',
              key: 'is_win',
              width: 120,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Ebg)
          })
          break
        case 'Wrttz':
          this.columns = [
            {
              title: '位置',
              key: 'name'
            },
            {
              title: '倍数',
              key: 'card_type_multiple'
            },
            {
              title: '玩家下注',
              key: 'bet'
            },
            {
              title: '牌型',
              align: 'center',
              key: 'poker_cards',
              render: (h, params) => {
                let conts = []
                if (params.row.poker_cards) {
                  conts.push(
                    params.row.poker_cards.map((item) => {
                      return (
                        {
                          2: (
                            <button
                              class="majiang"
                              style="letter-spacing: 3px;"
                            >
                              {item.Card.MaJiang.number}筒
                            </button>
                          ),
                          10: <button class="majiang">白板</button>
                        }[item.Card.MaJiang.cover] || ''
                      )
                    })
                  )
                  return <div>{conts}</div>
                }
              }
            },
            {
              title: '结果',
              key: 'is_win',
              align: 'center',
              render: (h, params) => {
                return params.row.is_win == true ? <span>下注赢</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData[0] = {
              name: '庄家',
              bet: '--',
              card_type_multiple: '--',
              poker_cards: item.SettlementDetail.Wrttz.banker_poker_cards
            }
            this.detialData.push(...item.SettlementDetail.Wrttz.details)
          })
          break
        case 'Bsxxl':
          this.columns = [
            { title: '投注大小', key: 'bet', align: 'center' },
            { title: '基础投注', key: 'base_bet', align: 'center' },
            { title: '投注总额', key: 'total_bet', align: 'center' },
            {
              title: '消除记录',
              key: 'player_details',
              align: 'center',
              render: (h, params) => {
                if (params.row.player_details) {
                  let baoshi = params.row.player_details.map((item) => {
                    let win = item.win_game_currency
                      ? ' 得分：' + item.win_game_currency
                      : ''

                    let model = item.gem_model
                      ? '/color/baoshi_lv' + item.gem_model + '.png'
                      : '/color/baoshi_lv0.png'
                    return (
                      <div style="line-height:28px">
                        <img
                          src={model}
                          height="26"
                          style="vertical-align: middle;"
                        />
                        {' × ' + item.gem_count + win}
                      </div>
                    )
                  })
                  return baoshi
                } else {
                  return <span>无消除</span>
                }
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Bsxxl)
          })
          break
        case 'Sss':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              width: 90,
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id} (玩家)</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '底注',
              key: 'bet_score'
            },
            {
              title: '牌列表',
              key: 'card_list',
              width: 475,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.card_list.map((item) => {
                    return (
                      <div>
                        <span>位置:{item.location}</span>
                        <span> 牌型:{item.card_type}</span>&emsp;
                        {item.cards.map((items) => {
                          return this.pokerCards(items.Card.Poker)
                        })}
                      </div>
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '位置比牌信息',
              key: 'result_list',
              width: 300,
              render: (h, params) => {
                return params.row.result_list.map((item) => {
                  return (
                    <div>
                      <span>位置：{item.location}</span>
                      <span> 基本水数：{item.base_num}</span>
                      <span> 额外水数：{item.extra_num}</span>
                    </div>
                  )
                })
              }
            },
            {
              title: '打枪信息',
              key: 'shoot_list',
              width: 200,
              render: (h, params) => {
                if (params.row.shoot_list) {
                  return params.row.shoot_list.map((item) => {
                    return (
                      <div class="sssys">
                        <div>打枪玩家：{item.shoot_player_id}</div>
                        <div> 击中玩家：{item.shoot_hit_player_id}</div>
                      </div>
                    )
                  })
                }
              }
            },
            { title: '输赢', key: 'delta' },
            {
              title: '结果',
              key: 'isWin',
              render: (h, params) => {
                return params.row.isWin == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Sss)
          })
          break
        case 'SssJs':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '底注',
              key: 'bet_score'
            },
            {
              title: '牌列表',
              key: 'card_list',
              width: 435,
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.card_list.map((item) => {
                    return (
                      <div>
                        <span>位置:{item.location}</span>
                        <span> 牌型:{item.card_type}</span>&emsp;
                        {item.cards.map((items) => {
                          return this.pokerCards(items.Card.Poker)
                        })}
                      </div>
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '位置比牌信息',
              key: 'result_list',
              width: 280,
              render: (h, params) => {
                return params.row.result_list.map((item) => {
                  return (
                    <div>
                      <span>位置：{item.location}</span>
                      <span> 基本水数：{item.base_num}</span>
                      <span> 额外水数：{item.extra_num}</span>
                    </div>
                  )
                })
              }
            },
            {
              title: '打枪信息',
              key: 'shoot_list',
              width: 200,
              render: (h, params) => {
                if (params.row.shoot_list) {
                  return params.row.shoot_list.map((item) => {
                    return (
                      <div class="sssys">
                        <div>打枪玩家：{item.shoot_player_id}</div>
                        <div> 击中玩家：{item.shoot_hit_player_id}</div>
                      </div>
                    )
                  })
                }
              }
            },
            { title: '输赢', key: 'delta' },
            {
              title: '结果',
              key: 'isWin',
              render: (h, params) => {
                return params.row.isWin == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.SssJs)
          })
          break
        case 'Qzxsz':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '角色',
              key: 'is_banker',
              render (h, params) {
                return params.row.is_banker ? <span>庄家</span> : ''
              }
            },
            {
              title: '选牌',
              align: 'center',
              width: 250,
              key: 'select_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.select_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            { title: '牌型', key: 'card_type' },
            { title: '倍数', key: 'card_type_multiple' },
            { title: '输赢分数', key: 'change_game_currency' },
            {
              title: '场馆',
              key: 'is_jin_hua',
              render: (h, params) => {
                return params.row.is_jin_hua == true ? (
                  <span>金花场</span>
                ) : (
                  <span>三公场</span>
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qzxsz)
          })
          break
        case 'qzxszsg':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '庄闲',
              key: 'game_currency'
            },
            {
              title: '初始牌',
              align: 'center',
              width: 200,
              key: 'pai_jiu_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.pai_jiu_cards.map((item) => {
                    return (
                      <button class="paijiu">
                        上&nbsp;{item.Card.PaiJiu.top_number} 下&nbsp;
                      {item.Card.PaiJiu.bottom_number}
                      </button>
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '选牌',
              align: 'center',
              width: 200,
              key: 'pai_jiu_cards',
              render: (h, params) => {
                let conts = []
                conts.push(
                  params.row.pai_jiu_cards.map((item) => {
                    return (
                      <button class="paijiu">
                        上&nbsp;{item.Card.PaiJiu.top_number} 下&nbsp;
                      {item.Card.PaiJiu.bottom_number}
                      </button>
                    )
                  })
                )
                return <div>{conts}</div>
              }
            },
            { title: '牌型', key: 'bet_multiple' },
            { title: '倍数', key: 'bet_multiple' },
            { title: '分数', key: 'bet_multiple' },
            {
              title: '结果',
              key: 'is_win',
              width: 150,
              render: (h, params) => {
                return params.row.is_win == true ? <span>赢家</span> : ''
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Qzpj)
          })
          break
        case 'Ermj':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '初始牌',
              key: 'game_currency'
            },
            {
              title: '打出牌',
              key: 'game_currency'
            },
            {
              title: '花牌',
              key: 'game_currency'
            },
            {
              title: '最终牌',
              key: 'game_currency'
            },
            {
              title: '输赢',
              key: 'game_currency'
            },
            {
              title: '分数',
              key: 'game_currency'
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Ermj)
          })
          break
        case 'Ttby':
          this.columns = [
            { title: '子弹数', key: 'number', align: 'center' },
            { title: '流水号', key: 'key0', align: 'center', width: 180 },
            { title: '子弹价值', key: 'key1', align: 'center', width: 100 },
            { title: '盈亏', key: 'key2', align: 'center' },
            { title: '场景倍数', key: 'key3', align: 'center' },
            {
              title: '射击时间',
              key: 'key4',
              align: 'center',
              width: 200,
              render: (h, params) => {
                return (
                  <span>
                    {new Date(params.row.key4 * 1000).toLocaleString()}
                  </span>
                )
              }
            },
            // { title: "鱼逻辑Id", key: "key5", align: "center" },
            {
              title: '鱼类型',
              key: 'key9',
              align: 'center',
              render: (h, params) => {
                return (
                  <div style="position: relative; height: 26px">
                    {Number(params.row.key7) >= 300 &&
                    Number(params.row.key7) < 400 ? (
                        <span>
                          <img
                            style={{
                              position: 'absolute',
                              zIndex: 0,
                              left: '50%',
                              transform: 'translateX(-50%) scale(1.5)'
                            }}
                            src={'/color/Ttby/huaquan.png'}
                            height="26"
                          />
                          <img
                            style={{
                              zIndex: 1,
                              position: 'relative'
                            }}
                            src={
                              '/color/Ttby/fish_type' +
                            (params.row.key9 == '0'
                              ? params.row.key7
                              : params.row.key9) +
                            '.png'
                            }
                            height="26"
                          />
                        </span>
                      ) : Number(params.row.key7) >= 400 &&
                      Number(params.row.key7) < 500 ? (
                          <img
                            src={'/color/Ttby/fish_type' + params.row.key7 + '.png'}
                            height="26"
                          />
                        ) : Number(params.row.key7) >= 500 &&
                      Number(params.row.key7) < 600 ? (
                            <img
                              src={'/color/Ttby/fish_type' + params.row.key7 + '.png'}
                              height="26"
                            />
                          ) : Number(params.row.key7) >= 600 &&
                      Number(params.row.key7) < 700 ? (
                              <img src={'/color/Ttby/fish_type47.png'} height="26" />
                            ) : Number(params.row.key7) >= 200 &&
                      Number(params.row.key7) < 300 ? (
                                <span>
                                  <img
                                    style={{
                                      position: 'absolute',
                                      left: '50%',
                                      transform: 'translateX(-50%)'
                                    }}
                                    src={'/color/Ttby/quan.png'}
                                    height="26"
                                  />
                                  <img
                                    src={
                                      '/color/Ttby/fish_type' +
                            (params.row.key9 == '0'
                              ? params.row.key7
                              : params.row.key9) +
                            '.png'
                                    }
                                    height="26"
                                  />
                                </span>
                              ) : (
                                <img
                                  src={
                                    '/color/Ttby/fish_type' +
                          (params.row.key9 == '0'
                            ? params.row.key7
                            : params.row.key9) +
                          '.png'
                                  }
                                  height="26"
                                />
                              )}
                  </div>
                )
              }
            },
            // {
            //   title: "特殊",
            //   key: "key7",
            //   align: "center",
            //   render: (h, params) => {
            //     let effect = Number(params.row.key7);
            //     if (effect > 700 && effect <= 800) {
            //       return <div>全屏冰冻</div>;
            //     }
            //     if (effect > 600 && effect <= 700) {
            //       return <div>炸弹</div>;
            //     }
            //     if (effect > 500 && effect <= 600) {
            //       return <div>大四喜</div>;
            //     }
            //     if (effect > 400 && effect <= 500) {
            //       return <div>大三元</div>;
            //     }
            //     if (effect > 300 && effect <= 400) {
            //       return <div>一网打尽</div>;
            //     }
            //     if (effect > 200 && effect <= 300) {
            //       return <div>奖金鱼</div>;
            //     }
            //     if (effect > 100 && effect <= 200) {
            //       return <div>龙龟特效</div>;
            //     }
            //     if (effect <= 100) {
            //       return <div>无</div>;
            //     }
            //   },
            // },
            {
              title: '命中',
              key: 'key8',
              align: 'center',
              render: (h, params) => {
                let model =
                  params.row.key8 == 'true'
                    ? '/color/xc_yy.png'
                    : '/color/xc_nn.png'
                return (
                  <div style="line-height:28px">
                    <img
                      src={model}
                      height="26"
                      style="vertical-align: middle;"
                    />
                  </div>
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            if (item.SettlementDetail.Ttby.result) {
              item.SettlementDetail.Ttby.result
                .split(',')
                .map((items, index) => {
                  let con = items.split('|')
                  let don = { number: index + 1 }
                  for (let i in con) {
                    don['key' + i] = con[i]
                  }
                  this.detialData.push(don)
                })
            }
          })
          break
        case 'Mrby':
          this.columns = [
            { title: '子弹数', key: 'number', align: 'center' },
            { title: '流水号', key: 'key0', align: 'center', width: 180 },
            { title: '子弹价值', key: 'key1', align: 'center', width: 100 },
            { title: '盈亏', key: 'key2', align: 'center' },
            { title: '场景倍数', key: 'key3', align: 'center' },
            {
              title: '射击时间',
              key: 'key4',
              align: 'center',
              width: 200,
              render: (h, params) => {
                return (
                  <span>
                    {new Date(params.row.key4 * 1000).toLocaleString()}
                  </span>
                )
              }
            },
            // { title: "鱼逻辑Id", key: "key5", align: "center" },
            {
              title: '鱼类型',
              key: 'key9',
              align: 'center',
              render: (h, params) => {
                return (
                  <div style="position: relative; height: 26px">
                    {Number(params.row.key6) == 1002 ? (
                      <span>美人鱼王宝箱</span>
                    ) : Number(params.row.key6) == 1001 ? (
                      <span>幸运宝箱</span>
                    ) : Number(params.row.key7) >= 300 &&
                      Number(params.row.key7) < 400 ? (
                        <img
                          src={
                            '/color/Mrby/fish_type' +
                          (params.row.key9 == '0'
                            ? params.row.key7
                            : params.row.key9) +
                          '.png'
                          }
                          height="26"
                        />
                      ) : Number(params.row.key7) >= 400 &&
                      Number(params.row.key7) < 500 ? (
                          <img
                            src={'/color/Mrby/fish_type' + params.row.key7 + '.png'}
                            height="26"
                          />
                        ) : Number(params.row.key7) >= 500 &&
                      Number(params.row.key7) < 600 ? (
                            <img
                              src={'/color/Mrby/fish_type' + params.row.key7 + '.png'}
                              height="26"
                            />
                          ) : Number(params.row.key7) >= 600 &&
                      Number(params.row.key7) < 700 ? (
                              <img src={'/color/Mrby/fish_type47.png'} height="26" />
                            ) : Number(params.row.key7) >= 200 &&
                      Number(params.row.key7) < 300 ? (
                                <span>
                                  <img
                                    style={{
                                      position: 'absolute',
                                      left: '50%',
                                      transform: 'translateX(-50%)'
                                    }}
                                    src={'/color/Mrby/quan.png'}
                                    height="26"
                                  />
                                  <img
                                    src={
                                      '/color/Mrby/fish_type' +
                            (params.row.key9 == '0'
                              ? params.row.key7
                              : params.row.key9) +
                            '.png'
                                    }
                                    height="26"
                                  />
                                </span>
                              ) : (
                                <img
                                  src={
                                    '/color/Mrby/fish_type' +
                          (params.row.key9 == '0'
                            ? params.row.key7
                            : params.row.key9) +
                          '.png'
                                  }
                                  height="26"
                                />
                              )}
                  </div>
                )
              }
            },
            // {
            //   title: "特殊",
            //   key: "key7",
            //   align: "center",
            //   render: (h, params) => {
            //     let effect = Number(params.row.key7);
            //     if (effect > 700 && effect <= 800) {
            //       return <div>全屏冰冻</div>;
            //     }
            //     if (effect > 600 && effect <= 700) {
            //       return <div>炸弹</div>;
            //     }
            //     if (effect > 500 && effect <= 600) {
            //       return <div>大四喜</div>;
            //     }
            //     if (effect > 400 && effect <= 500) {
            //       return <div>大三元</div>;
            //     }
            //     if (effect > 300 && effect <= 400) {
            //       return <div>一网打尽</div>;
            //     }
            //     if (effect > 200 && effect <= 300) {
            //       return <div>奖金鱼</div>;
            //     }
            //     if (effect > 100 && effect <= 200) {
            //       return <div>龙龟特效</div>;
            //     }
            //     if (effect <= 100) {
            //       return <div>无</div>;
            //     }
            //   },
            // },
            {
              title: '命中',
              key: 'key8',
              align: 'center',
              render: (h, params) => {
                let model =
                  params.row.key8 == 'true'
                    ? '/color/xc_yy.png'
                    : '/color/xc_nn.png'
                return (
                  <div style="line-height:28px">
                    <img
                      src={model}
                      height="26"
                      style="vertical-align: middle;"
                    />
                  </div>
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            if (item.SettlementDetail.Mrby.result) {
              item.SettlementDetail.Mrby.result
                .split(',')
                .map((items, index) => {
                  let con = items.split('|')
                  let don = { number: index + 1 }
                  for (let i in con) {
                    don['key' + i] = con[i]
                  }
                  this.detialData.push(don)
                })
            }
          })
          break
        case 'Sgj':
          this.columns = [
            {
              title: '全部下注',
              key: 'bet_details',
              render: (h, params) => {
                return params.row.bet_details.map((item) => {
                  return (
                    <div class="sgjxz">
                      <label>
                        <span>名称：</span>
                        {item.lottery_point_name}
                      </label>
                      <label>
                        <span> 倍数：</span>
                        {item.lottery_point_multiple}{' '}
                      </label>{' '}
                      <label>
                        <span>下注：</span>
                        {item.total_bet}
                      </label>
                    </div>
                  )
                })
              }
            },
            {
              title: '中奖下注',
              key: 'winning_point_names',
              render: (h, params) => {
                return (
                  <div class="sgjxz">
                    {params.row.winning_point_names.split(',').map((item) => {
                      return (
                        <span style="min-width:120px;text-align:left">
                          {item}
                        </span>
                      )
                    })}
                  </div>
                )
              }
            },
            {
              title: '特殊中奖名',
              key: 'winning_special_point_names',
              width: 200
            },
            {
              title: '猜大小',
              key: 'winning_special_point_names',
              width: 150,
              render: (h, params) => {
                return (
                  <div>{params.row.guess_size_win ? params.row.guess_size_win : 0}/{params.row.guess_size_loss ? params.row.guess_size_loss : 0} </div>
                )
              }
            },
            {
              title: '输赢总分',
              key: 'total_score',
              width: 200
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Sgj)
          })
          break
        case 'Lhd':
          this.columns = [
            {
              title: '玩家ID',
              key: 'player_id'
            },
            {
              title: '座位',
              key: 'seat_id',
              render: (h, params) => {
                return params.row.player_id == this.rowId ? (
                  <span>{params.row.seat_id}（玩家）</span>
                ) : (
                  <span>{params.row.seat_id}</span>
                )
              }
            },
            {
              title: '龙牌',
              key: 'long_card',
              render: (h, params) => {
                let conts = []
                conts.push(this.pokerCards(params.row.long_card.Card.Poker))
                return <div>{conts}</div>
              }
            },
            {
              title: '虎牌',
              key: 'hu_card',
              render: (h, params) => {
                let conts = []
                conts.push(this.pokerCards(params.row.hu_card.Card.Poker))
                return <div>{conts}</div>
              }
            },
            {
              title: '龙下注',
              render: (h, params) => {
                if (params.row.details) {
                  let item = params.row.details[0]
                  return (
                    <span>
                      {item.bet_score}
                      {item.is_win ? <label class="lhd-win">赢</label> : ''}
                    </span>
                  )
                }
              }
            },
            {
              title: '虎下注',
              render: (h, params) => {
                if (params.row.details) {
                  let item = params.row.details[1]
                  return (
                    <span>
                      {item.bet_score}
                      {item.is_win ? <label class="lhd-win">赢</label> : ''}
                    </span>
                  )
                }
              }
            },
            {
              title: '和下注',
              render: (h, params) => {
                if (params.row.details) {
                  let item = params.row.details[2]
                  return (
                    <span>
                      {item.bet_score}
                      {item.is_win ? <label class="lhd-win">赢</label> : ''}
                    </span>
                  )
                }
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Lhd)
          })
          break
        case 'Bjl':
          this.columns = [
            {
              title: '庄牌型',
              key: 'banker_cards',
              align: 'center',
              minWidth: 240,
              render: (h, params) => {
                if (params.index == 1) {
                  return <div>当前玩家ID</div>
                }

                let conts = []
                conts.push(
                  params.row.banker_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            },
            {
              title: '闲牌型',
              align: 'center',
              minWidth: 240,
              key: 'player_cards',
              render: (h, params) => {
                if (params.index == 1) {
                  return <div>{this.rowId}</div>
                }
                let conts = []
                conts.push(
                  params.row.player_cards.map((item) => {
                    return this.pokerCards(item.Card.Poker)
                  })
                )
                return <div>{conts}</div>
              }
            }
          ]
          let customColumns = ['闲', '庄', '闲对', '庄对', '大', '和', '小']

          if (iplatform == '--kaiyuan') {
            customColumns = ['闲', '庄', '闲对', '庄对', '和']
          }

          customColumns.map((title, tindex) => {
            this.columns.push({
              title,
              align: 'center',
              width: 100,
              render: (h, params) => {
                if (params.index == 1) {
                  return <div>{params.row.player_bet[tindex]}</div>
                }
                return (
                  <div>
                    {params.row.lottery_points.includes(tindex) ? '✔' : ''}
                  </div>
                )
              }
            })
          })

          this.detialData.push(this.rowInfo[0].SettlementDetail.Bjl)
          this.detialData.push(this.rowInfo[0].SettlementDetail.Bjl)
          break
        case 'Jcby':
          this.columns = [
            { title: '鱼编号', key: 'key7', align: 'center' },
            {
              title: '鱼',
              key: 'key1',
              align: 'center',
              width: 180,
              render: (h, params) => {
                let uid = params.row.key5 * 1
                // obj是鱼编号和图片的对应数据
                let obj = {
                  1: 0,
                  2: 1,
                  3: 2,
                  4: 3,
                  5: 4,
                  6: 5,
                  7: 6,
                  8: 7,
                  9: 8,
                  10: 9,
                  11: 10,
                  12: 11,
                  13: 12,
                  14: 13,
                  15: 14,
                  16: 15,
                  17: 16,
                  18: 17,
                  19: 18,
                  20: 19,
                  21: 20,
                  22: 21,
                  23: 22,
                  24: 22,
                  25: 22,
                  26: 23,
                  27: 24,
                  28: 25,
                  29: 26,
                  30: 27,
                  31: 28,
                  32: 29,
                  33: 30,
                  34: 33,
                  35: 34,
                  36: 35,
                  37: 36,
                  38: 37,
                  39: 38,
                  40: 39,
                  102: 102,
                  103: 103,
                  104: 104
                  // 102是红包   103是金蛋
                }
                let names = {
                  23: '镇妖金刚塔',
                  24: '如意金箍棒',
                  25: '无敌风火轮'
                }
                let name =
                  uid === 23 || uid === 24 || uid === 25 ? names[uid] : ''

                if (params.row.key5.includes('-')) {
                  return (
                    <span>
                      <img
                        src={`/icons/gameDetail_PTG0045_104.png`}
                        height="30"
                      />
                      <img
                        src={`/icons/gameDetail_PTG0045_${
                          obj[params.row.key5.split('-')[1]]
                        }.png`}
                        height="30"
                      />
                    </span>
                  )
                }

                return (
                  <span>
                    <img
                      src={`/icons/gameDetail_PTG0045_${obj[uid]}.png`}
                      height="30"
                    />
                    <span style="position: relative; top: -10px; left: -15px">
                      {name}
                    </span>
                  </span>
                )
              }
            },
            {
              title: '消耗子弹',
              key: 'key1',
              align: 'center',
              width: 100,
              render: (h, params) => {
                return (
                  <span>
                    {(params.row.key1 && (params.row.key1 * 1).toFixed(2)) || 0}
                  </span>
                )
              }
            },
            {
              title: '奖金',
              key: 'key2',
              align: 'center',
              render: (h, params) => {
                if (Number(params.row.key2) <= 0) {
                  return <span>0</span>
                }
                return (
                  <span>
                    {(params.row.key2 && (params.row.key2 * 1).toFixed(2)) || 0}
                  </span>
                )
              }
            },
            {
              title: '盈亏',
              key: 'key3',
              align: 'center',
              render: (h, params) => {
                return (
                  <div>
                    {(params.row.key2 * 1 - params.row.key1 * 1 &&
                      (params.row.key2 * 1 - params.row.key1 * 1).toFixed(2)) ||
                      0}
                  </div>
                )
              }
            }
          ]
          let ar = []
          this.rowInfo.map((item) => {
            if (item.SettlementDetail.Jcby.result) {
              item.SettlementDetail.Jcby.result
                .split(',')
                .map((items, index) => {
                  let con = items.split('|')
                  let don = { number: index + 1 }
                  for (let i in con) {
                    don['key' + i] = con[i]
                  }
                  ar.push(don)
                })
            }
          })

          var combineObjectInList = function (arr, item, list) {
            // 数组去除重复，item为重复判定项，list为重复记录需要累加的值的数组
            var obj = {}
            var a = []
            for (var i in arr) {
              if (!obj[arr[i][item]]) {
                obj[arr[i][item]] = copyObj(arr[i]) // 数组克隆
              } else if (obj[arr[i][item]]) {
                for (var j in list) {
                  obj[arr[i][item]][list[j]] =
                    parseFloat(obj[arr[i][item]][list[j]]) +
                    parseFloat(arr[i][list[j]])
                }
              }
            }
            for (var k in obj) {
              a.push(obj[k])
            }
            return a
          }
          var copyObj = function (obj) {
            // obj arr 对象的克隆（区分于指针赋值）
            if (obj.constructor == Array) {
              var a = []
              for (var i in obj) {
                a.push(obj[i])
              }
              return a
            } else {
              var o = {}
              for (var key in obj) {
                o[key] = obj[key]
              }
              return o
            }
          }
          this.detialData = combineObjectInList(ar, 'key5', ['key1', 'key2'])
          // 显示宝箱
          // if (this.detialData[0].key6 === undefined) {
          //   ar.sort(
          //     (a, b) => a.key2 * 10000 + a.key0 - (b.key2 * 10000 + b.key0)
          //   );
          //   this.detialData = ar;
          //   this.columns = [
          //     { title: "击杀时使用的炮倍", key: "key0", align: "center" },
          //     { title: "次数", key: "key1", align: "center" },
          //     { title: "等级", key: "key2", align: "center" },
          //     { title: "倍数", key: "key4", align: "center" },
          //     { title: "分数", key: "key5", align: "center" },
          //   ];
          // }

          break
        case 'Xydb':
          this.columns = [
            {
              title: '期数',
              key: 'round',
              align: 'center'
            },
            {
              title: '中奖号码',
              key: 'award_num',
              align: 'center'
            },
            {
              title: '奖金',
              key: 'detla',
              align: 'center'
            },
            {
              title: '下注玩家ID',
              key: 'player_id',
              align: 'center'
            },
            {
              title: '下注金额',
              key: 'bet',
              align: 'center'
            },
            {
              title: '占比',
              key: 'percentage',
              align: 'center',
              render: (h, params) => {
                return <span>{params.row.percentage * 1000}‰</span>
              }
            },
            {
              title: '是否中奖',
              key: 'is_lottery',
              align: 'center',
              render: (h, params) => {
                return <span>{params.row.is_lottery ? '是' : '否'}</span>
              }
            },
            {
              title: '购买号码',
              type: 'expand',
              key: 'lottery_num',
              align: 'center',
              render: (h, params) => {
                return (
                  <div style="display: flex;flex-wrap: wrap;">
                    {params.row.lottery_num.map((num) => {
                      return <span style="padding:10px">{num}</span>
                    })}
                  </div>
                )
              }
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Xydb)
          })
          break
        case 'Jdhb':
          this.columns = [
            {
              title: '用户ID',
              key: 'id',
              align: 'center'
            },
            {
              title: '盈亏',
              key: 'profitLoss',
              align: 'center'
            },
            {
              title: '发包金额',
              key: 'dist_score',
              align: 'center'
            },
            {
              title: '抢包金额',
              key: 'socre',
              align: 'center'
            },
            {
              title: '未抢退分',
              key: 'refundAmount',
              align: 'center'
            }
          ]
          this.rowInfo.map((item) => {
            this.detialData.push(item.SettlementDetail.Jdhb)
          })
          break
      }
    }
  }
}
</script>

<style>
.gogogo .ivu-table-cell {
  overflow: initial;
}

.ivu-table-border td {
  height: 54px;
}

.color-style-red {
  white-space: nowrap;
  color: red;
  font-size: 20px;
}
.color-style-black {
  white-space: nowrap;
  color: black;
  font-size: 20px;
}
.bets-header {
  line-height: 32px;
  font-weight: bold;
  border-bottom: 1px solid #988;
}
.bets-table > div {
  border-bottom: 1px solid #ccc;
  line-height: 32px;
  transition: 0.3s;
}
.bets-table > div:hover {
  background-color: antiquewhite;
  font-weight: bold;
}
.paijiu {
  background-color: #333;
  margin-right: 10px;
  color: #fff;
  padding: 2px 5px;
}
.majiang {
  background-color: #333;
  margin-right: 10px;
  color: #fff;
  padding: 2px 5px;
}
.zjh-table {
  width: 100%;
  border-left: 1px solid #ddd;
}
.zjh-table tr th,
.zjh-table tr td {
  text-align: center;
}
.ivu-table-body .ivu-table-row:not(:first-child) .zjh-table tr:first-child {
  display: none;
}
.sgjxz label {
  display: inline-block;
  min-width: 140px;
}
.sgjxz span {
  display: inline-block;
  min-width: 70px;
  font-weight: 600;
  text-align: right;
}
.sssys {
  border-bottom: gray 1px solid;
}
.sssys:last-child {
  border-bottom: none;
}
.lhd-win {
  float: right;
  color: #393;
}
</style>
