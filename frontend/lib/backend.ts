// /lib/backend.js

import axios, { AxiosError } from 'axios';
import { access } from 'fs';
import _, { get, set } from 'lodash';
import { useParams } from 'next/navigation';
import { useEffect, useState, useCallback } from 'react';

import { verifyRoleSignature } from './cryptoUtils';

export const thirdaiPlatformBaseUrl = typeof window !== 'undefined' ? window.location.origin : '';
export const deploymentBaseUrl = typeof window !== 'undefined' ? window.location.origin : '';

export function getAccessToken(redirectToLogin: boolean = true): string | null {
  const accessToken = localStorage.getItem('accessToken');

  if (!accessToken && redirectToLogin) {
    // Redirect to login page
    if (
      process.env.NEXT_PUBLIC_IDENTITY_PROVIDER &&
      process.env.NEXT_PUBLIC_IDENTITY_PROVIDER.toLowerCase().includes('keycloak')
    ) {
      window.location.href = '/login-keycloak';
    } else {
      window.location.href = '/login-email';
    }
    return null;
  }

  return accessToken;
}

export function getUsername(): string {
  const username = localStorage.getItem('username');
  if (!username) {
    throw new Error('Username is not available');
  }
  return username;
}

export interface Deployment {
  name: string;
  deployment_username: string;
  model_name: string;
  model_username: string;
  status: string;
  metadata: any;
  modelID: string;
}

export interface ApiResponse {
  status_code: number;
  message: string;
  data: Deployment[];
}

//TODO: ask @Peter to confirm the use-case of this function
export async function listDeployments(deployment_id: string): Promise<Deployment[]> {
  const accessToken = getAccessToken(); // Ensure this function is implemented to get the access token
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  try {
    const response = await axios.get<ApiResponse>(
      `${thirdaiPlatformBaseUrl}/api/deploy/list-deployments`,
      {
        params: { deployment_id },
      }
    );
    return response.data.data;
  } catch (error) {
    console.error('Error listing deployments:', error);
    alert('Error listing deployments:' + error);
    throw new Error('Failed to list deployments');
  }
}

// Update the base interface to match the API response structure

interface BaseStatusResponse {
  data: {
    model_identifier: string; // Changed from model_id to model_identifier
    messages: string[];
  };
}

interface BaseDeployStatusResponse {
  data: {
    model_id: string;
    messages: string[];
  };
}

interface DeployStatusResponse {
  status: string;
  errors: string[];
  warnings: string[];
}

export function getDeployStatus(model_id: string): Promise<DeployStatusResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .get<DeployStatusResponse>(`${thirdaiPlatformBaseUrl}/api/v2/deploy/${model_id}/status`)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.detail || 'Failed to get deployment status'));
        } else {
          reject(new Error('Failed to get deployment status'));
        }
      });
  });
}

interface TrainStatusResponse {
  status: 'pending' | 'running' | 'complete' | 'failed';
  errors: string[];
  warnings: string[];
}

export async function getTrainingStatus(modelId: string): Promise<TrainStatusResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  try {
    const response = await axios.get<TrainStatusResponse>(
      `${thirdaiPlatformBaseUrl}/api/v2/train/${encodeURIComponent(modelId)}/status`
    );

    if (!response.data) {
      throw new Error('Invalid response format');
    }

    return response.data;
  } catch (err) {
    if (axios.isAxiosError(err)) {
      throw new Error(err.response?.data?.detail || 'Failed to get training status');
    }
    throw err;
  }
}

interface LogEntry {
  stderr: string;
  stdout: string;
}

interface LogResponse {
  data: LogEntry[]; // Now it's an array of LogEntry objects
}

export function getTrainingLogs(modelIdentifier: string): Promise<LogResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .get(
        `${thirdaiPlatformBaseUrl}/api/train/logs?model_identifier=${encodeURIComponent(modelIdentifier)}`
      )
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

export function getDeploymentLogs(modelIdentifier: string): Promise<LogResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .get(
        `${thirdaiPlatformBaseUrl}/api/deploy/logs?model_identifier=${encodeURIComponent(modelIdentifier)}`
      )
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

interface StopResponse {
  data: {
    deployment_id: string;
  };
  status: string;
}

export function stopDeploy(values: {
  deployment_identifier: string;
  model_identifier: string;
}): Promise<StopResponse> {
  // Retrieve the access token from local storage
  const accessToken = getAccessToken();

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .post(
        `${thirdaiPlatformBaseUrl}/api/deploy/stop?deployment_identifier=${encodeURIComponent(values.deployment_identifier)}&model_identifier=${encodeURIComponent(values.model_identifier)}`
      )
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

interface DeploymentData {
  model_id: string;
  model_identifier: string;
  status: string;
}

interface DeploymentResponse {
  data: DeploymentData;
  message: string;
  status: string;
}

export function deployModel(values: {
  deployment_name: string;
  model_identifier: string;
  use_llm_guardrail?: boolean;
  token_model_identifier?: string;
}): Promise<DeploymentResponse> {
  const accessToken = getAccessToken();

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  let params;

  if (values.token_model_identifier) {
    params = new URLSearchParams({
      deployment_name: values.deployment_name,
      model_identifier: values.model_identifier,
      use_llm_guardrail: values.use_llm_guardrail ? 'true' : 'false',
      token_model_identifier: values.token_model_identifier,
    });
  } else {
    params = new URLSearchParams({
      deployment_name: values.deployment_name,
      model_identifier: values.model_identifier,
      use_llm_guardrail: values.use_llm_guardrail ? 'true' : 'false',
    });
  }

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/deploy/run?${params.toString()}`)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

interface TrainNdbParams {
  name: string;
  formData: FormData;
}

interface UploadResponse {
  upload_id: string;
}

export async function train_ndb({ name, formData }: TrainNdbParams): Promise<any> {
  const accessToken = getAccessToken();

  // Step 1: Upload files first
  const uploadResponse = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/upload-data`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: formData,
  });

  if (!uploadResponse.ok) {
    const error = await uploadResponse.json();
    throw new Error(error.detail || 'Failed to upload files');
  }

  const { upload_id } = (await uploadResponse.json()) as UploadResponse;
  console.log('Upload response:', upload_id); // Debug log

  // Step 2: Train NDB with upload ID
  const requestData = {
    model_name: name,
    model_options: {},
    data: {
      unsupervised_files: [
        {
          path: upload_id,
          location: 'upload',
        },
      ],
    },
  };

  const trainResponse = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/ndb`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(requestData),
  });

  if (!trainResponse.ok) {
    const error = await trainResponse.json();
    throw new Error(error.detail || 'Failed to train NDB model');
  }
  const trainData = await trainResponse.json();
  console.log('Train NDB response:', trainData);
  return trainData;
}

// src/interfaces/TrainNdbParams.ts
export interface JobOptions {
  allocation_cores: number;
  allocation_memory: number;
  // Add other JobOptions fields as necessary
}

export interface RetrainNdbParams {
  model_name: string;
  base_model_id: string;
  job_options: JobOptions;
}

export function retrain_ndb({
  model_name,
  base_model_id,
  job_options,
}: RetrainNdbParams): Promise<any> {
  // Retrieve the access token from local storage or any other storage mechanism
  const accessToken = getAccessToken();
  const requestBody = JSON.stringify({
    model_name,
    base_model_id,
    job_options,
  });

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/v2/train/ndb-retrain`, requestBody)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.message || 'Failed to retrain model'));
        } else {
          reject(new Error('Failed to retrain model'));
        }
      });
  });
}

