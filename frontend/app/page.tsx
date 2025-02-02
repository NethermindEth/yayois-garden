"use client";

import { useState, useEffect } from "react";
import { Header } from "@/components/header";
import { NFTCard } from "@/components/nft-card";
import { Footer } from "@/components/footer";
import { Logo } from "@/components/Logo";
import { getLastUpdate } from "@/lib/api";
import Link from "next/link";
import Image from "next/image";

const nfts = [
  {
    title: "NEON FACE",
    image: "/neonface.jpg",
    ethSpend: 8.1,
    creator: {
      name: "ALICE",
      avatar: "/avatar.jpg",
    },
  },
  {
    title: "BUBBLE MEN",
    image: "/bubblemen.jpg",
    ethSpend: 5.2,
    creator: {
      name: "BOB",
      avatar: "/avatar.jpg",
    },
  },
  {
    title: "COSMIC WHISPERS",
    image: "/cosmicwhispers.jpg",
    ethSpend: 4.3,
    creator: {
      name: "CHARLY",
      avatar: "/avatar.jpg",
    },
  },
  {
    title: "DREAMSCAPE",
    image: "/dreamscape.jpg",
    ethSpend: 4.1,
    creator: {
      name: "SEAN",
      avatar: "/avatar.jpg",
    },
  },
  {
    title: "MILKY QUARTZ",
    image: "/milky.jpg",
    ethSpend: 3.9,
    creator: {
      name: "SEAN",
      avatar: "/avatar.jpg",
    },
  },
  {
    title: "MARBLE SPACE",
    image: "/marble.jpg",
    ethSpend: 3.1,
    creator: {
      name: "SEAN",
      avatar: "/avatar.jpg",
    },
  },
];

export default function Home() {
  return (
    <main className="flex flex-col min-h-screen bg-black relative">
      <div className="absolute top-[-60px] z-0 w-full">
      <Image
          src="/HeroBackground.png"
          alt="Background"
          width={1920}
          height={290}
          priority
          className="w-full h-[290px] object-cover object-center"
          sizes="100vw"
        />
      </div>

      <div className="relative z-10 flex flex-col min-h-screen">
        <Header />
        <div className="flex-grow">
          <div className="relative pt-52 pb-20">
            <div className="container relative px-4 mx-auto text-center">
              <div className="mb-16 space-y-2">
                <h1 className="flex items-center justify-center space-x-4 text-4xl md:text-6xl font-extralight tracking-tighter">
                  <div className="w-16 h-16 md:w-20 md:h-20">
                    <Logo variant="colored" />
                  </div>
                  <span>YAYOI&apos;S GARDEN</span>
                </h1>
                <p className="text-xl md:text-2xl text-gray-400">
                  Where <span className="text-white">prompts</span> become{" "}
                  <span className="text-white">precious</span>
                </p>

                <div className="pt-4 rounded-lg">
                  <a
                    href="/create"
                    className="inline-block px-8 py-3 bg-white text-black font-medium rounded-full hover:bg-gray-100 transition-colors"
                  >
                    Start creating
                  </a>
                </div>
              </div>

              <div className="inline-block px-4 py-6 bg-black rounded-lg text-left flex items-center justify-between w-full">
                <LastUpdate />
                <Link
                  href="/leaderboard"
                  className="px-4 py-2 border border-white text-white rounded-full hover:bg-white hover:text-black transition-colors"
                >
                  Leaderboard
                </Link>
              </div>

              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {nfts.map((nft, index) => (
                  <NFTCard key={index} {...nft} />
                ))}
              </div>
            </div>
          </div>
        </div>
        <Footer />
      </div>
    </main>
  );
}

function LastUpdate() {
  const [lastUpdate, setLastUpdate] = useState({ date: "", message: "" });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    getLastUpdate()
      .then((data) => {
        setLastUpdate(data);
        setIsLoading(false);
      })
      .catch((err) => {
        console.error("Error in component:", err);
        setIsLoading(false);
      });
  }, []);

  if (isLoading) {
    return <code className="text-sm text-gray-400">Loading...</code>;
  }

  return (
    <div className="flex flex-col  items-start">
      <code className="text-sm text-gray-400 mb-1">
        Last update: {lastUpdate.date}
      </code>
      <code className="text-xl text-white">
        {lastUpdate.message}
        <span className="inline-block w-[0.5em] h-[1.2em] bg-white ml-[2px] animate-blink"></span>
      </code>
    </div>
  );
}
