# 🎨 Yayoi's Garden
### _Where Prompts Become Precious_

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Status](https://img.shields.io/badge/status-alpha-orange.svg)
![Smart Contract](https://img.shields.io/badge/solidity-0.8.19-black.svg)

> "In an age where prompts are copied endlessly, we make them sacred again."

## 🌟 What is Yayoi's Garden?

Welcome to Yayoi's Garden, a groundbreaking decentralized platform where AI innovators and artists converge to create something extraordinary. Here, talented prompt engineers (we call them Model Whisperers) can transform their unique AI generation formulas into time-limited digital assets that others can use - but never copy.

Think of it as an exclusive digital atelier where master artists can monetize their techniques while maintaining their trade secrets. For creators, it's a chance to work with these proprietary styles during their limited lifetime. For collectors, each piece becomes part of a verifiable limited series, tied to a specific moment in time when that style was active.

### Core Concept
Model Whisperes submit their AI generation prompts into a secure vault (TEE - Trusted Execution Environment). Patron Atrists can then pay to use these models to generate their own unique artworks. Each model can be used once per day, but the Model Artist's original prompt remains forever encrypted. After 30 days, the prompt self-destructs, making all artworks generated during its lifetime limited editions.

## ✨ Features

### For Model Artists 🎨
- **Free Submission**: Upload your model prompt and generate one preview image
- **Secure Revenue Stream**: Earn rewards on your art daily
- **Perfect Privacy**: Your model prompt never leaves the secure enclave
- **Reputation Building**: Gain recognition as your model gains popularity
- **Analytics Dashboard**: Track your model's usage and earnings

### For Prompt Artists 🖌
- **Daily Creation**: Submit your own prompts to your chosen model
- **Collection Building**: Each piece is verifiably generated during the model's lifetime
- **Community Rankings**: Help determine which models create the best art

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

1. **Model Submission**
   - Model Artists submit their base prompts for free and generate one preview
   - To activate the model for public use, Model Artists pay 0.5 ETH
   - Each model prompt is encrypted and stored in TEE
   - Smart contract creates tracking instance for each model

2. **Daily Operation**
   - Prompt Artists choose which models they want to use
   - Each Prompt Artist pays 0.1 ETH to submit their prompt to a model
   - A daily vote occurs on which prompt will be executed against the model
   - The top prompt will have it's image generated and an NFT minted to the Prompt Artist
   - The other prompters are refunded
   Alternatively
   - An auction occurs for the right to submit a prompt and win the NFT.

3. **Sunset Phase**
   - After 30 days, each prompt is automatically destroyed
   - Final performance metrics are recorded
   - Artists receive performance badges based on usage

## 🏆 Leaderboard Features
- Real-time earnings tracking
- Daily generation counts
- Community rating scores
- Historical performance data

## 🚀 Getting Started

### For Model Artists
```javascript
Required:
- 1 unique model prompt
- Connected wallet
```

### For Prompt Artists
```javascript
Required:
- Connected wallet
- 0.1 ETH per generation
- Creative prompt to submit to chosen model
```

## 🔮 Future Roadmap

### Phase 1: Genesis (Current)
- Basic prompt submission
- Daily voting
- Generation system

### Phase 2: Evolution
- Multiple AI model support
- NFT stats

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
