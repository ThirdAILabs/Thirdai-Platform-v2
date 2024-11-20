import React, { useState } from 'react';
import { TextField, Button, Paper, Typography, Box } from '@mui/material';

interface UserCreationFormProps {
  onUserCreated: () => void;
}

const UserCreationForm: React.FC<UserCreationFormProps> = ({ onUserCreated }) => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
  });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/users/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (!response.ok) throw new Error('Failed to create user');

      setFormData({ username: '', email: '', password: '' });
      onUserCreated();
    } catch (error) {
      console.error('Error creating user:', error);
      alert('Failed to create user');
    }
  };

  return (
    <Paper elevation={2} sx={{ p: 3, mb: 4, maxWidth: '100%' }}>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Create New User
      </Typography>
      <Box
        component="form"
        onSubmit={handleSubmit}
        sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}
      >
        <TextField
          label="Username"
          value={formData.username}
          onChange={(e) => setFormData((prev) => ({ ...prev, username: e.target.value }))}
          required
          size="small"
        />
        <TextField
          label="Email"
          type="email"
          value={formData.email}
          onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
          required
          size="small"
        />
        <TextField
          label="Password"
          type="password"
          value={formData.password}
          onChange={(e) => setFormData((prev) => ({ ...prev, password: e.target.value }))}
          required
          size="small"
        />
        <Button type="submit" variant="contained" color="primary" sx={{ mt: 1 }}>
          Create User
        </Button>
      </Box>
    </Paper>
  );
};

export default UserCreationForm;
