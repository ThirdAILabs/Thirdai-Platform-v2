import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import type { TrainReportData, LabelMetrics, ExampleCategories, TrainingExample } from '@/lib/backend';
import { ObjectDatabaseRecord, ClassifiedTokenDatabaseRecord } from '@/app/token-classification/[deploymentId]/jobs/[jobId]/(database-table)/types';
import {
  mockWorkflows,
  mockPredictionResponses,
  mockDeploymentStats,
  mockLabels,
  mockMetrics,
  mockExamples,
  mockTrainReport,
  mockGroups,
  mockTags,
  mockObjectRecords,
  mockClassifiedTokenRecords,
  loadMoreMockObjectRecords,
  loadMoreMockClassifiedTokenRecords,
  upvotes,
  associations,
  reformulations
} from './mock-data';

// Helper function to simulate API delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

// Helper function to format time
export const formatTime = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = seconds % 60;

  const parts = [];
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);
  if (remainingSeconds > 0 || parts.length === 0) parts.push(`${remainingSeconds}s`);

  return parts.join(' ');
};

// Helper function to format amount
export const formatAmount = (amount: number): string => {
  if (amount >= 1e9) return `${(amount / 1e9).toFixed(1)}B`;
  if (amount >= 1e6) return `${(amount / 1e6).toFixed(1)}M`;
  if (amount >= 1e3) return `${(amount / 1e3).toFixed(1)}K`;
  return amount.toString();
};

// Mock API functions
export const predict = async (text: string, mode: string = 'default') => {
  console.log("HERE")
  console.log('text', text);
  await delay(500);
  console.log('mockPredictionResponses', mockPredictionResponses[text]);
  return mockPredictionResponses[text] || mockPredictionResponses['default'];
};

export const getStats = async () => {
  await delay(300);
  return mockDeploymentStats;
};

export const insertSamples = async (samples: TrainingExample[]) => {
  await delay(1000);
  return { success: true, message: 'Samples inserted successfully' };
};

export const addLabels = async (labels: string[]) => {
  await delay(500);
  return { success: true, message: 'Labels added successfully' };
};

export const getWorkflows = async () => {
  await delay(800);
  return mockWorkflows;
};

export const retrainTokenClassifier = async (workflowId: string) => {
  await delay(2000);
  return { success: true, message: 'Retraining started successfully' };
};

export const trainUDTWithCSV = async (workflowId: string, csvData: string) => {
  await delay(2000);
  return { success: true, message: 'Training started successfully' };
};

export const getTrainReport = async (workflowId: string) => {
  await delay(1000);
  return mockTrainReport;
};

// React hooks for managing state
export const useLabels = () => {
  return {
    labels: mockLabels,
    isLoading: false,
    error: null
  };
};

export const useRecentSamples = () => {
  return {
    samples: mockExamples.true_positives,
    isLoading: false,
    error: null
  };
};

// Token classification functions
export const getObjectRecords = async () => {
  await delay(1000);
  return mockObjectRecords;
};

export const getClassifiedTokenRecords = async () => {
  await delay(1000);
  return mockClassifiedTokenRecords;
};

export const loadMoreObjectRecords = loadMoreMockObjectRecords;
export const loadMoreClassifiedTokenRecords = loadMoreMockClassifiedTokenRecords;

// Analytics functions
export const getUpvotes = async () => {
  await delay(800);
  return upvotes;
};

export const getAssociations = async () => {
  await delay(800);
  return associations;
};

export const getReformulations = async () => {
  await delay(800);
  return reformulations;
};

export function useTokenClassificationEndpoints() {
  const { deploymentId } = useParams();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const predict = async (text: string) => {
    setIsLoading(true);
    console.log("HERE")
    try {
      await delay(500);
      return mockPredictionResponses[text] || mockPredictionResponses['default'];
    } finally {
      setIsLoading(false);
    }
  };

  const getStats = async () => {
    setIsLoading(true);
    try {
      await delay(1000);
      return mockDeploymentStats;
    } finally {
      setIsLoading(false);
    }
  };

  const insertSample = async (sample: TrainingExample): Promise<void> => {
    setIsLoading(true);
    try {
      await delay(1000);
      // Mock implementation
    } finally {
      setIsLoading(false);
    }
  };

  const addLabel = async (label: string): Promise<void> => {
    setIsLoading(true);
    try {
      await delay(1000);
      // Mock implementation
    } finally {
      setIsLoading(false);
    }
  };

  const getLabels = async (): Promise<string[]> => {
    setIsLoading(true);
    try {
      await delay(1000);
      return mockLabels;
    } finally {
      setIsLoading(false);
    }
  };

  const getTextFromFile = async (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          resolve(event.target.result as string);
        } else {
          reject(new Error('Failed to read file'));
        }
      };
      reader.onerror = () => reject(new Error('Failed to read file'));
      reader.readAsText(file);
    });
  };

  return {
    isLoading,
    error,
    predict,
    getStats,
    insertSample,
    addLabel,
    getLabels,
    getTextFromFile
  };
}

export async function fetchWorkflows(): Promise<any[]> {
  await delay(800);
  return mockWorkflows;
} 