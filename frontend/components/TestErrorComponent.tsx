// components/TestErrorComponent.tsx
import React from 'react';

const TestErrorComponent: React.FC = () => {
  // Deliberately throw an error
  throw new Error('Test error for Sentry');

  return <div>This is a test component.</div>;
};

export default TestErrorComponent;
