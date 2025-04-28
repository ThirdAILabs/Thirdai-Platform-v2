import type { TrainReportData, LabelMetrics, ExampleCategories, TrainingExample } from '@/lib/backend';
import { ObjectDatabaseRecord, ClassifiedTokenDatabaseRecord } from '@/app/token-classification/[deploymentId]/jobs/[jobId]/(database-table)/types';

// ===== WORKFLOWS =====
export const mockWorkflows = [
  {
    model_id: '550e8400-e29b-41d4-a716-446655440000',
    model_name: 'PII',
    type: 'nlp-token',
    access: 'private',
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
        model_id: '7c9e6679-7425-40de-944b-e07fc1f90ae7',
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
    model_id: '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
    model_name: 'HIPAA15',
    type: 'nlp-token',
    access: 'private',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-04-10T09:15:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: '550e8400-e29b-41d4-a716-446655440001',
    attributes: {
      llm_provider: 'openai',
      default_mode: 'medical'
    },
    dependencies: [
      {
        model_id: '7c9e6679-7425-40de-944b-e07fc1f90ae8',
        model_name: 'Medical Base Model',
        type: 'nlp-token',
        sub_type: 'medical',
        username: 'demo_user'
      }
    ],
    size: '1.5GB',
    size_in_memory: '600MB'
  },
  {
    model_id: '6ba7b811-9dad-11d1-80b4-00c04fd430c8',
    model_name: 'HIPAA30',
    type: 'nlp-token',
    access: 'protected',
    train_status: 'complete',
    deploy_status: 'complete',
    publish_date: '2023-03-22T16:45:00Z',
    username: 'demo_user',
    user_email: 'demo@example.com',
    team_id: '550e8400-e29b-41d4-a716-446655440002',
    attributes: {
      llm_provider: 'openai',
      default_mode: 'finance'
    },
    dependencies: [
      {
        model_id: '7c9e6679-7425-40de-944b-e07fc1f90ae9',
        model_name: 'Finance Base Model',
        type: 'nlp-token',
        sub_type: 'finance',
        username: 'demo_user'
      }
    ],
    size: '1.8GB',
    size_in_memory: '700MB'
  },
  {
    model_id: '6ba7b812-9dad-11d1-80b4-00c04fd430c8',
    model_name: 'HIPAA20',
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
      default_mode: 'standard'
    },
    dependencies: [
      {
        model_id: '7c9e6679-7425-40de-944b-e07fc1f90aea',
        model_name: 'Base Token Classification Model',
        type: 'nlp-token',
        sub_type: 'base',
        username: 'demo_user'
      }
    ],
    size: '2.0GB',
    size_in_memory: '800MB'
  },
  {
    model_id: '6ba7b813-9dad-11d1-80b4-00c04fd430c8',
    model_name: 'HIPAA35',
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
      default_mode: 'medical'
    },
    dependencies: [
      {
        model_id: '7c9e6679-7425-40de-944b-e07fc1f90aeb',
        model_name: 'Medical Base Model',
        type: 'nlp-token',
        sub_type: 'medical',
        username: 'demo_user'
      }
    ],
    size: '1.6GB',
    size_in_memory: '650MB'
  }
];

