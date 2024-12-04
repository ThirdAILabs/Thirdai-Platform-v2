'use client';

import { useState, createContext } from 'react';

interface ClickContext {
  key: string;
  register: (key: string) => void;
  isOutside: (key: string) => boolean;
}

// Dummy default values. They will be correctly initialized by the time the context is
// used by the provider's children.
export const ClickContext = createContext<ClickContext>({
  key: '',
  register: (key: string) => {},
  isOutside: (key: string) => false,
});

export default function ClickWrapper({ children }: { children: React.ReactNode }) {
  const [registeredKey, register] = useState('');
  const isOutside = (key: string) => key === registeredKey;
  return (
    <ClickContext.Provider value={{ key: registeredKey, register, isOutside }}>
      {children}
    </ClickContext.Provider>
  );
}
