import React from 'react';
import { FaSave } from 'react-icons/fa';
import { Button } from '@mui/material';

interface SaveButtonProps {
  onClick: () => void;
  style?: React.CSSProperties;
}

export default function SaveButton({ onClick, style }: SaveButtonProps) {
  const defaultStyle = {
    width: '48px',
    height: '48px',
    minWidth: 'unset',
    padding: '12px',
    ...style
  };

  return (
    <Button
      variant="contained"
      color="primary"
      style={defaultStyle}
      onClick={onClick}
    >
      <FaSave style={{ fontSize: '24px' }} />
    </Button>
  );
}
