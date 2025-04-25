import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';

// Mock data for workflows
const mockWorkflows = [
  {
    model_id: 'workflow-1',
    model_name: 'Token Classification - General',
    type: 'nlp-token',
    access: 'public',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-05-15T10:30:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: null,
    attributes: {
      llm_provider: 'openai',
      default_mode: 'standard'
    },
    dependencies: [
      {
        model_id: 'base-model-1',
        model_name: 'Base Token Classification Model',
        type: 'nlp-token',
        sub_type: 'base',
        username: 'demo_user'
      }
    ],
    size: '1.2GB',
    size_in_memory: '500MB'
  },
  {
    model_id: 'workflow-2',
    model_name: 'Token Classification - Finance',
    type: 'nlp-token',
    access: 'private',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-04-10T09:15:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: 'team-1',
    attributes: {
      llm_provider: 'openai',
      default_mode: 'finance'
    },
    dependencies: [
      {
        model_id: 'base-model-2',
        model_name: 'Finance Base Model',
        type: 'nlp-token',
        sub_type: 'finance',
        username: 'demo_user'
      }
    ],
    size: '1.5GB',
    size_in_memory: '600MB'
  },
  {
    model_id: 'workflow-3',
    model_name: 'Token Classification - Medical',
    type: 'nlp-token',
    access: 'protected',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-03-22T16:45:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: 'team-2',
    attributes: {
      llm_provider: 'openai',
      default_mode: 'medical'
    },
    dependencies: [
      {
        model_id: 'base-model-3',
        model_name: 'Medical Base Model',
        type: 'nlp-token',
        sub_type: 'medical',
        username: 'demo_user'
      }
    ],
    size: '1.8GB',
    size_in_memory: '700MB'
  },
  {
    model_id: 'workflow-4',
    model_name: 'Token Classification - Legal',
    type: 'nlp-token',
    access: 'private',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-02-05T11:30:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: null,
    attributes: {
      llm_provider: 'openai',
      default_mode: 'legal'
    },
    dependencies: [
      {
        model_id: 'base-model-4',
        model_name: 'Legal Base Model',
        type: 'nlp-token',
        sub_type: 'legal',
        username: 'demo_user'
      }
    ],
    size: '2.0GB',
    size_in_memory: '800MB'
  },
  {
    model_id: 'workflow-5',
    model_name: 'Token Classification - Customer Support',
    type: 'nlp-token',
    access: 'public',
    train_status: 'starting',
    deploy_status: 'not_started',
    publish_date: '2023-01-18T14:20:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: null,
    attributes: {
      llm_provider: 'openai',
      default_mode: 'support'
    },
    dependencies: [
      {
        model_id: 'base-model-5',
        model_name: 'Support Base Model',
        type: 'nlp-token',
        sub_type: 'support',
        username: 'demo_user'
      }
    ],
    size: '1.6GB',
    size_in_memory: '650MB'
  }
];

// Mock data for predictions
const mockPredictionResponses: Record<string, any> = {
  'default': {
    prediction_results: {
      tokens: ['The', 'quick', 'brown', 'fox', 'jumps', 'over', 'the', 'lazy', 'dog', '.'],
      predicted_tags: [['O'], ['O'], ['COLOR'], ['ANIMAL'], ['O'], ['O'], ['O'], ['O'], ['ANIMAL'], ['O']],
      source_object: 'Sample text 1'
    },
    time_taken: 0.15
  },
  'finance': {
    prediction_results: {
      tokens: ['Apple', 'Inc', 'reported', 'revenue', 'of', '$', '90', '.', '3', 'billion', 'in', 'Q1', '2023', '.'],
      predicted_tags: [['COMPANY'], ['COMPANY'], ['O'], ['O'], ['O'], ['CURRENCY'], ['AMOUNT'], ['O'], ['AMOUNT'], ['AMOUNT'], ['O'], ['TIME'], ['YEAR'], ['O']],
      source_object: 'Financial news article'
    },
    time_taken: 0.18
  },
  'medical': {
    prediction_results: {
      tokens: ['The', 'patient', 'presented', 'with', 'fever', 'and', 'cough', 'for', '3', 'days', '.'],
      predicted_tags: [['O'], ['O'], ['O'], ['O'], ['SYMPTOM'], ['O'], ['SYMPTOM'], ['O'], ['DURATION'], ['DURATION'], ['O']],
      source_object: 'Medical record'
    },
    time_taken: 0.12
  }
};

