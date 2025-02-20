import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Chip,
  Stack,
} from '@mui/material';

interface ConfirmationDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  tokenTypes: string[];
}

const TokenTypeConfirmationDialog = ({
  open,
  onClose,
  onConfirm,
  tokenTypes,
}: ConfirmationDialogProps) => {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Confirm Token Types</DialogTitle>
      <DialogContent>
        <Typography sx={{ mb: 2 }}>The following token types were detected in your CSV:</Typography>
        <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mb: 2 }}>
          {tokenTypes.map((label) => (
            <Chip key={label} label={label} color="primary" sx={{ mb: 1 }} />
          ))}
        </Stack>
        <Typography>
          Are you sure you want to proceed with training using these token types?
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={onConfirm} variant="contained" color="primary">
          Proceed with Training
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default TokenTypeConfirmationDialog;
