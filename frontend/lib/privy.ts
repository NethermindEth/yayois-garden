import { type PrivyClientConfig } from '@privy-io/react-auth';

export const privyConfig: PrivyClientConfig = {
  appearance: {
    theme: 'dark',
    accentColor: '#676FFF',
    logo: '/logo.svg',
  },
  embeddedWallets: {
    createOnLogin: 'users-without-wallets',
  },
}; 