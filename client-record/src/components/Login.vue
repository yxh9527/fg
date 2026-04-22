<template>
  <div></div>
</template>

<script>
export default {
  data() {
    return {}
  },
  methods: {},
  created() {
    // function setCookie(name, value) {
    //   var Days = 30
    //   var exp = new Date()
    //   exp.setTime(exp.getTime() + Days * 24 * 60 * 60 * 1000)
    //   document.cookie = name + '=' + escape(value) + ';expires=' + exp.toGMTString()
    // }
    // setCookie('_ga', 'GA1.2.58166312.1553537741')
    // setCookie('_gid', 'GA1.2.2003815067.1553537741')
    // https://h5.fg-dev.99808.net/we_token?game_id=5004
    // https://h5.fg-sit.99808.net/we_token?game_id=6702
    // https://static.98078.net/we_token?game_id=5004
    // https://h5.cymsy.cn/we_token?game_id=5004
    let data = this.$route.query
    if (JSON.stringify(data) != '{}') {
      console.log(data)
      if (data.token) {
        Lockr.set('token', data.token)
        window.axios.defaults.headers.webToken = data.token
      }
      if (data.game_id) {
        Lockr.set('game_id', data.game_id)
      } else {
        Lockr.set('game_id', 0)
      }
      if (data.language) {
        Lockr.set('language', data.language)
      }
      if (data.timezone) {
        Lockr.set('timezone', data.timezone)
      }
      this.$router.replace('/slot_record')
    }
  }
}
</script>

<style>
</style>


// WEBPACK FOOTER //
// Login.vue?7bb09d60