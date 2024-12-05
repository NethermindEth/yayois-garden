# 🎨 Yayoi's Garden
### _Where Prompts Become Precious_

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Status](https://img.shields.io/badge/status-alpha-orange.svg)
![Smart Contract](https://img.shields.io/badge/solidity-0.8.19-black.svg)

> "In an age where prompts are copied endlessly, we make them sacred again."

## 🌟 What is Yayoi's Garden?

Yayoi's Garden is a decentralized platform that turns AI art prompts into time-limited digital artifacts. Think of it as a temporary art gallery where the brushes themselves are the masterpieces - visible in their effects but never in their essence.

### Core Concept
Artists submit their carefully crafted prompts into a secure vault (TEE - Trusted Execution Environment). These prompts can be used by the community to generate art, but the prompt itself remains forever encrypted. After 30 days, the prompt self-destructs, making all artworks generated during its lifetime limited editions.

## ✨ Features

### For Artists 🎨
- **Secure Revenue Stream**: Earn 0.1 ETH per artwork generation
- **Perfect Privacy**: Your prompt never leaves the secure enclave
- **Reputation Building**: Gain recognition through your prompt's performance
- **Analytics Dashboard**: Track your prompt's performance and earnings

### For Art Enthusiasts 🖼
- **Daily Generation**: One free generation per wallet per day
- **Voting Power**: Help choose the day's featured prompt
- **Collection Building**: Each piece is verifiably generated during the prompt's lifetime
- **Community Rankings**: Vote on your favorite results

## 🛠 Technical Architecture

### Secure Enclave (TEE)
```
┌──────────────────────────┐
│    Trusted Execution     │
│      Environment         │
│                         │
│  ┌─────────────────┐    │
│  │  Prompt Storage │    │
│  └─────────────────┘    │
│  ┌─────────────────┐    │
│  │   Generation    │    │
│  │     Engine      │    │
│  └─────────────────┘    │
└──────────────────────────┘
```

### Smart Contract Architecture
- **PromptRegistry.sol**: Handles prompt submission and lifecycle
- **VotingMechanism.sol**: Manages daily prompt selection
- **RevenueDistribution.sol**: Automatic artist payments
- **GenerationTracking.sol**: Limits and tracks daily generations

## 💫 How It Works

1. **Submission Phase**
   - Artist submits prompt + 0.5 ETH deposit
   - Prompt is encrypted and stored in TEE
   - Smart contract creates tracking instance

2. **Daily Operation**
   - Community votes on active prompts
   - Top prompt becomes "Prompt of the Day"
   - Users can generate one artwork per wallet
   - Artists earn automatically per generation

3. **Sunset Phase**
   - After 30 days, prompt is automatically destroyed
   - Final performance metrics are recorded
   - Artist receives performance badges

## 📊 Current Statistics
```
Total Prompts Submitted: 156
Active Prompts: 24
Total ETH Distributed: 1,456 ETH
Highest Daily Earnings: 12.4 ETH
Average Prompt Earnings: 9.3 ETH
```

## 🏆 Leaderboard Features
- Real-time earnings tracking
- Daily generation counts
- Community rating scores
- Historical performance data

## 🚀 Getting Started

### For Artists
```javascript
Required:
- 1 unique prompt
- 0.5 ETH deposit
- Connected wallet
```

### For Collectors
```javascript
Required:
- Connected wallet
- Daily: Select prompt & generate
- Optional: Participate in voting
```

## 🔮 Future Roadmap

### Phase 1: Genesis (Current)
- Basic prompt submission
- Daily voting
- Generation system

### Phase 2: Evolution
- Multiple AI model support
- Advanced prompt categories
- Community governance

### Phase 3: Ascension
- Cross-chain support
- Advanced analytics
- Collaborative prompts

## 🤝 Contributing

We welcome contributions! See our [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📜 License

MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🔗 Links
- [Documentation](https://docs.enigmavault.eth)
- [Discord](https://discord.gg/enigmavault)
- [Twitter](https://twitter.com/enigmavault)

---
*Built with 💜 by prompt artists, for prompt artists*

**Disclaimer: This project is not affiliated with Yayoi in any way**
