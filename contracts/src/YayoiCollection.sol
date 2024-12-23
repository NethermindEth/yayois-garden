// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/extensions/ERC721URIStorageUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/interfaces/IERC5267.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

import "./YayoiFactory.sol";

contract YayoiCollection is
    Initializable,
    ERC721URIStorageUpgradeable,
    OwnableUpgradeable,
    IERC5267,
    EIP712Upgradeable
{
    using SafeERC20 for IERC20;

    bytes32 private constant MINT_TYPEHASH = keccak256("Mint(address to,string uri)");

    uint256 public nextTokenId;

    YayoiFactory public factory;

    string public systemPromptUri;
    IERC20 public paymentToken;
    uint256 public promptSubmissionPrice;

    uint256 private constant PROTOCOL_FEE_BPS = 1000; // 10% fee in basis points

    event SetupCompleted(string systemPromptUri, address paymentToken, uint256 promptSubmissionPrice);
    event PromptSubmitted(address indexed submitter, uint256 indexed tokenId, string uri);
    event PromptSubmissionPriceUpdated(uint256 price);
    event PromptSuggested(address indexed sender, string prompt);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    struct InitializeParams {
        string name;
        string symbol;
        address factory;
        address owner;
        string systemPromptUri;
        address paymentToken;
        uint256 promptSubmissionPrice;
    }

    function initialize(InitializeParams memory params) public initializer {
        require(params.factory != address(0), "Invalid factory");

        __ERC721_init(params.name, params.symbol);
        __ERC721URIStorage_init();
        __Ownable_init(params.owner);
        __EIP712_init(params.name, "1");

        factory = YayoiFactory(payable(params.factory));
        systemPromptUri = params.systemPromptUri;
        paymentToken = IERC20(params.paymentToken);
        promptSubmissionPrice = params.promptSubmissionPrice;

        factory.registerSystemPromptUriHash(keccak256(bytes(params.systemPromptUri)));
        emit SetupCompleted(params.systemPromptUri, params.paymentToken, params.promptSubmissionPrice);
    }

    function mintGeneratedToken(address to, string memory uri, bytes memory signature) external payable {
        require(address(factory) != address(0), "Not initialized");

        uint256 tokenId = nextTokenId;
        nextTokenId = tokenId + 1;

        // Handle payment and protocol fee
        uint256 protocolFee = (promptSubmissionPrice * PROTOCOL_FEE_BPS) / 10000;
        uint256 collectionFee = promptSubmissionPrice - protocolFee;

        if (address(paymentToken) != address(0)) {
            paymentToken.safeTransferFrom(msg.sender, address(factory), protocolFee);
            paymentToken.safeTransferFrom(msg.sender, address(this), collectionFee);
        } else {
            require(msg.value >= promptSubmissionPrice, "Insufficient payment");

            // Forward protocol fee to factory
            (bool success1,) = payable(address(factory)).call{value: protocolFee}("");
            require(success1, "Protocol fee transfer failed");

            // Return only the excess above promptSubmissionPrice
            if (msg.value > promptSubmissionPrice) {
                (bool success2,) = payable(msg.sender).call{value: msg.value - promptSubmissionPrice}("");
                require(success2, "Excess ETH return failed");
            }
            // Collection fee remains in the contract automatically
        }

        // Verify signature
        bytes32 structHash = keccak256(abi.encode(MINT_TYPEHASH, to, keccak256(bytes(uri))));
        bytes32 hash = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(hash, signature);
        require(factory.isAuthorizedSigner(signer), "Invalid signature");

        // Mint NFT
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, uri);

        emit PromptSubmitted(msg.sender, tokenId, uri);
    }

    function suggestPrompt(string memory prompt) external {
        emit PromptSuggested(msg.sender, prompt);
    }

    function setPromptSubmissionPrice(uint256 _price) external onlyOwner {
        promptSubmissionPrice = _price;
        emit PromptSubmissionPriceUpdated(_price);
    }

    function sweepTokens(address token) external onlyOwner {
        if (token == address(0)) {
            (bool success,) = msg.sender.call{value: address(this).balance}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(msg.sender, IERC20(token).balanceOf(address(this)));
        }
    }

    function domainSeparator() external view returns (bytes32) {
        return _domainSeparatorV4();
    }

    receive() external payable {}
}
