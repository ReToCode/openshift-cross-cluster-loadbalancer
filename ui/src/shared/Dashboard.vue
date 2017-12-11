<template>
  <div>
    <div class="panel">
      <div class="panel-container flex50">
        <overall></overall>
      </div>
      <div class="panel-container flex50">
        <host-overview></host-overview>
      </div>
    </div>

    <!--Repeat a row for each host-->
    <div v-for="host in hosts">
        <host :host="host"></host>
    </div>
  </div>
</template>

<script>
  import Overall from "../parts/Overall.vue";
  import HostOverview from "../parts/HostOverview.vue";
  import Host from "../parts/Host.vue";

  export default {
    components: {
      Overall,
      HostOverview,
      Host
    },
    name: 'dashboard',
    computed: {
      hosts() {
        return this.$store.state.stats.hosts;
      }
    }
  }
</script>

<style>
  .panel {
    display: flex;
    flex-flow: row;
  }

  .flex50 {
    flex: 50%
  }

  .panel-container {
    min-height: 150px;

    background-color: #1f1d1d;
    position: relative;
    border: 1px solid #292929;
  }

  div {
    display: block;
  }
</style>
