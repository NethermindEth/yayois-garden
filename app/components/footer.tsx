import Link from 'next/link'
import { Logo } from '@/components/Logo'
import { Twitter, Instagram, DiscIcon as Discord } from 'lucide-react'

export function Footer() {
  return (
    <footer className="bg-black text-gray-300 py-8">
      <div className="container mx-auto px-4 mt-0">
        <div className="flex flex-col md:flex-row justify-between items-center">
          <div className="mb-4 md:mb-0 flex items-center">
            <div className="w-8 h-8 mr-2">
              <Logo variant="white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">Yayoi's Garden</h2>
              <p className="text-sm">Where prompts become precious</p>
            </div>
          </div>
          <div className="flex space-x-4">
            <Link href="https://twitter.com" target="_blank" rel="noopener noreferrer">
              <Twitter className="w-6 h-6 hover:text-white transition-colors" />
            </Link>
            <Link href="https://instagram.com" target="_blank" rel="noopener noreferrer">
              <Instagram className="w-6 h-6 hover:text-white transition-colors" />
            </Link>
            <Link href="https://discord.com" target="_blank" rel="noopener noreferrer">
              <Discord className="w-6 h-6 hover:text-white transition-colors" />
            </Link>
          </div>
        </div>
        <div className="mt-8 text-center text-sm">
          © {new Date().getFullYear()} Yayoi's Garden. All rights reserved.
        </div>
      </div>
    </footer>
  )
}

