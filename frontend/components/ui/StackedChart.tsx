import React from 'react';
import { Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ChartData,
  ChartOptions,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

interface StackedChartProps {
  data: ChartData<'bar'> | null;
}

const StackedChart: React.FC<StackedChartProps> = ({ data }) => {
  const options: ChartOptions<'bar'> = {
    responsive: true,
    plugins: {
      title: {
        display: true,
        text: 'User Click Distribution (Top 1 vs Top 2 vs Top 3 vs Top 4+)',
      },
      tooltip: {
        mode: 'index',
        intersect: false,
      },
      legend: {
        display: true,
        position: 'top',
      },
    },
    scales: {
      x: {
        stacked: true, // Enable stacking
      },
      y: {
        stacked: true, // Enable stacking
        beginAtZero: true,
        title: {
          display: true,
          text: 'Number of Clicks',
        },
      },
    },
  };

  return (
    <div>
      {data ? (
        <Bar data={data} options={options} />
      ) : (
        <p>Loading data...</p>
      )}
    </div>
  );
};

export default StackedChart;
