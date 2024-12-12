'use client'

import { useState } from 'react'
import { Header } from '@/components/header'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Footer } from '@/components/footer'
import { Paperclip, ArrowUp } from 'lucide-react'

export default function Create() {
  const [file, setFile] = useState<File | null>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files?.[0]) {
      setFile(e.target.files[0])
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    // Handle form submission here
    console.log('Form submitted')
  }

  return (
    <main className="flex flex-col min-h-screen bg-black">
      <Header />
      <div className="flex-grow container px-4 pt-32 pb-20">
        <div className="max-w-3xl mx-auto">
          <h1 className="text-4xl font-bold text-center mb-16">Create Your Prompt</h1>
          <div className="bg-gray-900/50 rounded-2xl p-6 backdrop-blur-sm">
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="space-y-4">
                <Input 
                  placeholder="Prompt name" 
                  className="bg-gray-950/50 border-gray-800 rounded-xl text-lg h-12"
                />
                <Textarea 
                  placeholder="Write your prompt here" 
                  className="min-h-[240px] bg-gray-950/50 border-gray-800 resize-none rounded-xl text-lg"
                />
                <div className="flex items-center justify-between px-4 py-3 bg-gray-950/50 border border-gray-800 rounded-xl">
                  <div className="flex items-center">
                    <input
                      id="file-upload"
                      type="file"
                      className="hidden"
                      onChange={handleFileChange}
                    />
                    <label 
                      htmlFor="file-upload" 
                      className="flex items-center space-x-2 cursor-pointer text-gray-400 hover:text-white transition-colors"
                    >
                      <Paperclip className="w-5 h-5" />
                      <span>{file ? file.name : 'Attach a file'}</span>
                    </label>
                  </div>
                  <Button 
                    type="submit"
                    className="bg-white text-black hover:bg-gray-100 rounded-xl px-6"
                  >
                    <span className="mr-2">Submit and deploy</span>
                    <ArrowUp className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
      <Footer />
    </main>
  )
}

