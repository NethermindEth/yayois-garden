import { Header } from '@/components/header'
import { Footer } from '@/components/footer'
import { NFTCard } from '@/components/nft-card'
import Image from 'next/image'
import { BackButton } from '@/components/BackButton'

// This is a mock function to simulate fetching user data
// In a real application, you would fetch this data from your backend
function getUserData(username: string) {
  return {
    name: username.toUpperCase(),
    avatar: "/avatar.jpg",
    bio: "Digital artist and NFT creator",
    totalArtworks: 42,
    totalEthSpent: 156.7
  }
}

// This is a mock function to simulate fetching user's NFTs
// In a real application, you would fetch this data from your backend
function getUserNFTs(username: string) {
  return [
    {
      title: "COSMIC DREAMS",
      image: "/cosmicwhispers.jpg",
      ethSpend: 5.6,
      creator: {
        name: username.toUpperCase(),
       avatar: "/avatar.jpg"
      }
    },
    {
      title: "NEON SUNSET",
      image: "/neonface.jpg",
      ethSpend: 4.2,
      creator: {
        name: username.toUpperCase(),
         avatar: "/avatar.jpg"
      }
    },
    {
      title: "DIGITAL OASIS",
      image: "/marble.jpg",
      ethSpend: 3.8,
      creator: {
        name: username.toUpperCase(),
        avatar: "/avatar.jpg"
      }
    }
  ]
}

type Params = Promise<{ username: string }>

export default async function UserProfile({ params }: { params: Params }) {
  const { username } = await params;
  const userData = getUserData(username)
  const userNFTs = getUserNFTs(username)

  return (
    <main className="flex flex-col min-h-screen bg-black">
      <Header />
      <div className="flex-grow container px-4 pt-32 pb-20">
        <div className="max-w-4xl mx-auto">
          <div className="mb-8">
            <BackButton />
            <div className="flex items-center space-x-6">
              <Image
                src={userData.avatar}
                alt={userData.name}
                width={128}
                height={128}
                className="rounded-full"
              />
              <div>
                <h1 className="text-3xl font-bold">{userData.name}</h1>
                <p className="text-gray-400 mt-2">{userData.bio}</p>
                <div className="flex space-x-4 mt-4">
                  <div>
                    <span className="text-gray-400">Total Artworks:</span>
                    <span className="ml-2 font-semibold">{userData.totalArtworks}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Total ETH Spent:</span>
                    <span className="ml-2 font-semibold">{userData.totalEthSpent} ETH</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <h2 className="text-2xl font-bold mb-6">Creations</h2>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {userNFTs.map((nft, index) => (
              <NFTCard key={index} {...nft} showCreator={false} />
            ))}
          </div>
        </div>
      </div>
      <Footer />
    </main>
  )
}

