import React, { useState } from 'react';
import { Button } from '@mui/material';
import { Input } from '@/components/ui/input';
import { retrainTokenClassifier } from '@/lib/backend';

interface RetrainButtonProps {
  modelId: string;
}

export default function RetrainButton({ modelId }: RetrainButtonProps) {
  const [newModelName, setNewModelName] = useState('');
  const [isRetraining, setIsRetraining] = useState(false);
  const [retrainError, setRetrainError] = useState('');

  const handleCreateNewModel = async () => {
    if (!newModelName) {
      setRetrainError('Please enter a name for the new model');
      return;
    }

    setIsRetraining(true);
    setRetrainError('');

    try {
      await retrainTokenClassifier({
        model_name: newModelName,
        base_model_id: modelId,
      });
      console.log('New model created successfully');
      setNewModelName(''); // Clear the input after successful creation
    } catch (error) {
      setRetrainError(
        error instanceof Error ? error.message : 'An error occurred while creating the new model'
      );
    } finally {
      setIsRetraining(false);
    }
  };

  return (
    <div className="flex flex-col space-y-4">
      <div className="flex space-x-4">
        <Input
          type="text"
          placeholder="New model name"
          value={newModelName}
          onChange={(e) => setNewModelName(e.target.value)}
          className="w-64"
        />
        <Button onClick={handleCreateNewModel} disabled={isRetraining} variant="contained">
          Create new model with feedback
        </Button>
      </div>
      {isRetraining && <p className="text-blue-500">Creating new model...</p>}
      {retrainError && <p className="text-red-500">{retrainError}</p>}
    </div>
  );
}
