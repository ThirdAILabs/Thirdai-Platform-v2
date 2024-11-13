import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
} from '@mui/material';

interface LabelConfirmationDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  labels: string[];
}

const LabelConfirmationDialog = ({ open, onClose, onConfirm, labels }: LabelConfirmationDialogProps) => {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Confirm Labels</DialogTitle>
      <DialogContent>
        <Typography variant="body1" gutterBottom>
          The following labels were detected in your CSV file:
        </Typography>
        <Box sx={{ mt: 2, mb: 2 }}>
          {labels.map((label, index) => (
            <Typography key={index} variant="body2" sx={{ mb: 1 }}>
              • {label}
            </Typography>
          ))}
        </Box>
        <Typography variant="body2" color="text.secondary">
          Please confirm that these are the correct labels for your classification task.
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} color="primary">
          Cancel
        </Button>
        <Button onClick={onConfirm} variant="contained" color="primary">
          Confirm & Train
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default LabelConfirmationDialog;