interface UploadResponse {
  upload_id: string;
}

export async function uploadDocument(files: FileList | File): Promise<UploadResponse> {
  if (!files) {
    throw new Error('No files provided');
  }

  const accessToken = getAccessToken();
  const formData = new FormData();

  // Handle single file upload
  if (files instanceof File) {
    formData.append('files', files, files.name);
    try {
      const response = await axios.post<UploadResponse>(
        `${thirdaiPlatformBaseUrl}/api/v2/train/upload-data`,
        formData,
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'multipart/form-data',
          },
        }
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data) {
        throw new Error(error.response.data.message || 'Upload failed');
      }
      throw new Error('Failed to upload file');
    }
  }

  // Handle FileList upload
  if (files.length === 0) {
    throw new Error('No files provided');
  }

  // Group files by their categories
  const categoryMap = new Map<string, File[]>();
  Array.from(files).forEach((file) => {
    const pathParts = file.webkitRelativePath.split('/');
    if (pathParts.length < 3) {
      throw new Error('Invalid folder structure. Files must be within category folders.');
    }

    const category = pathParts[1];
    if (!categoryMap.has(category)) {
      categoryMap.set(category, []);
    }
    categoryMap.get(category)?.push(file);
  });

  if (categoryMap.size === 0) {
    throw new Error('No valid categories found');
  }

  // Add files to FormData maintaining category structure
  categoryMap.forEach((categoryFiles, category) => {
    categoryFiles.forEach((file) => {
      formData.append('files', file, file.webkitRelativePath);
    });
  });

  try {
    const response = await axios.post<UploadResponse>(
      `${thirdaiPlatformBaseUrl}/api/v2/train/upload-data`,
      formData,
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(error.response.data.message || 'Upload failed');
    }
    throw new Error('Failed to upload files');
  }
}

interface TrainDocumentClassifierParams {
  modelName: string;
  files: FileList;
  testSplit?: number;
  nTargetClasses?: number; // Add this to accept the dynamic number of classes
}

export async function trainDocumentClassifier({
  modelName,
  files,
  testSplit = 0.1,
  nTargetClasses,
}: TrainDocumentClassifierParams): Promise<any> {
  const accessToken = getAccessToken();

  try {
    const formData = new FormData();

    // Add all document files to FormData
    Array.from(files).forEach((file) => {
      formData.append('files', file, file.webkitRelativePath);
    });

    // Prepare file info with webkitRelativePath to preserve directory structure
    const fileInfo = {
      supervised_files: Array.from(files).map((file) => ({
        filename: file.name,
        content_type: file.type,
        path: file.webkitRelativePath,
        location: 'local',
      })),
      test_files: [],
    };
    formData.append('file_info', JSON.stringify(fileInfo));

    // Model options for document classification
    const modelOptions = {
      model_type: 'udt',
      udt_options: {
        udt_sub_type: 'document',
        text_column: 'text',
        label_column: 'label',
        n_target_classes: nTargetClasses, // Use the dynamically passed number of classes
        word_limit: 1000, // Configure word limit
      },
      train_options: {
        test_split: testSplit,
      },
    };
    formData.append('model_options', JSON.stringify(modelOptions));

    // Job options
    const jobOptions = {
      allocation_cores: 2,
      allocation_memory: 16000,
    };
    formData.append('job_options', JSON.stringify(jobOptions));

    // Train the model
    const params = new URLSearchParams({ model_name: modelName });
    const response = await axios.post(
      `${thirdaiPlatformBaseUrl}/api/train/udt?${params.toString()}`,
      formData,
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'multipart/form-data',
        },
      }
    );

    if (response.data?.status === 'failed') {
      throw new Error(response.data.message || 'Failed to train model');
    }

    return response.data;
  } catch (error) {
    console.error('Training error:', error);
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(
        error.response.data.message || 'Failed to train document classification model'
      );
    }
    throw error instanceof Error
      ? error
      : new Error('Failed to train document classification model');
  }
}

export async function validateSentenceClassifierCSV(file: File) {
  const accessToken = getAccessToken();
  const formData = new FormData();
  formData.append('file', file);

  try {
    const response = await axios.post(
      `${thirdaiPlatformBaseUrl}/api/train/validate-text-classification-csv`,
      formData,
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'multipart/form-data',
        },
      }
    );

    return {
      valid: response.data.status === 'success',
      message: response.data.message,
      labels: response.data.data?.labels || [],
    };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const errorMessage = error.response?.data?.message || 'Failed to validate CSV';
      return {
        valid: false,
        message: errorMessage,
        labels: [],
      };
    }
    return {
      valid: false,
      message: 'Failed to validate CSV',
      labels: [],
    };
  }
}

interface TrainTextClassifierParams {
  modelName: string;
  file: File;
  labels: string[];
  testSplit?: number;
}

export function trainTextClassifierWithCSV({
  modelName,
  file,
  labels,
  testSplit = 0.1,
}: TrainTextClassifierParams): Promise<any> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  // Create FormData instance to handle file upload
  const formData = new FormData();
  formData.append('files', file);

  // Add file info with correct location type
  const fileInfo = {
    supervised_files: [
      {
        filename: file.name,
        content_type: file.type,
        path: file.name,
        location: 'local',
      },
    ],
    test_files: [], // No test files for now, will be split from training data
  };
  formData.append('file_info', JSON.stringify(fileInfo));

  // Model options for text classification with TextClassificationOptions
  const modelOptions = {
    model_type: 'udt',
    udt_options: {
      udt_sub_type: 'text',
      text_column: 'text', // Column containing the text
      label_column: 'label', // Column containing the label
      n_target_classes: labels.length, // Number of unique labels
      target_labels: labels, // Array of label names
    },
    train_options: {
      test_split: testSplit,
    },
  };
  formData.append('model_options', JSON.stringify(modelOptions));

  // Job options (using defaults)
  const jobOptions = {
    allocation_cores: 1,
    allocation_memory: 8000,
  };
  formData.append('job_options', JSON.stringify(jobOptions));

  // Create URL with query parameters
  const params = new URLSearchParams({
    model_name: modelName,
    base_model_identifier: '', // Empty string for new model
  });

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/train/udt?${params.toString()}`, formData)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const axiosError = err as AxiosError;
          if (axiosError.response && axiosError.response.data) {
            reject(
              new Error(
                (axiosError.response.data as any).detail ||
                  'Failed to train text classification model'
              )
            );
          } else {
            reject(new Error('Failed to train text classification model'));
          }
        } else {
          reject(new Error('Failed to train text classification model'));
        }
      });
  });
}

interface RetrainTokenClassifierParams {
  model_name: string;
  base_model_id: string;
}

export async function retrainTokenClassifier({
  model_name,
  base_model_id,
}: RetrainTokenClassifierParams): Promise<any> {
  const accessToken = getAccessToken();

  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/nlp-token-retrain`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model_name,
        base_model_id,
      }),
    });
    console.log('Response data in retrainTokenClassifier:', response);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.detail || 'Failed to retrain UDT model');
    }

    return await response.json();
  } catch (err) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error('Failed to retrain UDT model');
  }
}

