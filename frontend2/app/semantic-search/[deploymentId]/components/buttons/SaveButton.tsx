import React from 'react';
import { SaveAlt } from '@mui/icons-material';
import { Button } from '@mui/material';

interface SaveButtonProps {
  onClick: () => void;
  style?: React.CSSProperties;
}

const SaveButton: React.FC<SaveButtonProps> = ({ onClick, style }) => {
  return (
    <Button
      variant="contained"
      color="primary"
      onClick={onClick}
      style={style}
    >
      <SaveAlt />
    </Button>
  );
};

export default SaveButton; 