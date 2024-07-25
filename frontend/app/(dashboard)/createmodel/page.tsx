import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import ChooseProblem from './choose-model';
import { getModels } from '@/lib/db';

export default async function NewModelPage({
  searchParams
}: {
  searchParams: { q: string; offset: string };
}) {

  const search = searchParams.q ?? '';
  const offset = searchParams.offset ?? 0;
  const { models, newOffset, totalModels } = await getModels(
    search,
    Number(offset)
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Model</CardTitle>
        <CardDescription>Create a new model with a few simple steps.</CardDescription>
      </CardHeader>
      <CardContent>
        <ChooseProblem />
      </CardContent>
    </Card>
  );
}
