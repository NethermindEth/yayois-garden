'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Header } from '@/components/header'
import { Footer } from '@/components/footer'
import { Button } from '@/components/ui/button'

const leaderboardData = [
  { rank: 1, model: "Dreamscape", artist: "Alice", promptCount: 156 },
  { rank: 2, model: "Neon Nights", artist: "Bob", promptCount: 132 },
  { rank: 3, model: "Cosmic Whispers", artist: "Charlie", promptCount: 98 },
  { rank: 4, model: "Ethereal Echoes", artist: "Diana", promptCount: 87 },
  { rank: 5, model: "Quantum Canvas", artist: "Ethan", promptCount: 75 },
]

type TimeFrame = 'Daily' | 'Weekly' | 'Monthly'

export default function Leaderboard() {
  const [timeFrame, setTimeFrame] = useState<TimeFrame>('Daily')

  return (
    <main className="bg-black min-h-screen flex flex-col">
      <Header />
      <div className="container px-4 pt-32 pb-20 flex-grow flex flex-col">
        <div className="max-w-4xl mx-auto flex-grow">
          <h1 className="text-4xl font-bold text-center mb-12">Global Leaderboard</h1>
          
          <div className="bg-gray-900 rounded-lg overflow-hidden mb-8">
            <div className="flex">
              {(['Daily', 'Weekly', 'Monthly'] as TimeFrame[]).map((frame) => (
                <Button
                  key={frame}
                  onClick={() => setTimeFrame(frame)}
                  variant={timeFrame === frame ? 'default' : 'ghost'}
                  className={`flex-1 rounded-none h-12 ${
                    timeFrame === frame ? 'bg-black text-white' : 'text-gray-400'
                  }`}
                >
                  {frame}
                </Button>
              ))}
            </div>
          </div>

          <div className="bg-gray-900/50 rounded-lg overflow-hidden backdrop-blur-sm">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-800">
                  <th className="py-4 px-6 text-left text-sm font-semibold text-gray-400">Rank</th>
                  <th className="py-4 px-6 text-left text-sm font-semibold text-gray-400">Model</th>
                  <th className="py-4 px-6 text-left text-sm font-semibold text-gray-400">Artist</th>
                  <th className="py-4 px-6 text-right text-sm font-semibold text-gray-400">Prompt Count</th>
                </tr>
              </thead>
              <tbody>
                {leaderboardData.map((item) => (
                  <tr key={item.rank} className="border-b border-gray-800 last:border-b-0">
                    <td className="py-4 px-6 text-sm">{item.rank}</td>
                    <td className="py-4 px-6 text-sm">{item.model}</td>
                    <td className="py-4 px-6 text-sm">
                      <Link href={`/user/${item.artist.toLowerCase()}`} className="hover:text-purple-400 transition-colors">
                        {item.artist}
                      </Link>
                    </td>
                    <td className="py-4 px-6 text-sm text-right">{item.promptCount}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <Footer />
    </main>
  )
}

