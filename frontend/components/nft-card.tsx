import Image from 'next/image'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

interface NFTCardProps {
  title: string
  image: string
  ethSpend: number
  creator: {
    name: string
    avatar: string
  }
  showCreator?: boolean
}

export function NFTCard({ title, image, ethSpend, creator, showCreator = true }: NFTCardProps) {
  return (
    <div className="group relative overflow-hidden">
      <h3 className="text-lg font-regular text-white p-4">{title}</h3>
      <div className="group relative rounded-lg">
        <div className="relative aspect-square overflow-hidden w-full rounded-lg">
          <Image
            src={image}
            alt={title}
            fill
            sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
            className="object-cover transition-transform group-hover:scale-105"
          />
        </div>
        <div className="absolute inset-x-0 bottom-0 h-14 bg-black/50 backdrop-blur-sm flex items-center">
          <div className="flex items-center justify-between w-full">
            <span className="text-sm text-purple-100 flex items-center justify-center w-1/2">
              <Image 
                src="/eth32icon.png"
                alt="ETH"
                width={21}
                height={21}
                className="mr-1"
              />
              {ethSpend} ETH spend
            </span>
            <Button 
              className="w-1/2 h-14 bg-white text-black hover:bg-gray-200" 
              size="sm" 
              variant="secondary"
            >
              USE PROMPT
            </Button>
          </div>
        </div>
      </div>
      
      {showCreator && (
        <div className="flex items-center space-x-2 p-4">
          <Image
            src={creator.avatar}
            alt={creator.name}
            width={24}
            height={24}
            className="rounded-full"
          />
          <Link href={`/user/${creator.name.toLowerCase()}`} className="text-sm text-gray-300 hover:text-white transition-colors">
            Crafted by {creator.name}
          </Link>
        </div>
      )}
    </div>
  )
}

