<template>
  <div class="panel">
    <div class="panel-container">
      <line-chart :chart-data="connections"
                  :options="options"></line-chart>
    </div>
    <div class="panel-container">{{ stats }}</div>
  </div>
</template>

<script>
  export default {
    name: 'hostList',
    computed: {
      stats() {
        return this.$store.state.stats;
      },
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
    },
    data() {
      return {
        options: {
          responsive: true, maintainAspectRatio: false
        }
      }
    }
  }
</script>

<style>
  .panel {
    display: flex;
    flex-flow: row wrap;
  }

  .panel-container {
    flex: 1 1 auto;
    min-height: 250px;

    background-color: #1f1d1d;
    position: relative;
    border: 1px solid #292929;
  }

  div {
    display: block;
  }
</style>
