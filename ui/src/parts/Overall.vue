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
              backgroundColor: "rgba(226, 161, 8, 0.7)",
              data: this.$store.state.stats.overallConnections
            }
          ]
        }
      }
    }
  }
</script>

<style>
  a {
    text-decoration: none;
    color: #eff0f1;
  }
</style>
