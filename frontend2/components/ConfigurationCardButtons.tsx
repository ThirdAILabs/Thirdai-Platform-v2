'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { Plus } from 'lucide-react';
import { Typography } from '@mui/material';

interface ButtonTitleProps {
  children: React.ReactNode;
}

const ButtonTitle: React.FC<ButtonTitleProps> = ({ children }) => (
  <Typography 
    variant="h6" 
    className="text-gray-900"
    sx={{
      fontSize: '1.125rem',
      fontWeight: 500,
      lineHeight: 1.2
    }}
  >
    {children}
  </Typography>
);

// Base button-card component
interface BaseCardButtonProps {
  className?: string;
  children: React.ReactNode;
  onClick?: () => void;
  showDeleteButton?: boolean;
  onDelete?: () => void;
  isSelected?: boolean;
  disabled?: boolean;
  preserveStyles?: boolean;
}

const BaseCardButton: React.FC<BaseCardButtonProps> = ({
  className = '',
  children,
  onClick,
  showDeleteButton = false,
  onDelete,
  isSelected = false,
  disabled = false,
  preserveStyles = false,
}) => {
  return (
    <div className="relative group">
      <Button
        variant="outline"
        onClick={onClick}
        className={`
          h-[160px] w-48 
          p-6
          border border-gray-200
          hover:border-blue-300
          transition-all duration-200
          ${isSelected ? 'border-2 border-blue-500' : ''}
          ${disabled && !preserveStyles ? 'opacity-50' : ''}
          ${className}
        `}
        disabled={disabled}
      >
        {children}
      </Button>
      {showDeleteButton && onDelete && !disabled && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="absolute -top-2 -right-2 bg-white w-8 h-8 rounded-full shadow-md opacity-0 group-hover:opacity-100 transition-opacity duration-200 hover:bg-red-50 flex items-center justify-center"
        >
          <DeleteIcon className="text-red-500 w-5 h-5" />
        </button>
      )}
      {isSelected && (
        <div className="absolute -top-2 -right-2 bg-white rounded-full">
          <CheckCircleIcon className="text-blue-500 text-xl" />
        </div>
      )}
    </div>
  );
};

// Storage Option Button
interface StorageOptionButtonProps {
  title: string;
  description?: string;
  isSelected?: boolean;
  onClick?: () => void;
  showEditIcon?: boolean;
  disabled?: boolean;
  preserveStyles?: boolean;
}

export const StorageOptionButton: React.FC<StorageOptionButtonProps> = ({
  title,
  description,
  isSelected = false,
  onClick,
  showEditIcon = false,
  disabled = false,
  preserveStyles = false,
}) => {
  return (
    <BaseCardButton
      onClick={onClick}
      isSelected={isSelected}
      disabled={disabled}
      preserveStyles={preserveStyles}
    >
      <div className="h-full w-full flex flex-col">
        <div className="text-left">
          <ButtonTitle>{title}</ButtonTitle>
          {description && (
            <p className="text-sm text-gray-500 mt-2 w-full whitespace-normal break-words">{description}</p>
          )}
        </div>
        {showEditIcon && !disabled && (
          <div className="mt-auto self-end">
            <EditIcon className="text-gray-500" />
          </div>
        )}
      </div>
    </BaseCardButton>
  );
};

// Group button component
interface Group {
  name: string;
  definition: string;
}

interface GroupButtonProps {
  group: Group;
  onEdit: (name: string, definition: string) => void;
  onDelete: () => void;
}

export const GroupButton: React.FC<GroupButtonProps> = ({ group, onEdit, onDelete }) => {
  const handleEdit = () => {
    const newName = prompt('Group Name:', group.name);
    if (newName !== null) {
      const newDefinition = prompt('Group Definition:', group.definition);
      if (newDefinition !== null) {
        onEdit(newName, newDefinition);
      }
    }
  };

  return (
    <BaseCardButton
      onClick={handleEdit}
      showDeleteButton
      onDelete={onDelete}
    >
      <div className="h-full w-full">
        <div className="text-left">
          <ButtonTitle>{group.name}</ButtonTitle>
          <p className="text-sm text-gray-500 mt-2 w-full whitespace-normal break-words">
            {group.definition || 'Click to add definition'}
          </p>
        </div>
      </div>
    </BaseCardButton>
  );
};

// Add new group button component
export const AddGroupButton: React.FC<{ onClick: () => void }> = ({ onClick }) => {
  return (
    <BaseCardButton onClick={onClick}>
      <div className="h-full w-full flex flex-col items-center justify-center">
        <Plus className="h-12 w-12 mb-2 text-gray-900 transition-colors duration-200" />
        <ButtonTitle>Define new group</ButtonTitle>
      </div>
    </BaseCardButton>
  );
}; 