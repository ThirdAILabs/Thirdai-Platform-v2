import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button, TextField, IconButton, Box } from '@mui/material';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { CardDescription } from '@/components/ui/card';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import { trainSentenceClassifier } from '@/lib/backend';

type Category = {
  name: string;
  examples: { text: string }[];
  description: string;
};

type GeneratedData = {
  category: string;
  examples: string[];
};

interface SyntheticClassificationProps {
  workflowNames: string[];
  question: string;
  answer: string;
  onModelCreated?: (modelId: string) => void;
  modelName: string;
  stayOnPage?: boolean;
}

const predefinedChoices = ['POSITIVE_SENTIMENT', 'NEGATIVE_SENTIMENT'];

const SyntheticClassification = ({
  workflowNames,
  question,
  answer,
  onModelCreated,
  modelName,
  stayOnPage = false,
}: SyntheticClassificationProps) => {
  const [categories, setCategories] = useState<Category[]>([
    { name: '', examples: [{ text: '' }], description: '' },
  ]);
  const [isDataGenerating, setIsDataGenerating] = useState(false);
  const [generatedData, setGeneratedData] = useState<GeneratedData[]>([]);
  const [isCreating, setIsCreating] = useState(false);

  const router = useRouter();

  const handleCategoryChange = (
    index: number,
    field: keyof Category,
    value: string | { text: string }[]
  ) => {
    const updatedCategories = [...categories];
    if (field === 'examples') {
      updatedCategories[index][field] = value as { text: string }[];
    } else {
      updatedCategories[index][field] = value as string;
    }
    setCategories(updatedCategories);
  };

  const handleExampleChange = (categoryIndex: number, exampleIndex: number, value: string) => {
    const updatedCategories = [...categories];
    updatedCategories[categoryIndex].examples[exampleIndex].text = value;
    setCategories(updatedCategories);
  };

  const handleAddExample = (categoryIndex: number) => {
    const updatedCategories = [...categories];
    updatedCategories[categoryIndex].examples.push({ text: '' });
    setCategories(updatedCategories);
  };

  const handleRemoveExample = (categoryIndex: number, exampleIndex: number) => {
    const updatedCategories = [...categories];
    if (updatedCategories[categoryIndex].examples.length > 1) {
      updatedCategories[categoryIndex].examples.splice(exampleIndex, 1);
      setCategories(updatedCategories);
    }
  };

  const handleAddCategory = () => {
    setCategories([...categories, { name: '', examples: [{ text: '' }], description: '' }]);
  };

  const handleRemoveCategory = (index: number) => {
    if (categories.length > 1) {
      setCategories(categories.filter((_, i) => i !== index));
    }
  };

  const validateCategories = () => {
    return categories.every((category) => {
      return (
        category.name &&
        category.examples.length > 0 &&
        category.examples.every((ex) => ex.text) &&
        category.description &&
        !category.name.includes(' ') // Ensure no spaces in category names
      );
    });
  };

  const generateData = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();

    if (!validateCategories()) {
      alert(
        'All categories must have a name (without spaces), at least one example, and a description.'
      );
      return;
    }

    if (categories.length < 2) {
      alert('Please select at least two categories to proceed.');
      return;
    }

    if (isDataGenerating) return;

    try {
      setIsDataGenerating(true);

      const response = await fetch('/endpoints/generate-data-sentence-classification', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          question,
          answer,
          categories: categories.map((cat) => ({
            name: cat.name,
            example: cat.examples[0].text, // API expects single example
            description: cat.description,
          })),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Network response was not ok');
      }

      const result = await response.json();
      setGeneratedData(result.generatedExamples);
    } catch (error) {
      console.error('Error generating data:', error);
      alert('Error generating data:' + error);
    } finally {
      setIsDataGenerating(false);
    }
  };

  const handleCreateModel = async () => {
    if (!modelName) {
      alert('Please enter a model name.');
      return;
    }

    if (isCreating) return;

    try {
      setIsCreating(true);

      // Transform the data to match API expectations
      const apiCategories = categories.map((cat) => ({
        name: cat.name,
        example: cat.examples[0].text,
        description: cat.description,
      }));

      const modelResponse = await trainSentenceClassifier(modelName, question, apiCategories);

      console.log('Created text classification model:', modelResponse.model_id);

      if (onModelCreated) {
        onModelCreated(modelResponse.model_id);
      }

      if (!stayOnPage) {
        router.push('/');
      }
    } catch (error) {
      console.error('Error creating model:', error);
      alert('Failed to create model: ' + error);
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div>
      {generatedData.length === 0 && (
        <>
          <span className="block text-lg font-semibold">Specify Classes</span>
          <CardDescription>Define your classification categories</CardDescription>
          <div style={{ display: 'flex', flexDirection: 'column', marginTop: '10px' }}>
            {categories.map((category, categoryIndex) => (
              <div
                key={categoryIndex}
                style={{
                  marginBottom: '20px',
                  border: '1px solid #ccc',
                  padding: '10px',
                  borderRadius: '4px',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    marginBottom: '10px',
                  }}
                >
                  <TextField
                    style={{ width: '45%' }}
                    className="text-sm"
                    placeholder="Category Name"
                    value={category.name}
                    onChange={(e) => handleCategoryChange(categoryIndex, 'name', e.target.value)}
                    InputProps={{
                      inputProps: {
                        list: `category-options-${categoryIndex}`,
                      },
                    }}
                  />
                  <datalist id={`category-options-${categoryIndex}`}>
                    {predefinedChoices.map((choice, i) => (
                      <option key={i} value={choice} />
                    ))}
                  </datalist>
                  <TextField
                    style={{ width: '45%' }}
                    className="text-sm"
                    placeholder="What this category represents"
                    value={category.description}
                    onChange={(e) =>
                      handleCategoryChange(categoryIndex, 'description', e.target.value)
                    }
                  />
                </div>
                {category.examples.map((example, exampleIndex) => (
                  <div
                    key={exampleIndex}
                    style={{ display: 'flex', alignItems: 'center', marginBottom: '10px' }}
                  >
                    <TextField
                      style={{ flex: 1 }}
                      className="text-sm"
                      placeholder={`Example ${exampleIndex + 1}`}
                      value={example.text}
                      onChange={(e) =>
                        handleExampleChange(categoryIndex, exampleIndex, e.target.value)
                      }
                    />
                    <IconButton
                      onClick={() => handleRemoveExample(categoryIndex, exampleIndex)}
                      disabled={category.examples.length === 1}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </div>
                ))}
                <Button
                  variant="outlined"
                  startIcon={<AddIcon />}
                  onClick={() => handleAddExample(categoryIndex)}
                  style={{ marginRight: '10px' }}
                >
                  Add Example
                </Button>
                <Button
                  variant="contained"
                  color="error"
                  onClick={() => handleRemoveCategory(categoryIndex)}
                  disabled={categories.length === 1}
                >
                  Remove Category
                </Button>
              </div>
            ))}
            <Button
              variant="contained"
              style={{ marginTop: '10px', width: 'fit-content' }}
              onClick={handleAddCategory}
            >
              Add Category
            </Button>
            <Button
              variant="contained"
              color={isDataGenerating ? 'success' : 'primary'}
              style={{ marginTop: '30px' }}
              onClick={generateData}
            >
              {isDataGenerating ? 'Generating data...' : 'Generate data'}
            </Button>
          </div>
        </>
      )}

      {isDataGenerating && (
        <div className="flex justify-center mt-5">
          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      )}

      {!isDataGenerating && generatedData.length > 0 && (
        <div className="mt-5">
          <h3 className="mb-3 text-lg font-semibold">Generated Data</h3>
          <Table style={{ marginTop: '10px' }}>
            <TableHeader>
              <TableRow>
                <TableHead>Category</TableHead>
                <TableHead>Generated Examples</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {generatedData.map((data, index) => (
                <TableRow key={index}>
                  <TableCell className="font-medium" align="left">
                    {data.category}
                  </TableCell>
                  <TableCell className="font-medium" align="left">
                    <ul>
                      {data.examples.map((example, i) => (
                        <li key={i}>{example}</li>
                      ))}
                    </ul>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <div
            style={{
              display: 'flex',
              flexDirection: 'row',
              justifyContent: 'space-between',
              gap: '10px',
              marginTop: '20px',
            }}
          >
            <Button
              variant="outlined"
              style={{ width: '100%' }}
              onClick={() => setGeneratedData([])}
            >
              Redefine Categories
            </Button>
            <Button
              variant="contained"
              style={{ width: '100%' }}
              onClick={handleCreateModel}
              disabled={isCreating}
            >
              {isCreating ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-blue-500 mr-2"></div>
                  <span>Creating...</span>
                </div>
              ) : (
                'Create Model'
              )}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default SyntheticClassification;
