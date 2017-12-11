<template>
  <div class="navbar">
    <div class="navbar-inner">
      <ul class="nav">OpenShift Cross Cluster Loadbalancer</ul>
      <ul class="nav pull-right">
        <a href="https://github.com/ReToCode/openshift-cross-cluster-loadbalancer">By ReToCode</a>
      </ul>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'navbar',
    data() {
      return {}
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
  .pull-right {
    float: right!important;
  }
  .navbar .nav.pull-right {
    float: right;
    margin-right: 0;
  }
  .navbar .nav>li>a {
    color: #d8d9da;
  }
</style>
