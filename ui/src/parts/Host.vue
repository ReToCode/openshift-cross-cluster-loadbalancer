<template>
  <div class="panel">
    <div class="panel-container flex25">
      <line-chart :chart-data="connections" :height="200"></line-chart>
    </div>
    <div class="panel-container flex25">
      <line-chart :chart-data="activeConnections" :height="200"></line-chart>
    </div>
    <div class="panel-container flex25">
      <line-chart :chart-data="refusedConnections" :height="200"></line-chart>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'host',
    props: ['host'],
    computed: {
      connections() {
        return {
          labels: this.$store.state.stats.ticks,
          datasets: [
            {
              label: `Total Conn: ${this.host.clusterKey}-${this.host.hostIP}`,
              backgroundColor: 'rgba(226, 161, 8, 0.7)',
              data: this.host.stats.map(s => s.totalConnections)
            }
          ]
        }
      },
      activeConnections() {
        return {
          labels: this.$store.state.stats.ticks,
          datasets: [
            {
              label: `Active Conn: ${this.host.clusterKey}-${this.host.hostIP}`,
              backgroundColor: 'rgba(91, 156, 28, 0.8)',
              data: this.host.stats.map(s => s.activeConnections)
            }
          ]
        }
      },
      refusedConnections() {
        return {
          labels: this.$store.state.stats.ticks,
          datasets: [
            {
              label: `Refused Conn: ${this.host.clusterKey}-${this.host.hostIP}`,
              backgroundColor: 'rgba(223, 23, 27, 0.5)',
              data: this.host.stats.map(s => s.refusedConnections)
            }
          ]
        }
      }
    }
  }
</script>

<style>
  .flex25 {
    flex: 25%
  }
</style>


