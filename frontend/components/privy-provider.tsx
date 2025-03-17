'use client';

import { PrivyProvider } from '@privy-io/react-auth';
import { privyConfig } from '@/lib/privy';

export function PrivyClientProvider({ children }: { children: React.ReactNode }) {
  return (
    <PrivyProvider
      appId={process.env.NEXT_PUBLIC_PRIVY_APP_ID || ''}
      config={privyConfig}
    >
      {children}
    </PrivyProvider>
  );
} 