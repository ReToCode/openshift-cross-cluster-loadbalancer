<template>
  <bar-chart :chart-data="hosts"
             :height="300"
             :chart-options="options"></bar-chart>
</template>

<script>
  export default {
    name: 'host-overview',
    data() {
      return {
        options: {
          scales:
            {
              yAxes: [{
                stacked: true
              }],
              xAxes: [{
                stacked: true
              }]
            }
        }
      }
    },
    computed: {
      hosts() {
        let s = this.$store.state.stats;
        if (!s || !s.ticks || !s.healthyHosts || !s.unhealthyHosts) {
          s = {ticks: [], healthyHosts: [], unhealthyHosts: []}
        }
        return {
          labels: s.ticks.slice(Math.max(s.ticks.length - 5, 1)),
          datasets: [
            {
              label: "Healthy hosts",
              backgroundColor: 'rgba(91, 156, 28, 0.8)',
              data: s.healthyHosts.slice(Math.max(s.healthyHosts.length - 5, 1))
            },
            {
              label: "Unhealthy hosts",
              backgroundColor: 'rgba(237, 120, 53, 0.8)',
              data: s.unhealthyHosts.slice(Math.max(s.unhealthyHosts.length - 5, 1))
            }
          ]
        }
      }
    }
  }
</script>
