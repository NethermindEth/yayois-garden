'use client';

import Link from 'next/link'
import { Logo } from '@/components/Logo'
import { WalletButton } from '@/components/wallet-button'

export function Header() {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-black/90 backdrop-blur-sm">
      <div className="container flex items-center justify-between h-16 px-4">
        <Link href="/" className="flex items-center space-x-2">
          <div className="w-6 h-6">
            <Logo variant="colored" />
          </div>
          <span className="text-lg font-light">YAYOI&apos;S GARDEN</span>
        </Link>
        <nav className="hidden md:flex items-center space-x-6">
          <Link href="/create" className="text-sm hover:text-gray-300">
            Create
          </Link>
          <Link href="/gallery" className="text-sm hover:text-gray-300">
            Gallery
          </Link>
          <Link href="/leaderboard" className="text-sm hover:text-gray-300">
            Leaderboard
          </Link>
        </nav>
        <WalletButton />
      </div>
    </header>
  )
}

