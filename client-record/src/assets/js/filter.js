import Vue from 'vue'

export default (function () {
  Vue.filter('status', function (value) {
    if (value == 1) {
      return '启用'
    } else if (value == 0) {
      return '禁用'
    }
    return '未知状态'
  })

  Vue.filter('rules', function (value) {
    return value
  })

  Vue.filter('fileLink', function (value) {
    return window.imgUrl + value
  })

  Vue.filter('toolType', function (value) {
    if (value == 1) {
      return '系统工具'
    } else if (value == 2) {
      return '说明知道'
    }
    return ''
  })

  Vue.filter('numToString', function (value) {
    return value.toString()
  })

  Vue.filter('projectState', function (value) {
    switch (value) {
      case '1':
        return '售前项目'
      case '2':
        return '服务中项目'
      case '3':
        return '已结束项目'
      default:
        return ''
    }
  })

  Vue.filter('time', function (value) {
    const day = moment.unix(value)
    return moment(day).format('YYYY/MM/DD H:mm')
  })

  Vue.filter('date', function (value) {
    const day = moment.unix(value)
    return moment(day).format('YYYY/MM/DD')
  })

  Vue.filter('abstract', function (value) {
    if (value.length > 70) {
      return value.substr(0, 70) + '...'
    }
    return value
  })

  Vue.filter('posStatus', function (value) {
    switch (value) {
      case 1:
        return '在职'
      case 2:
        return '待入职'
      case 3:
        return '离职'
      default:
        return ''
    }
  })

  Vue.filter('template', function (value) {
    if (value == '') {
      return '上传'
    }
    return '上传更新'
  })
})()
