<template>
  <section class="content">
    <h1>Overall statistics</h1>

    <h4>Current open connections are:</h4>
    <p>
      <span class="tag is-success">Currently:</span>  {{ uiStats.currentConnections }}
    </p>
    <h4>Connected</h4>
    <p><a class="button is-info"
          v-bind:class="{'is-loading': isConnected}">Primary</a>
    </p>

  </section>
</template>

<script>
  export default {
    name: 'overallStats',
    computed: {
      uiStats() {
        return this.$store.state.stats;
      },
      isConnected() {
        return this.$store.state.socket.isConnected;
      }
    },
    data() {
      return {}
    }
  }
</script>
