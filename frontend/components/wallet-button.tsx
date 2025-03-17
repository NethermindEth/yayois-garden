'use client';

import { usePrivy } from '@privy-io/react-auth';
import { Button } from '@/components/ui/button';

export function WalletButton() {
  const { login, logout, authenticated, user } = usePrivy();

  if (!authenticated) {
    return (
      <Button 
        onClick={login}
        variant="outline" 
        className="bg-white text-black rounded hover:bg-gray-200"
      >
        Connect Wallet
      </Button>
    );
  }

  return (
    <Button
      onClick={logout}
      variant="outline"
      className="bg-white text-black rounded hover:bg-gray-200"
    >
      {user?.wallet?.address?.slice(0, 6)}...{user?.wallet?.address?.slice(-4)}
    </Button>
  );
} 