// ===== PREDICTIONS =====
export const mockPredictionResponses: Record<string, any> = {
  'default': {
    prediction_results: {
      tokens: ['The', 'patient', 'John', 'Smith', 'called', 'to', 'schedule', 'an', 'appointment', 'for', 'next', 'Monday', '.'],
      predicted_tags: [['O'], ['O'], ['NAME'], ['NAME'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['DATE'], ['O']],
      source_object: 'Sample text 1'
    },
    time_taken: 0.0032
  },
  'The patient John Smith called to schedule an appointment for next Monday.': {
    prediction_results: {
      tokens: ['The', 'patient', 'John', 'Smith', 'called', 'to', 'schedule', 'an', 'appointment', 'for', 'next', 'Monday', '.'],
      predicted_tags: [['O'], ['O'], ['NAME'], ['NAME'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['DATE'], ['O']],
      source_object: 'Sample text 1'
    },
    time_taken: 0.0032
  },
  'Patient ID: 123-45-6789, DOB: 01/15/1980, Phone: (555) 123-4567': {
    prediction_results: {
      tokens: ['Patient', 'ID', ':', '123-45-6789', ',', 'DOB', ':', '01/15/1980', ',', 'Phone', ':', '(', '555', ')', '123-4567'],
      predicted_tags: [['O'], ['O'], ['O'], ['SSN'], ['O'], ['O'], ['O'], ['DOB'], ['O'], ['O'], ['O'], ['O'], ['PHONE'], ['O'], ['PHONE']],
      source_object: 'Medical record'
    },
    time_taken: 0.0032
  },
  'Insurance claim #98765 for patient Jane Doe, policy number 987654321': {
    prediction_results: {
      tokens: ['Insurance', 'claim', '#', '98765', 'for', 'patient', 'Jane', 'Doe', ',', 'policy', 'number', '987654321'],
      predicted_tags: [['O'], ['O'], ['O'], ['CLAIM_ID'], ['O'], ['O'], ['NAME'], ['NAME'], ['O'], ['O'], ['O'], ['POLICY_NUMBER']],
      source_object: 'Insurance claim'
    },
    time_taken: 0.0032
  },
  // ALONE: CONTEXT_SENSITIVE_DATA
  // VACANT: CONTEXT_SENSITIVE_DATA
  'Customer : Also , I will be staying alone and my primary residence at 1742 Oakdale Avenue in Pasadena will be vacant for those three months . Is there anything else I need to do regarding my prescription ?': {
    prediction_results: {
      tokens: ['Customer', ':', 'Also', ',', 'I', 'will', 'be', 'staying', 'alone', 'and', 'my', 'primary', 'residence', 'at', '1742', 'Oakdale', 'Avenue', 'in', 'Pasadena', 'will', 'be', 'vacant', 'for', 'those', 'three', 'months', '.', 'Is', 'there', 'anything', 'else', 'I', 'need', 'to', 'do', 'regarding', 'my', 'prescription', '?'],
      predicted_tags: [['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['ADDRESS'], ['ADDRESS'], ['ADDRESS'], ['O'], ['LOCATION'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O'], ['O']],
      source_object: 'customer_chat_20231110.txt'
    },
    time_taken: 0.0032
  }
};

// ===== DEPLOYMENT STATS =====
export const mockDeploymentStats = {
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

// ===== LABELS =====
export const mockLabels = [
  'O',
  'NAME',
  'SSN',
  'DOB',
  'PHONE',
  'EMAIL',
  'ADDRESS',
  'POLICY_NUMBER',
  'CLAIM_ID',
  'DATE',
  'TIME',
  'MEDICAL_CONDITION',
  'MEDICATION',
  'PROVIDER_NAME',
  'FACILITY_NAME',
  'INSURANCE_ID'
];

// ===== TRAINING REPORTS =====
export const mockMetrics: LabelMetrics = {
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
  'SSN': {
    precision: 0.91,
    recall: 0.89,
    fmeasure: 0.90
  },
  'DOB': {
    precision: 0.94,
    recall: 0.93,
    fmeasure: 0.94
  },
  'PHONE': {
    precision: 0.87,
    recall: 0.86,
    fmeasure: 0.87
  },
  'EMAIL': {
    precision: 0.92,
    recall: 0.91,
    fmeasure: 0.92
  },
  'ADDRESS': {
    precision: 0.89,
    recall: 0.88,
    fmeasure: 0.89
  },
  'POLICY_NUMBER': {
    precision: 0.93,
    recall: 0.92,
    fmeasure: 0.93
  },
  'CLAIM_ID': {
    precision: 0.90,
    recall: 0.89,
    fmeasure: 0.90
  }
};

export const mockExamples: ExampleCategories = {
  true_positives: {
    'NAME': [
      { source: 'John Smith', target: 'NAME', predictions: 'NAME', index: 0 },
      { source: 'Jane Doe', target: 'NAME', predictions: 'NAME', index: 1 },
      { source: 'Dr. Robert Johnson', target: 'NAME', predictions: 'NAME', index: 2 }
    ],
    'SSN': [
      { source: '123-45-6789', target: 'SSN', predictions: 'SSN', index: 0 },
      { source: '(123) 45-6789', target: 'SSN', predictions: 'SSN', index: 1 },
      { source: '123.45.6789', target: 'SSN', predictions: 'SSN', index: 2 }
    ],
    'DOB': [
      { source: '01/15/1980', target: 'DOB', predictions: 'DOB', index: 0 },
      { source: '1980-01-15', target: 'DOB', predictions: 'DOB', index: 1 },
      { source: 'January 15, 1980', target: 'DOB', predictions: 'DOB', index: 2 }
    ],
    'PHONE': [
      { source: '(555) 123-4567', target: 'PHONE', predictions: 'PHONE', index: 0 },
      { source: '555-123-4567', target: 'PHONE', predictions: 'PHONE', index: 1 },
      { source: '555.123.4567', target: 'PHONE', predictions: 'PHONE', index: 2 }
    ]
  },
  false_positives: {
    'NAME': [
      { source: 'John', target: 'O', predictions: 'NAME', index: 0 },
      { source: 'Smith', target: 'O', predictions: 'NAME', index: 1 }
    ],
    'SSN': [
      { source: '123-4567', target: 'O', predictions: 'SSN', index: 0 },
      { source: '555-123', target: 'O', predictions: 'SSN', index: 1 }
    ],
    'DOB': [
      { source: '01/15', target: 'O', predictions: 'DOB', index: 0 },
      { source: '1980', target: 'O', predictions: 'DOB', index: 1 }
    ],
    'PHONE': [
      { source: '123-4567', target: 'O', predictions: 'PHONE', index: 0 },
      { source: '555-123', target: 'O', predictions: 'PHONE', index: 1 }
    ]
  },
  false_negatives: {
    'NAME': [
      { source: 'Michael Brown', target: 'NAME', predictions: 'O', index: 0 },
      { source: 'Sarah Wilson', target: 'NAME', predictions: 'O', index: 1 }
    ],
    'SSN': [
      { source: '987-65-4321', target: 'SSN', predictions: 'O', index: 0 },
      { source: '(987) 65-4321', target: 'SSN', predictions: 'O', index: 1 }
    ],
    'DOB': [
      { source: '03/22/1975', target: 'DOB', predictions: 'O', index: 0 },
      { source: '1975-03-22', target: 'DOB', predictions: 'O', index: 1 }
    ],
    'PHONE': [
      { source: '(987) 654-3210', target: 'PHONE', predictions: 'O', index: 0 },
      { source: '987-654-3210', target: 'PHONE', predictions: 'O', index: 1 }
    ]
  }
};

export const mockTrainReport: TrainReportData = {
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
    'SSN': {
      precision: 0.85,
      recall: 0.83,
      fmeasure: 0.84
    },
    'DOB': {
      precision: 0.88,
      recall: 0.87,
      fmeasure: 0.88
    },
    'PHONE': {
      precision: 0.80,
      recall: 0.78,
      fmeasure: 0.79
    }
  },
  after_train_metrics: mockMetrics,
  after_train_examples: mockExamples
};

// ===== TOKEN CLASSIFICATION DATA =====
export const mockGroups = ["Sensitive", "Review", "Safe"];

export const mockTags = ["NAME", "SSN", "DOB", "EMAIL", "PHONE", "ADDRESS", "POLICY_NUMBER", "CLAIM_ID", "DATE", "TIME", "MEDICAL_CONDITION", "MEDICATION", "PROVIDER_NAME", "FACILITY_NAME", "INSURANCE_ID"];

export const mockObjectRecords: ObjectDatabaseRecord[] = [
  {
    taggedTokens: [
      ['Agent', 'O'], [':', 'O'], ['Thank', 'O'], ['you', 'O'], ['for', 'O'], ['calling', 'O'], ['TechGuard', 'O'],
      ['Support', 'O'], ['.', 'O'], ['My', 'O'], ['name', 'O'], ['is', 'O'], ['Alex', 'O'], [',', 'O'], ['how', 'O'],
      ['can', 'O'], ['I', 'O'], ['help', 'O'], ['you', 'O'], ['today', 'O'], ['?', 'O']
    ],
    sourceObject: 'customer_chat_20231105_0.txt',
    groups: ['Safe']
  },
  {
    taggedTokens: [
      ['Customer', 'O'], [':', 'O'], ['Hi', 'O'], ['there', 'O'], ['.', 'O'], ['My', 'O'], ['name', 'O'], ['is', 'O'],
      ['Robert', 'NAME'], ['Chen', 'NAME'], ['and', 'O'], ["I'm", 'O'], ['having', 'O'], ['trouble', 'O'],
      ['accessing', 'O'], ['my', 'O'], ['account', 'O'], ['.', 'O'], ["I've", 'O'], ['been', 'O'], ['trying', 'O'],
      ['since', 'O'], ['yesterday', 'O'], ['.', 'O']
    ],
    sourceObject: 'customer_chat_20231105_1.txt',
    groups: ['Review']
  },
  {
    taggedTokens: [
      ['Agent', 'O'], [':', 'O'], ["I'm", 'O'], ['sorry', 'O'], ['to', 'O'], ['hear', 'O'], ['that', 'O'], [',', 'O'],
      ['Mr.', 'O'], ['Chen', 'NAME'], ['.', 'O'], ["I'd", 'O'], ['be', 'O'], ['happy', 'O'], ['to', 'O'], ['help', 'O'],
      ['you', 'O'], ['regain', 'O'], ['access', 'O'], ['.', 'O'], ['Could', 'O'], ['you', 'O'], ['please', 'O'],
      ['verify', 'O'], ['your', 'O'], ['account', 'O'], ['with', 'O'], ['your', 'O'], ['email', 'O'], ['address', 'O'],
      ['?', 'O']
    ],
    sourceObject: 'customer_chat_20231105_2.txt',
    groups: ['Review']
  }, 
  {
    taggedTokens: [
      ['Customer', 'O'], [':', 'O'], ['Sure', 'O'], [',', 'O'], ["it's", 'O'], ['robert.chen1982', 'EMAIL'], ['@', 'EMAIL'], ['gmail.com', 'EMAIL'], ['.', 'O']
    ], sourceObject: 'customer_chat_20231105_3.txt', groups: ['Review']
  }, { taggedTokens: [['Agent', 'O'], [':', 'O'], ['Thank', 'O'], ['you', 'O'], ['.', 'O'], ['And', 'O'], ['for', 'O'], ['additional', 'O'], ['verification', 'O'], [',', 'O'], ['could', 'O'], ['I', 'O'], ['have', 'O'], ['the', 'O'], ['last', 'O'], ['four', 'O'], ['digits', 'O'], ['of', 'O'], ['the', 'O'], ['phone', 'O'], ['number', 'O'], ['associated', 'O'], ['with', 'O'], ['the', 'O'], ['account', 'O'], ['?', 'O']], sourceObject: 'customer_chat_20231105_4.txt', groups: ['Safe'] }, { taggedTokens: [['Customer', 'O'], [':', 'O'], ['Yes', 'O'], [',', 'O'], ["it's", 'O'], ['5784', 'PHONE'], ['.', 'O']], sourceObject: 'customer_chat_20231105_5.txt', groups: ['Review'] }, { taggedTokens: [['Agent', 'O'], [':', 'O'], ['Perfect', 'O'], ['.', 'O'], ['I', 'O'], ['can', 'O'], ['see', 'O'], ['your', 'O'], ['account', 'O'], ['here', 'O'], ['.', 'O'], ['It', 'O'], ['looks', 'O'], ['like', 'O'], ['there', 'O'], ['were', 'O'], ['multiple', 'O'], ['failed', 'O'], ['login', 'O'], ['attempts', 'O'], ['from', 'O'], ['an', 'O'], ['unfamiliar', 'O'], ['IP', 'O'], ['address', 'O'], [',', 'O'], ['so', 'O'], ['our', 'O'], ['security', 'O'], ['system', 'O'], ['temporarily', 'O'], ['locked', 'O'], ['your', 'O'], ['account', 'O'], ['.', 'O'], ['Can', 'O'], ['you', 'O'], ['confirm', 'O'], ['your', 'O'], ['current', 'O'], ['address', 'O'], ['is', 'O'], ['still', 'O'], ['728', 'ADDRESS'], ['Maple', 'ADDRESS'], ['Street', 'ADDRESS'], [',', 'ADDRESS'], ['Apartment', 'ADDRESS'], ['4B', 'ADDRESS'], [',', 'ADDRESS'], ['San', 'ADDRESS'], ['Francisco', 'ADDRESS'], [',', 'ADDRESS'], ['CA', 'ADDRESS'], ['94107', 'ADDRESS'], ['?', 'O']], sourceObject: 'customer_chat_20231105_6.txt', groups: ['Sensitive'] }, { taggedTokens: [['Customer', 'O'], [':', 'O'], ['Yes', 'O'], [',', 'O'], ["that's", 'O'], ['correct', 'O'], ['.', 'O']], sourceObject: 'customer_chat_20231105_7.txt', groups: ['Safe'] }, { taggedTokens: [['Agent', 'O'], [':', 'O'], ['Great', 'O'], ['.', 'O'], ["I've", 'O'], ['reset', 'O'], ['your', 'O'], ['account', 'O'], ['access', 'O'], ['.', 'O'], ['You', 'O'], ['should', 'O'], ['receive', 'O'], ['a', 'O'], ['verification', 'O'], ['code', 'O'], ['at', 'O'], ['your', 'O'], ['email', 'O'], ['address', 'O'], ['shortly', 'O'], ['.', 'O'], ['Once', 'O'], ['you', 'O'], ['enter', 'O'], ['that', 'O'], ['code', 'O'], [',', 'O'], ["you'll", 'O'], ['be', 'O'], ['prompted', 'O'], ['to', 'O'], ['create', 'O'], ['a', 'O'], ['new', 'O'], ['password', 'O'], ['.', 'O'], ['Is', 'O'], ['there', 'O'], ['anything', 'O'], ['else', 'O'], ['I', 'O'], ['can', 'O'], ['help', 'O'], ['with', 'O'], ['today', 'O'], ['?', 'O']], sourceObject: 'customer_chat_20231105_8.txt', groups: ['Safe'] }, { taggedTokens: [['Customer', 'O'], [':', 'O'], ['Actually', 'O'], [',', 'O'], ['yes', 'O'], ['.', 'O'], ['I', 'O'], ['recently', 'O'], ['got', 'O'], ['a', 'O'], ['new', 'O'], ['credit', 'O'], ['card', 'O'], ['and', 'O'], ['need', 'O'], ['to', 'O'], ['update', 'O'], ['my', 'O'], ['billing', 'O'], ['information', 'O'], ['.', 'O'], ['The', 'O'], ['new', 'O'], ['card', 'O'], ['number', 'O'], ['is', 'O'], ['4832', 'CREDIT_CARD'], ['5691', 'CREDIT_CARD'], ['2748', 'CREDIT_CARD'], ['1035', 'CREDIT_CARD'], ['with', 'O'], ['expiration', 'O'], ['date', 'O'], ['09', 'EXPIRATION_DATE'], ['/', 'EXPIRATION_DATE'], ['27', 'EXPIRATION_DATE'], ['and', 'O'], ['security', 'O'], ['code', 'O'], ['382', 'CVV'], ['.', 'O']], sourceObject: 'customer_chat_20231105_9.txt', groups: ['Sensitive'] }, { taggedTokens: [['Customer', 'O'], [':', 'O'], ['That', 'O'], ['makes', 'O'], ['sense', 'O'], ['.', 'O'], ["I'll", 'O'], ['do', 'O'], ['that', 'O'], ['instead', 'O'], ['.', 'O'], ['My', 'O'], ['social', 'O'], ['security', 'O'], ['number', 'O'], ['is', 'O'], ['532', 'SSN'], ['-', 'SSN'], ['48', 'SSN'], ['-', 'SSN'], ['1095', 'SSN'], ['if', 'O'], ['you', 'O'], ['need', 'O'], ['that', 'O'], ['for', 'O'], ['verification', 'O'], ['.', 'O']], sourceObject: 'customer_chat_20231105_10.txt', groups: ['Sensitive'] },
  {
    taggedTokens: [
      ['Hi', 'O'], ['my', 'O'], ['name', 'O'], ['is', 'O'], ['John', 'NAME'], ['Smith', 'NAME'], ['and', 'O'],
      ['my', 'O'], ['phone', 'O'], ['number', 'O'], ['is', 'O'], ['555-123-4567', 'PHONE']
    ],
    sourceObject: 'customer_chat_20231105.txt',
    groups: ['Sensitive'],
  },
  {
    taggedTokens: [
      ['Please', 'O'], ['update', 'O'], ['my', 'O'], ['address', 'O'], ['to', 'O'],
      ['123', 'ADDRESS'], ['Main', 'ADDRESS'], ['Street', 'ADDRESS'], ['Apt', 'ADDRESS'], ['4B', 'ADDRESS']
    ],
    sourceObject: 'customer_chat_20231106.txt',
    groups: ['Safe'],
  },
  {
    taggedTokens: [
      ['My', 'O'], ['SSN', 'O'], ['is', 'O'], ['123-45-6789', 'SSN'], ['.', 'O'],
      ['Date', 'O'], ['of', 'O'], ['birth', 'O'], ['is', 'O'], ['01/15/1980', 'DOB']
    ],
    sourceObject: 'customer_chat_20231107.txt',
    groups: ['Sensitive', 'Sensitive'],
  },
  {
    taggedTokens: [
      ['I', 'O'], ['need', 'O'], ['help', 'O'], ['with', 'O'], ['my', 'O'], ['insurance', 'O'], ['claim', 'O'],
      ['987654321', 'CLAIM_ID']
    ],
    sourceObject: 'customer_chat_20231108.txt',
    groups: ['Sensitive'],
  },
  {
    taggedTokens: [
      ['You', 'O'], ['can', 'O'], ['email', 'O'], ['me', 'O'], ['at', 'O'],
      ['john.smith', 'EMAIL'], ['@', 'EMAIL'], ['healthcare.com', 'EMAIL']
    ],
    sourceObject: 'customer_chat_20231109.txt',
    groups: ['Safe'],
  },
  {
    taggedTokens: [
      ['My', 'O'], ['policy', 'O'], ['number', 'O'], ['is', 'O'],
      ['POL987654321', 'POLICY_NUMBER'], ['issued', 'O'], ['by', 'O'], ['BlueCross', 'O']
    ],
    sourceObject: 'customer_chat_20231110.txt',
    groups: ['Sensitive'],
  },
];

export const mockClassifiedTokenRecords: ClassifiedTokenDatabaseRecord[] = [{'token': 'Robert', 'tag': 'NAME', 'sourceObject': 'customer_chat_20231105_1.txt', 'groups': ['Review']}, {'token': 'Chen', 'tag': 'NAME', 'sourceObject': 'customer_chat_20231105_1.txt', 'groups': ['Review']}, {'token': 'Chen', 'tag': 'NAME', 'sourceObject': 'customer_chat_20231105_2.txt', 'groups': ['Review']}, {'token': 'robert.chen1982', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231105_3.txt', 'groups': ['Review']}, {'token': '@', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231105_3.txt', 'groups': ['Review']}, {'token': 'gmail.com', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231105_3.txt', 'groups': ['Review']}, {'token': '5784', 'tag': 'PHONE', 'sourceObject': 'customer_chat_20231105_5.txt', 'groups': ['Review']}, {'token': '728', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'Maple', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'Street', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': ',', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'Apartment', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': '4B', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': ',', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'San', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'Francisco', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': ',', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': 'CA', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': '94107', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231105_6.txt', 'groups': ['Sensitive']}, {'token': '4832', 'tag': 'CREDIT_CARD', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '5691', 'tag': 'CREDIT_CARD', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '2748', 'tag': 'CREDIT_CARD', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '1035', 'tag': 'CREDIT_CARD', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '09', 'tag': 'EXPIRATION_DATE', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '/', 'tag': 'EXPIRATION_DATE', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '27', 'tag': 'EXPIRATION_DATE', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '382', 'tag': 'CVV', 'sourceObject': 'customer_chat_20231105_9.txt', 'groups': ['Sensitive']}, {'token': '532', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive']}, {'token': '-', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive']}, {'token': '48', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive']}, {'token': '-', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive']}, {'token': '1095', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231105_10.txt', 'groups': ['Sensitive']}, {'token': 'John', 'tag': 'NAME', 'sourceObject': 'customer_chat_20231105.txt', 'groups': ['Sensitive']}, {'token': 'Smith', 'tag': 'NAME', 'sourceObject': 'customer_chat_20231105.txt', 'groups': ['Sensitive']}, {'token': '555-123-4567', 'tag': 'PHONE', 'sourceObject': 'customer_chat_20231105.txt', 'groups': ['Sensitive']}, {'token': '123', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231106.txt', 'groups': ['Safe']}, {'token': 'Main', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231106.txt', 'groups': ['Safe']}, {'token': 'Street', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231106.txt', 'groups': ['Safe']}, {'token': 'Apt', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231106.txt', 'groups': ['Safe']}, {'token': '4B', 'tag': 'ADDRESS', 'sourceObject': 'customer_chat_20231106.txt', 'groups': ['Safe']}, {'token': '123-45-6789', 'tag': 'SSN', 'sourceObject': 'customer_chat_20231107.txt', 'groups': ['Sensitive', 'Sensitive']}, {'token': '01/15/1980', 'tag': 'DOB', 'sourceObject': 'customer_chat_20231107.txt', 'groups': ['Sensitive', 'Sensitive']}, {'token': '987654321', 'tag': 'CLAIM_ID', 'sourceObject': 'customer_chat_20231108.txt', 'groups': ['Sensitive']}, {'token': 'john.smith', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231109.txt', 'groups': ['Safe']}, {'token': '@', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231109.txt', 'groups': ['Safe']}, {'token': 'healthcare.com', 'tag': 'EMAIL', 'sourceObject': 'customer_chat_20231109.txt', 'groups': ['Safe']}, {'token': 'POL987654321', 'tag': 'POLICY_NUMBER', 'sourceObject': 'customer_chat_20231110.txt', 'groups': ['Sensitive']}];

// Helper function for loading more mock data
export const makeLoadMoreMockData = <T>(records: T[]): () => Promise<T[]> => {
  return () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(records);
      }, 1000);
    });
  };
};

