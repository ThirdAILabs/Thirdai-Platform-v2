export interface ObjectDatabaseRecord {
  taggedTokens: [string, string][];
  sourceObject: string;
  groups: string[];
}

export interface ClassifiedTokenDatabaseRecord {
  token: string;
  tag: string;
  sourceObject: string;
  groups: string[];
}

export interface DatabaseTableProps {
  loadMoreObjectRecords: () => Promise<ObjectDatabaseRecord[]>;
  loadMoreClassifiedTokenRecords: () => Promise<ClassifiedTokenDatabaseRecord[]>;
  groups: string[];
  tags: string[];
}

export type ViewMode = 'object' | 'classified-token'; 