interface TrainUDTWithCSVParams {
  model_name: string;
  file: File;
  base_model_identifier: string;
  test_split?: number;
}

export function trainUDTWithCSV({
  model_name,
  file,
  base_model_identifier,
  test_split = 0.1,
}: TrainUDTWithCSVParams): Promise<any> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  // Create FormData instance to handle file upload
  const formData = new FormData();
  formData.append('files', file);

  // Add file info with correct location type
  const fileInfo = {
    supervised_files: [
      {
        filename: file.name,
        content_type: file.type,
        path: file.name,
        location: 'local',
      },
    ],
    test_files: [],
  };
  formData.append('file_info', JSON.stringify(fileInfo));

  // Simplified model options for token classification
  const modelOptions = {
    udt_options: {
      udt_sub_type: 'token',
      source_column: '',
      target_column: '',
      target_labels: [],
    },
    train_options: {
      test_split: test_split,
    },
  };
  formData.append('model_options', JSON.stringify(modelOptions));

  // Create URL with query parameters
  const params = new URLSearchParams({
    model_name,
    base_model_identifier,
  });

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/train/udt?${params.toString()}`, formData)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          const axiosError = err as AxiosError;
          if (axiosError.response && axiosError.response.data) {
            reject(
              new Error(
                (axiosError.response.data as any).detail || 'Failed to train UDT model with CSV'
              )
            );
          } else {
            reject(new Error('Failed to train UDT model with CSV'));
          }
        } else {
          reject(new Error('Failed to train UDT model with CSV'));
        }
      });
  });
}

interface APIResponse {
  status: string;
  message: string;
  data?: {
    valid: boolean;
    labels?: string[];
  };
}

interface ValidationResult {
  labels?: string[];
}

interface ValidationParams {
  upload_id: string;
  type: string;
}

export async function validateCSV({
  upload_id,
  type,
}: ValidationParams): Promise<ValidationResult> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/validate-trainable-csv`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ upload_id, type }),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.detail || 'Failed to validate CSV file');
    }
    const returnData = await response.json();
    return returnData;
  } catch (error) {
    return {
      labels: [],
    };
  }
}

interface TokenOptions {
  target_labels: string[];
  source_column: string;
  target_column: string;
  default_tag: string;
}

interface TrainTokenClassifierParams {
  model_name: string;
  base_model_id?: string | null;
  model_options: TokenOptions;
  data: {
    supervised_files: Array<{
      path: string;
      location: string;
    }>;
    test_files?: Array<{
      path: string;
      location: string;
    }>;
  };
  train_options?: {
    epochs?: number;
    learning_rate?: number;
    batch_size?: number;
    max_in_memory_batches?: number;
    test_split?: number;
  };
  job_options?: {
    allocation_cores?: number;
    allocation_memory?: number;
  };
}

export async function trainTokenClassifierWithCSV(
  params: TrainTokenClassifierParams
): Promise<TrainTokenClassifierResponse> {
  const accessToken = getAccessToken();

  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/nlp-token`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model_name: params.model_name,
        base_model_id: params.base_model_id,
        model_options: params.model_options,
        data: params.data,
        train_options: params.train_options || {},
        job_options: params.job_options || {},
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.detail || 'Failed to train token classifier');
    }

    const data = await response.json();
    return { model_id: data.model_id };
  } catch (error) {
    console.error('Error training token classifier:', error);
    throw error;
  }
}

// types.ts
export interface MetricValues {
  precision: number;
  recall: number;
  fmeasure: number;
}

export interface LabelMetrics {
  [labelName: string]: MetricValues;
}

export interface TrainingExample {
  source: string;
  target: string;
  predictions: string;
  index: number;
}

export interface LabelExamples {
  [labelName: string]: TrainingExample[];
}

export interface ExampleCategories {
  true_positives: LabelExamples;
  false_positives: LabelExamples;
  false_negatives: LabelExamples;
}

export interface TrainReportData {
  before_train_metrics: LabelMetrics;
  after_train_metrics: LabelMetrics;
  after_train_examples: ExampleCategories;
}

export interface TrainReportResponse {
  status: string;
  message: string;
  data: TrainReportData;
}

// api.ts
export async function getTrainReport(modelId: string): Promise<TrainReportResponse> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/train/${modelId}/report`, {
      method: 'GET',
      headers: {
        contentType: 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.detail || 'Failed to get training report');
    }
    return await response.json();
  } catch (err) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error('Failed to get training report');
  }
}

export interface EnterpriseSearchOptions {
  retrieval_id: string;
  guardrail_id?: string;
  nlp_classifier_id?: string;
  llm_provider?: string;
  default_mode?: string;
  model_name: string;
}

interface CreateWorkflowParams {
  workflow_name: string;
  options: EnterpriseSearchOptions;
}

export function create_enterprise_search_workflow({
  workflow_name,
  options,
}: CreateWorkflowParams): Promise<any> {
  const accessToken = getAccessToken();

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/v2/workflow/enterprise-search`, options)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.message || 'Failed to create workflow'));
        } else {
          reject(new Error('Failed to create workflow'));
        }
      });
  });
}

export interface Attributes {
  llm_provider?: string;
  default_mode?: string;
  retrieval_id?: string;
  guardrail_id?: string;
  nlp_classifier_id?: string;
}

interface Dependency {
  model_id: string;
  model_name: string;
  type: string;
  sub_type: string;
  username: string;
}

export interface Workflow {
  model_id: string;
  model_name: string;
  type: string;
  access: string;
  train_status: string;
  deploy_status: string;
  publish_date: string;
  username: string;
  user_email: string;
  team_id: string | null;
  attributes: Attributes;
  dependencies: Dependency[];
  size: string;
  size_in_memory: string;
}

export async function fetchWorkflows(): Promise<Workflow[]> {
  const accessToken = getAccessToken(); // Ensure this function retrieves the correct access token

  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/model/list`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.detail || 'Failed to fetch workflows');
    }

    const data = await response.json();
    return data; // Assuming the data is in the `data` field
  } catch (error) {
    console.error('Error fetching workflows:', error);
    throw error;
  }
}

function createModelIdentifier(username: string, model_name: string): string {
  return `${username}/${model_name}`;
}

interface StartWorkflowResponse {
  status_code: number;
  message: string;
  data: {
    models: { id: string; name: string }[];
  };
}

interface StartWorkflowRequest {
  deployment_name?: string;
  autoscaling_enabled: boolean;
  autoscaling_min?: number;
  autoscaling_max?: number;
  memory?: number;
}

export function start_workflow(
  model_id: string,
  autoscalingEnabled: boolean
): Promise<StartWorkflowResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  const requestBody: StartWorkflowRequest = {
    autoscaling_enabled: autoscalingEnabled,
  };

  return new Promise((resolve, reject) => {
    axios
      .post<StartWorkflowResponse>(
        `${thirdaiPlatformBaseUrl}/api/v2/deploy/${model_id.toString()}`,
        requestBody
      )
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.detail || 'Failed to start workflow'));
        } else {
          reject(new Error('Failed to start workflow'));
        }
      });
  });
}