export const loadMoreMockObjectRecords = makeLoadMoreMockData(mockObjectRecords);
export const loadMoreMockClassifiedTokenRecords = makeLoadMoreMockData(mockClassifiedTokenRecords);

// ===== ANALYTICS DATA =====
export const upvotes = [
  {
    query: 'What is the purpose of renovating the historic building?',
    upvote: 'Historic building undergoes renovation',
  },
  {
    query: 'What type of cuisine will the famous chef\u2019s new restaurant offer?',
    upvote: 'Famous chef opens new restaurant',
  },
  {
    query: 'What milestone has the local business recently achieved in sales?',
    upvote: 'Local business achieves milestone in sales',
  },
  {
    query: 'What are the objectives of the new Mars mission announced by the space agency?',
    upvote: 'International space agency announces mission to Mars',
  },
  {
    query: 'What new project is the tech company partnering on with the startup?',
    upvote: 'Tech company partners with startup for new project',
  },
  {
    query: 'What does the new healthcare legislation entail for the nation?',
    upvote: 'Nation passes new healthcare legislation',
  },
  {
    query: 'How will the new technology developed by the university impact its industry?',
    upvote: 'University develops cutting-edge technology',
  },
  {
    query: 'What are the goals of the government\u2019s educational reform?',
    upvote: 'Government implements new educational reform',
  },
  {
    query: 'When does the popular author\u2019s book tour begin?',
    upvote: 'Popular author announces book tour',
  },
  {
    query: 'What is the timeline for the new infrastructure project announced by the government?',
    upvote: 'Government announces infrastructure development project',
  },
  {
    query: 'How will the new healthcare legislation affect citizens?',
    upvote: 'Nation passes new healthcare legislation',
  },
  {
    query: 'What are some standout films at the international film festival?',
    upvote: 'International film festival showcases diverse films',
  },
  {
    query: "How can people participate in the environmental group's conservation efforts?",
    upvote: 'Environmental group launches conservation campaign',
  },
  {
    query: 'How does the film festival showcase diversity in its selections?',
    upvote: 'International film festival showcases diverse films',
  },
  {
    query: 'How is the author promoting their upcoming book tour?',
    upvote: 'Popular author announces book tour',
  },
  {
    query: 'What is the significance of the new public park for the city?',
    upvote: 'Major city opens new public park',
  },
  {
    query:
      'What are the highlights of the new technologies featured at the global tech conference?',
    upvote: 'Global tech conference features new innovations',
  },
  {
    query: 'What does the award win mean for the film director\u2019s career?',
    upvote: 'Film director wins prestigious award',
  },
  {
    query: 'How is the new educational reform being received by educators?',
    upvote: 'Government implements new educational reform',
  },
  {
    query: 'How are the latest innovations being presented at the tech conference?',
    upvote: 'Global tech conference features new innovations',
  },
  {
    query: "What is the significance of the university's record number of applicants?",
    upvote: 'University reports record number of applicants',
  },
  {
    query: "How does the new funding impact the tech startup's future plans?",
    upvote: 'Tech startup secures major funding',
  },
  {
    query: 'What changes are included in the new educational reform by the government?',
    upvote: 'Government implements new educational reform',
  },
  {
    query: 'What new innovations are being showcased at the global tech conference?',
    upvote: 'Global tech conference features new innovations',
  },
  {
    query: 'How does the university\u2019s new technology stand out in its field?',
    upvote: 'University develops cutting-edge technology',
  },
  {
    query: 'How can people participate in the charity event hosted by the local center?',
    upvote: 'Local community center hosts charity event',
  },
  {
    query: 'How is the new restaurant different from the chef\u2019s previous ventures?',
    upvote: 'Famous chef opens new restaurant',
  },
  {
    query: 'What types of films are featured at the international film festival?',
    upvote: 'International film festival showcases diverse films',
  },
  {
    query: 'How did the city celebrate the opening of the new public park?',
    upvote: 'Major city opens new public park',
  },
  {
    query: 'How was the award-winning film received by audiences and critics?',
    upvote: 'Film director wins prestigious award',
  },
  {
    query: 'Which new innovations are attracting attention at the tech conference?',
    upvote: 'Global tech conference features new innovations',
  },
  {
    query: 'How does the record number of applicants impact the university?',
    upvote: 'University reports record number of applicants',
  },
  {
    query: 'What goals does the new conservation campaign by the environmental group have?',
    upvote: 'Environmental group launches conservation campaign',
  },
  {
    query: 'Who are the investors behind the tech startup\u2019s major funding?',
    upvote: 'Tech startup secures major funding',
  },
  {
    query: 'What can fans expect from the author\u2019s book tour?',
    upvote: 'Popular author announces book tour',
  },
  {
    query: 'How is the new public park expected to benefit the major city?',
    upvote: 'Major city opens new public park',
  },
  {
    query: 'What will the tech startup use the new funding for?',
    upvote: 'Tech startup secures major funding',
  },
  {
    query: "What are the key provisions of the nation's new healthcare law?",
    upvote: 'Nation passes new healthcare legislation',
  },
  {
    query: 'What themes are explored in the famous artist\u2019s new collection?',
    upvote: 'Famous artist exhibits new collection',
  },
  {
    query: 'How will the infrastructure project impact the local community?',
    upvote: 'Government announces infrastructure development project',
  },
];

