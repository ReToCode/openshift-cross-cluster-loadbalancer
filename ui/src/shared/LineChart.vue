<script>
  import {Line, mixins} from 'vue-chartjs'

  const {reactiveProp} = mixins;

  export default {
    extends: Line,
    mixins: [reactiveProp],
    props: ['chartData', 'chartOptions'],
    computed: {
        options() {
          let opts = {};
          if (this.chartOptions) {
            opts = this.chartOptions;
          }

          opts.animation = {
            duration: 400
          };
          opts.responsive = true;
          opts.maintainAspectRatio = false;
          return opts;
      }
    },
    mounted() {
      this.renderChart(this.chartData, this.options)
    }
  }
</script>
