<template>
  <div class="panel">
    <div class="panel-container flex50">
      <overall></overall>
    </div>
    <div class="panel-container flex50">
      <host-overview></host-overview>
    </div>
  </div>
</template>

<script>
  import Overall from "../parts/Overall.vue";
  import HostOverview from "../parts/HostOverview.vue";

  export default {
    components: {
      Overall,
      HostOverview
    },
    name: 'dashboard'
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
    min-height: 250px;

    background-color: #1f1d1d;
    position: relative;
    border: 1px solid #292929;
  }



  div {
    display: block;
  }
</style>