export const associations = [
  {
    source: 'What activities are planned for the charity event at the community center?',
    target: 'Local community center hosts charity event',
  },
  {
    source: 'How is the local business celebrating its sales achievement?',
    target: 'Local business achieves milestone in sales',
  },
  {
    source: 'How will the renovation affect the historic building\u2019s appearance?',
    target: 'Historic building undergoes renovation',
  },
  {
    source: 'How will the charity event at the community center benefit the cause?',
    target: 'Local community center hosts charity event',
  },
  {
    source: 'What technologies will be used in the Mars mission by the space agency?',
    target: 'International space agency announces mission to Mars',
  },
  {
    source: 'How will the space agency\u2019s mission to Mars be conducted?',
    target: 'International space agency announces mission to Mars',
  },
  {
    source: 'What does the sales milestone mean for the future of the local business?',
    target: 'Local business achieves milestone in sales',
  },
  {
    source: 'How will the partnership between the tech company and the startup work?',
    target: 'Tech company partners with startup for new project',
  },
  {
    source: 'Where can viewers see the new art collection by the famous artist?',
    target: 'Famous artist exhibits new collection',
  },
  {
    source: 'Which world record did the sports star recently break?',
    target: 'Sports star breaks world record',
  },
  {
    source: 'How is the environmental group planning to execute their conservation campaign?',
    target: 'Environmental group launches conservation campaign',
  },
  {
    source: 'What charity is being supported by the event at the local community center?',
    target: 'Local community center hosts charity event',
  },
  {
    source: 'How is the new collection by the artist being received by critics?',
    target: 'Famous artist exhibits new collection',
  },
  {
    source: 'How will the new educational reform impact students and teachers?',
    target: 'Government implements new educational reform',
  },
  {
    source: 'What are the key aspects of the environmental group\u2019s new campaign?',
    target: 'Environmental group launches conservation campaign',
  },
  {
    source: 'How did the local business reach this significant sales milestone?',
    target: 'Local business achieves milestone in sales',
  },
  {
    source: 'Where is the new restaurant opened by the famous chef located?',
    target: 'Famous chef opens new restaurant',
  },
  {
    source: 'What are the highlights of the famous artist\u2019s latest exhibition?',
    target: 'Famous artist exhibits new collection',
  },
  {
    source: 'How did the film director\u2019s work lead to the award?',
    target: 'Film director wins prestigious award',
  },
  {
    source: 'What is the timeline for the international space agency\u2019s Mars mission?',
    target: 'International space agency announces mission to Mars',
  },
  {
    source: 'How did the sports star achieve the world record?',
    target: 'Sports star breaks world record',
  },
  {
    source: 'What cities will the popular author visit on their book tour?',
    target: 'Popular author announces book tour',
  },
  {
    source: 'What was the reaction to the sports star breaking the world record?',
    target: 'Sports star breaks world record',
  },
  {
    source: 'What are the objectives of the tech company\u2019s collaboration with the startup?',
    target: 'Tech company partners with startup for new project',
  },
  {
    source: 'How does the film festival contribute to the global film industry?',
    target: 'International film festival showcases diverse films',
  },
  {
    source: 'What are the main features of the famous chef\u2019s new restaurant?',
    target: 'Famous chef opens new restaurant',
  },
  {
    source: 'Which prestigious award did the film director recently win?',
    target: 'Film director wins prestigious award',
  },
  {
    source: "What are the goals of the government's infrastructure development project?",
    target: 'Government announces infrastructure development project',
  },
  {
    source:
      'What are the potential applications of the cutting-edge technology from the university?',
    target: 'University develops cutting-edge technology',
  },
  {
    source: 'How will the startup benefit from its partnership with the tech company?',
    target: 'Tech company partners with startup for new project',
  },
  {
    source: 'How did the new healthcare legislation pass in the nation?',
    target: 'Nation passes new healthcare legislation',
  },
  {
    source: 'What impact does the world record have on the sports star\u2019s career?',
    target: 'Sports star breaks world record',
  },
  {
    source: 'How long will the renovation of the historic building take?',
    target: 'Historic building undergoes renovation',
  },
  {
    source: 'What changes are being made during the renovation of the historic building?',
    target: 'Historic building undergoes renovation',
  },
  {
    source: 'What factors contributed to the record number of applicants at the university?',
    target: 'University reports record number of applicants',
  },
  {
    source: 'What new technology is being developed by the university?',
    target: 'University develops cutting-edge technology',
  },
  {
    source: 'How is the university handling the influx of applications?',
    target: 'University reports record number of applicants',
  },
  {
    source: 'What details are available about the new infrastructure development project?',
    target: 'Government announces infrastructure development project',
  },
  {
    source: 'What features are included in the new public park opened by the city?',
    target: 'Major city opens new public park',
  },
  {
    source: 'How much funding has the tech startup recently secured?',
    target: 'Tech startup secures major funding',
  },
];

