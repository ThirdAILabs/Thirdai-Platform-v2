'use client';

import { useEffect, useState } from 'react';
import {
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  Table
} from '@/components/ui/table';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import { Model } from './model';
import { SelectModel } from '@/lib/db';
import { useRouter } from 'next/navigation';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { fetchPublicModels, fetchPrivateModels } from "@/lib/backend"

// Define a type for the private model data structure
type PrivateModel = {
  access_level: string;
  domain: string;
  latency: string;
  model_name: string;
  num_params: string;
  publish_date: string;
  size: string;
  size_in_memory: string;
  thirdai_version: string;
  training_time: string;
  type: string;
  user_email: string;
  username: string;
};

// Map the private model data to the required model structure
const mapPrivateModelToSelectModel = (privateModel: PrivateModel, index: number): SelectModel => ({
  id: index + 1, // Use index as a unique identifier for demonstration
  imageUrl: '/thirdai-small.png', // Provide a default or dummy image URL
  name: privateModel.model_name,
  status: 'active', // Assuming all fetched models are active, replace with appropriate status if available
  trainedAt: new Date(privateModel.publish_date),
  description: `Model by ${privateModel.username}`,
  deployEndpointUrl: null, // Provide a default or dummy endpoint URL
  onDiskSizeKb: privateModel.size,
  ramSizeKb: privateModel.size_in_memory,
  numberParameters: Number(privateModel.num_params),
  rlhfCounts: 0, // Replace with the appropriate value if available
  modelType: 'ner model', // Adjust the model type based on expected literals
});

export function ModelsTable({
  models,
  offset,
  totalModels
}: {
  models: SelectModel[];
  offset: number;
  totalModels: number;
}) {
  let router = useRouter();
  let modelsPerPage = 5;

  function prevPage() {
    router.back();
  }

  function nextPage() {
    router.push(`/?offset=${offset}`, { scroll: false });
  }

  const [privateModels, setPrivateModels] = useState<SelectModel[]>([])

  useEffect(() => {
    async function getModels() {
        try {
          const publicModels = await fetchPublicModels('');
          console.log('publicModels', publicModels)

          const response = await fetchPrivateModels('');
          const privateModels: PrivateModel[] = response.data; // Extract the data field
          console.log('privateModels', privateModels)

          const mappedModels = privateModels.map(mapPrivateModelToSelectModel);

          console.log('mappedModels', mappedModels)
          setPrivateModels(mappedModels)
        } catch (err) {
          if (err instanceof Error) {
              console.log(err.message);
          } else {
              console.log('An unknown error occurred');
          }
      }
    }

    getModels();
  }, []);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Models</CardTitle>
        <CardDescription>
          Manage your models and view their performance.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="hidden w-[100px] sm:table-cell">
                <span className="sr-only">Image</span>
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="hidden md:table-cell">Type</TableHead>
              <TableHead className="hidden md:table-cell">Trained at</TableHead>
              <TableHead className="hidden md:table-cell">Description</TableHead>
              <TableHead className="hidden md:table-cell">Endpoint</TableHead>
              <TableHead>
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {models.map((model) => (
              <Model key={model.id} model={model} />
            ))}

            {privateModels.map((model) => (
                <Model key={model.id} model={model} />
            ))}
          </TableBody>
        </Table>
      </CardContent>
      <CardFooter>
        <form className="flex items-center w-full justify-between">
          <div className="text-xs text-muted-foreground">
            Showing{' '}
            <strong>
              {Math.min(offset - modelsPerPage, totalModels) + 1}-{offset}
            </strong>{' '}
            of <strong>{totalModels}</strong> models
          </div>
          <div className="flex">
            <Button
              formAction={prevPage}
              variant="ghost"
              size="sm"
              type="submit"
              disabled={offset === modelsPerPage}
            >
              <ChevronLeft className="mr-2 h-4 w-4" />
              Prev
            </Button>
            <Button
              formAction={nextPage}
              variant="ghost"
              size="sm"
              type="submit"
              // disabled={offset + modelsPerPage > totalModels}
            >
              Next
              <ChevronRight className="ml-2 h-4 w-4" />
            </Button>
          </div>
        </form>
      </CardFooter>
    </Card>
  );
}
