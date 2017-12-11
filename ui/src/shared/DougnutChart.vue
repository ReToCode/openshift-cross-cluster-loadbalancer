<script>
  import {Doughnut, mixins} from 'vue-chartjs'

  const {reactiveProp} = mixins;

  export default {
    extends: Doughnut,
    mixins: [reactiveProp],
    props: ['chartData'],
    mounted() {
      this.renderChart(this.chartData, {responsive: true, maintainAspectRatio: false})
    }
  }
</script>
