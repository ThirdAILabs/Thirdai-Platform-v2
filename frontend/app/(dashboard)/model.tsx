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
      {/* <TableCell className="hidden md:table-cell">{`$${model.price}`}</TableCell> */}
      {/* <TableCell className="hidden md:table-cell">{model.stock}</TableCell> */}
      <TableCell className="hidden md:table-cell">
        {model.trainedAt.toLocaleDateString()}
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
