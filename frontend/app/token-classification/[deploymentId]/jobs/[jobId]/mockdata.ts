import { ObjectDatabaseRecord, ClassifiedTokenDatabaseRecord } from "./(database-table)/types";

export const mockGroups = ["Reject", "Sensitive", "Safe"];

export const mockTags = ["VIN", "NAME", "SSN", "DOB", "EMAIL", "PHONE", "ADDRESS", "CREDITCARD", "PASSPORT", "LICENSE"];

export const mockObjectRecords: ObjectDatabaseRecord[] = [
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
    groups: ['Reject', 'Sensitive'],
  },
  {
    taggedTokens: [
      ['I', 'O'], ['need', 'O'], ['help', 'O'], ['with', 'O'], ['my', 'O'], ['credit', 'O'], ['card', 'O'],
      ['4111-1111-1111-1111', 'CREDITCARD']
    ],
    sourceObject: 'customer_chat_20231108.txt',
    groups: ['Sensitive'],
  },
  {
    taggedTokens: [
      ['You', 'O'], ['can', 'O'], ['email', 'O'], ['me', 'O'], ['at', 'O'],
      ['john.smith', 'EMAIL'], ['@', 'EMAIL'], ['email.com', 'EMAIL']
    ],
    sourceObject: 'customer_chat_20231109.txt',
    groups: ['Safe'],
  },
  {
    taggedTokens: [
      ['My', 'O'], ['driver', 'O'], ['license', 'O'], ['number', 'O'], ['is', 'O'],
      ['D1234567', 'LICENSE'], ['issued', 'O'], ['in', 'O'], ['California', 'O']
    ],
    sourceObject: 'customer_chat_20231110.txt',
    groups: ['Sensitive'],
  },
];

export const mockClassifiedTokenRecords: ClassifiedTokenDatabaseRecord[] = [
  {
    token: 'John',
    tag: 'NAME',
    sourceObject: 'customer_chat_20231105.txt',
    groups: ['Sensitive'],
  },
  {
    token: '555-123-4567',
    tag: 'PHONE',
    sourceObject: 'customer_chat_20231105.txt', 
    groups: ['Sensitive'],
  },
  {
    token: '123-45-6789',
    tag: 'SSN',
    sourceObject: 'customer_chat_20231107.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    token: 'john.smith@email.com',
    tag: 'EMAIL',
    sourceObject: 'customer_chat_20231109.txt',
    groups: ['Safe'],
  },
];

const makeLoadMoreMockData = <T>(records: T[]): () => Promise<T[]> => {
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