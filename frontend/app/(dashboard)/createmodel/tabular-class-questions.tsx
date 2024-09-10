import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { CardDescription } from '@/components/ui/card';

interface TabularClassificationQuestionsProps {
  workflowNames: string[];
  onCreateModel?: (modelID: string) => void;
  stayOnPage?: boolean;
  appName?: string;
};

const TabularClassificationQuestions: React.FC<TabularClassificationQuestionsProps> = ({
  workflowNames,
  onCreateModel,
  stayOnPage,
  appName,
}) => {
  const [modelName, setModelName] = useState<string>(appName || '');
  const [warningMessage, setWarningMessage] = useState<string>('');
  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const handleCsvFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]; // Access the first file
    if (file && file.type === 'text/csv') {
      setCsvFile(file);
      setWarningMessage('');
    } else {
      setWarningMessage('Please upload a valid CSV file.');
    }
  };

  useEffect(() => {
    if (appName && workflowNames.includes(appName)) {
      setWarningMessage("An app with the same name already exists. Please choose a different name.");
    } else {
      setWarningMessage('');
    }
  }, [appName, workflowNames]);

  const submit = async () => {
    if (!modelName) {
      setWarningMessage('Please give the app a name.');
      return;
    }

    if (!csvFile) {
      setWarningMessage('Please upload a CSV file.');
      return;
    }

    setIsLoading(true);
    
    // Logic to handle form submission, such as API call to create the app or workflow
    try {
      // Example logic
      // const formData = new FormData();
      // formData.append('file', csvFile);
      // const response = await submitFormData(formData);

      if (onCreateModel) {
        // Assuming `response.data.model_id` is returned from the API
        onCreateModel('model_id_example');
      }

      if (!stayOnPage) {
        // Redirect to a different page if required
        // router.push("/");
      }
    } catch (error) {
      console.error('Error submitting the form:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div>
      <span className="block text-lg font-semibold">App Name</span>
      <Input
        className="text-md"
        value={modelName}
        onChange={(e) => {
          const name = e.target.value;
          const regexPattern = /^[\w-]+$/;
          let warningMessage = '';

          if (name.includes(' ')) {
            warningMessage = 'The app name cannot contain spaces. Please remove the spaces.';
          } else if (name.includes('.')) {
            warningMessage = "The app name cannot contain periods ('.'). Please remove the periods.";
          } else if (!regexPattern.test(name)) {
            warningMessage = 'The app name can only contain letters, numbers, underscores, and hyphens. Please modify the name.';
          } else if (workflowNames.includes(name)) {
            warningMessage = 'An app with the same name already exists. Please choose a different name.';
          }

          setWarningMessage(warningMessage);
          setModelName(name);
        }}
        placeholder="Enter app name"
        style={{ marginTop: '10px' }}
        disabled={!!appName && !workflowNames.includes(modelName)}
      />

      {warningMessage && (
        <span style={{ color: 'red', marginTop: '10px' }}>{warningMessage}</span>
      )}

      <span className="block text-lg font-semibold" style={{ marginTop: '20px' }}>Upload CSV File</span>
      <CardDescription>Upload a CSV file for tabular classification.</CardDescription>

      <div style={{ marginTop: '10px' }}>
        <Input type="file" accept=".csv" onChange={handleCsvFileChange} />
        {csvFile && (
          <span style={{ marginTop: '10px', display: 'block' }}>
            Selected file: {csvFile.name}
          </span>
        )}
      </div>

      <div className="flex justify-start">
        <Button
          onClick={submit}
          style={{ marginTop: '30px', width: '100%' }}
          disabled={isLoading || !csvFile || !!warningMessage}
        >
          {isLoading ? (
            <div className="flex items-center">
              <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-blue-500 mr-2"></div>
              <span>Creating...</span>
            </div>
          ) : (
            'Create'
          )}
        </Button>
      </div>
    </div>
  );
};

export default TabularClassificationQuestions;
