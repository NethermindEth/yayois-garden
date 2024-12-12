'use client'

import { Button } from '@/components/ui/button'
import { ChevronLeft } from 'lucide-react'
import { useRouter } from 'next/navigation'

export function BackButton() {
  const router = useRouter()

  const handleBack = () => {
    router.back()
  }

  return (
    <Button variant="ghost" className="mb-4" onClick={handleBack}>
      <ChevronLeft className="mr-2 h-4 w-4" />
      Back
    </Button>
  )
}

