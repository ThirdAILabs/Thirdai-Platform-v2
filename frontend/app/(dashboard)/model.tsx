import Image from 'next/image';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { MoreHorizontal } from 'lucide-react';
import { TableCell, TableRow } from '@/components/ui/table';
import { SelectModel } from '@/lib/db';
import { deleteModel } from './actions';
import { deployModel } from '@/lib/backend';

export function Model({ model }: { model: SelectModel }) {
  return (
    <TableRow>
      <TableCell className="hidden sm:table-cell">
        <Image
          alt="Model image"
          className="aspect-square rounded-md object-cover"
          height="64"
          src={model.imageUrl}
          width="64"
        />
      </TableCell>
      <TableCell className="font-medium">{model.name}</TableCell>
      <TableCell>
        <Badge variant="outline" className="capitalize">
          {model.status}
        </Badge>
      </TableCell>
      <TableCell className="hidden md:table-cell">{model.modelType}</TableCell>
      <TableCell className="hidden md:table-cell">
        {model.trainedAt.toLocaleDateString()}
      </TableCell>
      <TableCell className="hidden md:table-cell">{model.description}</TableCell>
      <TableCell className="hidden md:table-cell">
        <button type="button" 
                onClick={()=>{
                  const username = 'peter'; // Retrieve the username dynamically if needed
                  const modelIdentifier = `${username}/${model.name}`;
                  console.log('model.name', model.name)
                  console.log('modelIdentifier', modelIdentifier)

                  deployModel({ deployment_name: model.name, model_identifier: modelIdentifier })
                    .then((data) => {
                      console.log('Deployment response:', data);

                      const baseUrl = `${window.location.protocol}//${window.location.host}`;
                      const newUrl = `${baseUrl}/search?id=${data.data.deployment_id}`;
                      window.open(newUrl, '_blank');
                    })
                    .catch((error) => {
                      console.error('Error deploying model:', error);
                    });

                }}
                className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-full text-sm p-2.5 text-center inline-flex items-center me-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
          <svg className="w-4 h-4" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
          <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
          </svg>
          <span className="sr-only">Go to endpoint</span>
        </button>
      </TableCell>
      <TableCell>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button aria-haspopup="true" size="icon" variant="ghost">
              <MoreHorizontal className="h-4 w-4" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem>Edit</DropdownMenuItem>
            <DropdownMenuItem>
              <form action={deleteModel}>
                <button type="submit">Delete</button>
              </form>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
}