interface StopWorkflowResponse {
  status_code: number;
  message: string;
}

export function stop_workflow(model_id: string): Promise<StopWorkflowResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .delete<StopWorkflowResponse>(`${thirdaiPlatformBaseUrl}/api/v2/deploy/${model_id}`)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.detail || 'Failed to stop workflow'));
        } else {
          reject(new Error('Failed to stop workflow'));
        }
      });
  });
}

interface DeleteWorkflowResponse {
  status_code: number;
  message: string;
}

export async function delete_workflow(model_id: string): Promise<DeleteWorkflowResponse> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
  return new Promise((resolve, reject) => {
    axios
      .delete<DeleteWorkflowResponse>(
        `${thirdaiPlatformBaseUrl}/api/v2/model/${model_id.toString()}`
      )
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        console.error('Error deleting workflow:', err);
        alert('Error deleting workflow:' + err);
        reject(new Error('Failed to delete workflow'));
      });
  });
}

interface ModelDependency {
  model_id: string;
  model_name: string;
  type: string;
  username: string;
}

interface ModelDetails {
  model_id: string;
  model_name: string;
  type: string;
  access: string;
  train_status: string;
  deploy_status: string;
  publish_date: string;
  user_email: string;
  username: string;
  team_id: string | null;
  attributes: Record<string, string>;
  dependencies: ModelDependency[];
}

export async function getWorkflowDetails(model_id: string): Promise<ModelDetails> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  try {
    const response = await axios.get<ModelDetails>(
      `${thirdaiPlatformBaseUrl}/api/v2/model/${model_id}`
    );
    return response.data;
  } catch (err) {
    const error = err as any;
    if (error.response?.data) {
      throw new Error(error.response.data.detail || 'Failed to fetch model details');
    }
    throw new Error('Failed to fetch model details');
  }
}

export async function userEmailLogin(
  email: string,
  password: string,
  setAccessToken: (token: string | null | undefined) => void
): Promise<any> {
  try {
    const apiUrl = `${thirdaiPlatformBaseUrl}/api/v2/user/login`;

    const base64Credentials = btoa(`${email}:${password}`);
    const response = await fetch(apiUrl, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        Authorization: `Basic ${base64Credentials}`,
      },
    });

    if (!response.ok) {
      const errorMessage = await response.text();
      throw new Error(`Request failed: ${response.status} - ${errorMessage}`);
    }

    const data = await response.json();
    const accessToken = data.access_token;

    if (accessToken) {
      setAccessToken(accessToken); // Set the token in context first
    } else {
      throw new Error('No access token received from server');
    }

    return data;
  } catch (error) {
    console.error('Error logging in:', error);
    setAccessToken(null);
    throw error;
  }
}

//TODO: check once before merging....
export async function SyncKeycloakUser(
  accessToken: string,
  setAccessToken: (token: string | null | undefined) => void
): Promise<any> {
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/login-with-token`, {
      method: 'POST',
      headers: {
        contentType: 'application/json',
      },
      body: JSON.stringify({ access_token: accessToken }),
    });

    if (!response.ok) {
      const errorMessage = await response.text();
      throw new Error(`Request failed: ${response.status} - ${errorMessage}`);
    }
    const data = await response.json();
    if (accessToken) {
      setAccessToken(accessToken); // Set the token in context first
    } else {
      throw new Error('No access token received from server');
    }

    return data;
  } catch (error) {
    console.error('Error logging in:', error);
    setAccessToken(null);
    throw error;
  }
}

export async function userRegister(
  email: string,
  password: string,
  username: string
): Promise<any> {
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password, username }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to register user');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error during user registration:', error);
    throw error;
  }
}

interface TokenClassificationExample {
  name: string;
  example: string;
  description: string;
}

function tokenClassifierDatagenForm(modelGoal: string, categories: Category[]) {
  const tags = categories.map((category) => ({
    name: category.name,
    examples: category.examples.map((ex) => ex.text),
    description: category.description,
  }));
  return {
    sub_type: 'token',
    task_prompt: modelGoal,
    tags: tags,
  };
}

interface TrainTokenClassifierResponse {
  model_id: string;
}

type Example = {
  text: string;
};

type Category = {
  name: string;
  examples: Example[];
  description: string;
};
interface Tag {
  name: string;
  examples?: string[];
  description?: string;
  status?: 'uninserted';
}

interface TokenOptionsDatagen {
  tags: Tag[];
  num_sentences_to_generate?: number;
  num_samples_per_tag?: number;
  templates_per_sample?: number;
}

interface DatagenRequest {
  model_name: string;
  base_model_id?: string | null;
  task_prompt: string;
  llm_provider?: string;
  test_size?: number;
  token_options: TokenOptionsDatagen;
  train_options?: {
    epochs?: number;
    learning_rate?: number;
    batch_size?: number;
    max_in_memory_batches?: number;
    test_split?: number;
  };
  job_options?: {
    allocation_cores?: number;
    allocation_memory?: number;
  };
}
export function trainTokenClassifier(
  modelName: string,
  modelGoal: string,
  categories: Category[]
): Promise<TrainTokenClassifierResponse> {
  // Retrieve the access token from local storage
  const accessToken = getAccessToken();

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  const requestData: DatagenRequest = {
    model_name: modelName,
    task_prompt: modelGoal,
    token_options: {
      tags: categories.map((category) => ({
        name: category?.name,
        examples: category?.examples.map((ex) => ex.text),
        description: category?.description,
      })),
      num_sentences_to_generate: 2000,
      num_samples_per_tag: 10,
      templates_per_sample: 4,
    },
  };

  console.log('Request data:', requestData);
  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/v2/train/nlp-datagen`, requestData)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response && err.response.data) {
          reject(new Error(err.response.data.detail || 'Failed to run model'));
        } else {
          reject(new Error('Failed to run model'));
        }
      });
  });
}

interface SentenceClassificationExample {
  name: string;
  example: string;
  description: string;
}

function sentenceClassifierDatagenForm(examples: SentenceClassificationExample[]) {
  const labels = examples.map((example) => ({
    name: example.name,
    examples: [example.example],
    description: example.description,
  }));

  const numSentences = 10_000;
  return {
    sub_type: 'text',
    samples_per_label: Math.max(Math.ceil(numSentences / labels.length), 50),
    labels: labels,
  };
}

interface TrainSentenceClassifierResponse {
  status_code: number;
  message: string;
  model_id: string;
  user_id: string;
}

interface TextOptions {
  labels: {
    name: string;
    examples?: string[];
    description?: string;
    status?: 'uninserted';
  }[];
  samples_per_label: number;
  sub_type: string;
}

export interface NLPDatagenRequest {
  model_name: string;
  base_model_id?: string | null;
  task_prompt: string;
  llm_provider?: string;
  test_size?: number;
  text_options: TextOptions;
  train_options?: {
    epochs?: number;
    learning_rate?: number;
    batch_size?: number;
    max_in_memory_batches?: number;
    test_split?: number;
  };
  job_options?: {
    allocation_cores?: number;
    allocation_memory?: number;
  };
}