export const reformulations = [
  {
    original: 'How did scientists find the new planet?',
    reformulations: ['Which new planet did scientists recently discover?'],
  },
  {
    original: 'How did Apple describe their newest iPhone model?',
    reformulations: ['Apple announces new iPhone model'],
  },
  {
    original: 'What does the new international trade agreement involve?',
    reformulations: [
      'International trade agreement signed',
      'What are the key terms of the international trade agreement?',
    ],
  },
  {
    original: 'Stock market reaches all-time high',
    reformulations: [
      'How did the stock market achieve a record high?',
      "What factors contributed to the stock market's all-time high?",
      'What is the new all-time high for the stock market?',
      'What led to the stock market reaching an all-time high?',
    ],
  },
  {
    original: 'What severe weather conditions is the nation currently experiencing?',
    reformulations: [
      'Nation experiences severe weather conditions',
      'What caused the nation to face extreme weather conditions?',
    ],
  },
  {
    original: 'What are the details of the government\u2019s investment in renewable energy?',
    reformulations: [
      'How is the government planning to invest in renewable energy?',
      'How will the government\u2019s renewable energy investment impact the sector?',
      'What renewable energy projects is the government investing in?',
      'Government invests in renewable energy',
    ],
  },
  {
    original: 'Which world leaders are attending the climate summit?',
    reformulations: [
      'World leaders meet for climate summit',
      'How are world leaders addressing climate issues at the summit?',
      'What are the key topics at the climate summit involving global leaders?',
      'What is the agenda for the climate summit with world leaders?',
    ],
  },
  {
    original: "University's sports team wins championship",
    reformulations: [
      'What were the key moments in the university\u2019s sports team\u2019s championship victory?',
      "How did the university's sports team achieve the championship win?",
      "What does the championship win mean for the university's sports team?",
      "Which championship did the university's sports team win?",
    ],
  },
  {
    original: 'University announces new research center',
    reformulations: [
      'What are the goals of the university\u2019s newly announced research center?',
      'What will the new research center at the university specialize in?',
      'What is the focus of the new research center announced by the university?',
    ],
  },
  {
    original: 'Which celebrity recently married in a private ceremony?',
    reformulations: [
      'What details are known about the celebrity\u2019s private wedding?',
      'Celebrity marries in a private ceremony',
    ],
  },
  {
    original: 'How is the new electric vehicle being received by the public?',
    reformulations: [
      'Automaker unveils electric vehicle',
      'What new electric vehicle has the automaker unveiled?',
      'What are the features of the new electric vehicle from the automaker?',
      'How did the automaker describe their latest electric vehicle?',
    ],
  },
  {
    original: 'How did the major film perform at the festival?',
    reformulations: [
      'What award did the major film receive at the festival?',
      'Major film wins award at festival',
    ],
  },
  {
    original: 'Author releases bestselling novel',
    reformulations: [
      'How has the new novel from the author become a bestseller?',
      'What is the latest bestselling novel released by the author?',
    ],
  },
  {
    original: 'What innovative product did the tech company launch?',
    reformulations: [
      'How did the tech company unveil their new groundbreaking product?',
      'Tech company launches innovative product',
      'How is the new product from the tech company described?',
      'What are the features of the latest product introduced by the tech company?',
    ],
  },
  {
    original: 'Which team claimed the national championship title?',
    reformulations: [
      'What was the outcome for the local team in the national championship?',
      'Local team wins national championship',
    ],
  },
  {
    original: 'What is the new economic policy announced by the government?',
    reformulations: ['How did the government describe their latest economic policy?'],
  },
  {
    original: 'What are the economists predicting about GDP growth?',
    reformulations: ['What is the expected impact on the economy based on GDP growth predictions?'],
  },
  {
    original: 'What are the highlights of the international music festival in the city?',
    reformulations: [
      'Which artists are performing at the city\u2019s international music festival?',
    ],
  },
  {
    original: 'Popular TV show returns for new season',
    reformulations: [
      'When does the popular TV show return for its new season?',
      'How is the new season of the TV show being promoted?',
    ],
  },
  {
    original: 'How did the healthcare provider describe their latest vaccine?',
    reformulations: [
      'Healthcare provider announces new vaccine',
      'What are the details of the new vaccine introduced by the healthcare provider?',
      'What new vaccine has the healthcare provider announced?',
    ],
  },
];
