import type { TrainReportData, LabelMetrics, ExampleCategories, TrainingExample } from '@/lib/backend';
import { useState, useEffect } from 'react';

// Mock data for training reports
const mockMetrics: LabelMetrics = {
  'O': {
    precision: 0.95,
    recall: 0.92,
    fmeasure: 0.93
  },
  'NAME': {
    precision: 0.88,
    recall: 0.85,
    fmeasure: 0.86
  },
  'PHONE': {
    precision: 0.91,
    recall: 0.89,
    fmeasure: 0.90
  },
  'EMAIL': {
    precision: 0.94,
    recall: 0.93,
    fmeasure: 0.94
  },
  'ADDRESS': {
    precision: 0.87,
    recall: 0.86,
    fmeasure: 0.87
  }
};

const mockExamples: ExampleCategories = {
  true_positives: {
    'NAME': [
      { source: 'John Smith', target: 'NAME', predictions: 'NAME', index: 0 },
      { source: 'Jane Doe', target: 'NAME', predictions: 'NAME', index: 1 },
      { source: 'Robert Johnson', target: 'NAME', predictions: 'NAME', index: 2 }
    ],
    'PHONE': [
      { source: '555-123-4567', target: 'PHONE', predictions: 'PHONE', index: 0 },
      { source: '(555) 123-4567', target: 'PHONE', predictions: 'PHONE', index: 1 },
      { source: '555.123.4567', target: 'PHONE', predictions: 'PHONE', index: 2 }
    ],
    'EMAIL': [
      { source: 'john.smith@example.com', target: 'EMAIL', predictions: 'EMAIL', index: 0 },
      { source: 'jane.doe@example.com', target: 'EMAIL', predictions: 'EMAIL', index: 1 },
      { source: 'robert.johnson@example.com', target: 'EMAIL', predictions: 'EMAIL', index: 2 }
    ],
    'ADDRESS': [
      { source: '123 Main St', target: 'ADDRESS', predictions: 'ADDRESS', index: 0 },
      { source: '456 Oak Ave', target: 'ADDRESS', predictions: 'ADDRESS', index: 1 },
      { source: '789 Pine Rd', target: 'ADDRESS', predictions: 'ADDRESS', index: 2 }
    ]
  },
  false_positives: {
    'NAME': [
      { source: 'John', target: 'O', predictions: 'NAME', index: 0 },
      { source: 'Smith', target: 'O', predictions: 'NAME', index: 1 }
    ],
    'PHONE': [
      { source: '123-4567', target: 'O', predictions: 'PHONE', index: 0 },
      { source: '555-123', target: 'O', predictions: 'PHONE', index: 1 }
    ],
    'EMAIL': [
      { source: 'john.smith@', target: 'O', predictions: 'EMAIL', index: 0 },
      { source: '@example.com', target: 'O', predictions: 'EMAIL', index: 1 }
    ],
    'ADDRESS': [
      { source: 'Main St', target: 'O', predictions: 'ADDRESS', index: 0 },
      { source: 'Oak Ave', target: 'O', predictions: 'ADDRESS', index: 1 }
    ]
  },
  false_negatives: {
    'NAME': [
      { source: 'Michael Brown', target: 'NAME', predictions: 'O', index: 0 },
      { source: 'Sarah Wilson', target: 'NAME', predictions: 'O', index: 1 }
    ],
    'PHONE': [
      { source: '555-987-6543', target: 'PHONE', predictions: 'O', index: 0 },
      { source: '(555) 987-6543', target: 'PHONE', predictions: 'O', index: 1 }
    ],
    'EMAIL': [
      { source: 'michael.brown@example.com', target: 'EMAIL', predictions: 'O', index: 0 },
      { source: 'sarah.wilson@example.com', target: 'EMAIL', predictions: 'O', index: 1 }
    ],
    'ADDRESS': [
      { source: '321 Elm St', target: 'ADDRESS', predictions: 'O', index: 0 },
      { source: '654 Maple Ave', target: 'ADDRESS', predictions: 'O', index: 1 }
    ]
  }
};

const mockTrainReport: TrainReportData = {
  before_train_metrics: {
    'O': {
      precision: 0.90,
      recall: 0.88,
      fmeasure: 0.89
    },
    'NAME': {
      precision: 0.82,
      recall: 0.80,
      fmeasure: 0.81
    },
    'PHONE': {
      precision: 0.85,
      recall: 0.83,
      fmeasure: 0.84
    },
    'EMAIL': {
      precision: 0.88,
      recall: 0.87,
      fmeasure: 0.88
    },
    'ADDRESS': {
      precision: 0.80,
      recall: 0.78,
      fmeasure: 0.79
    }
  },
  after_train_metrics: mockMetrics,
  after_train_examples: mockExamples
};

