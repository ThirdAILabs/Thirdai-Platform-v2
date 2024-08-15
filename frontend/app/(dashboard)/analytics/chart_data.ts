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
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Hours per Day',
          data: [2.5, 3.2, 2.8, 3.5, 4.1, 3.9, 4.5],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Hours per Month',
          data: [75, 90, 86, 105, 128, 121, 135],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
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
          data: [5, 6, 7, 8, 10, 9, 11],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [2, 3, 2, 3, 4, 3, 4],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Reformulated Queries',
          data: [18, 15, 17, 14, 13, 12, 5],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
  "bad": {
    usageDuration: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Hours per Day',
          data: [2.5, 3.2, 2.8, 3.5, 4.1, 3.9, 4.5],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Hours per Month',
          data: [75, 90, 86, 105, 128, 121, 135],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
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
          data: [5, 6, 7, 8, 10, 9, 11],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [2, 3, 2, 3, 4, 3, 4],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Reformulated Queries',
          data: [18, 15, 17, 14, 13, 12, 20],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
  "default": {
    usageDuration: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Hours per Day',
          data: [2.5, 3.2, 2.8, 3.5, 4.1, 3.9, 4.5],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
        {
          label: 'Hours per Month',
          data: [75, 90, 86, 105, 128, 121, 135],
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.2)',
        },
      ],
    },
    usageFrequency: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
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
          data: [5, 6, 7, 8, 10, 9, 11],
          borderColor: 'rgb(255, 159, 64)',
          backgroundColor: 'rgba(255, 159, 64, 0.2)',
        },
        {
          label: 'Associates per Day',
          data: [2, 3, 2, 3, 4, 3, 4],
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
        },
      ],
    },
    reformulatedQueries: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          label: 'Reformulated Queries',
          data: [18, 15, 17, 14, 13, 12, 10],
          borderColor: 'rgb(255, 205, 86)',
          backgroundColor: 'rgba(255, 205, 86, 0.2)',
        },
      ],
    }
  },
};