export function trainSentenceClassifier(
  modelName: string,
  modelGoal: string,
  examples: SentenceClassificationExample[]
): Promise<TrainSentenceClassifierResponse> {
  const accessToken = getAccessToken();

  const requestData: NLPDatagenRequest = {
    model_name: modelName,
    task_prompt: modelGoal,
    llm_provider: 'openai',
    test_size: 0.1,
    text_options: {
      ...sentenceClassifierDatagenForm(examples), // Spread operator to include all properties
    },
    train_options: {
      epochs: 5,
      learning_rate: 0.0001,
      batch_size: 1000,
      test_split: 0.2,
    },
  };

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/v2/train/nlp-datagen`, requestData, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
      })
      .then((res) => {
        console.log(res);
        resolve(res.data);
      })
      .catch((err) => {
        if (err.response?.data) {
          reject(new Error(err.response.data.detail || 'Failed to run model'));
        } else {
          reject(new Error('Failed to run model'));
        }
      });
  });
}

function useAccessToken() {
  const [accessToken, setAccessToken] = useState<string | undefined>();
  useEffect(() => {
    const accessToken = localStorage.getItem('accessToken');
    if (!accessToken) {
      throw new Error('Access token is not available');
    }
    setAccessToken(accessToken);
  }, []);

  return accessToken;
}

interface UseLabelsOptions {
  deploymentUrl: string;
  maxRecentLabels?: number;
}

interface UseLabelsResult {
  allLabels: Set<string>;
  recentLabels: string[];
  error: Error | null;
  isLoading: boolean;
  refresh: () => Promise<void>;
}

export function useLabels({
  deploymentUrl,
  maxRecentLabels = 5,
}: UseLabelsOptions): UseLabelsResult {
  const [allLabels, setAllLabels] = useState<Set<string>>(new Set());
  const [recentLabels, setRecentLabels] = useState<string[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    try {
      const accessToken = getAccessToken();
      axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

      const response = await axios.get<{ data: string[] }>(`${deploymentUrl}/get_labels`);
      const labels = response.data.data;

      setAllLabels((prevLabels) => {
        const newLabels = new Set(prevLabels);
        labels.forEach((label: string) => {
          if (!prevLabels.has(label)) {
            newLabels.add(label);
            setRecentLabels((prev) => [label, ...prev].slice(0, maxRecentLabels));
          }
        });
        return newLabels;
      });

      setError(null);
    } catch (err) {
      console.error('Error fetching labels:', err);
      setError(err instanceof Error ? err : new Error('An unknown error occurred'));
    } finally {
      setIsLoading(false);
    }
  }, [deploymentUrl, maxRecentLabels]);

  return { allLabels, recentLabels, error, isLoading, refresh };
}

interface Sample {
  tokens: string[];
  tags: string[];
}

interface UseRecentSamplesOptions {
  deploymentUrl: string;
  maxRecentSamples?: number;
}

interface UseRecentSamplesResult {
  recentSamples: Sample[];
  error: Error | null;
  isLoading: boolean;
  refresh: () => Promise<void>;
}

export function useRecentSamples({
  deploymentUrl,
  maxRecentSamples = 5,
}: UseRecentSamplesOptions): UseRecentSamplesResult {
  const [recentSamples, setRecentSamples] = useState<Sample[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    try {
      const accessToken = getAccessToken();
      axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

      const response = await axios.get<{ data: Sample[] }>(`${deploymentUrl}/get_recent_samples`);
      setRecentSamples(response.data.data.slice(0, maxRecentSamples));
      setError(null);
    } catch (err) {
      console.error('Error fetching recent samples:', err);
      setError(err instanceof Error ? err : new Error('An unknown error occurred'));
    } finally {
      setIsLoading(false);
    }
  }, [deploymentUrl, maxRecentSamples]);

  return { recentSamples, error, isLoading, refresh };
}

export interface TokenClassificationResult {
  query_text: string;
  tokens: string[];
  predicted_tags: string[][];
}
export interface PredictionResponse {
  prediction_results: TokenClassificationResult;
  time_taken: number;
}

export interface InsertSamplePayload {
  tokens: string[];
  tags: string[];
}

export function useTokenClassificationEndpoints() {
  const accessToken = useAccessToken();
  const params = useParams();
  // console.log(params);
  const workflowId = params.deploymentId as string;
  const [workflowName, setWorkflowName] = useState<string>('');
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>();

  // console.log('PARAMS', params);

  useEffect(() => {
    const init = async () => {
      const modelInfo = await getWorkflowDetails(workflowId);
      setWorkflowName(modelInfo.model_name);
      setDeploymentUrl(`${thirdaiPlatformBaseUrl}/${modelInfo.model_id}`);
    };
    init();
  }, []);

  const predict = async (query: string): Promise<PredictionResponse> => {
    // Set the default authorization header for axios
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    const model_id = workflowId;
    console.log('Deployment URL:', deploymentUrl);
    try {
      const response = await axios.post(`${deploymentUrl}/predict`, {
        text: query,
        top_k: 1,
      });
      return response.data.data;
    } catch (error) {
      console.error('Error predicting tokens:', error);
      alert('Error predicting tokens:' + error);
      throw new Error('Failed to predict tokens');
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

  const getStats =
    deploymentUrl &&
    (async (): Promise<DeploymentStats> => {
      axios.defaults.headers.common.Authorization = `Bearer ${getAccessToken()}`;
      try {
        console.log(deploymentUrl);
        const response = await axios.get(`${deploymentUrl}/stats`);
        return {
          system: {
            header: ['Name', 'Description'],
            rows: [
              ['CPU', '12 vCPUs'],
              ['CPU Model', 'Intel(R) Xeon(R) CPU E5-2680 v3 @ 2.50GHz'],
              ['Memory', '64 GB RAM'],
              ['System Uptime', formatTime(response.data.data.uptime)],
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
                formatAmount(response.data.data.past_hour.tokens_identified),
                formatAmount(response.data.data.past_hour.queries_ingested),
                formatAmount(response.data.data.past_hour.queries_ingested_bytes) + 'B',
              ],
              [
                'Total',
                formatAmount(response.data.data.total.tokens_identified),
                formatAmount(response.data.data.total.queries_ingested),
                formatAmount(response.data.data.total.queries_ingested_bytes) + 'B',
              ],
            ],
          },
        };
      } catch (error) {
        console.error('Error fetching stats:', error);
        alert('Error fetching stats:' + error);
        throw new Error('Error fetching stats.');
      }
    });

  const insertSample = async (sample: InsertSamplePayload): Promise<void> => {
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    try {
      await axios.post(`${deploymentUrl}/insert_sample`, sample);
    } catch (error) {
      console.error('Error inserting sample:', error);
      alert('Error inserting sample:' + error);
      throw new Error('Failed to insert sample');
    }
  };

  const addLabel = async (labels: {
    tags: { name: string; description: string }[];
  }): Promise<void> => {
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    try {
      await axios.post(`${deploymentUrl}/add_labels`, labels);
    } catch (error) {
      console.error('Error adding label:', error);
      alert('Error adding label:' + error);
      throw new Error('Failed to add label');
    }
  };

  const getLabels = async (): Promise<string[]> => {
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    try {
      const response = await axios.get(`${deploymentUrl}/get_labels`);
      return response.data.data;
    } catch (error) {
      console.error('Error fetching labels:', error);
      alert('Error fetching labels:' + error);
      throw new Error('Failed to fetch labels');
    }
  };

  const getTextFromFile = async (file: File): Promise<string[]> => {
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await axios.post(`${deploymentUrl}/get-text`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error parsing file:', error);
      alert('Error parsing file:' + error);
      throw new Error('Failed to parse file');
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

interface TextClassificationResult {
  query_text: string;
  predicted_classes: [string, number][];
}

interface PredictionClass {
  class: string;
  score: number;
}

interface PredictionResult {
  status: string;
  message: string;
  data: {
    prediction_results: {
      query_text: string;
      predicted_classes: PredictionClass[];
    };
    time_taken: number;
  };
}

export function useTextClassificationEndpoints() {
  const accessToken = useAccessToken();
  const params = useParams();
  const workflowId = params.deploymentId as string;
  const [workflowName, setWorkflowName] = useState<string>('');
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>();

  console.log('PARAMS', params);

  useEffect(() => {
    const init = async () => {
      const modelInfo = await getWorkflowDetails(workflowId);
      setWorkflowName(modelInfo.model_name);
      setDeploymentUrl(`${thirdaiPlatformBaseUrl}/${modelInfo.model_id}`);
    };
    init();
  }, []);

  const predict = async (query: string): Promise<PredictionResult> => {
    // Set the default authorization header for axios
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

    try {
      const response = await axios.post<PredictionResult>(`${deploymentUrl}/predict`, {
        text: query,
        top_k: 5,
      });

      // Return the full response data structure
      return response.data;
    } catch (error) {
      console.error('Error predicting tokens:', error);
      alert('Error predicting tokens:' + error);
      throw new Error('Failed to predict tokens');
    }
  };

  const getTextFromFile = async (file: File): Promise<string[]> => {
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await axios.post(`${deploymentUrl}/get-text`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error parsing file:', error);
      alert('Error parsing file:' + error);
      throw new Error('Failed to parse file');
    }
  };

  return {
    workflowName,
    getTextFromFile,
    predict,
  };
}

export function useSentimentClassification(workflowId: string | null) {
  const accessToken = useAccessToken(); // Assuming this function exists
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>();

  useEffect(() => {
    if (!workflowId) return;

    const init = async () => {
      try {
        setDeploymentUrl(`${deploymentBaseUrl}/${workflowId}`);
      } catch (error) {
        console.error('Error fetching sentiment model details:', error);
        alert('Error fetching sentiment model details: ' + error);
      }
    };

    init();
  }, [workflowId, accessToken]);

  // Function to predict sentiment based on the input query
  const predictSentiment = async (query: string): Promise<TextClassificationResult> => {
    if (!deploymentUrl) {
      throw new Error('Sentiment classifier deployment URL not set');
    }

    try {
      // Corrected the key from 'query' to 'text'
      const response = await axios.post(`${deploymentUrl}/predict`, { text: query, top_k: 5 });
      return response.data.data;
    } catch (error) {
      console.error('Error predicting sentiment:', error);
      alert('Error predicting sentiment: ' + error);
      throw new Error('Failed to predict sentiment');
    }
  };

  // Return the predict function
  return {
    predictSentiment,
  };
}

export async function piiDetect(
  query: string,
  workflowId: string
): Promise<TokenClassificationResult> {
  console.log('workflowId in piiDetect:', workflowId);
  try {
    // Corrected the key from 'query' to 'text'
    const response = await axios.post(`${deploymentBaseUrl}/${workflowId}/predict`, {
      text: query,
      top_k: 1,
    });
    return response.data.data;
  } catch (error) {
    console.error('Error performing pii detection:', error);
    alert('Error performing pii detection: ' + error);
    throw new Error('Failed to perform pii detection');
  }
}

export interface DeploymentStatsTable {
  header: string[];
  rows: string[][];
}

export interface DeploymentStats {
  system: DeploymentStatsTable;
  throughput: DeploymentStatsTable;
}

//// Admin access dashboard functions /////

// Define the response types for models, teams, and users
interface ModelResponse {
  access_level: string;
  domain: string;
  latency: string;
  model_id: string;
  model_name: string;
  num_params: string;
  publish_date: string;
  size: string;
  size_in_memory: string;
  sub_type: string;
  team_id: string;
  thirdai_version: string;
  training_time: string;
  type: string;
  user_email: string;
  username: string;
}

interface UserTeamInfo {
  team_id: string;
  team_name: string;
  team_admin: boolean;
}

interface UserResponse {
  email: string;
  admin: boolean;
  id: string;
  teams: UserTeamInfo[];
  username: string;
  verified: boolean;
}

interface TeamResponse {
  id: string;
  name: string;
}

export async function fetchAllModels(): Promise<ModelResponse[]> {
  const accessToken = getAccessToken(); // Ensure this function exists and correctly retrieves the access token
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/model/list`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to fetch models');
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching models:', error);
    throw error;
  }
}

