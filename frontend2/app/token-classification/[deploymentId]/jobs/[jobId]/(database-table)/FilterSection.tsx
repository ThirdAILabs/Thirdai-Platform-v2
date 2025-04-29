import { useRef, useState } from 'react';
import { ChevronDown, ChevronUp, ListFilter } from 'lucide-react';

interface FilterSectionProps {
  groups: string[];
  tags: string[];
  groupFilters: Record<string, boolean>;
  tagFilters: Record<string, boolean>;
  onGroupFilterChange: (filterKey: string) => void;
  onTagFilterChange: (filterKey: string) => void;
  onSelectAllGroups: () => void;
  onDeselectAllGroups: () => void;
  onSelectAllTags: () => void;
  onDeselectAllTags: () => void;
}

export function FilterSection({
  groups,
  tags,
  groupFilters,
  tagFilters,
  onGroupFilterChange,
  onTagFilterChange,
  onSelectAllGroups,
  onDeselectAllGroups,
  onSelectAllTags,
  onDeselectAllTags,
}: FilterSectionProps) {
  const [isGroupsExpanded, setIsGroupsExpanded] = useState(true);
  const [isTagsExpanded, setIsTagsExpanded] = useState(true);
  const [showShadow, setShowShadow] = useState(false);
  const filterScrollRef = useRef<HTMLDivElement>(null);

  const handleFilterScroll = () => {
    if (filterScrollRef.current) {
      setShowShadow(filterScrollRef.current.scrollTop > 0);
    }
  };

  const toggleGroups = () => setIsGroupsExpanded(!isGroupsExpanded);
  const toggleTags = () => setIsTagsExpanded(!isTagsExpanded);

  return (
    <div className="w-64 flex flex-col border-r relative">
      {/* Fixed Filter Header */}
      <div className="sticky top-0 p-6 pb-2 pt-4 z-10">
        <div className="flex items-center gap-2">
          <ListFilter className="h-5 w-5" />
          <span className="flex font-medium h-[40px] items-center">
            Filter
          </span>
        </div>
      </div>

      {/* Scrollable Filter Content with Shadow */}
      <div
        ref={filterScrollRef}
        className="flex-1 overflow-y-auto"
        onScroll={handleFilterScroll}
        style={{
          boxShadow: showShadow
            ? 'inset 0 4px 6px -4px rgba(0, 0, 0, 0.1)'
            : 'none',
        }}
      >
        <div className="p-6 pt-4 space-y-6">
          {/* Groups Section */}
          <div>
            <div
              className="flex items-center justify-between text-sm text-gray-600 mb-2 cursor-pointer hover:text-gray-800"
              onClick={toggleGroups}
            >
              <span>Groups</span>
              {isGroupsExpanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </div>
            {isGroupsExpanded && (
              <>
                <div className="flex gap-2 mb-2">
                  <button
                    onClick={onSelectAllGroups}
                    className="text-xs text-blue-600 hover:text-blue-800"
                  >
                    Select All
                  </button>
                  <span className="text-gray-300">|</span>
                  <button
                    onClick={onDeselectAllGroups}
                    className="text-xs text-blue-600 hover:text-blue-800"
                  >
                    Deselect All
                  </button>
                </div>
                <div className="space-y-2">
                  {groups.map((filter) => (
                    <label key={filter} className="flex items-center">
                      <input
                        type="checkbox"
                        checked={groupFilters[filter]}
                        onChange={() => onGroupFilterChange(filter)}
                        className="mr-2"
                      />
                      {filter}
                    </label>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Tags Section */}
          <div>
            <div
              className="flex items-center justify-between text-sm text-gray-600 mb-2 cursor-pointer hover:text-gray-800"
              onClick={toggleTags}
            >
              <span>Tags</span>
              {isTagsExpanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </div>
            {isTagsExpanded && (
              <>
                <div className="flex gap-2 mb-2">
                  <button
                    onClick={onSelectAllTags}
                    className="text-xs text-blue-600 hover:text-blue-800"
                  >
                    Select All
                  </button>
                  <span className="text-gray-300">|</span>
                  <button
                    onClick={onDeselectAllTags}
                    className="text-xs text-blue-600 hover:text-blue-800"
                  >
                    Deselect All
                  </button>
                </div>
                <div className="space-y-2">
                  {tags.map((filter) => (
                    <label key={filter} className="flex items-center">
                      <input
                        type="checkbox"
                        checked={tagFilters[filter]}
                        onChange={() => onTagFilterChange(filter)}
                        className="mr-2"
                      />
                      {filter}
                    </label>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
} 