<template>
  <line-chart
    :height="300"
    :chart-data="connections"></line-chart>
</template>

<script>
  export default {
    name: 'overall',
    computed: {
      connections() {
        return {
          labels: this.$store.state.stats.ticks,
          datasets: [
            {
              label: "Overall connections",
              backgroundColor: '#e3a108',
              data: this.$store.state.stats.overallConnections
            }
          ]
        }
      }
    }
  }
</script>


