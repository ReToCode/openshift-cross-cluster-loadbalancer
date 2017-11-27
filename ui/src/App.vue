<template>
  <div id="app">
    <div class="container">
      <section class="hero is-dark">
        <div class="hero-body">
          <div class="container">
            <h1 class="title">
              OpenShift: Smart cross cluster loadbalancer
            </h1>
            <h2 class="subtitle">
              A smart load balancer that is aware of multiple OpenShift clusters and balances traffic based on the routes.
            </h2>
          </div>
        </div>
      </section>
      <br>
      <overall-stats></overall-stats>
      <host-list></host-list>
    </div>
    <footer class="footer">
      <div class="container">
        <div class="content has-text-centered">
          <p>
            <strong><a href="https://github.com/ReToCode/openshift-cross-cluster-loadbalancer">OpenShift-Cross-Cluster-Loadbalancer</a></strong> by <a href="https://github.com/ReToCode">Reto Lehmann</a>. The source code is licensed
            <a href="http://opensource.org/licenses/mit-license.php">MIT</a>.
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script>
  export default {
    name: 'app',
    data() {
      return {}
    }
  }
</script>



