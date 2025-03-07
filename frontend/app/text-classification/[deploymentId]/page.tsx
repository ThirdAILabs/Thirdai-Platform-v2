'use client';

import { useState, useRef } from 'react';
import { Container, Box, CircularProgress, Typography, Alert } from '@mui/material';
import { Tabs } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import * as _ from 'lodash';
import { useTextClassificationEndpoints } from '@/lib/backend';
import { Card, CardHeader, CardTitle } from '@/components/ui/card';
import { parseCSV, parseExcel, parseTXT } from '@/utils/fileParsingUtils';
import InferenceTimeDisplay from '@/components/ui/InferenceTimeDisplay';
import ThumbsUpButton from './thumbsUpButton';
import ExpandingInput from '@/components/ui/ExpandingInput';

interface ParsedData {
  type: 'csv' | 'pdf' | 'other';
  content: string;
  rows?: { label: string; content: string }[];
  pdfParagraphs?: string[];
}

export default function Page() {
  const { workflowName, predict, getTextFromFile } = useTextClassificationEndpoints();
  const [inputText, setInputText] = useState('');
  const [predictions, setPredictions] = useState<[string, number][]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [fileError, setFileError] = useState<string | null>(null);
  const [parsedData, setParsedData] = useState<ParsedData | null>(null);
  const [processingTime, setProcessingTime] = useState<number | undefined>();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const MAX_FILE_SIZE = 1024 * 1024; // 1MB in bytes

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    setFileError(null);
    setPredictions([]);

    if (file) {
      if (file.size > MAX_FILE_SIZE) {
        setFileError('File size exceeds 1MB. Please use the API for larger files.');
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        return;
      }

      const fileExtension = file.name.split('.').pop()?.toLowerCase();
      setIsLoading(true);

      try {
        let parsed: ParsedData;
        if (fileExtension === 'csv') {
          parsed = await parseCSV(file);
        } else if (['pdf', 'docx', 'html'].includes(fileExtension ?? '')) {
          const content = await getTextFromFile(file);
          parsed = {
            type: 'pdf',
            content: content.join('\n'),
            pdfParagraphs: content,
          };
        } else if (['xls', 'xlsx'].includes(fileExtension ?? '')) {
          const excelRows = await parseExcel(file);
          parsed = {
            type: 'csv',
            content: excelRows.map((row) => row.content).join('\n\n'),
            rows: excelRows,
          };
        } else {
          parsed = { type: 'other', content: await parseTXT(file) };
        }

        setInputText(parsed.content);
        setParsedData(parsed);
        handleRun(parsed.content);
      } catch (error) {
        console.error('Error processing file:', error);
        setFileError('Error processing file. Please try again.');
      } finally {
        setIsLoading(false);
      }
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleRun = async (text: string = inputText) => {
    if (text === '') {
      return;
    }
    setIsLoading(true);
    try {
      const result = await predict(text);
      console.log('result', result);
      setPredictions(
        result.data.prediction_results.predicted_classes.map((predClass) => [
          predClass.class,
          Math.floor(predClass.score * 100),
        ])
      );
      setProcessingTime(result.data.time_taken);
    } catch (error) {
      console.error('Error during prediction:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const renderPredictions = () => (
    <Box mt={4} display="flex" gap={4}>
      {/* Left side - Predictions */}
      <div style={{ flex: 2 }}>
        {predictions.map((prediction, index) => (
          <Card
            key={index}
            className={`${index > 0 ? 'mt-4' : ''} bg-white hover:bg-gray-50 transition-colors`}
          >
            <div className="flex justify-between items-center p-6">
              <div>
                <p className="text-sm text-muted-foreground mb-1">class</p>
                <CardTitle className="text-xl font-bold">
                  {prediction[0].replace(/_/g, ' ')}
                </CardTitle>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center min-w-[100px]">
                <p className="text-sm text-muted-foreground mb-1">score</p>
                <CardTitle className="text-xl">{prediction[1]}</CardTitle>
              </div>
              <ThumbsUpButton inputText={inputText} prediction={prediction[0]} />
            </div>
          </Card>
        ))}
      </div>

      {/* Right side - Inference Time */}
      <div style={{ flex: 1 }}>
        {processingTime !== undefined && <InferenceTimeDisplay processingTime={processingTime} />}
      </div>
    </Box>
  );

  return (
    <div
      className="bg-muted"
      style={{
        width: '100%',
        display: 'flex',
        justifyContent: 'center',
        height: '100vh',
      }}
    >
      <Tabs defaultValue="interact" style={{ width: '100%' }}>
        <div style={{ position: 'fixed', top: '20px', left: '20px' }}>
          <div className="text-muted-foreground" style={{ fontSize: '16px' }}>
            Text Classification
          </div>
          <div style={{ fontWeight: 'bold', fontSize: '24px' }}>{workflowName}</div>
        </div>
        <Container
          style={{
            textAlign: 'center',
            paddingTop: '20vh',
            width: '70%',
            minWidth: '400px',
            maxWidth: '800px',
          }}
        >
          <Box display="flex" flexDirection="column" width="100%">
            <Box display="flex" justifyContent="center" alignItems="center" width="100%">
              <ExpandingInput onSubmit={handleRun} onFileChange={handleFileChange} />
            </Box>

            {fileError && (
              <Alert severity="error" style={{ marginTop: '8px' }}>
                {fileError}
              </Alert>
            )}
          </Box>

          {isLoading ? (
            <Box mt={4} display="flex" justifyContent="center">
              <CircularProgress />
            </Box>
          ) : (
            predictions.length > 0 && renderPredictions()
          )}
        </Container>
      </Tabs>
    </div>
  );
}
