<template>
  <div id="app">
    <navbar></navbar>
    <dashboard></dashboard>
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

<style>
  body {
    font-family: Open Sans, Helvetica, Arial, sans-serif;
    font-size: 1rem;
    line-height: 1.5;
    color: #d8d9da;
    background-color: #141414;
  }

  a:visited {
    text-decoration: none;
    color: #d8d9da;
  }

  a:hover {
    text-decoration: none;
    color: #eff0f1;
  }

  a:focus {
    text-decoration: none;
    color: #d8d9da;
  }

  a:hover, a:active {
    text-decoration: none;
    color: #eff0f1
  }
</style>


