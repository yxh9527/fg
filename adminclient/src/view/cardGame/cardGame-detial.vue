 <template>
  <div>
    <!-- 按实际情况判断是否展示庄家信息 -->
    <Table v-if="bankerShow" :columns="columnsBank" :data="bankerData" border></Table>

    <!-- 展示玩家详情 -->
    <Table :columns="columns" :data="detialData" border></Table>
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
      detialData: [],

      // 庄家配置信息
      bankerShow: false,
      columnsBank: [],
      bankerData: []
    }
  },
  methods: {
    pokerCards (item) {
      if (item.number > 10 || item.number == 1) {
        item.number =
          { 1: 'A', 11: 'J', 12: 'Q', 13: 'K' }[item.number] || item.number
      }
      let cards = {
        1: (
          <span class="color-style-black">
            <img src="/color/hei.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        ),
        2: (
          <span class="color-style-red">
            <Icon type="md-heart" size="23" />
            {item.number}
            &emsp;
          </span>
        ),
        3: (
          <span class="color-style-black">
            <img src="/color/hua.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        ),
        4: (
          <span class="color-style-red">
            <img src="/color/fang.png" height="18" />
            &nbsp;
            {item.number}
            &emsp;
          </span>
        )
      }[item.color] || (
        <span>
          {item.color}&emsp;{item.number}
        </span>
      )
      return cards
    }
  },

  mounted () {
    const fwType = Object.keys(this.rowInfo[0].SettlementDetail)[0]
    this.detialData = []
    this.banker = []
    this.columns = []
    if (fwType == 'slhb') {
      this.columns = [
        {
          title: '玩家ID',
          key: 'player_id'
        },
        {
          title: '金额-雷号',
          key: ''
        },
        {
          title: '抢包金额',
          key: ''
        },
        {
          title: '结果',
          key: ''
        },
        {
          title: '结算',
          key: ''
        }
      ]
      this.rowInfo.map(item => {})
    }

    if (fwType == 'nnhb') {
      this.columns = [
        {
          title: '玩家ID',
          key: 'player_id'
        },
        {
          title: '金额-包数',
          key: ''
        },
        {
          title: '抢包牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '结算',
          key: ''
        }
      ]
      this.rowInfo.map(item => {})
    }

    if (fwType == 'qznn') {
      this.columns = [
        {
          title: '玩家ID',
          key: 'player_id'
        },
        {
          title: '抢庄金额',
          key: ''
        },
        {
          title: '下注金额',
          key: ''
        },
        {
          title: '抢包牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '结算',
          key: ''
        }
      ]
      this.rowInfo.map(item => {})
    }

    if (fwType == 'qznn') {
      this.columns = [
        {
          title: '玩家ID',
          key: 'player_id'
        },
        {
          title: '抢庄金额',
          key: ''
        },
        {
          title: '下注金额',
          key: ''
        },
        {
          title: '抢包牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '牌型',
          align: 'center',
          key: 'poker_cards',
          width: 250,
          render: (h, parmas) => {
            let conts = []
            conts.push(
              parmas.row.poker_cards.map(item => {
                return this.pokerCards(item.Card.Poker)
              })
            )
            return <div>{conts}</div>
          }
        },
        {
          title: '结算',
          key: ''
        }
      ]
      this.rowInfo.map(item => {})
    }
  }
}
</script>

<style>
.ivu-table-border td {
  height: 34px;
}
.color-style-red {
  color: red;
  font-size: 20px;
}
.color-style-black {
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
</style>