export async function fetchAllTeams(): Promise<{ data: TeamResponse[] }> {
  const accessToken = getAccessToken();

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .get(`${thirdaiPlatformBaseUrl}/api/v2/team/list`)
      .then((res) => {
        resolve(res);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

export async function fetchAllUsers(): Promise<{ data: UserResponse[] }> {
  const accessToken = getAccessToken();

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .get(`${thirdaiPlatformBaseUrl}/api/v2/user/list`)
      .then((res) => {
        resolve(res);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

export async function verifyUser(user_id: string): Promise<void> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/${user_id}/verify`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        contentType: 'application/json',
      },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => {});
      alert('Error verifying user:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to verify user');
    }
  } catch (error) {
    console.error('Error verifying user:', error);
    alert('Error verifying user:' + error);
    throw error;
  }
}
// MODEL //

export async function updateModelAccessLevel(
  model_identifier: string,
  access_level: 'private' | 'protected' | 'public',
  team_id?: string
): Promise<void> {
  const accessToken = getAccessToken(); // Ensure this function is implemented elsewhere in your codebase

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  const params = new URLSearchParams({ model_identifier, access_level });

  if (access_level === 'protected' && team_id) {
    params.append('team_id', team_id);
  }

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/model/update-access-level?${params.toString()}`)
      .then(() => {
        resolve();
      })
      .catch((err) => {
        console.error('Error updating model access level:', err);
        alert('Error updating model access level:' + err);
        reject(err);
      });
  });
}

export async function deleteModel(model_id: string): Promise<void> {
  const accessToken = getAccessToken(); // Ensure this function is implemented elsewhere in your codebase

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
  return new Promise((resolve, reject) => {
    axios
      .delete(`${thirdaiPlatformBaseUrl}/api/v2/model/${model_id.toString()}`)
      .then(() => {
        resolve();
      })
      .catch((err) => {
        console.error('Error deleting model:', err);
        alert('Error deleting model:' + err);
        reject(err);
      });
  });
}

// TEAM //

interface CreateTeamResponse {
  team_id: string;
}

export async function createTeam(name: string): Promise<CreateTeamResponse | null> {
  const accessToken = getAccessToken(); // Make sure this function is implemented elsewhere in your codebase
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/team/create`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        contentType: 'application/json',
      },
      body: JSON.stringify({ name }),
    });
    console.log('response in create team:', response);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      alert('Error creating team:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to create team');
    }
    const data = await response.json();
    console.log('data in create team:', data);
    return data;
  } catch (error) {
    console.error('Error creating team:', error);
    alert('Error creating team:' + error);
    return null;
  }
}

export async function addUserToTeam(user_id: string, team_id: string): Promise<void> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(
      `${thirdaiPlatformBaseUrl}/api/v2/team/${team_id}/users/${user_id}`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          contentType: 'application/json',
        },
      }
    );
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      alert('Error adding user to team:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to add user to team');
    }
  } catch (error) {
    console.error('Error adding user to team:', error);
    alert('Error adding user to team:' + error);
  }
}

export async function assignTeamAdmin(user_id: string, team_id: string): Promise<void> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(
      `${thirdaiPlatformBaseUrl}/api/v2/team/${team_id}/admins/${user_id}`,
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          contentType: 'application/json',
        },
      }
    );
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      alert('Error assigning team admin:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to assign team admin');
    }
  } catch (error) {
    console.error('Error assigning team admin:', error);
    alert('Error assigning team admin:' + error);
  }
}

export async function removeTeamAdmin(user_id: string, team_id: string) {
  const accessToken = getAccessToken();

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
  return new Promise((resolve, reject) => {
    axios
      .delete(`${thirdaiPlatformBaseUrl}/api/v2/team/${team_id}/admins/${user_id}`)
      .then((res) => {
        resolve(res.data);
      })
      .catch((err) => {
        reject(err);
      });
  });
}

export async function deleteUserFromTeam(user_id: string, team_id: string): Promise<void> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .delete(`${thirdaiPlatformBaseUrl}/api/v2/team/${team_id}/users/${user_id}`)
      .then(() => {
        resolve();
      })
      .catch((err) => {
        console.error('Error removing user from team:', err);
        alert('Error removing user from team:' + err);
        reject(err);
      });
  });
}

export async function deleteTeamById(team_id: string): Promise<void> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/team/${team_id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
  } catch (error) {
    console.error('Error deleting team:', error);
    alert('Error deleting team:' + error);
  }
}

// USER //

export async function deleteUserAccount(user_id: string): Promise<void> {
  const accessToken = getAccessToken(); // Ensure this function is implemented elsewhere in your codebase
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/${user_id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        contentType: 'application/json',
      },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => {});
      alert('Error deleting user account:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to delete user account');
    }
  } catch (error) {
    console.error('Error deleting user account:', error);
    alert('Error deleting user account:' + error);
  }
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
}

export interface AddUserPayload {
  username: string;
  email: string;
  password: string;
}

export interface AddUserResponse {
  user_id: string;
}

export async function addUser(userData: AddUserPayload): Promise<AddUserResponse | null> {
  const accessToken = getAccessToken();
  console.log('accessToken in add user', accessToken);
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/create`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(userData),
    });
    if (!response.ok) {
      console.error('Error adding user');
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error adding user:', error);
    return null;
  }
}

