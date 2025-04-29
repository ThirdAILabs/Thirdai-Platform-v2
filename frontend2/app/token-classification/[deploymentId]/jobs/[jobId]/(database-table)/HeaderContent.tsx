import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import SaveButton from '../../../../../semantic-search/[deploymentId]/components/buttons/SaveButton';
import { ViewMode } from './types';

interface HeaderContentProps {
  viewMode: ViewMode;
  query: string;
  onQueryChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onViewModeChange: (value: ViewMode) => void;
  onSave: () => void;
}

export function HeaderContent({
  viewMode,
  query,
  onQueryChange,
  onViewModeChange,
  onSave,
}: HeaderContentProps) {
  // Cast the ViewMode to string for compatibility with the Tabs component
  const handleViewModeChange = (value: string) => {
    onViewModeChange(value as ViewMode);
  };

  return (
    <div className="p-6 pb-2 pt-4">
      <div className="flex items-center space-x-4">
        <div className="font-medium">View By</div>
        <Tabs value={viewMode} onValueChange={handleViewModeChange}>
          <TabsList>
            <TabsTrigger value="object">Object</TabsTrigger>
            <TabsTrigger value="classified-token">
              Classified Token
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="font-medium pl-2">Query</div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <div className="flex-1">
              <Input
                type="text"
                placeholder="Enter query..."
                value={query}
                onChange={onQueryChange}
              />
            </div>
            <SaveButton
              onClick={onSave}
              style={{
                width: '40px',
                height: '40px',
                minWidth: '40px',
                padding: '8px',
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
} 