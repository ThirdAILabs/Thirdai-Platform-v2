import { ObjectDatabaseRecord, ClassifiedTokenDatabaseRecord } from "./(database-table)/types";

export const mockGroups = ["Reject", "Sensitive", "Safe"];

export const mockTags = ["vin", "vun", "name", "ssn", "dob", "email", "phone", "address", "creditCard", "passport", "license"];

export const mockObjectRecords: ObjectDatabaseRecord[] = [
  {
    taggedTokens: [['1234567890', 'vin'], ['1234567890', 'vun'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin']],
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    taggedTokens: [['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin']],
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    taggedTokens: [['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin']],
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    taggedTokens: [['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin'], ['1234567890', 'vin']],
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
];

export const mockClassifiedTokenRecords: ClassifiedTokenDatabaseRecord[] = [
  {
    token: '1234567890',
    tag: 'vin',
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    token: '1234567890',
    tag: 'vin',
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    token: '1234567890',
    tag: 'vin',
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
  },
  {
    token: '1234567890',
    tag: 'vin',
    sourceObject: '/path/to/s3/bucket/file.txt',
    groups: ['Reject', 'Sensitive'],
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