export async function promoteUserToGlobalAdmin(user_id: string): Promise<void> {
  const accessToken = getAccessToken();
  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/${user_id}/admin`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        contentType: 'application/json',
      },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => {});
      alert('Error promoting user:' + errorData.detail);
      throw new Error(errorData.detail || 'Failed to promote user');
    }
  } catch (error) {
    console.error('Error promoting user:', error);
    alert('Error promoting user:' + error);
  }
}

export async function updateModel(modelIdentifier: string): Promise<void> {
  const accessToken = getAccessToken(); // Ensure this function is implemented elsewhere in your codebase

  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  const params = new URLSearchParams({ model_identifier: modelIdentifier });

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/model/update-model?${params.toString()}`)
      .then(() => {
        resolve();
      })
      .catch((err) => {
        console.error('Error updating model:', err);
        alert('Error updating model:' + err);
        reject(err);
      });
  });
}

export interface Team {
  team_id: string;
  team_name: string;
  team_admin: boolean;
}

export interface User {
  id: string;
  username: string;
  email: string;
  global_admin: boolean;
  teams: Team[];
  role_signature: string;
}
interface APIUserResponse {
  id: string;
  username: string;
  email: string;
  admin: boolean;
  teams: {
    team_id: string;
    team_name: string;
    team_admin: boolean;
  }[];
  role_signature: string;
}
export async function accessTokenUser(accessToken: string | null) {
  if (accessToken === null) {
    return null;
  }

  try {
    const response = await fetch(`${thirdaiPlatformBaseUrl}/api/v2/user/info`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }

    const data = (await response.json()) as APIUserResponse;
    const transformedData: User = {
      id: data.id,
      username: data.username,
      email: data.email,
      global_admin: data.admin,
      teams: data.teams.map((team) => ({
        team_id: team.team_id,
        team_name: team.team_name,
        team_admin: team.team_admin,
      })),
      role_signature: data.role_signature,
    };

    const expectedPayload = {
      global_admin: data.admin,
      teams: data.teams.map((team) => ({
        team_id: team.team_id,
        team_name: team.team_name,
        team_admin: team.team_admin,
      })),
    };

    // Await the asynchronous verification.
    const isValid = await verifyRoleSignature(expectedPayload, data.role_signature);
    if (!isValid) {
      console.error('Role signature verification failed');
      alert('Authorization failed. Please try again.');
      return null;
    }

    return transformedData;
  } catch (error) {
    console.error('Error fetching user info:', error);
    return null;
  }
}

export async function fetchAutoCompleteQueries(modelId: string, query: string) {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  const params = new URLSearchParams({ model_id: modelId, query });

  try {
    const response = await axios.get(`${deploymentBaseUrl}/cache/suggestions?${params.toString()}`);

    return response.data; // Assuming the backend returns the data directly
  } catch (err) {
    console.error('Error fetching autocomplete suggestions:', err);
    throw err; // Re-throwing the error to handle it in the component
  }
}

export async function fetchCachedGeneration(modelId: string, query: string) {
  const accessToken = getAccessToken();
  axios.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

  const params = new URLSearchParams({ model_id: modelId, query });

  try {
    const response = await axios.get(`${deploymentBaseUrl}/cache/query?${params.toString()}`);
    return response.data.cached_response; // Assuming the backend returns the data directly
  } catch (err) {
    console.error('Error fetching cached generation:', err);
    throw err; // Re-throwing the error to handle it in the component
  }
}

export async function temporaryCacheToken(modelId: string) {
  const accessToken = getAccessToken();
  axios.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

  const params = new URLSearchParams({ model_id: modelId });

  try {
    const response = await axios.get(`${deploymentBaseUrl}/cache/token?${params.toString()}`);
    return response.data.access_token; // Assuming the backend returns the data directly
  } catch (err) {
    console.error('Error getting temporary cache access token:', err);
    throw err; // Re-throwing the error to handle it in the component
  }
}

