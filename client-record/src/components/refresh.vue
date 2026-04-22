<template>
	<div></div>
</template>
<script>
export default {
  created() {
    this.$alert(this.$t('message.current_page_expired'), '', {
      confirmButtonText: this.$t('message.btn_ok'),
      callback: action => {
        window.opener = null
        window.close()
      }
    })
  }
}
</script>


// WEBPACK FOOTER //
// refresh.vue?38895ead