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
              backgroundColor: '#e3a108',
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
              backgroundColor: '#5b9c1c',
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
              backgroundColor: '#df171b',
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


