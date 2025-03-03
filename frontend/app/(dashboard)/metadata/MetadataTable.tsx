'use client';

import React, { useState, useEffect } from 'react';
import {
  TableContainer,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Paper,
  IconButton,
  Typography,
  CircularProgress,
} from '@mui/material';
import EditIcon from '@mui/icons-material/Edit';
import EditMetadataModal from './EditMetadataModal';
import {
  getDocumentMetadata,
  getWorkflowDetails,
  deploymentBaseUrl,
  getSources,
  Source,
} from '@/lib/backend';
import { useSearchParams } from 'next/navigation';

interface MetadataAttribute {
  attribute_name: string;
  value: string | number;
}

interface DocumentMetadata {
  document_id: string;
  document_name: string;
  metadata_attributes: MetadataAttribute[];
}

const MetadataTable: React.FC = () => {
  const searchParams = useSearchParams();
  const workflowId = searchParams.get('workflow_id');
  const username = searchParams.get('username');

  const [documents, setDocuments] = useState<DocumentMetadata[]>([]);
  const [selectedDocument, setSelectedDocument] = useState<DocumentMetadata | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deploymentUrl, setDeploymentUrl] = useState<string>('');
  const [rawSources, setRawSources] = useState<Source[]>([]);

  useEffect(() => {
    const fetchMetadata = async () => {
      if (!workflowId) return;

      try {
        // First get the workflow details to get the dependencies
        const workflowDetails = await getWorkflowDetails(workflowId);
        console.log('Workflow details:', workflowDetails);

        if (workflowDetails.type !== 'enterprise-search') {
          throw new Error('This workflow is not an enterprise-search type');
        }

        // Get the deployment URL using the first dependency for enterprise-search
        const firstDependency = workflowDetails.dependencies?.[0];
        if (!firstDependency?.model_id) {
          throw new Error('No dependency model found');
        }
        const deploymentUrl = `${deploymentBaseUrl}/${firstDependency.model_id}`;
        setDeploymentUrl(deploymentUrl);
        console.log('Deployment URL:', deploymentUrl);

        // Fetch sources
        const sources = await getSources(deploymentUrl);
        console.log('Sources retrieved:', sources);

        setRawSources(sources);

        if (sources.length === 0) {
          throw new Error('No sources found for this model');
        }

        // Fetch metadata for each source
        const documentsData = await Promise.all(
          sources.map(async (source) => {
            try {
              console.log(`Processing source: ${source.source_id} - ${source.source}`);

              // Extract filename from the full path
              const pathParts = source.source.split('/');
              const fileName = pathParts[pathParts.length - 1] || source.source;

              // Create a default document object with empty metadata
              const documentObj: DocumentMetadata = {
                document_id: source.source_id,
                document_name: fileName,
                metadata_attributes: [],
              };

              try {
                // Attempt to fetch metadata
                const metadataResponse = await getDocumentMetadata(deploymentUrl, source.source_id);
                console.log(`Metadata for ${fileName}:`, metadataResponse);

                // Check if we have metadata
                if (metadataResponse && metadataResponse.data) {
                  // Transform the metadata into our document format
                  documentObj.metadata_attributes = Object.entries(metadataResponse.data).map(
                    ([key, value]) => ({
                      attribute_name: key,
                      value: value,
                    })
                  );
                } else {
                  // No metadata available, but not an error
                  documentObj.metadata_attributes = [
                    { attribute_name: 'Info', value: 'No metadata available for this document' },
                  ];
                }
              } catch (err) {
                // Type-safe error handling
                const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
                console.warn(`Error fetching metadata: ${errorMessage}`);

                // Handle specific error cases
                if (errorMessage.includes('404')) {
                  documentObj.metadata_attributes = [
                    { attribute_name: 'Info', value: 'No metadata found for this document' },
                  ];
                } else if (errorMessage.includes('401') || errorMessage.includes('403')) {
                  documentObj.metadata_attributes = [
                    { attribute_name: 'Error', value: 'Permission denied accessing metadata' },
                  ];
                } else {
                  documentObj.metadata_attributes = [
                    { attribute_name: 'Error', value: errorMessage || 'Failed to fetch metadata' },
                  ];
                }
              }

              return documentObj;
            } catch (sourceErr) {
              console.error(`Error processing source ${source.source_id}:`, sourceErr);
              return {
                document_id: source.source_id,
                document_name: source.source.split('/').pop() || 'Unknown document',
                metadata_attributes: [
                  { attribute_name: 'Error', value: 'Error processing document' },
                ],
              };
            }
          })
        );

        console.log('All documents data:', documentsData);
        setDocuments(documentsData);
      } catch (err) {
        console.error('Error fetching metadata:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch metadata');
      } finally {
        setIsLoading(false);
      }
    };

    fetchMetadata();
  }, [workflowId]);

  const handleEdit = (document: DocumentMetadata) => {
    setSelectedDocument(document);
    setIsModalOpen(true);
  };

  const handleUpdate = (documentId: string, updatedAttributes: MetadataAttribute[]) => {
    const updatedDocuments = documents.map((doc) =>
      doc.document_id === documentId ? { ...doc, metadata_attributes: updatedAttributes } : doc
    );
    setDocuments(updatedDocuments);
    setIsModalOpen(false);
    setSelectedDocument(null);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <CircularProgress />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-red-500 text-center p-4">
        <Typography variant="h6">Error: {error}</Typography>
      </div>
    );
  }

  return (
    <>
      <TableContainer component={Paper}>
        <Table aria-label="metadata table">
          <TableHead>
            <TableRow>
              <TableCell>
                <strong>Document Name</strong>
              </TableCell>
              <TableCell>
                <strong>Metadata Attributes</strong>
              </TableCell>
              <TableCell align="center">
                <strong>Actions</strong>
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {documents.length > 0 ? (
              documents.map((doc) => (
                <TableRow key={doc.document_id}>
                  <TableCell component="th" scope="row">
                    {doc.document_name}
                  </TableCell>
                  <TableCell>
                    {doc.metadata_attributes.length > 0 ? (
                      <ul className="list-disc pl-5">
                        {doc.metadata_attributes.map((attr, index) => (
                          <li key={index}>
                            <Typography variant="body2" component="span" fontWeight="bold">
                              {attr.attribute_name}:
                            </Typography>{' '}
                            {attr.value}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <Typography variant="body2" color="text.secondary">
                        No metadata attributes found
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <IconButton
                      color="primary"
                      aria-label={`edit metadata for ${doc.document_name}`}
                      onClick={() => handleEdit(doc)}
                    >
                      <EditIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={3} align="center">
                  <Typography variant="body1">No documents found</Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {isModalOpen && selectedDocument && deploymentUrl && (
        <EditMetadataModal
          document={selectedDocument}
          deploymentUrl={deploymentUrl}
          onClose={() => setIsModalOpen(false)}
          onSave={(documentId, updatedAttributes) => handleUpdate(documentId, updatedAttributes)}
        />
      )}
    </>
  );
};

export default MetadataTable;
