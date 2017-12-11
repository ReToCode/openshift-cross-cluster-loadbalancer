<template>
  <div class="navbar">
    <div class="navbar-inner">
      <ul>
        <li class="nav">OpenShift Cross Cluster Loadbalancer</li>
        <li class="nav">
          <a class="pointer" v-on:click="resetStats">
            Reset stats
          </a>
        </li>
        <li class="nav pull-right">
          By <a href="https://github.com/ReToCode/openshift-cross-cluster-loadbalancer">ReToCode</a>
        </li>
      </ul>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'navbar',
    methods: {
      resetStats: function(evt) {
        this.$http.post('/ui/resetstats',{})
      }
    }
  }
</script>

<style>
  .navbar {
    overflow: visible;
    position: relative;
    z-index: 110;
  }

  .navbar-inner {
    min-height: 52px;
    padding-right: 1rem;
    background-color: #292929;
    border-bottom: 1px solid #141414;
  }

  .nav {
    margin: 0;
    list-style: none;
  }

  .navbar .nav {
    position: relative;
    left: 0;
    float: left;

    overflow: hidden;
    display: block;
    margin: 0;
    color: #d6d6d6;
    font-size: 1.25rem;
    padding: 1rem 1rem .75rem;
    font-style: normal;
    font-weight: 400;
    font-variant: normal;
    text-transform: none;
    line-height: 1;
    -webkit-font-smoothing: antialiased;
  }

  .pointer {
    cursor: pointer;
  }

  .pull-right {
    float: right !important;
  }

  .navbar .nav.pull-right {
    float: right;
    margin-right: 0;
  }

  .navbar .nav > li > a {
    color: #d8d9da;
  }
</style>
