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

    bytes32 public promptId;
    IERC20 public paymentToken;
    uint256 public promptSubmissionPrice;

    uint256 private constant PROTOCOL_FEE_BPS = 1000; // 10% fee in basis points

    event SetupCompleted(bytes32 promptId, address paymentToken, uint256 promptSubmissionPrice);
    event PromptSubmitted(address indexed submitter, uint256 indexed tokenId, string uri);
    event PromptSubmissionPriceUpdated(uint256 price);
    event PromptSuggested(address indexed sender, string prompt);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(string calldata name, string calldata symbol, address _factory, address _owner)
        public
        initializer
    {
        require(_factory != address(0), "Invalid factory");

        __ERC721_init(name, symbol);
        __ERC721URIStorage_init();
        __Ownable_init(_owner);
        __EIP712_init(name, "1");

        factory = YayoiFactory(payable(_factory));
        _transferOwnership(_factory);
    }

    function initializeOwner(address owner) external {
        require(msg.sender == address(factory), "Only factory");
        require(promptId == bytes32(0), "Already initialized");
        _transferOwnership(owner);
    }

    function setup(bytes32 _promptId, address _paymentToken, uint256 _promptSubmissionPrice) external onlyOwner {
        require(promptId == bytes32(0), "Already setup");
        require(_promptId != bytes32(0), "Invalid prompt ID");
        require(!factory.isPromptIdUsed(_promptId), "Prompt ID already used");

        promptId = _promptId;
        paymentToken = IERC20(_paymentToken);
        promptSubmissionPrice = _promptSubmissionPrice;

        factory.registerPromptId(_promptId);
        emit SetupCompleted(_promptId, _paymentToken, _promptSubmissionPrice);
    }

    function mintGeneratedToken(address to, string memory uri, bytes memory signature) external payable {
        require(promptId != bytes32(0), "Not initialized");

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

    receive() external payable {}

    function DOMAIN_SEPARATOR() external view returns (bytes32) {
        return _domainSeparatorV4();
    }

    function systemPromptId() external view returns (bytes32) {
        return promptId;
    }

    function eip712Domain()
        public
        view
        override(EIP712Upgradeable, IERC5267)
        returns (
            bytes1 fields,
            string memory name,
            string memory version,
            uint256 chainId,
            address verifyingContract,
            bytes32 salt,
            uint256[] memory extensions
        )
    {
        return (0x0f, "Yayoi", "1", block.chainid, address(this), bytes32(0), new uint256[](0));
    }
}