// Mock data for deployment stats
const mockDeploymentStats = {
  system: {
    header: ['Name', 'Description'],
    rows: [
      ['CPU', '12 vCPUs'],
      ['CPU Model', 'Intel(R) Xeon(R) CPU E5-2680 v3 @ 2.50GHz'],
      ['Memory', '64 GB RAM'],
      ['System Uptime', '5 days 12 hours 30 minutes 15 seconds'],
    ],
  },
  throughput: {
    header: [
      'Time Period',
      'Tokens Identified',
      'Queries Ingested',
      'Queries Ingested Size',
    ],
    rows: [
      [
        'Past hour',
        '1.2M',
        '50K',
        '2.5MB',
      ],
      [
        'Total',
        '15.7B',
        '1.2M',
        '45.3MB',
      ],
    ],
  },
};

// Mock data for labels
const mockLabels = ['O', 'PERSON', 'ORGANIZATION', 'LOCATION', 'DATE', 'TIME', 'MONEY', 'PERCENT', 'PRODUCT', 'EVENT', 'WORK_OF_ART', 'LAW', 'LANGUAGE', 'FACILITY', 'ANIMAL', 'PLANT', 'COLOR', 'COMPANY', 'AMOUNT', 'CURRENCY', 'YEAR', 'SYMPTOM', 'DURATION'];

export function useTokenClassificationEndpoints() {
  const params = useParams();
  const workflowId = params.deploymentId as string;
  const [workflowName, setWorkflowName] = useState<string>('Mock Token Classification Model');
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>(`https://mock-api.example.com/${workflowId}`);

  useEffect(() => {
    // Simulate API call to get workflow details
    const init = async () => {
      // In a real implementation, this would fetch from the API
      setWorkflowName(`Mock Token Classification Model (${workflowId})`);
      setDeploymentUrl(`https://mock-api.example.com/${workflowId}`);
    };
    init();
  }, [workflowId]);

  const predict = async (query: string): Promise<any> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // Return a random prediction response based on the query
    const queryLower = query.toLowerCase();
    if (queryLower.includes('finance') || queryLower.includes('money') || queryLower.includes('stock')) {
      return mockPredictionResponses['finance'];
    } else if (queryLower.includes('medical') || queryLower.includes('patient') || queryLower.includes('symptom')) {
      return mockPredictionResponses['medical'];
    } else {
      return mockPredictionResponses['default'];
    }
  };

  const formatTime = (timeSeconds: number) => {
    const timeMinutes = Math.floor(timeSeconds / 60);
    const timeHours = Math.floor(timeMinutes / 60);
    const timeDays = Math.floor(timeHours / 24);
    return `${timeDays} days ${timeHours % 24} hours ${timeMinutes % 60} minutes ${timeSeconds % 60} seconds`;
  };

  const formatAmount = (amount: number) => {
    if (amount < 1000) {
      return amount.toString();
    }
    let suffix = '';
    if (amount >= 1000000000) {
      amount /= 1000000000;
      suffix = ' B';
    } else if (amount >= 1000000) {
      amount /= 1000000;
      suffix = ' M';
    } else {
      amount /= 1000;
      suffix = ' K';
    }
    let amountstr = amount.toString();
    if (amountstr.includes('.')) {
      const [wholes, decimals] = amountstr.split('.');
      const decimalsLength = 3 - Math.min(3, wholes.length);
      amountstr = decimalsLength ? wholes + '.' + decimals.substring(0, decimalsLength) : wholes;
    }
    return amountstr + suffix;
  };

  const getStats = async () => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 800));
    return mockDeploymentStats;
  };

  const insertSample = async (sample): Promise<void> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 600));
    console.log('Mock: Inserting sample', sample);
    // In a real implementation, this would send data to the API
  };

  const addLabel = async (labels: {
    tags: { name: string; description: string }[];
  }): Promise<void> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    console.log('Mock: Adding labels', labels);
    // In a real implementation, this would send data to the API
  };

  const getLabels = async (): Promise<string[]> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockLabels;
  };

  const getTextFromFile = async (file: File): Promise<string[]> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    console.log('Mock: Processing file', file.name);
    
    // Return mock text based on file name
    if (file.name.includes('finance')) {
      return [
        'Apple Inc reported revenue of $90.3 billion in Q1 2023.',
        'Microsoft Corp announced a new AI product yesterday.',
        'Tesla stock price increased by 5% after the earnings call.'
      ];
    } else if (file.name.includes('medical')) {
      return [
        'The patient presented with fever and cough for 3 days.',
        'Previous medical history includes hypertension and diabetes.',
        'Current medications: Lisinopril 10mg daily, Metformin 500mg twice daily.'
      ];
    } else {
      return [
        'The quick brown fox jumps over the lazy dog.',
        'She sells seashells by the seashore.',
        'All that glitters is not gold.'
      ];
    }
  };

  return {
    workflowName,
    predict,
    insertSample,
    addLabel,
    getLabels,
    getTextFromFile,
    getStats,
  };
}

// Add the fetchWorkflows function
export async function fetchWorkflows(): Promise<any[]> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 800));
  
  // Return mock workflows
  return mockWorkflows;
} 