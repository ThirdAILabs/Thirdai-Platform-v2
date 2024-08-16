interface Dataset {
  label: string;
  data: number[];
  borderColor: string;
  backgroundColor: string;
}

interface ChartData {
  labels: string[];
  datasets: Dataset[]
}

interface ModelChartData {
  usageDuration: ChartData;
  usageFrequency: ChartData;
  reformulatedQueries: ChartData;
}

export const chartData: Record<string, ModelChartData> = {
  "good": {
    usageDuration: {
      labels: ['-7d', '-6d', '-5d', '-4d', '-3d', '-2d', '-1d'],
      datasets: [
        {
          label: 'First result',
          data: [0.41484274932444826, 0.38484170128942097, 0.4140113332016118, 0.4189991875746669, 0.40050614862180595, 0.39044126228709064, 0.38984451072568044],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Second result',
          data: [0.2986441375731919, 0.3046771817430856, 0.2760520970055765, 0.29289221615668126, 0.30697655450429445, 0.29982029001230054, 0.31286904723901926],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
        {
          label: 'Third result',
          data: [0.18601091696452068, 0.20664136794667037, 0.21168107850089374, 0.19068793294226452, 0.19450608428992563, 0.2063434408738992, 0.20113552215390693],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Fourth result',
          data: [0.0813220143314663, 0.08486517754104242, 0.07646379243699299, 0.07701904963445688, 0.07949816920419694, 0.0843735582914763, 0.07432716203807613],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Fifth result',
          data: [0.019180181806372758, 0.018974571479780742, 0.021791698854924914, 0.020401613691930463, 0.01851304337977704, 0.019021448535233368, 0.02182375784331717],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['-7d', '-6d', '-5d', '-4d', '-3d', '-2d', '-1d'],
      datasets: [
        {
          label: 'Queries per Day',
          data: [25, 30, 28, 35, 40, 39, 45],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Clicks per Day',
          data: [15, 18, 20, 22, 24, 26, 28],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Upvotes per Day',
          data: [8, 7, 9, 8, 7, 6, 5],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [4, 4, 3, 3, 3, 2, 2],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['-7d', '-6d', '-5d', '-4d', '-3d', '-2d', '-1d'],
      datasets: [
        {
          label: 'Reformulations per Query',
          data: [3.5, 2.8, 3.1, 2.7, 2.8, 3.2, 3.3],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
  "bad": {
    usageDuration: {
      labels: ['-13d', '-11d', '-9d', '-7d', '-5d', '-3d', '-1d'],
      datasets: [
        {
          label: 'First result',
          data: [0.11197484995988896, 0.0986436164085194, 0.10838208845921726, 0.09601322665586973, 0.10777317283719022, 0.10169813863511774, 0.10930642437882511],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Second result',
          data: [0.2715609029119882, 0.2788901114387518, 0.2763283745626661, 0.2838865280978141, 0.2723396521020153, 0.25799510679744936, 0.2527330922401086],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
        {
          label: 'Third result',
          data: [0.3210601442822079, 0.3434425016483842, 0.32149393590689523, 0.3313959346677103, 0.33539110840911784, 0.3424920284350734, 0.33490398791025594],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Fourth result',
          data: [0.21360721204648866, 0.1909163872247643, 0.208354619909204, 0.2076855865724631, 0.20428209375225545, 0.21294349814682897, 0.2116374717935489],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Fifth result',
          data: [0.08179689079942627, 0.08810738327958031, 0.0854409811620173, 0.08101872400614281, 0.0802139728994213, 0.08487122798553048, 0.0914190236772614],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['-13d', '-11d', '-9d', '-7d', '-5d', '-3d', '-1d'],
      datasets: [
        {
          label: 'Queries per Day',
          data: [22, 24, 19, 23, 23, 24, 20],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Clicks per Day',
          data: [13, 15, 14, 12, 10, 11, 12],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Upvotes per Day',
          data: [7, 6, 7, 6, 6, 5, 6],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [4, 3, 4, 5, 2, 3, 4],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['-13d', '-11d', '-9d', '-7d', '-5d', '-3d', '-1d'],
      datasets: [
        {
          label: 'Reformulations per Query',
          data: [5.2, 5.9, 6.1, 5.7, 5.5, 5.9, 5.3],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
  "default": {
    usageDuration: {
      labels: ['-19d', '-16d', '-13d', '-10d', '-7d', '-4d', '-1d'],
      datasets: [
        {
          label: 'First result',
          data: [0.11650485436893203, 0.16363636363636364, 0.23364485981308414, 0.29090909090909095, 0.32432432432432434, 0.3457943925233645, 0.3679245283018868],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Second result',
          data: [0.3106796116504854, 0.3090909090909091, 0.2710280373831776, 0.2545454545454545, 0.2882882882882883, 0.31775700934579443, 0.33018867924528306],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
        {
          label: 'Third result',
          data: [0.2621359223300971, 0.23636363636363636, 0.2429906542056075, 0.23636363636363636, 0.21621621621621623, 0.21495327102803738, 0.2075471698113208],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Fourth result',
          data: [0.21359223300970875, 0.18181818181818182, 0.15887850467289721, 0.13636363636363635, 0.09909909909909911, 0.08411214953271029, 0.07547169811320756],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Fifth result',
          data: [0.0970873786407767, 0.10909090909090909, 0.09345794392523366, 0.08181818181818182, 0.07207207207207207, 0.03738317757009346, 0.01886792452830189],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['-19d', '-16d', '-13d', '-10d', '-7d', '-4d', '-1d'],
      datasets: [
        {
          label: 'Queries per Day',
          data: [27, 31, 28, 33, 37, 38, 42],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Clicks per Day',
          data: [14, 15, 14, 17, 24, 25, 29],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Upvotes per Day',
          data: [10, 8, 7, 8, 7, 9, 11],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [4, 3, 5, 3, 4, 3, 4],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['-19d', '-16d', '-13d', '-10d', '-7d', '-4d', '-1d'],
      datasets: [
        {
          label: 'Reformulations per Query',
          data: [5.2, 4.4, 4.8, 3.9, 3.5, 2.8, 2.6],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
  "empty": {
    usageDuration: {
      labels: ['-1d', '-0d'],
      datasets: [
        {
          label: 'First result',
          data: [0, 0],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Second result',
          data: [0, 0],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
        {
          label: 'Third result',
          data: [0, 0],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Fourth result',
          data: [0, 0],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Fifth result',
          data: [0, 0],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['-1d', '-0d'],
      datasets: [
        {
          label: 'Queries per Day',
          data: [0, 0],
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
        },
        {
          label: 'Clicks per Day',
          data: [0, 0],
          borderColor: 'rgb(153, 102, 255)',
          backgroundColor: 'rgba(153, 102, 255, 0.2)',
        },
        {
          label: 'Upvotes per Day',
          data: [0, 0],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [0, 0],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['-1d', '-0d'],
      datasets: [
        {
          label: 'Reformulations per Query',
          data: [0, 0],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
};
