'use client';

import { PrivyClientProvider } from '@/components/privy-provider';
import { Header } from '@/components/header';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <PrivyClientProvider>
      <Header />
      {children}
    </PrivyClientProvider>
  );
} 