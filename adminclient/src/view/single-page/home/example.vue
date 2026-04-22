<template>
  <div ref="dom"></div>
</template>

<script>
import * as echarts from 'echarts'
import { on, off } from '@/libs/tools'
export default {
  name: 'serviceRequests',
  data () {
    return {
      dom: null
    }
  },
  props: {
    value: Object,
    barData: Object,
    columns: Number,
    legend: Array,
    type: String,
    text: String,
    subtext: String,
    color: String
  },
  methods: {
    resize () {
      this.dom.resize()
    },
    getDay (m) {
      let t = new Date()
      let d = new Date(t.getFullYear(), t.getMonth() + m, 0)
      let arr = []
      for (let day = 1; day <= d.getDate(); day++) {
        arr.push(day)
      }
      return arr
    }
  },
  mounted () {
    const option = {
      title: {
        text: this.text,
        subtext: this.subtext,
        x: 'left'
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
          label: {
            backgroundColor: '#6a7985'
          }
        }
      },
      legend: {
        data: this.legend,
        x: 'right'
      },
      grid: {
        left: '1.2%',
        right: '1%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: [
        {
          type: 'category',
          data: {
            3: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
            4: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
            5: this.getDay(1),
            6: this.getDay(0)
          }[this.columns] || [
            '00时-02时',
            '02时-04时',
            '04时-06时',
            '06时-08时',
            '08时-10时',
            '10时-12时',
            '12时-14时',
            '14时-16时',
            '16时-18时',
            '18时-20时',
            '20时-22时',
            '22时-24时'
          ]
        }
      ],
      yAxis: [
        {
          type: 'value'
        }
      ],
      series: this.barData.bet
        ? [
          {
            name: this.legend ? this.legend : '总点数',
            type: this.type,
            data: this.barData.sum,
            barWidth: '20px',
            color: this.color
          }
        ]
        : {
          name: this.legend,
          type: this.type,
          barWidth: '20px',
          color: this.color,
          data: this.barData.sum
        }
    }
    this.$nextTick(() => {
      this.dom = echarts.init(this.$refs.dom)
      on(window, 'resize', this.resize)
      this.dom.setOption(option)
    })
  },
  beforeDestroy () {
    off(window, 'resize', this.resize)
  }
}
</script>
