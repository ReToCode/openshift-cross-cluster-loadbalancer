<template>
  <section class="content">
    <h1>List of connected OpenShift router hosts</h1>
    <b-table
      :data="!hasData ? [] : uiStats.hostList"
      :striped="isStriped"
      :narrowed="isNarrowed">

      <template slot-scope="props">
        <b-table-column label="Router Host IP">
          {{ props.row.hostIP }}
        </b-table-column>
        <b-table-column label="Healthy">
          {{ props.row.stats.healthy }}
        </b-table-column>
        <b-table-column label="Tot. Conn.">
          {{ props.row.stats.totalConnections }}
        </b-table-column>
        <b-table-column label="Active Conn.">
          {{ props.row.stats.activeConnections }}
        </b-table-column>
        <b-table-column label="Refused Conn.">
          {{ props.row.stats.refusedConnections }}
        </b-table-column>
      </template>
    </b-table>
  </section>
</template>

<script>
  export default {
    name: 'hostList',
    computed: {
      uiStats() {
        return this.$store.state.uiStats;
      },
      hasData() {
        const stats = this.$store.state.uiStats;
        return stats && stats.hostList && stats.hostList.length;
      }
    },
    data() {
      return {
        isStriped: true,
        isNarrowed: true
      }
    }
  }
</script>