export async function fetchFeedback(modelId: string) {
  const accessToken = getAccessToken();
  console.log('fetchFeedback', modelId);
  try {
    const response = await axios({
      method: 'get',
      url: `${deploymentBaseUrl}/api/deploy/feedbacks`,
      params: {
        model_identifier: modelId,
      },
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    return response?.data?.data;
  } catch (error) {
    console.error('Error getting Feedback Response:', error);
    throw error;
  }
}

interface ModelOptions {
  text_column: string;
  label_column: string;
  n_target_classes: number;
  delimiter?: string;
}

interface DataFile {
  path: string;
  location: string;
}

interface TrainOptions {
  epochs?: number;
  learning_rate?: number;
  batch_size?: number;
  max_in_memory_batches?: number;
  test_split?: number;
}

interface NLPTextTrainRequest {
  model_name: string;
  doc_classification?: boolean;
  base_model_id?: string | null;
  model_options: ModelOptions;
  data: {
    supervised_files: DataFile[];
    test_files?: DataFile[];
  };
  train_options?: TrainOptions;
  job_options?: JobOptions;
}

interface NLPTextTrainResponse {
  model_id: string;
}

export async function trainNLPTextModel(params: {
  model_name: string;
  uploadId: string;
  textColumn?: string;
  labelColumn?: string;
  nTargetClasses: number;
  baseModelId?: string;
  trainOptions?: TrainOptions;
  doc_classification?: boolean;
}): Promise<NLPTextTrainResponse> {
  const accessToken = getAccessToken();

  const payload: NLPTextTrainRequest = {
    model_name: params.model_name,
    doc_classification: params.doc_classification === true,
    base_model_id: params.baseModelId || null,
    model_options: {
      text_column: params.doc_classification ? 'text' : params.textColumn || 'text',
      label_column: params.doc_classification ? 'label' : params.labelColumn || 'label',
      n_target_classes: params.nTargetClasses,
      delimiter: ',',
    },
    data: {
      supervised_files: [
        {
          path: params.uploadId,
          location: 'upload',
        },
      ],
    },
    train_options: params.trainOptions || {
      epochs: 5,
      learning_rate: 0.0001,
      batch_size: 1000,
      test_split: 0.2,
    },
    job_options: {
      allocation_cores: 4,
      allocation_memory: 2000,
    },
  };

  try {
    const response = await axios.post<NLPTextTrainResponse>(
      `${thirdaiPlatformBaseUrl}/api/v2/train/nlp-text`,
      payload,
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
      }
    );

    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data) {
      throw new Error(error.response.data.message || 'Failed to train NLP model');
    }
    throw new Error('Failed to train NLP model');
  }
}

// Knowledge extraction use case

interface KnowledgeExtractionQuestion {
  question: string;
  keywords?: string[];
}

interface KnowledgeExtractionParams {
  model_name: string;
  questions: KnowledgeExtractionQuestion[];
  llm_provider: string;
  advanced_indexing?: boolean;
  rerank?: boolean;
  generate_answers?: boolean;
}

export function create_knowledge_extraction(params: KnowledgeExtractionParams): Promise<any> {
  const accessToken = getAccessToken();
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  return new Promise((resolve, reject) => {
    axios
      .post(`${thirdaiPlatformBaseUrl}/api/v2/workflow/knowledge-extraction`, params)
      .then((res) => resolve(res.data))
      .catch((err) => {
        reject(
          new Error(err.response?.data?.detail || 'Failed to create knowledge extraction model')
        );
      });
  });
}

export interface Question {
  question_id: string;
  question_text: string;
  keywords: string[];
}
interface Reference {
  text: string;
  source: string;
}

export interface QuestionResult {
  question_id: string;
  question: string;
  answer: string;
  references: Reference[];
}

interface ReportContent {
  report_id: string;
  results: QuestionResult[];
}

interface Document {
  name: string;
  path: string;
  location: string;
}

export interface Report {
  report_id: string;
  status: 'complete' | 'pending' | 'failed' | 'queued' | 'in_progress';
  submitted_at: string;
  updated_at: string;
  content?: ReportContent;
  documents: Document[];
}

interface ReportStatus {
  data: {
    report_id: string;
    status: string;
    submitted_at: string;
    updated_at: string;
  }[];
  status: string;
  message: string;
}

export function useKnowledgeExtractionEndpoints(workflowId: string | null) {
  const accessToken = getAccessToken();
  const deploymentUrl = workflowId ? `${deploymentBaseUrl}/${workflowId}` : undefined;

  const listReports = async (): Promise<ReportStatus['data']> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    const response = await axios.get<ReportStatus>(`${deploymentUrl}/reports`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
    return response.data.data;
  };

  const createReport = async (files: File[]): Promise<string> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');

    const formData = new FormData();
    const documents = files.map((file) => ({
      name: file.name,
      path: file.name,
      location: 'local',
    }));
    formData.append('documents', JSON.stringify({ documents }));
    files.forEach((file) => formData.append('files', file));

    try {
      const accessToken = getAccessToken();
      const response = await axios.post(`${deploymentUrl}/report/create`, formData, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data.data.report_id;
    } catch (error) {
      console.error('Error creating report:', error);
      throw new Error('Failed to create report');
    }
  };

  const getReport = async (reportId: string): Promise<Report> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    try {
      const accessToken = getAccessToken();
      const response = await axios.get<{ data: Report }>(`${deploymentUrl}/report/${reportId}`, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error getting report:', error);
      throw new Error('Failed to get report');
    }
  };

  const deleteReport = async (reportId: string): Promise<void> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    try {
      const accessToken = getAccessToken();
      await axios.delete(`${deploymentUrl}/report/${reportId}`, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
    } catch (error) {
      console.error('Error deleting report:', error);
      throw new Error('Failed to delete report');
    }
  };

  const addQuestion = async (question: string, keywords?: string[]): Promise<string> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    try {
      const accessToken = getAccessToken();
      // Change to send as application/json with a request body instead of query params
      const response = await axios.post(
        `${deploymentUrl}/questions`, 
        { question, keywords }, // Send as request body
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'application/json'
          },
        }
      );
      return response.data.data.question_id;
    } catch (error) {
      console.error('Error adding question:', error);
      throw new Error('Failed to add question');
    }
  };

  const getQuestions = async (): Promise<Question[]> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    try {
      const accessToken = getAccessToken();
      const response = await axios.get<{
        status: string;
        message: string;
        data: Question[];
      }>(`${deploymentUrl}/questions`, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      return response.data.data;
    } catch (error) {
      console.error('Error getting questions:', error);
      throw new Error('Failed to get questions');
    }
  };

  const deleteQuestion = async (questionId: string): Promise<void> => {
    if (!deploymentUrl) throw new Error('Knowledge extraction deployment URL not set');
    try {
      const accessToken = getAccessToken();
      await axios.delete(`${deploymentUrl}/questions/${questionId}`, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
    } catch (error) {
      console.error('Error deleting question:', error);
      throw new Error('Failed to delete question');
    }
  };

  const editQuestion = async (
    questionId: string,
    newQuestion: string,
    keywords?: string[]
  ): Promise<void> => {
    await deleteQuestion(questionId);
    await addQuestion(newQuestion, keywords);
  };

  return {
    listReports,
    createReport,
    getReport,
    deleteReport,
    addQuestion,
    getQuestions,
    deleteQuestion,
    editQuestion,
  };
}