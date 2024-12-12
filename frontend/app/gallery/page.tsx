import { Header } from '@/components/header'
import { NFTCard } from '@/components/nft-card'
import { Button } from '@/components/ui/button'
import { Footer } from '@/components/footer'

// Reusing the same NFTs data for demonstration
const nfts = [
  {
    title: "NEON FACE",
    image: "/neonface.jpg",
    ethSpend: 8.1,
    creator: {
      name: "ALICE",
       avatar: "/avatar.jpg"
    }
  },
  // ... other NFTs
]

export default function Gallery() {
  return (
    <main className="flex flex-col min-h-screen bg-black">
      <Header />
      <div className="flex-grow container px-4 pt-32 pb-20">
        <div className="flex justify-between items-center mb-8">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold">Gallery</h1>
            <p className="text-gray-400">Explore the latest AI-generated artworks</p>
          </div>
          <div className="flex gap-4">
            <Button variant="outline">Latest</Button>
            <Button variant="outline">Popular</Button>
            <Button variant="outline">Trending</Button>
          </div>
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {nfts.map((nft, index) => (
            <NFTCard key={index} {...nft} />
          ))}
        </div>
      </div>
      <Footer />
    </main>
  )
}