// Mock implementation of retrainTokenClassifier
export async function retrainTokenClassifier({
  model_name,
  base_model_id,
}: {
  model_name: string;
  base_model_id: string;
}): Promise<any> {
  console.log(`Mock: Retraining token classifier with model_name: ${model_name}, base_model_id: ${base_model_id}`);
  
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  return {
    status: 'success',
    message: 'Model retraining initiated successfully',
    model_id: `mock-${Date.now()}`,
    model_name: model_name
  };
}

// Mock implementation of trainUDTWithCSV
export function trainUDTWithCSV({
  model_name,
  file,
  base_model_identifier,
  test_split = 0.1,
}: {
  model_name: string;
  file: File;
  base_model_identifier: string;
  test_split?: number;
}): Promise<any> {
  console.log(`Mock: Training UDT with CSV - model_name: ${model_name}, base_model_identifier: ${base_model_identifier}, test_split: ${test_split}`);
  console.log(`File details: ${file.name}, ${file.size} bytes, ${file.type}`);
  
  // Simulate API delay
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        status: 'success',
        message: 'Model training with CSV initiated successfully',
        model_id: `mock-csv-${Date.now()}`,
        model_name: model_name
      });
    }, 1000);
  });
}

// Mock implementation of getTrainReport
export async function getTrainReport(modelId: string): Promise<{ data: TrainReportData }> {
  console.log(`Mock: Getting training report for model ID: ${modelId}`);
  
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 800));
  
  return {
    data: mockTrainReport
  };
}

// Mock implementation of useLabels hook
export function useLabels({
  deploymentUrl,
  maxRecentLabels = 5,
}: {
  deploymentUrl: string;
  maxRecentLabels?: number;
}): {
  allLabels: Set<string>;
  recentLabels: string[];
  error: Error | null;
  isLoading: boolean;
  refresh: () => Promise<void>;
} {
  const [allLabels, setAllLabels] = useState<Set<string>>(new Set());
  const [recentLabels, setRecentLabels] = useState<string[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const fetchLabels = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Mock labels
      const mockLabels = ['NAME', 'PHONE', 'EMAIL', 'ADDRESS', 'O', 'DATE', 'LOCATION', 'ORGANIZATION'];
      setAllLabels(new Set(mockLabels));
      setRecentLabels(mockLabels.slice(0, maxRecentLabels));
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch labels'));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLabels();
  }, [deploymentUrl, maxRecentLabels]);

  return {
    allLabels,
    recentLabels,
    error,
    isLoading,
    refresh: fetchLabels
  };
}

// Mock implementation of useRecentSamples hook
export function useRecentSamples({
  deploymentUrl,
  maxRecentSamples = 5,
}: {
  deploymentUrl: string;
  maxRecentSamples?: number;
}): {
  recentSamples: { tokens: string[]; tags: string[] }[];
  error: Error | null;
  isLoading: boolean;
  refresh: () => Promise<void>;
} {
  const [recentSamples, setRecentSamples] = useState<{ tokens: string[]; tags: string[] }[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const fetchSamples = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Mock samples
      const mockSamples = [
        { 
          tokens: ['John', 'Smith', 'contacted', 'us', 'at', '555-123-4567'], 
          tags: ['NAME', 'NAME', 'O', 'O', 'O', 'PHONE'] 
        },
        { 
          tokens: ['Please', 'email', 'Jane', 'Doe', 'at', 'jane.doe@example.com'], 
          tags: ['O', 'O', 'NAME', 'NAME', 'O', 'EMAIL'] 
        },
        { 
          tokens: ['Our', 'office', 'is', 'located', 'at', '123', 'Main', 'St'], 
          tags: ['O', 'O', 'O', 'O', 'O', 'ADDRESS', 'ADDRESS', 'ADDRESS'] 
        },
        { 
          tokens: ['Robert', 'Johnson', 'will', 'visit', 'on', 'January', '15th'], 
          tags: ['NAME', 'NAME', 'O', 'O', 'O', 'DATE', 'DATE'] 
        },
        { 
          tokens: ['The', 'meeting', 'is', 'at', 'Google', 'headquarters'], 
          tags: ['O', 'O', 'O', 'O', 'ORGANIZATION', 'O'] 
        }
      ];
      
      setRecentSamples(mockSamples.slice(0, maxRecentSamples));
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch recent samples'));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchSamples();
  }, [deploymentUrl, maxRecentSamples]);

  return {
    recentSamples,
    error,
    isLoading,
    refresh: fetchSamples
  };
}

// Mock implementation of getLabels
export async function getLabels(): Promise<string[]> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 500));
  
  // Return a predefined set of labels
  return [
    'NAME', 
    'PHONE', 
    'EMAIL', 
    'ADDRESS', 
    'DATE', 
    'ORGANIZATION', 
    'LOCATION', 
    'O'
  ];
} 